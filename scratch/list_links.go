package main

import (
	"fmt"
	"net/http"
	"github.com/PuerkitoBio/goquery"
)

func main() {
	resp, _ := http.Get("https://dlskits.com/")
	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		fmt.Printf("Link: %s Text: %s\n", href, s.Text())
	})
}
