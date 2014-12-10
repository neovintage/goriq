package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/mattes/migrate/migrate"
)

var (
	ErrTableDoesNotExist = errors.New("table does not exist")
	ErrNoPreviousVersion = errors.New("no previous version found")
	config               Config
)

type Auth struct {
	Username string
	Password string
}

type Config struct {
	Auth     Auth           `json:"api"`
	Database DatabaseConfig `json:"database"`
}

var Commands = []cli.Command{
	commandMigrate,
	commandExtract,
}

var commandMigrate = cli.Command{
	Name:  "migrate",
	Usage: "",
	Description: `
`,
	Action: doMigrate,
}

var commandExtract = cli.Command{
	Name:  "extract",
	Usage: "",
	Description: `
Will check the database if there are any pending migrations, migrate, and then
extract all of the data from RIQ into the database.
`,
	Action: doExtract,
}

func debug(v ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		log.Println(v...)
	}
}

func assert(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
		os.Exit(2)
	}
}

func assert_true(b bool, msg string) {
	if b {
		log.Fatal("Error: ", msg)
		os.Exit(2)
	}
}

func parseConfig(c *cli.Context) {
	configFile := c.GlobalString("config")
	if configFile == "" {
		log.Fatal("Missing Config File")
		os.Exit(2)
	}

	content, err := ioutil.ReadFile(configFile)
	assert(err)

	err = json.Unmarshal(content, &config)
	assert(err)
}

func doMigrate(c *cli.Context) {
	parseConfig(c)
	migrationPath := config.Database.Migrations
	if config.Database.Migrations == "" {
		migrationPath = "./migrations"
	}
	allErrors, ok := migrate.UpSync(config.Database.connectionString(), migrationPath)
	if !ok {
		for _, e := range allErrors {
			log.Fatal(e)
		}
		os.Exit(2)
	}
	log.Println("Migrations Complete")
}

func doExtract(c *cli.Context) {
	doMigrate(c)
	doTransformation()
}
