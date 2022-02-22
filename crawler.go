package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// checkKeywordMatch checks if the given keyword is part of the content on the page
// (not an attribute, tag, or comment) does account for content hidden via CSS (and/or JS).
// Matching is case-insensitive.
func (wc *WebCrawler) checkKeywordMatch(doc *goquery.Document, keyword string) (bool, string) {
	// Remove any script and style tags, else these will show up within the text content.
	doc.Find("script").Remove()
	doc.Find("style").Remove()

	// Use goquery to "render" (for lack of a better term) the page content to plain text.
	contentText := doc.Text()

	// Early return if we have no content.
	if len(contentText) == 0 {
		return false, ""
	}

	startIndex := strings.Index(strings.ToLower(contentText), strings.ToLower(keyword))

	// Calcuate the indexes for getting the context around the match.
	contextSize := 6
	contextStartIndex := startIndex - contextSize
	contextEndIndex := startIndex + len(keyword) + contextSize
	if contextStartIndex < 0 {
		contextStartIndex = 0
	}
	if contextEndIndex >= len(contentText) {
		contextEndIndex = len(contentText) - 1
	}

	// Remove any newlines and whitespace suffix/prefixes from the output context.
	context := contentText[contextStartIndex:contextEndIndex]
	context = strings.ReplaceAll(context, "\n", "")
	context = strings.Trim(context, " ")

	return startIndex >= 0, context
}

func dedupeList(input []string) (output []string) {
	tmpMap := make(map[string]bool)
	for _, val := range input {
		tmpMap[val] = true
	}

	for key, _ := range tmpMap {
		output = append(output, key)
	}

	return
}

// crawlPage returns a list of found URLs, if the page contains content matching the keyword, and context text around the match (if any).
func (wc *WebCrawler) crawlPage(sourceURL string, keyword string) (foundURLs []string, isKeywordMatch bool, matchContext string, err error) {
	// Request the page
	resp, err := http.Get(sourceURL)
	if err != nil {
		return nil, false, "", err
	}
	defer resp.Body.Close()

	// Warn for non-200 status code. We may need to do something different here. (retry logic?)
	if resp.StatusCode != 200 {
		fmt.Printf("WARN: got status code %d for %s\n", resp.StatusCode, sourceURL)
	}

	// Parse the HTML into a goquery document type
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, false, "", err
	}

	// Check if the page is a keyword match while we have the document.
	isKeywordMatch, matchContext = wc.checkKeywordMatch(doc, keyword)

	// Parse the source URL (used to fix relative links later)
	rootURL, err := url.Parse(sourceURL)
	if err != nil {
		return nil, false, "", err
	}

	// Find the "a" tags via a CSS selector and collect a list a of URLs from the page.
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		tagURL, exists := s.Attr("href")
		if exists && len(tagURL) > 0 {
			// Skip URL hash fragments
			// (e.g. links to section on page)
			if tagURL[0] == '#' {
				return
			}

			// Relative URL, prepend the schema and hostname
			if tagURL[0] == '/' {
				relativeURL, err := url.Parse(tagURL)
				if err != nil {
					return
				}

				tagURL = rootURL.ResolveReference(relativeURL).String()
			}

			foundURLs = append(foundURLs, tagURL)
		}
	})

	return dedupeList(foundURLs), isKeywordMatch, matchContext, nil
}

// WebCrawler is a type which contains web crawling functionality
type WebCrawler struct {
}

// PageResult contains information about a single keyword-based page crawl.
type PageResult struct {
	URL            string
	IsKeywordMatch bool
	MatchContext   string
}

func NewWebCrawler() *WebCrawler {
	return &WebCrawler{}
}

// PerformKeywordCrawl crawls starting at the provided URL for pages with content matching the given keyword.
func (wc *WebCrawler) PerformKeywordCrawl(startURL string, keyword string, depth int) (TotalCrawled int, Matches []*PageResult, err error) {
	results := make(map[string]*PageResult)
	crawlURLs := []string{startURL}

	// Iterate sequentially over each crawled URL, handling the concept of "depth" as each iteration of the loop.
	// TODO: This would significantly benefit from parallelism.
	for i := 0; i < depth; i++ {
		var newURLs []string
		for _, crawlURL := range crawlURLs {
			foundURLs, isKeywordMatch, matchContext, err := wc.crawlPage(crawlURL, keyword)
			if err != nil {
				return 0, nil, err
			}

			// Store the results from the page crawl.
			results[crawlURL] = &PageResult{
				URL:            crawlURL,
				IsKeywordMatch: isKeywordMatch,
				MatchContext:   matchContext,
			}

			// Queue the new URLs if they haven't already been crawled AND are on the same domain as the current URL.
			for _, foundURL := range foundURLs {
				// If it's not in the results, we haven't crawled it.
				if _, ok := results[foundURL]; !ok {
					// Compare the hostnames.
					currentURLInfo, err := url.Parse(crawlURL)
					if err != nil {
						continue
					}

					foundURLInfo, err := url.Parse(foundURL)
					if err != nil {
						continue
					}

					if foundURLInfo.Hostname() == currentURLInfo.Hostname() {
						newURLs = append(newURLs, foundURL)
					}
				}
			}
		}

		// After all the current URLs have been crawled, crawl the new URLs.
		crawlURLs = newURLs
	}

	// Convert results map into slice of keyword matches.
	var matches []*PageResult
	for _, result := range results {
		if result.IsKeywordMatch {
			matches = append(matches, result)
		}
	}

	return len(results), matches, nil
}
