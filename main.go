package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"sync"
)

type config struct {
	pages              map[string]int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           int
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	cfg.concurrencyControl <- struct{}{}
	defer cfg.wg.Done()
	allURLs := func() []string {
		defer func() {
			<-cfg.concurrencyControl
		}()
		cfg.mu.Lock()
		if len(cfg.pages) >= cfg.maxPages {
			cfg.mu.Unlock()
			return []string{}
		}
		cfg.mu.Unlock()
		baseURL, err := url.Parse(cfg.baseURL.String())
		if err != nil {
			fmt.Println(err)
			return []string{}
		}

		currURL, err := url.Parse(rawCurrentURL)
		if err != nil {
			fmt.Println(err)
			return []string{}
		}

		if baseURL.Hostname() != currURL.Hostname() {
			return []string{}
		}

		cfg.mu.Lock()
		if ok := cfg.addPageVisit(rawCurrentURL); !ok {
			return []string{}
		}

		currHTML, err := getHTML(rawCurrentURL)
		if err != nil {
			fmt.Println(err)
			return []string{}
		}
		fmt.Println(currHTML)

		allURLs, err := getURLsFromHTML(currHTML, rawCurrentURL)
		if err != nil {
			fmt.Println(err)
			return []string{}
		}
		return allURLs
	}()

	for _, u := range allURLs {
		cfg.wg.Add(1)
		go cfg.crawlPage(u)
	}
}

func (cfg *config) addPageVisit(rawCurrentURL string) (isFirst bool) {
	defer cfg.mu.Unlock()
	normURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if _, ok := cfg.pages[normURL]; ok {
		cfg.pages[normURL]++
		return false
	}
	cfg.pages[normURL] = 1
	return true
}

func getHTML(rawURL string) (string, error) {
	res, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return "", errors.New(string(res.StatusCode))
	}
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func printReport(pages map[string]int, baseURL string) {
	fmt.Printf("=============================\nREPORT for %s\n=============================\n", baseURL)
	keys := make([]string, 0, len(pages))
	for key := range pages {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return pages[keys[i]] > pages[keys[j]] })
	for _, key := range keys {
		fmt.Printf("Found %v internal links to %v\n", pages[key], key)
	}
}

func main() {
	maxConcurrency := 10
	maxPages := 10
	args := os.Args[1:]
	switch args_len := len(args); args_len {
	case 0:
		fmt.Println("no website provided")
		os.Exit(1)
	case 2:
		maxConcurrency, _ = strconv.Atoi(args[1])
	case 3:
		maxConcurrency, _ = strconv.Atoi(args[1])
		maxPages, _ = strconv.Atoi(args[2])
	default:
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}

	BASE_URL, err := url.Parse(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("starting crawl of: %v\n", BASE_URL)

	cfg := config{
		pages:              make(map[string]int),
		baseURL:            BASE_URL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
		maxPages:           maxPages,
	}
	cfg.wg.Add(1)
	go cfg.crawlPage(cfg.baseURL.String())
	cfg.wg.Wait()
	close(cfg.concurrencyControl)

	for k, v := range cfg.pages {
		fmt.Println(k, v)
	}
	printReport(cfg.pages, args[0])
}
