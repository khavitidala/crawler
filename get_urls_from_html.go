package main

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	res := []string{}
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return []string{}, nil
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					u, err := url.Parse(a.Val)
					if err != nil {
						continue
					}
					if u.Hostname() == "" {
						raw, err := url.Parse(rawBaseURL)
						if err != nil {
							continue
						}
						uri := &url.URL{Host: raw.Hostname(), Path: a.Val, Scheme: raw.Scheme}
						res = append(res, uri.String())
						continue
					}
					res = append(res, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return res, nil
}
