package main

import (
	"context"
	"html"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
)

const (
	TITLE = "JBoss Tools Aggregated Feed"
	LINK  = "https://tools.jboss.org"
)

var (
	urls = map[string]string{
		"JBoss Tools":       "https://tools.jboss.org/blog/news.atom",
		"JBoss Blogs":       "https://www.jboss.org/atom.xml",
		"Red Hat Developer": "https://developers.redhat.com/blog/feed/",
		"Quarkus":           "https://quarkus.io/feed",
	}
)

type Items []*gofeed.Item

func (a Items) Len() int { return len(a) }

func (a Items) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a Items) Less(i, j int) bool { return !(*a[i].PublishedParsed).Before(*a[j].PublishedParsed) }

func main() {
	parentCtx := context.Background()

	// fetch lates info from feeds
	items := []*gofeed.Item{}
	for name, url := range urls {
		log.Printf("checking: %s @ %s\n", name, url)
		ctx, cancel := context.WithTimeout(parentCtx, 15*time.Second)
		fp := gofeed.NewParser()
		feed, err := fp.ParseURLWithContext(url, ctx)
		if err != nil {
			log.Fatalf("failed to parse feed (%s): %s", url, err)
		}
		cancel()

		// save items
		for i := range feed.Items {
			feed.Items[i].Title = name + ": " + feed.Items[i].Title
			items = append(items, feed.Items[i])
		}
	}

	// sort items
	sort.Sort(Items(items)) // newest first

	// create new feed
	feed := &feeds.Feed{
		Title:   TITLE,
		Link:    &feeds.Link{Href: LINK},
		Updated: time.Now(),
	}
	for i := range items {
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       items[i].Title,
			Link:        &feeds.Link{Href: items[i].Link},
			Description: items[i].Description,
			Created:     *items[i].PublishedParsed,
			Id:          items[i].GUID,
		})
	}

	data, err := feed.ToAtom()
	if err != nil {
		log.Fatalf("failed to generate atom feed file: %s", err)
	}

	// remove <?xml version="1.0" encoding="UTF-8"?>
	data = strings.Replace(data, `<?xml version="1.0" encoding="UTF-8"?>`, "", 1)
	// replace double &&
	data = strings.Replace(data, `&&`, "&#38;&#38;", 1)
	

	f, err := os.OpenFile("rss.xml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := f.Write([]byte(html.UnescapeString(data))); err != nil {
		log.Fatalf("failed to write rss file: %s", err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

}
