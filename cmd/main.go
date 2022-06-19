package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/tapestrylabs/fabric"
	"github.com/urfave/cli/v2"
)

func main() {
	config := fabric.NewConfig()

	runService := fabric.NewRunService(config)
	defer runService.Close()

	artifactoryService := fabric.NewArtifactoryService(config)
	defer artifactoryService.Close()

	sqlService := fabric.NewSqlService(config)

	app := &cli.App{
		Name:  "fabric",
		Usage: "deploy serverless containers",
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "initialize new project",
				Action: func(ctx *cli.Context) error {
					err := artifactoryService.CreateArtifactory(ctx.Context)
					if err != nil {
						return err
					}
					// Do more stuf here like scaffold out new project
					return nil
				},
			},
			{
				Name:    "database",
				Aliases: []string{"db"},
				Usage:   "invoke actions for a database",
				Subcommands: []*cli.Command{
					{
						Name:    "create",
						Aliases: []string{"c"},
						Usage:   "creates a new database instance",
						Action: func(ctx *cli.Context) error {
							for _, v := range config.Organization.Services {
								err := sqlService.CreateSqlInstance(ctx.Context, &v.Service)
								if err != nil {
									return err
								}
							}
							return nil
						},
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "lists database instances",
						Action: func(ctx *cli.Context) error {
							instances, err := sqlService.ListSqlInstances(ctx.Context)
							if err != nil {
								return err
							}

							for _, v := range instances {
								fmt.Printf("instance: %s, type: %s, state: %s\n", v.Name, v.DatabaseVersion, v.State)
							}

							return nil
						},
					},
				},
			},
			{
				Name:    "service",
				Aliases: []string{"s"},
				Usage:   "invoke actions for a service",
				Subcommands: []*cli.Command{
					{
						Name:    "deploy",
						Aliases: []string{"d"},
						Usage:   "deploy service",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Required: false,
							},
						},
						Action: func(ctx *cli.Context) error {
							if ctx.String("name") == "" {
								for _, v := range config.Organization.Services {
									url, err := runService.DeployServiceSequence(ctx.Context, &v.Service)
									if err != nil {
										return err
									}
									fmt.Printf("%s: %s\n", v.Service.Name, *url)
								}
							} else {
								var (
									foundService *fabric.Service
									serviceName  string = ctx.String("name")
								)
								for _, v := range config.Organization.Services {
									if strings.EqualFold(v.Service.Name, serviceName) {
										foundService = &v.Service
									}
								}
								if foundService == nil {
									return errors.New(fmt.Sprintf("%s is not registered in fabric.yaml", serviceName))
								}
								url, err := runService.DeployServiceSequence(ctx.Context, foundService)
								if err != nil {
									return err
								}
								fmt.Printf("%s: %s\n", foundService.Name, *url)
							}

							return nil
						},
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "list all deployed services",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "region",
								Required: true,
							},
						},
						Action: func(ctx *cli.Context) error {
							region := ctx.String("region")
							services, err := runService.ListDeployedServices(ctx.Context, region)
							if err != nil {
								return err
							}

							if len(services) == 0 {
								fmt.Printf("no services deployed in region %s\n", region)
							} else {
								fmt.Printf("services deployed in region %s:\n", region)
								for _, v := range services {
									_, name := path.Split(v.Name)
									_, revision := path.Split(v.LatestReadyRevision)
									fmt.Printf("service: %s, latest revision: %s\n", name, revision)
								}
							}

							return nil
						},
					},
					{
						Name:    "remove",
						Aliases: []string{"rm"},
						Usage:   "remove deployed service",
						Action: func(ctx *cli.Context) error {
							if ctx.String("name") == "" {
								for _, v := range config.Organization.Services {
									return runService.DeleteRunService(ctx.Context, &v.Service)
								}
							} else {
								var (
									foundService *fabric.Service
									serviceName  string = ctx.String("name")
								)
								for _, v := range config.Organization.Services {
									if strings.EqualFold(v.Service.Name, serviceName) {
										foundService = &v.Service
									}
								}
								if foundService == nil {
									return errors.New(fmt.Sprintf("%s is not registered in fabric.yaml", serviceName))
								}
								return runService.DeleteRunService(ctx.Context, foundService)
							}

							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
