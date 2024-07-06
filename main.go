package main

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"unicode"

	"github.com/gocolly/colly/v2"
)

func main() {
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("freedns.afraid.org"),
		colly.Async(true), // Enable asynchronous requests
	)

	// Set custom headers
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		r.Headers.Set("Accept-Encoding", "gzip, deflate, br, zstd")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Cache-Control", "no-cache")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Host", "freedns.afraid.org")
		r.Headers.Set("Pragma", "no-cache")
		r.Headers.Set("Sec-Ch-Ua", `"Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"`)
		r.Headers.Set("Sec-Ch-Ua-Mobile", "?0")
		r.Headers.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "none")
		r.Headers.Set("Sec-Fetch-User", "?1")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	})

	// Slice to store domains
	var subdomains []string
	var subdomainsMutex sync.Mutex // Mutex to protect the slice

	// On every <tr> element
	c.OnHTML("tr", func(e *colly.HTMLElement) {
		// Check if the second column contains "public"
		if e.ChildText("td:nth-child(2)") == "public" {
			// Extract subdomain from the first column
			subdomain := e.ChildText("td:nth-child(1) a[href*='/subdomain/edit.php']")
			if subdomain != "" {
				subdomainsMutex.Lock()
				subdomains = append(subdomains, subdomain)
				subdomainsMutex.Unlock()
			}
		}
	})

	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %v failed with response: %v. Error: %v", r.Request.URL, r, err)
	})

	// Use WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start scraping from page 1 to 270
	for i := 1; i <= 270; i++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			url := fmt.Sprintf("https://freedns.afraid.org/domain/registry/page-%d.html", page)
			err := c.Visit(url)
			if err != nil {
				log.Fatalf("Failed to visit %s: %v", url, err)
			}
		}(i)
	}

	// Wait for all requests to finish
	wg.Wait()

	// Wait for scraping to finish
	c.Wait()

	// Sort subdomains by length first, then alphabetically, and put domains with numbers at the end
	sort.Slice(subdomains, func(i, j int) bool {
		if len(subdomains[i]) != len(subdomains[j]) {
			return len(subdomains[i]) < len(subdomains[j])
		}
		containsNumberI := containsNumber(subdomains[i])
		containsNumberJ := containsNumber(subdomains[j])
		if containsNumberI != containsNumberJ {
			return !containsNumberI
		}
		return subdomains[i] < subdomains[j]
	})

	// Print all domains
	fmt.Println("Total domains found:", len(subdomains))
	count := 1
	for _, subdomain := range subdomains {
		fmt.Println("|", count, "|", subdomain, "|")
		count++
	}
}

func containsNumber(s string) bool {
	for _, c := range s {
		if unicode.IsDigit(c) {
			return true
		}
	}
	return false
}
