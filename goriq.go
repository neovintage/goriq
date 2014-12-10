package main

import (
  "os"

  "github.com/codegangsta/cli"
)

func main() {
  app := cli.NewApp()
  app.Name = "goriq"
  app.Version = Version
  app.Usage = ""
  app.Author = "Rimas Silkaitis"
  app.Email = "rimas@chartio.com"
  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name: "config",
      Value: "config.json",
      Usage: "Configuration for goriq",
    },
  }
  app.Commands = Commands
  app.Action = doExtract
  app.Run(os.Args)
}
