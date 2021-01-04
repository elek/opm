package main

import (
	"github.com/elek/go-utils/kv"
	"github.com/elek/opm/util"
)

func init() {
	//runner.App.Commands = append(runner.App.Commands, cli.Command{
	//	Name: "iterate",
	//	Flags: []cli.Flag{
	//		cli.BoolFlag{
	//			Name:  "read",
	//			Usage: "Read all the records",
	//		},
	//	},
	//	Action: func(c *cli.Context) error {
	//		return iterate(c.Args().Get(0), c.Bool("read"))
	//	},
	//})
	//
	//runner.App.Commands = append(runner.App.Commands, cli.Command{
	//	Name: "copy",
	//	Action: func(c *cli.Context) error {
	//		return copy(c.Args().Get(0), c.Args().Get(1))
	//	},
	//})
}

func copy(from string, to string) error {
	source, err := kv.Create(from)
	if err != nil {
		return err
	}
	dest, err := kv.Create(to)
	if err != nil {
		return err
	}
	progress := util.CreateProgress()
	return source.IterateAll(func(key string) error {
		data, err := source.Get(key)
		progress.Increment()
		if err != nil {
			return err
		}
		return dest.Put(key, data)
	})

}

func iterate(dir string, read bool) error {
	from, err := kv.Create(dir)
	if err != nil {
		return err
	}
	progress := util.CreateProgress()
	err = from.IterateAll(func(key string) error {
		progress.Increment()
		if read {
			_, err := from.Get(key)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	progress.End()
	return nil
}
