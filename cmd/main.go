package main

import (
	"fmt"
	"tch-muzik/internal/services/tch"
)

func main() {
	fmt.Println("TCH Muzik")

	tchSvc := tch.NewTchService()

	if song, err := tchSvc.GetTchMusicInfo(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(song)
	}

}
