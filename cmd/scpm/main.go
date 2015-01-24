package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/gophergala/scpm"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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
		cli.StringFlag{
			Name:  "in",
			Value: "../../README.md",
			Usage: "/path/to/file or folder",
		},
		cli.StringSliceFlag{
			Name:  "path",
			Value: &cli.StringSlice{"gronpipmaster.ru:/tmp/ss/readme.md", "gronpipmaster.ru:/tmp/ss/readme.md"},
			Usage: "user@example.com:/path/to",
		},
	}
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	app := cli.NewApp()
	app.Name = "scpm"
	// app.EnableBashCompletion = true
	app.Author = "gronpipmaster"
	app.Email = "gronpipmaster@gmail.com"
	app.Version = version
	app.Usage = "Copy files over ssh protocol to multiple servers."
	app.Flags = globalFlags
	app.Action = action
	app.Run(os.Args)
}

func action(ctx *cli.Context) {
	hosts := []scpm.Host{}
	for _, host := range ctx.GlobalStringSlice("path") {
		h, err := scpm.NewHost(host, ctx.GlobalString("identity"), ctx.GlobalInt("port"))
		if err != nil {
			log.Fatalln(err)
		}
		hosts = append(hosts, h)
	}

	scp, err := scpm.New(
		hosts,
		ctx.GlobalDuration("timeout"),
		ctx.GlobalString("in"),
	)
	if err != nil {
		log.Fatalln(err)
	}
	//Create chanels for wait quit signal
	quit := make(chan bool)
	//Create chanel for wait system signals
	osSigs := make(chan os.Signal, 1)
	//Kill - 3
	signal.Notify(osSigs, syscall.SIGQUIT)
	//Ctrl + C
	signal.Notify(osSigs, os.Interrupt)
	//Init and run gc
	go scp.Run(quit)
	for {
		select {
		case <-osSigs:
			fmt.Println("Quit signal, wait stop.")
			quit <- true //send signal stop app
		case <-quit:
			return
		}
	}
}
