package main

import (
	"net/url"
	"strings"
)

func normalizeURL(inputUrl string) (string, error) {
	u, err := url.Parse(inputUrl)
	if err != nil {
		return inputUrl, err
	}
	return u.Hostname() + strings.TrimSuffix(u.Path, "/"), nil
}
