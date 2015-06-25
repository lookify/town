package main

import (
  "os"
  "log" // was github.com/Sirupsen/logrus
  "github.com/codegangsta/cli"
  "github.com/lookify/town/version"
)

const DEFAULT_CLUSTER_NAME = "town"

func main() {
  app := cli.NewApp()
  app.Name = "town"
  app.Usage = "town orchestartion tool for docker"
  app.Version = version.VERSION + " (" + version.GITCOMMIT + ")"
  app.Author = ""
  app.Email = ""

  app.Flags = []cli.Flag{
    cli.BoolFlag{
      Name:   "debug",
      Usage:  "debug mode",
      EnvVar: "DEBUG",
    },
  }

  // logs
  app.Before = func(c *cli.Context) error {
    /*
    log.SetOutput(os.Stderr)
    if c.Bool("debug") {
      log.SetLevel(log.DebugLevel)
    }
    */
    return nil
  }

  app.Commands = []cli.Command{
    {
      Name:      "restart",
      ShortName: "re",
      Usage:     "restart a cluster",
      Action: func(c *cli.Context) {
        var name = c.Args().First()
        if len(name) == 0 {
          name = DEFAULT_CLUSTER_NAME
        }

        town := NewTown()
        town.ReadFile(name)
        town.Connect()
        town.Provision(false)
        town.StopContainers(false)
        town.RemoveContainers(false)
        town.CreateContainers(false)
      },
    },
    {
      Name:      "run",
      ShortName: "r",
      Usage:     "run a cluster",
      Action: func(c *cli.Context) {
        var name = c.Args().First()
        if len(name) == 0 {
          name = DEFAULT_CLUSTER_NAME
        }

        town := NewTown()
        town.ReadFile(name)
        town.Connect()
        town.Provision(true)
        town.StopContainers(true)
        town.RemoveContainers(true)
        town.CreateContainers(true)
      },
    },
    {
      Name:      "stop",
      ShortName: "s",
      Usage:     "stop a cluster",
      Action: func(c *cli.Context) {
        var name = c.Args().First()
        if len(name) == 0 {
          name = DEFAULT_CLUSTER_NAME
        }

        town := NewTown()
        town.ReadFile(name)
        town.Connect()
        town.Provision(false)
        town.StopContainers(false)
        town.RemoveContainers(false)
      },
    },
    {
      Name:      "deploy",
      ShortName: "d",
      Usage:     "deploy a cluster",
      Action: func(c *cli.Context) {
        var name = c.Args().First()
        if len(name) == 0 {
          name = DEFAULT_CLUSTER_NAME
        }

        town := NewTown()
        town.ReadFile(name)
        // town.Connect()
        // town.Provision(false)
        // town.StopContainers(false)
        // town.RemoveContainers(false)
      },
    },
    {
      Name:      "ps",
      ShortName: "p",
      Usage:     "display information about running cluster",
      Action: func(c *cli.Context) {
        var name = c.Args().First()
        if len(name) == 0 {
          name = DEFAULT_CLUSTER_NAME
        }

        town := NewTown()
        town.ReadFile(name)
        town.Connect()
        town.Provision(false)
        town.Info()
      },
    },
  }

  if err := app.Run(os.Args); err != nil {
    log.Println("Error: ", err)
  }
}