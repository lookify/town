package main

import (
  "os"
  log "github.com/Sirupsen/logrus"
  "github.com/codegangsta/cli"
//  "regexp"
)

func main() {
  app := cli.NewApp()
  app.Name = "town"
  app.Usage = "docker orchestration"
  app.Version = "0.0.1"
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
    log.SetOutput(os.Stderr)
    if c.Bool("debug") {
      log.SetLevel(log.DebugLevel)
    }
    return nil
  }

  app.Commands = []cli.Command{
    {
      Name:      "rerun",
      ShortName: "rerun",
      Usage:     "rerun a cluster",
      Action: func(c *cli.Context) {
        cluster := NewCluster()
        cluster.ReadFile()
        cluster.Stop()
        cluster.Run()
      },
    },
    {
      Name:      "run",
      ShortName: "run",
      Usage:     "run a cluster",
      Action: func(c *cli.Context) {
        cluster := NewCluster()
        cluster.ReadFile()
        //cluster.StopChanged()
        cluster.ResetChanged()
        // ,-3074457345618258603,3074457345618258602
// ([^,]+,?)+

        // r, _ := regexp.Compile("\\$\\{SCALE_NUM:(.+)\\}")
        // match := r.FindAllStringSubmatch("CASSANDRA_TOKEN=${SCALE_NUM:-9223372036854775808,-3074457345618258603,3074457345618258602}", -1)
        // if len(match) > 0 {
        //   //name = match[1]
        //   for i, m := range match {
        //     for x, n := range m {
        //       log.Println( i, ", ", x,  ": ", n )
        //     }
        //   }
        // }

        /*
        discovery := &token.TokenDiscoveryService{}
        discovery.Initialize("", 0)
        token, err := discovery.CreateCluster()
        if err != nil {
          log.Fatal(err)
        }
        fmt.Println(token)
        */
      },
    },
    {
      Name:      "stop",
      ShortName: "stop",
      Usage:     "stop a cluster",
      Action: func(c *cli.Context) {
        cluster := NewCluster()
        cluster.ReadFile()
        cluster.Stop()
      },
    },
    {
      Name:      "hosts",
      ShortName: "hosts",
      Usage:     "update hosts",
      Action: func(c *cli.Context) {
        cluster := NewCluster()
        cluster.ReadFile()
        // cluster.UpdateHosts()
      },
    },
    {
      Name:      "tree",
      ShortName: "tree",
      Usage:     "display cluster tree",
      Action: func(c *cli.Context) {
        cluster := NewCluster()
        cluster.ReadFile()
        cluster.PrintTree()
      },
    },
  }

  if err := app.Run(os.Args); err != nil {
    log.Fatal(err)
  }
}