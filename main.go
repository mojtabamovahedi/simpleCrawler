package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"golang.org/x/net/html"
)

var baseURL string = "https://gobyexample.com"
var setLinks map[uint32]string

func main() {
	setLinks = make(map[uint32]string)
	var wg sync.WaitGroup

	server := new(http.Server)
	server.ReadTimeout = 5 * time.Second
	server.WriteTimeout = 5 * time.Second

	callHttpRequest(baseURL, &wg)

	wg.Wait()
}

func isURL(str string) bool {
	_, err := url.ParseRequestURI(str)
	return err == nil
}

func isBaseUrl(str string) bool {
	url, err := url.Parse(str)
	if err != nil {
		return false
	}
	return url.Host == strings.Split(baseURL, "//")[1]
}

func isDuplicate(str string) bool {
	_, exist := setLinks[hash(str)]
	return exist
}

func callHttpRequest(url string, wg *sync.WaitGroup) {
	wg.Add(1)

	resp, err := http.Get(url)

	if err != nil {
		log.Default().Println("url is", url, "and error is", err)
		wg.Done()
		return
	}
	fmt.Println("# '" + url + "'")
	setLinks[hash(url)] = url
	time.Sleep(time.Millisecond * 2000)

	defer resp.Body.Close()

	links := getLinks(resp.Body)

	for _, url := range links {
		if isBaseUrl(url) && !isDuplicate(url) {
			go callHttpRequest(url, wg)
		}
	}

	wg.Done()
}

func getLinks(body io.Reader) (links []string) {
	z := html.NewTokenizer(body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.EndTagToken:
			token := z.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						if isURL(attr.Val) {
							links = append(links, attr.Val)
						} else {
							links = append(links, baseURL+"/"+attr.Val)
						}
					}

				}
			}

		}
	}
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
