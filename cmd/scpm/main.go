package main

import (
	"github.com/codegangsta/cli"
	"log"
	"os"
	"time"
)

var (
	version     string
	globalFlags = []cli.Flag{
		cli.DurationFlag{
			Name:  "timeout, t",
			Value: 59 * time.Second,
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 22,
		},
		cli.StringFlag{
			Name:  "identity, i",
			Usage: "ssh identity to use for connecting to the host",
		},
		cli.StringSliceFlag{
			Name:  "path",
			Value: &cli.StringSlice{},
			Usage: "user@example.com:/path/to",
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "scpm"
	// app.EnableBashCompletion = true
	app.Author = "gronpipmaster"
	app.Email = "gronpipmaster@gmail.com"
	app.Version = version
	app.Usage = "Copy files over ssh protocol to multiple servers."
	app.Flags = globalFlags
	app.Action = func(c *cli.Context) {
		log.Println(c.GlobalStringSlice("path"))
	}

	app.Run(os.Args)
}
