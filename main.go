package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/olekukonko/tablewriter"
)

type feedItem struct {
	Feed string
	Last string
}

type config struct {
	Exec  string
	Feeds []feedItem
}

func main() {
	home, ok := os.LookupEnv("HOME")
	if !ok {
		panic(errors.New("HOME is not set"))
	}

	xdgConfigHome, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		xdgConfigHome = fmt.Sprintf("%v/.config", home)
	}

	var config string
	var standalone bool
	var interval int

	flag.StringVar(&config, "config", fmt.Sprintf("%v/rssd/config.json", xdgConfigHome), "path to config file")
	flag.BoolVar(&standalone, "standalone", false, "whether rssd should loop on it's own")
	flag.IntVar(&interval, "interval", 5, "interval in minutes for standalone mode")
	flag.Parse()

	if standalone {
		for {
			d(config)
			time.Sleep(time.Duration(interval) * time.Minute)
		}
	} else {
		d(config)
		os.Exit(0)
	}
}

func d(c string) {
	err := initConfig(c)
	if err != nil {
		panic(err.Error())
	}

	if len(flag.Args()) == 0 {
		err := synchronize(c)
		if err != nil {
			panic(err.Error())
		}
		return
	}

	if flag.Arg(0) == "add-feed" {
		if len(flag.Args()) < 2 {
			fmt.Fprintln(os.Stderr, "insufficient number of arguments")
			os.Exit(2)
		}
		err := addFeed(c, flag.Arg(1))
		if err != nil {
			panic(err.Error())
		}
		os.Exit(0)
	}

	if flag.Arg(0) == "list-feed" {
		err := listFeed(c)
		if err != nil {
			panic(err.Error())
		}
		os.Exit(0)
	}

	if flag.Arg(0) == "remove-feed" {
		i, err := strconv.Atoi(flag.Arg(1))
		if err != nil {
			panic(err.Error())
		}
		err = removeFeed(c, i)
		if err != nil {
			panic(err.Error())
		}
		os.Exit(0)
	}

	if flag.Arg(0) == "set-exec" {
		if len(flag.Args()) < 2 {
			fmt.Fprintln(os.Stderr, "insufficient number of arguments")
			os.Exit(2)
		}
		err := setExec(c, flag.Arg(1))
		if err != nil {
			panic(err.Error())
		}
		os.Exit(0)
	}

	os.Exit(1)
}

func synchronize(p string) error {
	c, err := readConfig(p)
	if err != nil {
		return err
	}

	for i, v := range c.Feeds {
		f, err := getFeed(v.Feed)
		if err != nil {
			return err
		}

		if f.Items[0].Link != v.Last {
			s := c.Exec
			for q, r := range map[string]string{
				"&title":            f.Title,
				"&desc":             f.Description,
				"&lang":             f.Language,
				"&item_title":       f.Items[0].Title,
				"&item_link":        f.Items[0].Link,
				"&item_pubDate":     f.Items[0].Published,
				"&item_desc":        f.Items[0].Description,
				"&item_authorName":  f.Items[0].Author.Name,
				"&item_authorEmail": f.Items[0].Author.Email,
			} {
				s = strings.ReplaceAll(s, q, r)
				s = os.ExpandEnv(s)
			}

			err = exec.Command("sh", "-c", s).Run()
			if err != nil {
				return err
			}

			v.Last = f.Items[0].Link
			c.Feeds[i] = v
		}
	}

	err = writeConfig(p, c)
	if err != nil {
		return err
	}

	return nil
}

func setExec(p string, e string) error {
	c, err := readConfig(p)
	if err != nil {
		return err
	}

	c.Exec = e

	err = writeConfig(p, c)
	if err != nil {
		return err
	}

	return nil
}

func getFeed(url string) (*gofeed.Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	defer cancel()

	f, err := (gofeed.NewParser()).ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func listFeed(p string) error {
	s, err := readConfig(p)
	if err != nil {
		return err
	}

	t := tablewriter.NewWriter(os.Stdout)
	t.SetHeader([]string{"Index", "Feed", "Last"})

	for i, v := range s.Feeds {
		t.Append([]string{fmt.Sprint(i), v.Feed, v.Last})
	}

	t.Render()

	return nil
}

func addFeed(p string, feed string) error {
	s, err := readConfig(p)
	if err != nil {
		return err
	}

	flag := false
	for _, v := range s.Feeds {
		if v.Feed == feed {
			flag = true
		}
	}
	if flag {
		return errors.New("duplicate feed")
	}

	f, err := getFeed(feed)
	if err != nil {
		return err
	}

	s.Feeds = append(s.Feeds, feedItem{feed, f.Items[0].Link})

	err = writeConfig(p, s)
	if err != nil {
		return err
	}

	return nil
}

func removeFeed(p string, i int) error {
	c, err := readConfig(p)
	if err != nil {
		return err
	}

	if len(c.Feeds) <= i {
		return errors.New("invalid index")
	}

	fmt.Println("removed: ", c.Feeds[i].Feed)

	// https://stackoverflow.com/a/37335777/11143333
	c.Feeds = append(c.Feeds[:i], c.Feeds[i+1:]...)

	err = writeConfig(p, c)
	if err != nil {
		return err
	}

	return nil
}

func readConfig(p string) (*config, error) {
	f, err := os.OpenFile(p, os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	d := make([]byte, 1000)
	_, err = f.Read(d)
	if err != nil {
		return nil, err
	}

	d = bytes.Trim(d, "\x00")

	var s config
	err = json.Unmarshal(d, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func writeConfig(p string, c *config) error {
	f, err := os.OpenFile(p, os.O_RDWR, 0600)
	if err != nil {
		return err
	}

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = f.Truncate(0)
	if err != nil {
		return err
	}

	_, err = f.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func initConfig(p string) error {
	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(p), 0755)
		if err != nil {
			return err
		}

		f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return err
		}

		c, err := json.Marshal(config{})
		if err != nil {
			return err
		}

		f.Truncate(0)
		_, err = f.Write(c)
		if err != nil {
			return err
		}
	}

	return nil
}
