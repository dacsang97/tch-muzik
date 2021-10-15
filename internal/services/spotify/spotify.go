package spotify

import (
	"fmt"
	"net/http"
	"strings"
	"tch-muzik/internal/model"
	"time"

	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const (
	loginServerAddress = "127.0.0.1:51197"
	loginCallbackPath  = "/callback"
	state              = "2021-10-10T00:40:29+07:00"
)

type SpotifyService struct {
	Auth spotify.Authenticator
}

func NewSpotifyService(clientId string, secretKey string) *SpotifyService {
	// init Spotify client
	auth := spotify.NewAuthenticator(
		"http://"+loginServerAddress+loginCallbackPath,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopeUserReadCurrentlyPlaying,
		spotify.ScopeUserTopRead,
	)
	auth.SetAuthInfo(viper.GetString("client_id"), viper.GetString("secret_key"))
	return &SpotifyService{auth}
}

func (svc *SpotifyService) GetToken() *oauth2.Token {
	var token *oauth2.Token
	accessToken := viper.GetString("access_token")
	accessTokenExpiry := viper.GetTime("access_token_expiry")
	if accessToken == "" || accessTokenExpiry.Before(time.Now()) {
		var err error
		token, err = svc.login()
		if err != nil {
			log.Fatal().Stack().Err(errors.WithStack(err)).Send()
		}

		viper.Set("access_token", token.AccessToken)
		viper.Set("access_token_expiry", token.Expiry)
		viper.Set("refresh_token", token.RefreshToken)
		if err := viper.WriteConfig(); err != nil {
			log.Fatal().Stack().Err(errors.WithStack(err)).Send()
		}
	} else {
		refreshToken := viper.GetString("refresh_token")
		token = &oauth2.Token{
			AccessToken:  accessToken,
			Expiry:       accessTokenExpiry,
			RefreshToken: refreshToken,
		}
	}
	return token
}

func (svc *SpotifyService) login() (*oauth2.Token, error) {
	tokenChan := make(chan *oauth2.Token)
	errChan := make(chan error)

	http.HandleFunc(loginCallbackPath, func(w http.ResponseWriter, r *http.Request) {
		token, err := svc.Auth.Token(state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			errChan <- err
		}

		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			errChan <- errors.New("state not match")
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `Success! Back to CLI.<script>setTimeout(window.close, 5000);</script>`)
		tokenChan <- token
	})
	go http.ListenAndServe(loginServerAddress, nil)

	url := svc.Auth.AuthURL(state)
	if err := browser.OpenURL(url); err != nil {
		log.Fatal().Err(errors.WithStack(err)).Send()
	}
	select {
	case err := <-errChan:
		return nil, err
	case token := <-tokenChan:
		return token, nil
	}
}

func (svc *SpotifyService) PlaySong(client *spotify.Client, song *model.Song) error {
	var availableGenreSeeds []string
	searchCountry := "VN"
	searchLimit := 1
	searchResult, err := client.SearchOpt("track:"+song.Name+" "+"artist:"+song.Artist, spotify.SearchTypeTrack, &spotify.Options{Country: &searchCountry, Limit: &searchLimit})
	if err != nil {
		return err
	}
	var track *spotify.SimpleTrack
	if searchResult.Tracks.Total > 0 {
		track = &searchResult.Tracks.Tracks[0].SimpleTrack
	} else {
		log.Info().Str("song", song.Name).Msg("song not found")
		if len(availableGenreSeeds) == 0 {
			topArtists, err := client.CurrentUsersTopArtists()
			if err == nil && topArtists != nil {
				availableGenreSeeds = topArtists.Artists[0].Genres[:4]
			} else {
				log.Error().Stack().Err(errors.WithStack(err)).Send()
				availableGenreSeeds = []string{"pop"}
			}
		}

		recommendations, err := client.GetRecommendations(spotify.Seeds{
			Genres: availableGenreSeeds,
		}, nil, nil)
		if err != nil {
			return err
		}
		if len(recommendations.Tracks) == 0 {
			return errors.New("no song")
		}
		track = &recommendations.Tracks[0]
		log.Info().Msg("Recommend a song in genre " + strings.Join(availableGenreSeeds, ","))
	}

	log.Info().Str("play", track.Name).Send()

	playOptions := &spotify.PlayOptions{
		URIs: []spotify.URI{
			track.URI,
		},
	}

	playerState, err := client.PlayerState()
	if err != nil {
		return err
	}
	if playerState.Device.ID != "" {
		devices, err := client.PlayerDevices()
		if err != nil {
			return err
		}
		if len(devices) == 0 {
			return errors.New("no device")
		}
		playOptions.DeviceID = &devices[0].ID
	}

	if err := client.PlayOpt(playOptions); err != nil {
		return err
	}
	return nil
}
