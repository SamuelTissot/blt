package main

import (
	"log"
	"os"
	"sort"
	"time"

	"github.com/SamuelTissot/blt/cmd"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Version = "0.0.4"
	app.Name = "blt"
	app.Usage = "blt, the bacon lettuce tomato of testing. Seriously; Binary Load Tester"
	app.Compiled = time.Now()
	app.Copyright = "(c) 2018 Samuel Tissot"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Samuel Tissot",
			Email: "tissotjobin@gmail.com",
		},
	}

	// global flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "target, t",
			Usage: "target `FILE`",
		},
		cli.BoolFlag{
			Name:  "nocache",
			Usage: "if present it will send the header 'Cache-Control: no-cache'",
		},
		cli.StringFlag{
			Name:  "auth",
			Usage: "Will include a basic auth header with the request",
		},
	}

	app.Commands = []cli.Command{
		cmd.Find,
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
