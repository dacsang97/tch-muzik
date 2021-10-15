package model

import "time"

type Song struct {
	Name   string    `json:"name"`
	Starts time.Time `json:"starts"`
	Ends   time.Time `json:"ends"`
	Type   string    `json:"type"`
	FileID int       `json:"file_id"`
	Track  string    `json:"track"`
	Artist string    `json:"artist"`
	Image  string    `json:"image"`
	Album  string    `json:"album"`
}

type MusicInfo struct {
	Previous      Song      `json:"previous"`
	Current       Song      `json:"current"`
	Next          Song      `json:"next"`
	SchedulerTime time.Time `json:"schedulerTime"`
	Expire        int       `json:"expire"`
}
