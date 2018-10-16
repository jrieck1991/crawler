package crawler

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

const (
	cache = "cache"
)

// StartCrawl begins crawling for urls on given search engine domain and URL
func StartCrawl(queries map[string]string) ([]string, error) {
	//URLs := []SearchResult{}
	var URLs []string

	for engine, query := range queries {
		log.Printf("crawling %s", engine)

		result := getLinks(engine, query)

		if len(result) == 0 {
			log.Printf("No URLs gathered from %s", engine)
			continue
		}

		for _, u := range result {
			URLs = append(URLs, u)
		}
	}

	return URLs, nil
}

// getLinks collects urls from google and bing
func getLinks(engine, query string) []string {
	// paginate and collect urls
	pageCollector := colly.NewCollector(
		colly.CacheDir(cache),
	)

	// gather url pages
	searchURLs, err := getPages(pageCollector, query)
	if err != nil {
		log.Fatalln(err)
	}

	// setup url collector
	var URLs []string
	urlCollector := colly.NewCollector(
		colly.CacheDir(cache),
	)

	urlCollector.OnHTML("a", func(e *colly.HTMLElement) {
		URLs = append(URLs, e.Attr("href"))
	})

	urlCollector.OnError(func(r *colly.Response, err error) {
		log.Println(r.Request.URL, string(r.Body), err)
	})

	// if results fit in a single page
	if len(searchURLs) == 0 {
		if err := urlCollector.Visit(query); err != nil {
			log.Fatalln(err)
		}
		return URLs
	}

	// collect URLs for all given urls
	for _, absoluteURL := range searchURLs {
		urlCollector.Visit(absoluteURL)
	}

	return URLs
}

// getPages iterates through all pages of a web search and returns those urls
func getPages(pageCollector *colly.Collector, searchURL string) ([]string, error) {
	var searchURLs []string

	pageCollector.OnHTML("a", func(e *colly.HTMLElement) {
		time.Sleep(time.Duration(rand.Intn(30)) * time.Microsecond)
		if strings.Contains(e.Text, "Next") {
			absoluteURL := e.Request.AbsoluteURL(e.Attr("href"))
			searchURLs = append(searchURLs, absoluteURL)
			e.Request.Visit(absoluteURL)
		}
	})

	pageCollector.OnError(func(r *colly.Response, err error) {
		log.Println(r.Request.URL, string(r.Body), err)
	})

	if err := pageCollector.Visit(searchURL); err != nil {
		return searchURLs, err
	}

	return searchURLs, nil
}
