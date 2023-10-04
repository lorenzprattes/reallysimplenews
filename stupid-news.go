package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mmcdole/gofeed"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Port      int      `yaml:"Port"`
	SiteTitle string   `yaml:"SiteTitle"`
	FeedURLs  []string `yaml:"FeedURLs"`
	Num_items int      `yaml:"Num_items"`
}

var config Config

type Page struct {
	Title string
	Body  []Feed
}

type Feed struct {
	Title string
	Items []Item
}

type Item struct {
	Title    string
	Link     string
	Comments string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/main.html")
	if err != nil {
		panic(err)
	}
	page := Page{Title: config.SiteTitle, Body: getFeeds()}

	err = t.Execute(w, page)
	if err != nil {
		panic(err)
	}
}

func main() {
	config = loadConfig()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", indexHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
	fmt.Println(err)
}

func getFeeds() []Feed {
	var config = loadConfig()
	var feeds []Feed
	//loop through a string array and print it
	for _, feedurl := range config.FeedURLs {
		fp := gofeed.NewParser()
		feed, err := fp.ParseURL(feedurl)
		if err != nil {
			panic(err)
		}
		feeds = append(feeds, Feed{Title: feed.Title, Items: getItems(feed, config.Num_items)})
	}
	return feeds
}

// a function that exdtracts the title of the first 15 items in an rss feed and returns them as a slice of strings	/
func getItems(feed *gofeed.Feed, num_items int) []Item {
	var items []Item
	for i := 0; i < num_items; i++ {
		if strings.Contains(feed.Title, "Hacker News") {
			var first_split = strings.SplitAfter(feed.Items[i].Description, "Comments URL: <a href=\"")
			var second_split = strings.Split(first_split[1], "\"")
			var comments_link = second_split[0]
			items = append(items, Item{Title: feed.Items[i].Title, Link: feed.Items[i].Link, Comments: comments_link})
		} else {
			items = append(items, Item{Title: feed.Items[i].Title, Link: feed.Items[i].Link, Comments: ""})
		}
	}
	return items
}

func loadConfig() Config {
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}
	// Decode the YAML data into a struct
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Error decoding YAML: %v", err)
	}
	return config
}
