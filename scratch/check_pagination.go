package main

import (
	"fmt"
	"net/http"
	"strings"
	"github.com/PuerkitoBio/goquery"
)

func main() {
	resp, _ := http.Get("https://dlskits.com/")
	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	
	doc.Find("a.next, a.next-page, a.page-numbers.next, .pagination a").Each(func(_ int, s *goquery.Selection) {
		text := strings.ToLower(s.Text())
		href, _ := s.Attr("href")
		fmt.Printf("Link: %s Text: %s\n", href, text)
	})
}
