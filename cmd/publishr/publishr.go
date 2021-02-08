package main

import (
	"fmt"
	"github.com/analogj/go-util/utils"
	"github.com/packagrio/publishr/pkg"
	"github.com/packagrio/publishr/pkg/config"
	"github.com/packagrio/publishr/pkg/version"
	"github.com/urfave/cli"
	"log"
	"os"
	"time"
)

var goos string
var goarch string

func main() {
	app := &cli.App{
		Name:     "publishr",
		Usage:    "Language agnostic tool to ",
		Version:  version.VERSION,
		Compiled: time.Now(),
		Authors: []cli.Author{
			{
				Name:  "Jason Kulatunga",
				Email: "jason@thesparktree.com",
			},
		},
		Before: func(c *cli.Context) error {

			packagrUrl := "github.com/packagrio/publishr"

			versionInfo := fmt.Sprintf("%s.%s-%s", goos, goarch, version.VERSION)

			subtitle := packagrUrl + utils.LeftPad2Len(versionInfo, " ", 53-len(packagrUrl))

			fmt.Fprintf(c.App.Writer, fmt.Sprintf(utils.StripIndent(
				`
			 ____   __    ___  __ _   __    ___  ____ 
			(  _ \ / _\  / __)(  / ) / _\  / __)(  _ \
			 ) __//    \( (__  )  ( /    \( (_ \ )   /
			(__)  \_/\_/ \___)(__\_)\_/\_/ \___/(__\_)
			%s

			`), subtitle))
			return nil
		},

		Commands: []cli.Command{
			{
				Name:  "start",
				Usage: "Start publishr pipeline",
				Action: func(c *cli.Context) error {

					configuration, _ := config.Create()
					configuration.Set(config.PACKAGR_SCM, c.String("scm"))
					configuration.Set(config.PACKAGR_PACKAGE_TYPE, c.String("package_type"))

					if c.IsSet("local_branch") {
						configuration.Set(config.PACKAGR_SCM_LOCAL_BRANCH, c.String("local_branch"))
					}
					if c.IsSet("remote_branch") {
						configuration.Set(config.PACKAGR_SCM_REMOTE_BRANCH, c.String("remote_branch"))
					}

					fmt.Println("package type:", configuration.GetString(config.PACKAGR_PACKAGE_TYPE))
					fmt.Println("scm:", configuration.GetString(config.PACKAGR_SCM))

					pipeline := pkg.Pipeline{}
					err := pipeline.Start(configuration)
					if err != nil {
						fmt.Printf("FATAL: %+v\n", err)
						os.Exit(1)
					}

					return nil
				},

				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "scm",
						Value: "generic",
						Usage: "The scm for the code, for setting additional SCM specific metadata",
					},

					&cli.StringFlag{
						Name:  "package_type",
						Value: "generic",
						Usage: "The type of package being built.",
					},
					&cli.StringFlag{
						Name:  "remote_branch",
						Value: "master",
						Usage: "The destination branch to push changes to",
					},
					&cli.BoolFlag{
						Name:  "dry_run",
						Usage: "When dry run is enabled, no data is written to file system",
					},
					//&cli.StringSliceFlag{
					//	Name: "upload_artifact",
					//	Usage: "Path to a release artifact"
					//},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}
