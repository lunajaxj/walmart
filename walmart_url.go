package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/xuri/excelize/v2"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Wal struct {
	url    string
	id     string
	typez  string
	value  string
	brand  string //"name":"VicTsing"},"offers":{  品牌
	query  string //*[@id="maincontent"]//div[@data-testid="sticky-buy-box"]//div/p//span//text()  标签，多行相加
	title  string //"name":"Gymax 5 Piece Dining Set Glass Top Table & 4 Upholstered Chairs Kitchen Room Furniture","sku":  标题
	score  string //(4.5)  评分
	review string //"totalReviewCount":1187}  评论数量
	price  string //aria-hidden="false">$22.98<	价格
	////div/div/span[@class="lh-title"]//text()  卖家+配送
	seller       string   //卖家
	delivery     string   //配送
	deliveryDate string   //配送时间
	variant1     string   //变体1 :</span><span aria-hidden="true" class="ml1">(.*?)</span>
	variant2     string   //变体1 :</span><span aria-hidden="true" class="ml1">(*?)</span>
	otherIds     []string //变体id
}

var keysm = make(map[string]string)

var res = make(map[string][]Wal)
var keywords []string
var lock sync.Mutex
var lockk sync.Mutex
var wg = sync.WaitGroup{}
var wgg = sync.WaitGroup{}
var wggg = sync.WaitGroup{}
var mux sync.Mutex
var ch = make(chan int, 4)
var chh = make(chan int, 5)
var ids []string

var count int

func main() {
	log.Println("自动化脚本-walmart-链接获取信息")
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("链接.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 5 {
			keywords = append(keywords, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}

	}
	log.Println("有", len(keywords), "个链接任务")
	for _, v := range keywords {
		ch <- 1
		wg.Add(1)
		go crawler(v, 1, 1)
	}
	wg.Wait()
	log.Println("完成")

}

func save(keys string) {
	mux.Lock()
	defer mux.Unlock()
	fileName := "out.xlsx"
	xlsx, err := excelize.OpenFile("out.xlsx")
	if err != nil {
		xlsx = excelize.NewFile()
	}
	defer xlsx.Close()
	if count == 0 {
		if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"链接", "id", "商品码类型", "商品码值", "品牌", "标签", "标题", "评分", "评论数量", "价格", "卖家", "配送", "变体1", "变体2", "变体id", "到达时间"}); err != nil {
			log.Println(err)
		}
	}
	count += 2
	for _, v := range res[keys] {
		other := ""
		for i := range v.otherIds {
			if i == 0 {
				other = v.otherIds[i]
				continue
			}
			other = other + "," + v.otherIds[i]
		}
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(count), &[]interface{}{v.url, v.id, v.typez, v.value, v.brand, v.query, v.title, v.score, v.review, v.price, v.seller, v.delivery, v.variant1, v.variant2, other, v.deliveryDate}); err != nil {
			log.Println(err)
		}
		count++
	}

	xlsx.Save()
	if err != nil {
		xlsx.SaveAs(fileName)
	}
	res[keys] = nil
}

func exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var IsC = false
var IsC2 = true

func init() {
	rand.Seed(time.Now().UnixNano()) // 初始化随机数生成器
}

func generateRandomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func extractMaxPages(result string) ([]int, error) {
	maxPageRegex := regexp.MustCompile("\"maxPage\":([0-9]+?),")
	maxPageMatches := maxPageRegex.FindAllStringSubmatch(result, -1)

	if len(maxPageMatches) == 0 {
		return nil, fmt.Errorf("max pages not found")
	}

	var maxPages []int
	for _, match := range maxPageMatches {
		max, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, err
		}
		maxPages = append(maxPages, max)
	}
	return maxPages, nil
}

func crawler(keyword string, xc int, page int) {

	defer func() {
		if xc == 1 {
			<-ch
			wg.Done()
		}
	}()
	for page <= 25 {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		proxyUrl, _ := url.Parse("http://l752.kdltps.com:15818")
		tr.Proxy = http.ProxyURL(proxyUrl)
		basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("t19932187800946:wsad123456"))
		tr.ProxyConnectHeader = http.Header{}
		tr.ProxyConnectHeader.Add("Proxy-Authorization", basicAuth)

		client := &http.Client{Timeout: 15 * time.Second, Transport: tr}
		urll := keyword
		if page != 1 {
			if strings.Contains(urll, "?") {
				urll = keyword + "&page=" + strconv.Itoa(page)
			} else {
				urll = keyword + "?page=" + strconv.Itoa(page)
			}
		}
		request, _ := http.NewRequest("GET", urll, nil)
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		request.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
		request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		request.Header.Set("Cache-Control", "max-age=0")
		request.Header.Set("Sec-Ch-Ua", `"Not)A;Brand";v="99", "Google Chrome";v="127", "Chromium";v="127"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", `document`)
		request.Header.Set("Sec-Fetch-Mode", `navigate`)
		request.Header.Set("Sec-Fetch-Site", `same-origin`)
		request.Header.Set("Sec-Fetch-User", `?1`)
		request.Header.Set("Upgrade-Insecure-Requests", `1`)

		var isc2 = IsC2
		if IsC2 {
			request.Header.Set("Cookie", generateRandomString(10))
		}
		response, err := client.Do(request)
		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始：" + keyword)
				continue
			} else if strings.Contains(err.Error(), "441") {
				log.Println("代理超频！暂停10秒后继续...")
				time.Sleep(time.Second * 10)
				continue
			} else if strings.Contains(err.Error(), "440") {
				log.Println("代理宽带超频！暂停5秒后继续...")
				time.Sleep(time.Second * 5)
				continue
			} else {
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同链接连续出现请联系我，重新开始：" + keyword)
				continue
			}
		}
		result := ""
		if response.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(response.Body) // gzip解压缩
			if err != nil {
				log.Println("解析body错误，重新开始")
				continue
			}
			defer reader.Close()
			con, err := io.ReadAll(reader)
			if err != nil {
				log.Println("gzip解压错误，重新开始")
				continue
			}
			result = string(con)
		} else {
			dataBytes, err := io.ReadAll(response.Body)
			if err != nil {
				if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "Service Unavailable") {
					log.Println("代理IP无效，自动切换中")
					log.Println("连续出现代理IP无效请联系我，重新开始")
				} else {
					log.Println("错误信息：" + err.Error())
					log.Println("出现错误，如果同id连续出现请联系我，重新开始")
				}
				continue
			}
			defer response.Body.Close()
			result = string(dataBytes)
		}
		//log.Println(result)

		cw1 := regexp.MustCompile("(is not valid JSON)").FindAllStringSubmatch(result, -1)
		cw2 := regexp.MustCompile("(The requested URL was rejected. Please consult with your administrator)").FindAllStringSubmatch(result, -1)
		if len(cw1) > 0 || len(cw2) > 0 {
			log.Println("搜索内容错误，跳过该标题：" + keyword)
			return
		}
		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
		if len(fk) > 0 {
			log.Println("被风控，更换IP重新开始")
			IsC2 = !isc2
			continue
		}
		//doc, err := htmlquery.Parse(strings.NewReader(result))
		if err != nil {
			log.Println("错误信息：" + err.Error())
			return
		}
		//log.Println(keyword+page)
		//log.Println(result)
		resultStr := ""
		resultS := regexp.MustCompile("(\"layoutEnum\":\"GRID\"}].+?\"pageMetadata\")").FindAllStringSubmatch(result, -1)
		allString := regexp.MustCompile("(There were no search results for)").FindAllString(result, -1)
		if len(allString) > 0 {
			log.Println("链接:" + keyword + " 第" + strconv.Itoa(page) + "页 无搜索结果")
			return
		}
		if len(resultS) == 0 {
			log.Println("被风控，更换IP重新开始")
			continue
		}

		//maxPages定义
		maxPages, err := extractMaxPages(result)
		if err != nil {
			log.Println("Error extracting max pages:", err)
			return
		}
		// 根据当前页和最大页数选择结果字符串
		if len(maxPages) > 1 && page >= maxPages[0] {
			resultStr = resultS[1][1] // 使用第二个结果集
		} else {
			resultStr = resultS[0][1] // 默认使用第一个结果集
		}
		//id
		id := regexp.MustCompile("usItemId\":\"([0-9]+?)\"").FindAllStringSubmatch(resultStr, -1)
		if err != nil {
			//log.Println("无标签")
		} else {
			if res[keyword] == nil {
				res[keyword] = []Wal{}
			}
			for i, _ := range id {
				res[keyword] = append(res[keyword], Wal{url: keyword, id: id[i][1]})

			}
			wgg.Wait()
		}

		//最大分页
		for _, max := range maxPages {
			if page == max {
				// ...处理逻辑...
				log.Println("链接:" + keyword + " 第" + strconv.Itoa(page) + "页 完成 " + strconv.Itoa(len(id)) + "个")
				log.Println("链接:", keyword, "到达页尾")
				for i1, ss := range res[keyword] {
					wggg.Add(1)
					chh <- 1
					go crawlerId(ss.id, keyword, i1)
				}
				wggg.Wait()
				save(keyword)
				break
			}
		}
		log.Println("链接:", keyword, " 第", page, "页 完成", len(id), "个")
		page++
	}
}

func crawlerId(id, key string, i1 int) {

	//配置代理
	defer func() {
		wggg.Done()
		<-chh
	}()
	for i := 0; i < 16; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}

		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l753.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		request, _ := http.NewRequest("PUT", "https://www.walmart.com/ip/"+id, nil)

		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		request.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
		request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		request.Header.Set("Cache-Control", "max-age=0")
		request.Header.Set("Sec-Ch-Ua", `"Not)A;Brand";v="99", "Google Chrome";v="127", "Chromium";v="127"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", `document`)
		request.Header.Set("Sec-Fetch-Mode", `navigate`)
		request.Header.Set("Sec-Fetch-Site", `same-origin`)
		request.Header.Set("Sec-Fetch-User", `?1`)
		request.Header.Set("Upgrade-Insecure-Requests", `1`)
		var isc = IsC
		if IsC {
			request.Header.Set("Cookie", generateRandomString(10))
		}
		response, err := client.Do(request)

		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始：" + id)
			} else if strings.Contains(err.Error(), "441") {
				log.Println("代理超频！暂停10秒后继续...")
				time.Sleep(time.Second * 10)
			} else if strings.Contains(err.Error(), "440") {
				log.Println("代理宽带超频！暂停5秒后继续...")
				time.Sleep(time.Second * 5)
			} else {
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + id)
			}

			continue
		}
		result := ""
		if response.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(response.Body) // gzip解压缩
			if err != nil {
				log.Println("解析body错误，重新开始：" + id)
				continue
			}
			defer reader.Close()
			con, err := io.ReadAll(reader)
			if err != nil {
				log.Println("gzip解压错误，重新开始：" + id)
				continue
			}
			result = string(con)
		} else {
			dataBytes, err := io.ReadAll(response.Body)
			if err != nil {
				if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "Service Unavailable") {
					log.Println("代理IP无效，自动切换中")
					log.Println("连续出现代理IP无效请联系我，重新开始：" + id)
				} else {
					log.Println("错误信息：" + err.Error())
					log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + id)
				}
				continue
			}
			defer response.Body.Close()
			result = string(dataBytes)
		}
		if strings.Contains(result, "This page could not be found.") {
			res[key][i1].typez = "该商品不存在"
			log.Println("id:" + id + "商品不存在")
			return
		}

		//upc与upc类型
		upc := regexp.MustCompile("upc\":\"(.{4,30}?)\"").FindAllStringSubmatch(result, -1)
		gtin := regexp.MustCompile("gtin13\":\"(.{4,30}?)\"").FindAllStringSubmatch(result, -1)
		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
		if len(upc) > 0 {
			res[key][i1].value = upc[0][1]
			res[key][i1].typez = "upc"
		} else if len(gtin) > 0 {
			res[key][i1].value = gtin[0][1]
			res[key][i1].typez = "gtin"
		} else if len(fk) > 0 {
			log.Println("id:" + id + " 被风控,更换IP继续")
			IsC = !isc
			continue
		} else {
			res[key][i1].value = ""
			res[key][i1].typez = "ean"
			//log.Println("id:" + id + " 获取为空，默认为ean")
		}

		doc, err := htmlquery.Parse(strings.NewReader(result))
		if err != nil {
			log.Println("错误信息：" + err.Error())

			continue
		}
		doc1, err := goquery.NewDocumentFromReader(strings.NewReader(result))
		if err != nil {
			log.Println("解析HTML错误：", err)
			return
		}

		//品牌
		brand := regexp.MustCompile("\"brand\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		if len(brand) == 0 {
			log.Println("品牌获取失败id：" + id)
		} else {
			res[key][i1].brand = brand[0][1]
		}

		//标签
		queryStr := ""
		doc1.Find("#maincontent > section > main > div.flex.undefined.flex-column.h-100 > div:nth-child(2) > div > div.w_aoqv.w_wRee.w_fdPt > div > div:nth-child(2) > div > div > div:nth-child(1) > div.flex.items-center.mv2.flex-wrap > div > div > span").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			if !strings.Contains(queryStr, text) {
				queryStr += text + " "
			}
		})
		res[key][i1].query = queryStr

		//query, err := htmlquery.QueryAll(doc, "//*[@id=\"maincontent\"]//div[@data-testid=\"sticky-buy-box\"]//div/p//span//text()")
		//query, err := htmlquery.QueryAll(doc, "//div[@class=\"h2 relative mr2\"]/span/text()")
		//if err != nil {
		//	log.Println("无标签")
		//} else {
		//	queryStr := ""
		//	for _, v := range query {
		//		text := htmlquery.InnerText(v)
		//		if !strings.Contains(queryStr, text) {
		//			queryStr += text + " "
		//
		//		}
		//	}
		//	res[key][i1].query = queryStr
		//}
		//if res[key][i1].query == "" {
		//	queryf := regexp.MustCompile("2\" aria-hidden=\"false\">(.*?)</span>").FindAllStringSubmatch(result, -1)
		//	queryStr := ""
		//	for _, v := range queryf {
		//		queryStr += v[1] + " "
		//
		//	}
		//	res[key][i1].query = queryStr
		//}
		//标题
		title := regexp.MustCompile("\"productName\":\"(.+?)\",").FindAllStringSubmatch(result, -1)
		if len(title) == 0 {
			log.Println("获取失败id："+id, "重新请求")
			i += 2
			continue
		} else {
			res[key][i1].title = strings.Replace(title[0][1], "\\u0026", "&", -1)
		}

		//评分
		score := regexp.MustCompile("[(]([\\d][.][\\d])[)]").FindAllStringSubmatch(result, -1)
		if len(score) == 0 {
			log.Println("评分获取失败id：" + id)
		} else {
			res[key][i1].score = score[0][1]
		}

		//评论数量
		review := regexp.MustCompile("\"totalReviewCount\":(\\d+)}").FindAllStringSubmatch(result, -1)
		if len(review) == 0 {
			log.Println("评论数量获取失败id：" + id)
		} else {
			res[key][i1].review = review[0][1]
		}

		//价格
		price := regexp.MustCompile("<span itemprop=\"price\".*?.{0,20}(\\$[.,\\d]+).{0,20}?</span>").FindAllStringSubmatch(result, -1)
		if len(price) == 0 {
			log.Println("价格获取失败id：" + id)
		} else {
			res[key][i1].price = price[0][1]
		}

		//卖家与配送
		//fulfilled := regexp.MustCompile(">Fulfilled by (.*?)</div>|>Fulfilled by .*?>(.*?)</a>?").FindAllStringSubmatch(result, -1)
		//sold := regexp.MustCompile(">Sold by ([^<]*?)</div>|>Sold by .*?>(.*?)</a>?").FindAllStringSubmatch(result, -1)
		//shipped := regexp.MustCompile("<div>Sold and shipped by ([^/]*?)</div>|<div>Sold and shipped by.*?>(.*?)</a>?").FindAllStringSubmatch(result, -1)
		//if len(fulfilled) != 0 && len(sold) != 0 {
		//	res[key][i1].seller = sold[0][1]
		//	res[key][i1].delivery = fulfilled[0][1]
		//} else if len(shipped) != 0 {
		//	res[key][i1].seller = shipped[0][1]
		//	res[key][i1].delivery = shipped[0][1]
		//} else {
		//卖家与配送
		all, err := htmlquery.QueryAll(doc, "//div/div/span[@class=\"lh-title\"]//text()")
		//log.Println(result)
		if err != nil {
			log.Println("卖家与配送获取失败")
		} else {
			for i, v := range all {
				sv := htmlquery.InnerText(v)
				if strings.Contains(sv, "Sold by") {
					res[key][i1].seller = htmlquery.InnerText(all[i+1])
					continue
				}
				if strings.Contains(sv, "Fulfilled by") {
					res[key][i1].delivery = strings.Replace(sv, "Fulfilled by ", "", -1)
					if len(res[key][i1].delivery) < 3 && len(all) > i+1 {
						res[key][i1].delivery = htmlquery.InnerText(all[i+1])
					}
					continue
				}
				if strings.Contains(sv, "Sold and shipped by") {
					res[key][i1].seller = htmlquery.InnerText(all[i+1])
					res[key][i1].delivery = res[key][i1].seller
					break
				}
			}
		}
		if res[key][i1].seller == "" {
			seller := regexp.MustCompile("\"sellerDisplayName\":\"(.*?)\"").FindAllStringSubmatch(result, -1)
			if len(seller) > 0 {
				res[key][i1].seller = seller[0][1]
			}
		}
		//}

		//配送时间
		deliveryDate := regexp.MustCompile("\"fulfillmentText\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		if len(deliveryDate) == 0 {
			log.Println("配送时间获取失败id：" + id)
		} else {
			res[key][i1].deliveryDate = deliveryDate[0][1]
		}

		//变体
		variant := regexp.MustCompile(":</span><span class=\"ml1\">(.*?)</span>").FindAllStringSubmatch(result, -1)
		if len(variant) == 0 {
			log.Println("变体获取失败id：" + id)
		} else if len(variant) == 1 {
			res[key][i1].variant1 = variant[0][1]
		} else if len(variant) == 2 {
			res[key][i1].variant1 = variant[0][1]
			res[key][i1].variant2 = variant[1][1]
		}

		allString := regexp.MustCompile("\",\"usItemId\":\"([0-9]+?)\"").FindAllStringSubmatch(result, -1)
		for i := range allString {
			//if v[1]!= vv{
			res[key][i1].otherIds = append(res[key][i1].otherIds, allString[i][1])
			//}
		}

		log.Println("id:" + res[key][i1].id + "完成")
		return
	}
}
