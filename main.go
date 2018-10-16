package main

import (
	"fmt"
	"log"
	"regexp"
	"search_engine_crawler/crawler"
	"search_engine_crawler/filter"
	"search_engine_crawler/reporter"
)

const (
	searchTerm = "jrieck1991"
)

func main() {
	// base queries
	queries := map[string]string{
		"Google":     fmt.Sprintf("https://www.google.com/search?q=site%%3A%s&start=", searchTerm),
		"Bing":       fmt.Sprintf("https://www.bing.com/search?q=site%%3A%s&start=", searchTerm),
		"DuckDuckGo": fmt.Sprintf("https://www.duckduckgo.com/html/?q=site%%3A%s", searchTerm),
	}

	// begin crawling given searches
	results, err := crawler.StartCrawl(queries)
	if err != nil {
		log.Fatalln(err)
	}

	// filter by category
	categories := []filter.Category{
		filter.Category{
			Name:    "jack",
			Regexp:  regexp.MustCompile(`jack`),
			Allowed: true,
		},
		filter.Category{
			Name:    "webcache",
			Regexp:  regexp.MustCompile(`webcache`),
			Allowed: false,
		},
		filter.Category{
			Name:        "http",
			Regexp:      regexp.MustCompile(`[^http{s}]`),
			Allowed:     false,
			Description: "deny if link doesn't contain http or https",
		},
	}

	// filter urls
	text, err := filter.Results(results, categories)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(text)

	// send email report
	if err := reporter.SendReport(text, queries); err != nil {
		log.Fatalln(err)
	}
}

// set email creds with env vars
// Add search URLS to categories
