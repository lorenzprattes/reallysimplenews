package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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
	Links []string
}

type Feed struct {
	Title string
	Link  string
	Items []Item
}

type Item struct {
	Title    string
	Link     string
	Comments string
}

var client = http.Client{
	Timeout: 2 * time.Second,
}

func initLinksFromCookies(w http.ResponseWriter, r *http.Request) ([]string, error) {
	var links []string
	cookie_read, err := r.Cookie("feeds")
	if err != nil {
		fmt.Printf("Cannot get cookie: %v\n", err)
		jsonArray, err := json.Marshal(config.FeedURLs)
		if err != nil {
			return nil, err
		}

		cookie := http.Cookie{
			Name:    "feeds",
			Value:   url.QueryEscape(string(jsonArray)),
			Path:    "/",
			Expires: time.Now().Add(24 * time.Hour * 399), // Expire in 1 day
		}

		http.SetCookie(w, &cookie)
	} else {
		jsonArray, err := url.QueryUnescape(cookie_read.Value)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(jsonArray), &links)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Array from cookie: %v\n", links)
	}
	return links, nil
}

func addLinkToCookies(w http.ResponseWriter, r *http.Request, link string) error {
	links, err := initLinksFromCookies(w, r)
	if err != nil {
		return err
	}
	for _, l := range links {
		if l == link {
			//TODO disabled temporarily
			fmt.Print("")
			//return errors.New("link already exists in feeds")
		}
	}

	links = append(links, link)
	jsonArray, err := json.Marshal(links)
	if err != nil {
		return err
	}
	cookie := http.Cookie{
		Name:    "feeds",
		Value:   url.QueryEscape(string(jsonArray)),
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour * 399),
	}
	http.SetCookie(w, &cookie)
	return nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	links, err := initLinksFromCookies(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("templates/main.html")
	if err != nil {
		panic(err)
	}
	page := Page{Title: config.SiteTitle, Body: getFeeds(links), Links: config.FeedURLs}

	err = t.Execute(w, page)
	if err != nil {
		panic(err)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	links, err := initLinksFromCookies(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := template.ParseFiles("templates/edit.html")
	if err != nil {
		panic(err)
	}
	page := Page{Title: config.SiteTitle, Body: getFeeds(links), Links: config.FeedURLs}

	err = t.Execute(w, page)
	if err != nil {
		panic(err)
	}
}

func ensureHttpsUrl(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	return "https://" + url
}

func checkLink(link string) (string, error) {
	fmt.Println("Link" + link)
	parsedUrl, err := url.Parse(link)
	if err != nil {
		fmt.Println("Error with link parsing")
		return "", err
	}
	link = ensureHttpsUrl(parsedUrl.String())
	resp, err := client.Get(link)
	if err != nil {
		fmt.Println("Error with reaching server")
		return "", err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	fmt.Println(contentType)
	if strings.Contains(contentType, "application/rss+xml") || strings.Contains(contentType, "application/xml") || strings.Contains(contentType, "text/xml") {
		fmt.Println("Link is valid")
		return link, nil
	}
	fmt.Println("Error with link final")
	return "", err
}

func getPostContents(w http.ResponseWriter, r *http.Request) (string, error) {
	if r.Method != "POST" {
		http.Error(w, "Method is not supported.", http.StatusMethodNotAllowed)
		return "", errors.New("Method is not supported.")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		fmt.Println("Error reading request body")
		return "", errors.New("Error reading request body")
	}
	defer r.Body.Close()
	receivedText := string(body)
	fmt.Printf("Received link: %s\n", receivedText)
	return receivedText, nil
}

func addLinkHandler(w http.ResponseWriter, r *http.Request) {
	receivedText, err := getPostContents(w, r)
	if err != nil {
		return
	}
	link, err := checkLink(receivedText)
	if err != nil {
		http.Error(w, "Error with link", http.StatusInternalServerError)
		fmt.Println("Error with link")
		return
	}
	fmt.Println("Link will be added:", link)
	//TODO add check if already exists
	addLinkToCookies(w, r, link)
}

func removeLinkHandler(w http.ResponseWriter, r *http.Request) {
	receivedText, err := getPostContents(w, r)
	if err != nil {
		return
	}
	links, err := initLinksFromCookies(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, l := range links {
		if l == receivedText {
			links = append(links[:i], links[i+1:]...)
			break
		}
	}
	jsonArray, err := json.Marshal(links)
	if err != nil {
		return
	}
	cookie := http.Cookie{
		Name:    "feeds",
		Value:   url.QueryEscape(string(jsonArray)),
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour * 399),
	}
	http.SetCookie(w, &cookie)
}

func changeOrderHandler(w http.ResponseWriter, r *http.Request) {
	receivedText, err := getPostContents(w, r)
	if err != nil {
		return
	}
	var split = strings.Split(receivedText, "http")[1:]
	fmt.Println("Split: ", split)
	if len(split) != 2 {
		http.Error(w, "Error with links", http.StatusBadRequest)
	}
	for i, v := range split {
		fmt.Println(v)
		split[i] = "http" + v
	}
	for _, v := range split {
		fmt.Println(v)
	}
	links, err := initLinksFromCookies(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Links before exchanging: ", links)
	var index0, index1 = -1, -1
	for i, v := range links {
		if v == split[0] {
			index0 = i
		} else if v == split[1] {
			index1 = i
		}
		if index0 != -1 && index1 != -1 {
			break
		}
	}
	fmt.Println("Indexes: ", index0, index1)
	if index0 == -1 || index1 == -1 {
		http.Error(w, "Error with links", http.StatusBadRequest)
		return
	}

	links[index0], links[index1] = links[index1], links[index0]
	fmt.Println("Links after exchanging: ", links)
	jsonArray, err := json.Marshal(links)
	if err != nil {
		return
	}
	cookie := http.Cookie{
		Name:    "feeds",
		Value:   url.QueryEscape(string(jsonArray)),
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour * 399),
	}
	http.SetCookie(w, &cookie)
}

func main() {
	config = loadConfig()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/addlink", addLinkHandler)
	http.HandleFunc("/removelink", removeLinkHandler)
	http.HandleFunc("/changeorder", changeOrderHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
	fmt.Println(err)
}

func getFeeds(links []string) []Feed {
	var config = loadConfig()
	var feeds []Feed
	//loop through a string array and print it
	for _, feedurl := range links {
		fp := gofeed.NewParser()
		feed, err := fp.ParseURL(feedurl)
		if err != nil {
			fmt.Println("Feed url: " + feedurl + " is not valid")
			fmt.Println(err)
			continue
			panic(err)
		}
		feeds = append(feeds, Feed{Title: feed.Title, Link: feedurl, Items: getItems(feed, config.Num_items)})
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
