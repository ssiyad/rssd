package main

import (
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
		fmt.Println(errors.New("HOME is not set"))
		os.Exit(1)
	}

	configDir, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		configDir = fmt.Sprintf("%v/.config", home)
	}

	var config string
	var standalone bool
	var interval int

	flag.StringVar(&config, "config", fmt.Sprintf("%v/rssd/config.json", configDir), "path to config file")
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
		fmt.Println(err)
		os.Exit(1)
	}

	if len(flag.Args()) == 0 {
		err := synchronize(c)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
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
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if flag.Arg(0) == "list-feed" {
		err := listFeed(c)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if flag.Arg(0) == "remove-feed" {
		i, err := strconv.Atoi(flag.Arg(1))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = removeFeed(c, i)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
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
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	os.Exit(1)
}

func synchronize(p string) error {
	c, err := readConfig(p)
	if err != nil {
		return fmt.Errorf("synchronize -> %v", err)
	}

	for i, v := range c.Feeds {
		f, err := getFeed(v.Feed)
		if err != nil {
			return fmt.Errorf("synchronize -> %v", err)
		}

		for _, item := range f.Items {
			if item.Link == v.Last {
				break
			}

			s := c.Exec
			s = os.ExpandEnv(s)

			_, err = os.Stat(s)
			if err == nil {
				e, err := os.ReadFile(s)
				if err != nil {
					return fmt.Errorf("synchronize -> %v", err)
				}
				s = string(e)
			}

			var itemAuthorName string
			var itemAuthorEmail string

			if item.Authors != nil {
				itemAuthorName = item.Authors[0].Name
				itemAuthorEmail = item.Authors[0].Email
			}

			for q, r := range map[string]string{
				"&title":            f.Title,
				"&desc":             f.Description,
				"&lang":             f.Language,
				"&item_title":       item.Title,
				"&item_link":        item.Link,
				"&item_pubDate":     item.Published,
				"&item_desc":        item.Description,
				"&item_authorName":  itemAuthorName,
				"&item_authorEmail": itemAuthorEmail,
			} {
				s = strings.ReplaceAll(s, q, r)
			}

			err = exec.Command("sh", "-c", s).Run()
			if err != nil {
				return fmt.Errorf("synchronize -> %v", err)
			}
		}
		v.Last = f.Items[0].Link
		c.Feeds[i] = v
	}

	err = writeConfig(p, c)
	if err != nil {
		return fmt.Errorf("synchronize -> %v", err)
	}

	return nil
}

func setExec(p string, e string) error {
	c, err := readConfig(p)
	if err != nil {
		return fmt.Errorf("setExec -> %v", err)
	}

	c.Exec = e

	err = writeConfig(p, c)
	if err != nil {
		return fmt.Errorf("setExec -> %v", err)
	}

	return nil
}

func getFeed(url string) (*gofeed.Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	defer cancel()

	f, err := (gofeed.NewParser()).ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, fmt.Errorf("getFeed -> %v", err)
	}

	return f, nil
}

func listFeed(p string) error {
	s, err := readConfig(p)
	if err != nil {
		return fmt.Errorf("listFeed -> %v", err)
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
		return fmt.Errorf("addFeed -> %v", err)
	}

	flag := false
	for _, v := range s.Feeds {
		if v.Feed == feed {
			flag = true
		}
	}
	if flag {
		return fmt.Errorf("addFeed -> %v", err)
	}

	f, err := getFeed(feed)
	if err != nil {
		return fmt.Errorf("addFeed -> %v", err)
	}

	s.Feeds = append(s.Feeds, feedItem{feed, f.Items[0].Link})

	err = writeConfig(p, s)
	if err != nil {
		return fmt.Errorf("addFeed -> %v", err)
	}

	return nil
}

func removeFeed(p string, i int) error {
	c, err := readConfig(p)
	if err != nil {
		return fmt.Errorf("removeFeed -> %v", err)
	}

	if len(c.Feeds) <= i {
		return fmt.Errorf("removeFeed -> %v", err)
	}

	fmt.Println("removed: ", c.Feeds[i].Feed)

	// https://stackoverflow.com/a/37335777/11143333
	c.Feeds = append(c.Feeds[:i], c.Feeds[i+1:]...)

	err = writeConfig(p, c)
	if err != nil {
		return fmt.Errorf("removeFeed -> %v", err)
	}

	return nil
}

func readConfig(p string) (*config, error) {
	f, err := os.OpenFile(p, os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("readConfig -> %v", err)
	}

	defer f.Close()

	fileStat, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("readConfig -> %v", err)
	}
	d := make([]byte, fileStat.Size())
	_, err = f.Read(d)
	if err != nil {
		return nil, fmt.Errorf("readConfig -> %v", err)
	}

	var s config
	err = json.Unmarshal(d, &s)
	if err != nil {
		return nil, fmt.Errorf("readConfig -> %v", err)
	}

	return &s, nil
}

func writeConfig(p string, c *config) error {
	f, err := os.OpenFile(p, os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("writeConfig -> %v", err)
	}

	b, err := json.MarshalIndent(c, "", "	")
	if err != nil {
		return fmt.Errorf("writeConfig -> %v", err)
	}

	err = f.Truncate(0)
	if err != nil {
		return fmt.Errorf("writeConfig -> %v", err)
	}

	_, err = f.Write(b)
	if err != nil {
		return fmt.Errorf("writeConfig -> %v", err)
	}

	return nil
}

func initConfig(p string) error {
	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(p), 0755)
		if err != nil {
			return fmt.Errorf("initConfig -> %v", err)
		}

		f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return fmt.Errorf("initConfig -> %v", err)
		}

		c, err := json.Marshal(config{})
		if err != nil {
			return fmt.Errorf("initConfig -> %v", err)
		}

		f.Truncate(0)
		_, err = f.Write(c)
		if err != nil {
			return fmt.Errorf("initConfig -> %v", err)
		}
	}

	return nil
}
