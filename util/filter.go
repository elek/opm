package util

import (
	"github.com/elek/go-utils/kv"
	"github.com/urfave/cli"
	"strings"
)

func init() {
	_ = cli.Command{
		Name: "filter",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "prefix",
				Value: "",
				Usage: "Prefix in the source store to iterate over",
			},
			cli.StringFlag{
				Name:  "filter",
				Value: "",
				Usage: "Filter which should be included by the key.",
			},
		},
		Action: func(c *cli.Context) error {

			sourceStore, err := kv.Create(c.Args().Get(0))
			if err != nil {
				return err
			}

			destinationStore, err := kv.Create(c.Args().Get(1))
			if err != nil {
				return err
			}
			filter := c.String("filter")
			return sourceStore.IterateSubTree(c.String("prefix"), func(key string) error {
				val, err := sourceStore.Get(key)
				if err != nil {
					return err
				}
				if len(filter) > 0 && !strings.Contains(string(val), filter) {
					return nil
				}
				return destinationStore.Put(key, val)
			})
		},
	}

}
