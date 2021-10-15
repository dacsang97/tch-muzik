package cmd

import (
	"os"
	"tch-muzik/internal/services/spotify"
	"tch-muzik/internal/services/tch"
	"time"

	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	envPrefix  = "tch"
	configName = ".tchmuzik"
	configType = "yaml"
)

var (
	ClientID  string
	SecretKey string
	rootCmd   = &cobra.Command{
		Use:   "tch",
		Short: "An application help to search & play TCH Muzik on Spotify.",
		Run:   Run,
	}
)

func Run(cmd *cobra.Command, args []string) {
	spotify := spotify.NewSpotifyService(ClientID, SecretKey)
	token := spotify.GetToken()
	client := spotify.Auth.NewClient(token)
	tch := tch.NewTchService()

	for ; true; <-time.Tick(5 * time.Second) {
		playerState, err := client.PlayerState()
		if err != nil {
			log.Error().Stack().Err(errors.WithStack(err)).Send()
			continue
		}
		if playerState.Playing {
			continue
		}
		song, err := tch.GetTchMusicInfo()
		if err != nil {
			log.Error().Stack().Err(errors.WithStack(err)).Send()
			continue
		}
		if song == nil {
			log.Error().Stack().Err(errors.New("can not fetch song")).Send()
			continue
		}
		log.Info().Str("tch_song", song.Name).Send()
		if err := spotify.PlaySong(&client, song); err != nil {
			log.Error().Stack().Err(errors.WithStack(err)).Send()
			continue
		}
	}
}

type TchCli struct {
}

func NewTchCli() *TchCli {
	return &TchCli{}
}

func (cli *TchCli) Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&ClientID, "client_id", "c", "", "Spotify Client ID")
	rootCmd.PersistentFlags().StringVarP(&SecretKey, "secret_key", "s", "", "Spotify Secret Key")

	viper.BindPFlag("client_id", rootCmd.PersistentFlags().Lookup("client_id"))
	viper.BindPFlag("secret_key", rootCmd.PersistentFlags().Lookup("secret_key"))
}

func initConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Stack().Err(errors.WithStack(err)).Send()
	}

	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".tchmuzik")
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if _, isConfigFileNotFoundError := err.(viper.ConfigFileNotFoundError); isConfigFileNotFoundError {
			configPath := filepath.Join(home, configName+"."+configType)
			file, createErr := os.Create(configPath)
			if createErr != nil {
				log.Fatal().Stack().Err(createErr).Send()
			}
			file.Close()
		}
	}

	if viper.GetString("client_id") == "" {
		log.Fatal().Strs("missing_fields", []string{"client_id"}).Msg("required fields missing")
	}

	if viper.GetString("secret_key") == "" {
		log.Fatal().Strs("missing_fields", []string{"secret_key"}).Msg("required fields missing")
	}

	if err := viper.WriteConfig(); err != nil {
		log.Fatal().Stack().Err(errors.WithStack(err)).Msg("can not save config")
	}

}
