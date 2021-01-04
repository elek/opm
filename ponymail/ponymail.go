package ponymail

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"net/http"
	"os"
)

var PonymailCommands []*cli.Command

func RegisterPonymail(command cli.Command) {
	if PonymailCommands == nil {
		PonymailCommands = make([]*cli.Command, 0)
	}
	PonymailCommands = append(PonymailCommands, &command)
}

func CreatePonymailCommand() *cli.Command {
	return &cli.Command{
		Name:        "ponymail",
		Description: "Retrieve and extract information from Apache mailing list archive",
		Subcommands: PonymailCommands,
	}
}

func httpCall(query string) ([]byte, error) {
	client := &http.Client{}
	url := "https://lists.apache.org/api/" + query
	log.Debug().Msg("Downloading " + url)
	req, err := http.NewRequest("GET", url, nil)
	if cookie, found := os.LookupEnv("PONYMAIL_COOKIE"); found {
		req.Header.Add("Cookie", "ponymail="+cookie)
	}
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode > 300 {
		return nil, errors.New("Couldn't retrieve " + url + " status: " + resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}