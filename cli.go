package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

var (
	version = "1.0.0"
)

func main() {
	app := cli.NewApp()
	app.Name = "SecNginX"
	app.Usage = "Build and setup an secure and minimnal NginX webserver and submit your certificates to all of Chrome's CT log servers"

	app.Version = version

	app.Commands = []cli.Command{
		{
			Name:   "install",
			Usage:  "Build and install NginX and create basic NginX file structure",
			Action: start,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "without-brotli-module",
					Usage: "Compile NginX without the brotli module",
				},
				cli.BoolFlag{
					Name:  "without-cors-module",
					Usage: "Compile NginX without the CORS filter module",
				},
				cli.BoolFlag{
					Name:  "without-dynamic-tls-records",
					Usage: "Compile NginX without applying the Dynamic TLS Records patch",
				},
				cli.BoolFlag{
					Name:  "without-headers-more-module",
					Usage: "Compile NginX without the headers-more module",
				},
				cli.BoolFlag{
					Name:  "without-ct-module",
					Usage: "Compile NginX without the CT module",
				},
				cli.BoolFlag{
					Name:  "without-cookie-flag-module",
					Usage: "Compile NginX without the cookie-flag module",
				},
				cli.BoolFlag{
					Name:  "upgrade",
					Usage: "Only compile and install NginX, do not change the nginx data",
				},
			},
		},
		{
			Name:   "submit-ct",
			Usage:  "Submit the given public certificate to some of Chrome's Certificate Transparency Log Servers",
			Action: submitCT,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input",
					Usage: "Path of the public certificate to submit",
				},
				cli.StringFlag{
					Name:  "output",
					Usage: "Path of the folder to output all generated .sct files to",
				},
				cli.StringFlag{
					Name:  "filename",
					Usage: "Optional filename for the created .sct files: <CT Log Server>.<filename>.sct",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
