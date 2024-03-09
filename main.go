package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"dario.cat/mergo"
	"github.com/BurntSushi/toml"
	"github.com/urfave/cli/v2"
)

type (
	Config struct {
		Categories map[string]Category `toml:"categories"`
	}

	Category struct {
		Editor    string `toml:"editor"`
		Stubs     string `toml:"stubs"`
		Path      string `toml:"path"`
		Filename  string `toml:"filename"`
		Extension string `toml:"ext"`
	}
)

func ParseDate(value string) (time.Time, error) {
	datePattern := "(?P<Date>(?:0?[1-9])|(?:[12][1-9])|(?:3[01]))"
	monthPattern := "(?P<Month>(?:0?[1-9])|(?:1[0-2]))"
	yearPattern := "(?P<Year>(?:[0-9]{2})|(?:[0-9]{4}))"
	separatorPattern := "(?:[.-_/])"
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

func (c Config) ResolveCategory(name string) (Category, error) {
	category, found := c.Categories[name]

	if !found {
		return Category{}, errors.New("category not found")
	}

	defaultCategory, found := c.Categories["default"]

	if !found {
		return Category{}, errors.New("default category not found")
	}

	mergo.Merge(&category, defaultCategory)

	return category, nil
}

func main() {
	var configPath string
	var config Config
	var silent bool

	app := &cli.App{
		Name:  "notes",
		Usage: "Manage personal notes",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"C"},
				Value:       "./config.toml",
				Destination: &configPath,
			},
			&cli.BoolFlag{
				Name:        "silent",
				Aliases:     []string{"s"},
				Value:       false,
				Destination: &silent,
			},
		},
		Before: func(ctx *cli.Context) error {
			if _, err := toml.DecodeFile(configPath, &config); err != nil {
				if !silent {
					log.Fatalln(err)
				} else {
					return errors.New("failed decode config")
				}
			}

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "new",
				Aliases: []string{"n"},
				Usage:   "Generate new note",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "category",
						Aliases: []string{"c"},
						Value:   "default",
					},
					&cli.StringFlag{
						Name:    "sub-directory",
						Aliases: []string{"D"},
					},
					&cli.StringFlag{
						Name:        "date",
						Aliases:     []string{"d"},
						DefaultText: "today",
					},
					&cli.BoolFlag{
						Name:    "first-directory",
						Aliases: []string{"f"},
					},
					&cli.BoolFlag{
						Name:    "new",
						Aliases: []string{"n"},
					},
					&cli.BoolFlag{
						Name:    "open-editor",
						Aliases: []string{"e"},
					},
				},
				Action: func(ctx *cli.Context) error {
                    var category Category
                    var err error

					if category, err = config.ResolveCategory(ctx.String("category")); err != nil {
						if !silent {
							log.Fatalln(errors.New("category not found"))
						} else {
							return nil
						}
					}

					var date time.Time

					if ctx.String("date") != "" {
						if date, err = ParseDate(ctx.String("date")); err != nil {
							if !silent {
								log.Fatalln(err)
							} else {
								return nil
							}
						}
					} else {
						date = time.Now().Local()
					}

					path := filepath.Join(category.Path, date.Format(category.Filename)+category.Extension)
					folder := filepath.Dir(path)

					if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
						return errors.New("note already exist")
					}

					if err := os.MkdirAll(folder, os.ModePerm); err != nil {
						log.Fatalln(err)
					}

					if _, err := os.Create(path); err != nil {
						log.Fatalln(err)
					}

					log.Println("note " + path + " created")

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
