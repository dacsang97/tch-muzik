package tch

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"tch-muzik/internal/model"
	"time"
)

const musicInfoEndpoint = "https://api.thecoffeehouse.com/api/get_music_info"

type TchService struct {
}

func NewTchService() *TchService {
	return &TchService{}
}

func (tch *TchService) GetTchMusicInfo() (*model.Song, error) {
	httpClient := http.Client{
		Timeout: time.Second * 5,
	}
	req, err := http.NewRequest(http.MethodGet, musicInfoEndpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "github.com:dacsang97/tch+muzik+v1")

	res, getErr := httpClient.Do(req)
	if getErr != nil {
		return nil, err
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, err
	}

	var musicInfo model.MusicInfo
	if jsonErr := json.Unmarshal(body, &musicInfo); jsonErr != nil {
		return nil, err
	}
	return &musicInfo.Current, nil
}
