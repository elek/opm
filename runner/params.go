package runner

import (
	"github.com/elek/go-utils/kv"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
)

func CreateRepo(c *cli.Context) (kv.KV, error) {
	var err error
	path := ""
	if c.String("repo") != "" {
		path = c.String("repo")
	} else {
		path, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	log.Info().Msg("Opening local data repository: " + path)
	return kv.Create(path)
}

func DestDir(c *cli.Context) (string, error) {

	path, err := os.Getwd()
	if err != nil {
		return "", err
	}
	log.Info().Msg("Destination directory: " + path)
	return path, nil
}
