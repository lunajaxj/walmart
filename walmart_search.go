package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
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

type Mode struct {
	keyword  string
	stock    string
	id       string //div[@data-testid="list-view"]/preceding-sibling::a[@link-identifier]/@link-identifier
	title    string //div[@data-testid="list-view"]//div[@class="relative"]//img/@alt
	typez    string
	value    string
	brand    string //"name":"VicTsing"},"offers":{  品牌
	query    string //*[@id="maincontent"]//div[@data-testid="sticky-buy-box"]//div/p//span//text()  标签，多行相加
	score    string //(4.5)  评分
	review   string //"totalReviewCount":1187}  评论数量
	price    string //aria-hidden="false">$22.98<	价格
	category string
	////div/div/span[@class="lh-title"]//text()  卖家+配送
	seller       string   //卖家
	delivery     string   //配送
	deliveryDate string   //配送时间
	variant1     string   //变体1 :</span><span aria-hidden="true" class="ml1">(.*?)</span>
	variant2     string   //变体1 :</span><span aria-hidden="true" class="ml1">(*?)</span>
	otherId      []string //","usItemId":"[0-9]+?"
}

var keysm = make(map[string]string)

var res = make(map[string][]Mode)
var keywords []string
var lock sync.Mutex
var lockk sync.Mutex
var wg = sync.WaitGroup{}
var wgg = sync.WaitGroup{}
var wggg = sync.WaitGroup{}
var mux sync.Mutex
var ch = make(chan int, 5)
var chh = make(chan int, 4)
var ids []string

var count int

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
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("无法打开日志文件: %v", err)
	}
	defer logFile.Close()
	// 将日志输出到文件和控制台
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)                                    // 设置日志输出
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile) // 设置日志格式
	log.Println("自动化脚本-walmart-关键词获取信息")
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("关键词_search.txt")
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 0 {
			split := strings.Split(strings.TrimSpace(string(lineB)), ",")
			if len(split) != 2 {
				log.Println("错误: 关键词格式错误", string(lineB))
				continue
			}
			keysm[split[0]] = split[1]
			keywords = append(keywords, split[0])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("读取文件时出现错误", err)
		}
	}
	log.Println("keywords=", keywords)
	for _, v := range keywords {
		ch <- 1
		wg.Add(1)
		go crawler(v, 1)
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
		if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"关键词", "id", "商品码类型", "商品码值", "品牌", "标签", "标题", "评分", "评论数量", "价格", "卖家", "配送", "变体1", "变体2", "变体id", "到达时间", "库存", "类目"}); err != nil {
			log.Println(err)
		}
	}
	num := count + 2
	count += len(res[keys])
	for i := range res[keys] {
		other := ""
		for ii, v := range res[keys][i].otherId {
			if ii == 0 {
				other = v
				continue
			}
			other = other + "," + v
		}
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{res[keys][i].keyword, res[keys][i].id, res[keys][i].typez, res[keys][i].value, res[keys][i].brand, res[keys][i].query, res[keys][i].title, res[keys][i].score, res[keys][i].review, res[keys][i].price, res[keys][i].seller, res[keys][i].delivery, res[keys][i].variant1, res[keys][i].variant2, other, res[keys][i].deliveryDate, res[keys][i].stock, res[keys][i].category}); err != nil {
			log.Println(err)
		}
		num++
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

func crawler(keyword string, page int) {

	defer func() {

		<-ch
		wg.Done()

	}()
	var cou int
	atoi, _ := strconv.Atoi(keysm[keyword])
	for page <= atoi && cou < 16 {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		proxyUrl, _ := url.Parse("http://l752.kdltps.com:15818")

		tr.Proxy = http.ProxyURL(proxyUrl)
		basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("t19932187800946:wsad123456"))
		tr.ProxyConnectHeader = http.Header{}
		tr.ProxyConnectHeader.Add("Proxy-Authorization", basicAuth)

		client := &http.Client{Timeout: 15 * time.Second, Transport: tr}
		var k = strings.Replace(url.QueryEscape(keyword), "%20", "+", -1)
		urll := ""
		if page != 1 {
			urll = k + "&page=" + strconv.Itoa(page)
		} else {
			urll = k
		}
		urll = strings.Replace(urll, "%20", "+", -1)
		request, _ := http.NewRequest("GET", "https://www.walmart.com/search?q="+urll, nil)
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36")
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")
		request.Header.Set("Accept-Language", "zh")
		request.Header.Set("Sec-Ch-Ua", `"Not.A/Brand";v="8", "Chromium";v="114", "Google Chrome";v="114"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", `document`)
		request.Header.Set("Sec-Fetch-Mode", `navigate`)
		request.Header.Set("Sec-Fetch-Site", `none`)
		request.Header.Set("Sec-Fetch-User", `?1`)
		request.Header.Set("Upgrade-Insecure-Requests", `1`)
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")
		var isc2 = IsC2
		if IsC2 {
			request.Header.Set("Cookie", generateRandomString(10))
		}

		response, err := client.Do(request)
		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始：" + keyword)
			} else if strings.Contains(err.Error(), "441") {
				log.Println("代理超频！暂停10秒后继续...")
				time.Sleep(time.Second * 10)
			} else if strings.Contains(err.Error(), "440") {
				log.Println("代理宽带超频！暂停5秒后继续...")
				time.Sleep(time.Second * 5)
			} else {
				log.Println("错误信息：" + err.Error())
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同关键词连续出现请联系我，重新开始：" + keyword)
			}
			cou++
			continue
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
			time.Sleep(time.Second * 1)
			cou++
			continue
		}
		//doc, err := htmlquery.Parse(strings.NewReader(result))
		if err != nil {
			log.Println("错误信息：" + err.Error())
			return
		}

		//id
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
		if err != nil {
			log.Fatalf("Failed to parse HTML: %v", err)
		}
		var itemDetails ScriptContent
		doc.Find("script").Each(func(i int, s *goquery.Selection) {
			if id, exists := s.Attr("id"); exists && id == "__NEXT_DATA__" {
				scriptContent := s.Text()
				if err := json.Unmarshal([]byte(scriptContent), &itemDetails); err != nil {
					log.Fatalf("Failed to unmarshal JSON: %v", err)
				}
			}
		})

		cnt := 0
		if len(itemDetails.Props.PageProps.InitialData.SearchResult.ItemStacks) > 0 {
			for _, item := range itemDetails.Props.PageProps.InitialData.SearchResult.ItemStacks[0].Items {
				cnt++
				log.Printf("Processing item: %v", item) // 打印每个item的信息
				var linePrice string
				if priceInfoMap, ok := item.PriceInfo.(map[string]interface{}); ok {
					if lp, exists := priceInfoMap["linePrice"]; exists {
						linePrice = lp.(string)
					}
				}
				id := item.UsItemID
				price := linePrice
				title := item.Name

				lock.Lock()
				if res[keyword] == nil {
					res[keyword] = []Mode{}
				}
				if !IsContain(res[keyword], id) {
					res[keyword] = append(res[keyword], Mode{keyword: keyword, id: id, title: title, price: price})
					log.Printf("Added item ID: %s, Title: %s", id, title) // 打印添加的商品信息
				} else {
					log.Printf("Duplicate item ID: %s", id) // 打印重复商品信息
				}
				lock.Unlock()
			}
		} else {
			log.Printf("关键词: %s 第 %d 页 没有搜索结果", keyword, page)
			// 如果你需要在没有结果的情况下做其他处理，可以在这里添加逻辑
		}

		//最大分页
		var max int
		maxPage := regexp.MustCompile("\"maxPage\":([0-9]+?),").FindAllStringSubmatch(result, -1)
		log.Println("MAXPAGE=", maxPage)
		if len(maxPage) != 0 {
			max, _ = strconv.Atoi(maxPage[0][1])
		}
		log.Println("关键词:" + keyword + " 第" + strconv.Itoa(page) + "页 完成 " + strconv.Itoa(cnt) + "个")
		if page == max || page == atoi {
			log.Println("关键词:" + keyword + "到达页尾")
			for i1, ss := range res[keyword] {
				wggg.Add(1)
				chh <- 1
				go getOtherId(ss.id, keyword, i1)
			}
			//log.Println(res[keyword])
			wggg.Wait()
			save(keyword)
			return
		}
		page++
		cou++
	}
}

func getOtherId(vv, i1 string, i2 int) []string {
	defer func() {
		wggg.Done()
		<-chh
	}()

	for i := 0; i < 16; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}
		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		request, _ := http.NewRequest("PUT", "https://www.walmart.com/ip/"+vv, nil)

		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")
		request.Header.Set("Accept-Language", "zh")
		request.Header.Set("Sec-Ch-Ua", `"Not.A/Brand";v="8", "Chromium";v="114", "Google Chrome";v="114"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", `document`)
		request.Header.Set("Sec-Fetch-Mode", `navigate`)
		request.Header.Set("Sec-Fetch-Site", `none`)
		request.Header.Set("Sec-Fetch-User", `?1`)
		request.Header.Set("Upgrade-Insecure-Requests", `1`)
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")
		var isc = IsC
		if IsC {
			request.Header.Set("Cookie", generateRandomString(10))
		}
		response, err := client.Do(request)
		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("错误打印：", err.Error()+"重新开始："+vv)
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
				log.Println("超频或其他错误，重新开始：" + vv)
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
		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
		if len(fk) > 0 {
			log.Println("被风控，更换IP重新开始：" + vv)
			IsC = !isc
			continue

		}

		//upc与upc类型
		upc := regexp.MustCompile("upc\":\"(.{4,30}?)\"").FindAllStringSubmatch(result, -1)
		gtin := regexp.MustCompile("gtin13\":\"(.{4,30}?)\"").FindAllStringSubmatch(result, -1)
		if len(upc) > 0 {
			res[i1][i2].value = upc[0][1]
			res[i1][i2].typez = "upc"
		} else if len(gtin) > 0 {
			res[i1][i2].value = gtin[0][1]
			res[i1][i2].typez = "gtin"
		} else {
			res[i1][i2].value = ""
			res[i1][i2].typez = "ean"
			//log.Println("id:"+id+" 获取为空，默认为ean")
		}

		doc, err := htmlquery.Parse(strings.NewReader(result))
		if err != nil {
			log.Println("错误信息：" + err.Error())
			return nil
		}

		//品牌
		brand := regexp.MustCompile("\"brand\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		if len(brand) == 0 {
			//log.Println("品牌获取失败id："+id)
		} else {
			res[i1][i2].brand = brand[0][1]
		}

		//标签

		match := regexp.MustCompile(`"badges":\{"flags":(\[[\s\S]*?\])`).FindStringSubmatch(result)
		// Check if a match was found
		if len(match) < 2 {
			fmt.Println("No match found")
		} else {
			// Extract and parse the JSON
			flagsJSON := match[1]
			var flags []struct {
				Text string `json:"text"`
			}
			// Parse the JSON data
			if err := json.Unmarshal([]byte(flagsJSON), &flags); err != nil {
				fmt.Println("Error parsing JSON:", err)
			}

			// Extract and print the 'text' attribute from each flag
			queryStr := ""
			for _, flag := range flags {
				queryStr += flag.Text
			}
			fmt.Println("label:", queryStr)
			res[i1][i2].query = queryStr
		}
		//库存
		stock := regexp.MustCompile(`("message":"Currently out of stock")`).FindAllStringSubmatch(result, -1)
		if len(stock) == 0 {
			res[i1][i2].stock = "有库存"
		} else {
			res[i1][i2].stock = "无库存"
		}

		//标题
		title := regexp.MustCompile("\"productName\":\"(.+?)\",").FindAllStringSubmatch(result, -1)
		if len(title) == 0 {
		} else {
			res[i1][i2].title = strings.Replace(title[0][1], "\\u0026", "&", -1)
		}

		//评分
		score := regexp.MustCompile("[(]([\\d][.][\\d])[)]").FindAllStringSubmatch(result, -1)
		if len(score) == 0 {
			//log.Println("评分获取失败id："+id)
		} else {
			res[i1][i2].score = score[0][1]
		}

		//评论数量

		review := regexp.MustCompile("\"totalReviewCount\":(\\d+)").FindAllStringSubmatch(result, -1)
		if len(review) == 0 {
			//log.Println("评论数量获取失败id："+id)
		} else {
			res[i1][i2].review = review[0][1]
		}

		//价格

		price := regexp.MustCompile(`"best[^{]+?,"priceDisplay":"([^"]+)"`)
		price1 := price.FindAllString(result, -1)
		if len(price1) > 0 {
			//log.Println(result)
			// Check if the matched string contains "priceDisplay"
			if strings.Contains(price1[0], `"priceDisplay":"`) {
				// Split the string to isolate the part after "priceDisplay":"
				parts := strings.Split(price1[0], `"priceDisplay":"`)
				if len(parts) > 1 {
					// Further split to get just the value before the closing quote
					valueParts := strings.Split(parts[1], `"`)
					if len(valueParts) > 0 {
						//fmt.Println("Extracted Value:", valueParts[0]) // Should print "Now $16.99"
						reg := regexp.MustCompile(`[^\d.]`)
						numericValue := reg.ReplaceAllString(valueParts[0], "")
						fmt.Println("Numeric Value:", numericValue) // Should print "16.99"
						res[i1][i2].price = numericValue
					} else {
						fmt.Println("No value extracted after priceDisplay")
					}
				} else {
					fmt.Println("No priceDisplay part found in string")
				}
			} else {
				fmt.Println("String does not contain priceDisplay")
			}
		} else {
			fmt.Println("No matches found or result is empty")
			res[i1][i2].price = "" // 如果是空的，赋值空字符串
		}

		//类目
		category := regexp.MustCompile(`categoryName":"(.+?)",`).FindAllStringSubmatch(result, -1)
		if len(category) == 0 {
			//log.Println("价格获取失败id："+id)
		} else {
			res[i1][i2].category = strings.Replace(category[0][1], `\u0026`, "&", -1)
		}

		//卖家与配送
		all, err := htmlquery.QueryAll(doc, "//div/div/span[@class=\"lh-title\"]//text()")
		//log.Println(result)
		if err != nil {
			//log.Println("卖家与配送获取失败")
		} else {
			for i, v := range all {
				sv := htmlquery.InnerText(v)
				if strings.Contains(sv, "Sold by") {
					res[i1][i2].seller = htmlquery.InnerText(all[i+1])
					continue
				}
				if strings.Contains(sv, "Fulfilled by") {
					res[i1][i2].delivery = strings.Replace(sv, "Fulfilled by ", "", -1)
					if len(res[i1][i2].delivery) < 3 && len(all) > i+1 {
						res[i1][i2].delivery = htmlquery.InnerText(all[i+1])
					}
					continue
				}
				if strings.Contains(sv, "Sold and shipped by") {
					res[i1][i2].seller = htmlquery.InnerText(all[i+1])
					res[i1][i2].delivery = res[i1][i2].seller
					break
				}
			}
		}
		if res[i1][i2].seller == "" {
			seller := regexp.MustCompile("\"sellerDisplayName\":\"(.*?)\"").FindAllStringSubmatch(result, -1)
			if len(seller) > 0 {
				res[i1][i2].seller = seller[0][1]
			}
		}
		//配送时间
		deliveryDate := regexp.MustCompile("\"fulfillmentText\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		if len(deliveryDate) == 0 {
			//log.Println("配送时间获取失败id："+id)
		} else {
			res[i1][i2].deliveryDate = deliveryDate[0][1]
		}
		//变体
		variant := regexp.MustCompile(":</span><span class=\"ml1\">(.*?)</span>").FindAllStringSubmatch(result, -1)
		if len(variant) == 0 {
			//log.Println("变体获取失败id："+id)
		} else if len(variant) == 1 {
			res[i1][i2].variant1 = variant[0][1]
		} else if len(variant) == 2 {
			res[i1][i2].variant1 = variant[0][1]
			res[i1][i2].variant2 = variant[1][1]
		}

		allString := regexp.MustCompile("\",\"usItemId\":\"([0-9]+?)\"").FindAllStringSubmatch(result, -1)
		otherIds := []string{}
		for _, v := range allString {
			//if v[1]!= vv{
			otherIds = append(otherIds, v[1])
			//}
		}
		if len(allString) == 0 {
			return nil
		} else {
			res[i1][i2].otherId = otherIds
			log.Println("id完成：" + vv)
			return otherIds
		}

	}
	return nil

}
func IsContain(eachItem []Mode, item string) bool {
	lockk.Lock()
	defer lockk.Unlock()
	it := false
	for _, eachItem := range eachItem {
		if eachItem.id == item {
			it = true
		}
	}
	return it
}
