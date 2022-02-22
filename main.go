package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	// Parse the command line args.
	var startURL string
	var keyword string
	var crawlDepth int = 2

	if len(os.Args) <= 2 {
		fmt.Println("Not enough arguments!")
		fmt.Printf("Usage: %s <URL> <Keyword> [depth]\n", os.Args[0])
		os.Exit(1)
	}
	startURL = os.Args[1]
	keyword = os.Args[2]

	if len(os.Args) > 3 {
		var err error
		crawlDepth, err = strconv.Atoi(os.Args[3])
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("Starting crawler with URL:%s, Keyword:%s, Depth:%d\n", startURL, keyword, crawlDepth)

	// Create the web crawler and start crawling
	crawler := NewWebCrawler()
	totalCrawled, matches, err := crawler.PerformKeywordCrawl(startURL, keyword, crawlDepth)
	if err != nil {
		panic(err)
	}

	// Output the results
	fmt.Printf("Crawled %d pages. Found %d pages with the term '%s'\n", totalCrawled, len(matches), keyword)
	for _, match := range matches {
		fmt.Printf("%s => '%s'\n", match.URL, match.MatchContext)
	}
}
