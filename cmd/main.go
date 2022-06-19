package main

import (
	"log"
	"os"

	"github.com/tapestrylabs/fabric"
	"github.com/urfave/cli/v2"
)

func main() {
	config := fabric.NewConfig()

	app := &cli.App{
		Name:  "Fabric",
		Usage: "Deploy serverless containers",
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize new organization",
				Action: func(ctx *cli.Context) error {
					err := fabric.CreateArtifactory(ctx.Context, config)
					if err != nil {
						return err
					}
					// Do more stuf here like scaffold out new project
					return nil
				},
			},
			{
				Name:    "deploy",
				Aliases: []string{"d"},
				Usage:   "Build and push image",
				Action: func(ctx *cli.Context) error {
					for _, v := range config.Organization.Services {
						var err error
						err = fabric.PushImage(ctx.Context, v.Service.Name, config)
						if err != nil {
							return err
						}

						err = fabric.UpsertRunService(ctx.Context, v.Service.Name, config)
						if err != nil {
							return err
						}

						service, err := fabric.GetRunService(ctx.Context, v.Service.Name, config)
						if err != nil {
							return err
						}

						log.Println(service.Uri)

						if v.Service.Public {
							err = fabric.SetIamPolicy(ctx.Context, v.Service.Name, config)
							if err != nil {
								return err
							}
						}
					}

					return nil
				},
			},
			{
				Name:    "remove",
				Aliases: []string{"r"},
				Usage:   "Remove deployed infrastructure",
				Action: func(ctx *cli.Context) error {
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
