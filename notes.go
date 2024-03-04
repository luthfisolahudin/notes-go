package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/urfave/cli/v2"
)

type Config struct {
	Daily struct {
		Path      string `toml:"path"`
		Filename  string `toml:"filename"`
		Extension string `toml:"ext"`
	} `toml:"daily"`

	Project struct {
		Path      string `toml:"path"`
		Filename  string `toml:"filename"`
		Extension string `toml:"ext"`
	} `toml:"project"`
}

func parseDate(value string) (time.Time, error) {
	datePattern := "(?P<Date>(?:0?[1-9])|(?:[12][1-9])|(?:3[01]))"
	monthPattern := "(?P<Month>(?:0?[1-9])|(?:1[0-2]))"
	yearPattern := "(?P<Year>(?:[0-9]{2})|(?:[0-9]{4}))"
	separatorPattern := "(?:[ -])"
	now := time.Now().Local()

	re := regexp.MustCompile("^" + datePattern + "$")
	if re.MatchString(value) {
		date, _ := strconv.Atoi(value)

		return time.Date(now.Year(), now.Month(), date, 0, 0, 0, 0, time.Local), nil
	}

	re = regexp.MustCompile(
		"^" + monthPattern + separatorPattern + datePattern + "$",
	)
	if re.MatchString(value) {
		matches := re.FindStringSubmatch(value)
		month, _ := strconv.Atoi(matches[re.SubexpIndex("Month")])
		date, _ := strconv.Atoi(matches[re.SubexpIndex("Date")])

		return time.Date(now.Year(), time.Month(month), date, 0, 0, 0, 0, time.Local), nil
	}

	re = regexp.MustCompile(
		"^" + yearPattern + separatorPattern + monthPattern + separatorPattern + datePattern + "$",
	)
	if re.MatchString(value) {
		matches := re.FindStringSubmatch(value)
		year, _ := strconv.Atoi(matches[re.SubexpIndex("Year")])
		month, _ := strconv.Atoi(matches[re.SubexpIndex("Month")])
		date, _ := strconv.Atoi(matches[re.SubexpIndex("Date")])

		return time.Date(year, time.Month(month), date, 0, 0, 0, 0, time.Local), nil
	}

	return time.Time{}, errors.New("unrecognized date")
}

func main() {
	const DefaultConfigPath = "./config.toml"
	var config Config

	if _, err := toml.DecodeFile(DefaultConfigPath, &config); err != nil {
		log.Panic(err)
	}

	app := &cli.App{
		Name:  "notes",
		Usage: "Manage personal notes",
		Commands: []*cli.Command{
			{
				Name:      "daily",
				Aliases:   []string{"d"},
				Usage:     "Generate daily note",
				ArgsUsage: "[date]",
				Action: func(ctx *cli.Context) error {
					var date time.Time

					if ctx.Args().Present() {
						var err error

						if date, err = parseDate(strings.Join(ctx.Args().Slice(), " ")); err != nil {
							log.Panic(err)

							return nil
						}
					} else {
						date = time.Now().Local()
					}

					path := filepath.Join(config.Daily.Path, date.Format(config.Daily.Filename)+config.Daily.Extension)
					folder := filepath.Dir(path)

					if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
						log.Fatal(errors.New("note already exist"))
					}

					if err := os.MkdirAll(folder, os.ModePerm); err != nil {
						log.Panic(err)
					}

					if _, err := os.Create(path); err != nil {
						log.Panic(err)
					}

					log.Print("note " + path +  " created")

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Panic(err)
	}
}
