package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Item struct {
	ID       string `json:"id"`
	UsItemID string `json:"usItemId"`
	Name     string `json:"name"`
	//Type      string      `json:"type"`
	PriceInfo interface{} `json:"priceInfo"` // 使用interface{}来处理未知的嵌套结构
}

type ItemStacks struct {
	Items []Item `json:"items"`
}

type SearchResult struct {
	ItemStacks []ItemStacks `json:"itemStacks"`
}

type InitialData struct {
	SearchResult SearchResult `json:"searchResult"`
}

type PageProps struct {
	InitialData InitialData `json:"initialData"`
}

type Props struct {
	PageProps PageProps `json:"pageProps"`
}

type ScriptContent struct {
	Props Props `json:"props"`
}

func main() {
	targetURL := "https://www.walmart.com/search?q=Electric+Callus+Remover+for+Feet"
	proxyURL := "http://127.0.0.1:51599"

	// Parse the proxy URL
	proxyURLParsed, err := url.Parse(proxyURL)
	if err != nil {
		log.Fatalf("Failed to parse proxy URL: %v", err)
	}

	// Setting up the proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURLParsed),
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Set up headers
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	// Find and parse the script tag
	var itemDetails ScriptContent
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if id, exists := s.Attr("id"); exists && id == "__NEXT_DATA__" {
			scriptContent := s.Text()
			if err := json.Unmarshal([]byte(scriptContent), &itemDetails); err != nil {
				log.Fatalf("Failed to unmarshal JSON: %v", err)
			}
		}
	})

	// Print the entire parsed JSON for inspection
	//fmt.Printf("%+v\n", itemDetails)

	// Extract and print item details
	cnt := 0
	for _, item := range itemDetails.Props.PageProps.InitialData.SearchResult.ItemStacks[0].Items {
		cnt++
		var linePrice string
		if priceInfoMap, ok := item.PriceInfo.(map[string]interface{}); ok {
			if lp, exists := priceInfoMap["linePrice"]; exists {
				linePrice = lp.(string)
			}
		}
		fmt.Printf("item=%s |us_itemID=%s |item_name=%s |item_price=%s\n", item.ID, item.UsItemID, item.Name, linePrice)

	}

	fmt.Println(cnt)
}
