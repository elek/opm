package ponymail

import (
	"github.com/elek/go-utils/json"
	"github.com/elek/go-utils/kv"
	"os"
	"path"
	"strconv"
	"time"
)

func init() {
	//runner.App.Commands = append(runner.App.Commands, cli.Command{
	//	Name: "mail-lists-combine",
	//	Action: func(c *cli.Context) error {
	//		store, err := kv.Create(c.Args().Get(0))
	//		if err != nil {
	//			return err
	//		}
	//		//return maillists(c.Args().Get(0))
	//		return combine(store, "hadoop.apache.org", "private")
	//	},
	//})

}

func combine(k kv.KV, domain string, list string) error {
	listPath := path.Join("ponymail/lists", domain, list)
	archviePath := path.Join("ponymail/archive", domain, list)
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
	f, err := os.Create( "/tmp/" + list + "." + domain + ".mbox")
	if err != nil {
		return err
	}
	for month := from; month <= to; month++ {
		downloadYear := month / 12
		downloadMonth := month%12 + 1
		archiveOfMonth := path.Join(archviePath, strconv.Itoa(downloadYear), strconv.Itoa(downloadMonth))
		if k.Contains(archiveOfMonth) {
			content, err := k.Get(archiveOfMonth)
			if err != nil {
				return err
			}
			f.Write(content)
		}
	}
	f.Close()
	return nil

}
