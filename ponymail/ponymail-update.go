package ponymail

import (
	"fmt"
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"path"
	"strconv"
	"time"
)

func init() {
	cmd := cli.Command{
		Name: "update",
		Action: func(c *cli.Context) error {
			store, err := kv.Create(c.Args().Get(0))
			if err != nil {
				return err
			}
			return updateLists(store)
		},
	}
	RegisterPonymail(cmd)
}

func updateLists(store kv.KV) error {
	domains, err := store.List(path.Join("ponymail", "lists"))
	if err != nil {
		return err
	}
	for _, domainPrefix := range domains {
		lists, err := store.List(domainPrefix)
		if err != nil {
			return err
		}

		for _, listPrefix := range lists {
			domain := path.Base(path.Dir(listPrefix))
			list := path.Base(listPrefix)
			log.Info().Msg("Downloading " + domain + "/" + list)
			err = retrieveList(store, domain, list)
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func retrieveList(k kv.KV, domain string, list string) error {
	listPath := path.Join("ponymail", "lists", domain, list)
	archviePath := path.Join("ponymail", "archive", domain, list)
	desc, err := json.AsJson(k.Get(listPath))
	if err != nil {
		return err
	}
	from := json.MN(desc, "firstYear")*12 + json.MN(desc, "firstMonth") - 1
	to := json.MN(desc, "lastYear")*12 + json.MN(desc, "lastMonth") - 1
	lastMonth := time.Now().Year()*12 + int(time.Now().Month()) - 1 - 1
	if lastMonth < to {
		to = lastMonth
	}
	for month := from; month <= to; month++ {
		downloadYear := month / 12
		downloadMonth := month%12 + 1
		archiveOfMonth := path.Join(archviePath, strconv.Itoa(downloadYear), strconv.Itoa(downloadMonth))
		if !k.Contains(archiveOfMonth) {
			query := fmt.Sprintf("mbox.lua?list=%s@%s&date=%d-%d", list, domain, downloadYear, downloadMonth)
			body, err := httpCall(query)
			if err != nil {
				return err
			}
			err = k.Put(archiveOfMonth, body)
			if err != nil {
				return err
			}
			time.Sleep(3 * time.Second)

		}

	}
	return nil

}
