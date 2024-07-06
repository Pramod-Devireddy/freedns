package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const baseURL = "https://freedns.afraid.org/domain/registry/page-"

func main() {
	// domains := []string{}

	for i := 1; i <= 270; i++ {
		url := fmt.Sprintf("%s%d.html", baseURL, i)

		pageDomains, err := fetchDomains(url)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("****** Page-%d ********\n", i)

		for _, domain := range pageDomains {
			fmt.Println(domain)
		}

		fmt.Println()

		time.Sleep(1 * time.Second)
		// domains = append(domains, pageDomains...)
	}

}

func fetchDomains(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request returned status %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTML parsing failed: %v", err)
	}

	var pageDomains []string

	// Recursive function to traverse the HTML nodes and extract domain names
	// based on the specific structure (assuming domain links contain the word "subdomain")

	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.Contains(attr.Val, "/subdomain/edit") {
					pageDomains = append(pageDomains, n.FirstChild.Data)
					break
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	return pageDomains, nil
}
