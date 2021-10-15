package main

import (
	"tch-muzik/internal/cmd"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	// init log
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}

func main() {
	cli := cmd.NewTchCli()
	cli.Execute()
}
