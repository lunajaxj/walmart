package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
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

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/xuri/excelize/v2"
)

var mu sync.Mutex
var file *excelize.File
var num int

func getRandomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
		//"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0.2 Safari/605.1.15",
		//"Mozilla/5.0 (Linux; Android 10; Pixel 3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.127 Mobile Safari/537.36",
		// 可以添加更多的 User-Agent 字符串
	}

	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}

func getRandomDownlink() string {
	return fmt.Sprintf("%.1f", rand.Float64()*2+1) // 1-3 MB
}

func getRandomDpr() string {
	return fmt.Sprintf("%.1f", rand.Float64()*1+1) // 1-2
}

var res []Wal

type Wal struct {
	id                string
	typez             string
	stock             string //库存
	value             string //商品码值
	brand             string //"name":"VicTsing"},"offers":{  品牌
	query             string //*[@id="maincontent"]//div[@data-testid="sticky-buy-box"]//div/p//span//text()  标签，多行相加
	title             string //"name":"Gymax 5 Piece Dining Set Glass Top Table & 4 Upholstered Chairs Kitchen Room Furniture","sku":  标题
	score             string //(4.5)  评分
	review            string //"totalReviewCount":1187}  评论数量
	price             string //aria-hidden="false">$22.98<	价格
	category          string
	seller            string   //卖家
	delivery          string   //配送
	deliveryDate      string   //配送时间
	variant1          string   //变体1 :</span><span aria-hidden="true" class="ml1">(.*?)</span>
	variant2          string   //变体1 :</span><span aria-hidden="true" class="ml1">(*?)</span>
	otherIds          []string //变体id
	startingFrom      string   //>Starting from \$([^<]+)<
	moreSellerOptions string   //More seller options \((\d+)\)
	availableQuantity string
	crossedPrice      string //划线价
	freeFreight       string //自发货运费
}

var idStores []string
var ids []string
var wg = sync.WaitGroup{}
var ch = make(chan int, 5) //频率修改

func main() {
	// 创建日志文件
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// 日志同时输出到文件和控制台
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	log.Println("WM升级版-ID&店铺ID采集信息-OUT不覆盖")
	log.Println("开始执行...")

	// 创建句柄
	fi, err := os.Open("id_storeid.txt")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer fi.Close()

	r := bufio.NewReader(fi)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Read error: %v", err)
			}
			break
		}
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			parts := strings.Split(line, "|")
			if len(parts) == 2 {
				id := strings.TrimSpace(parts[0])
				idStore := strings.TrimSpace(parts[1])
				ids = append(ids, id)
				idStores = append(idStores, idStore)
				log.Printf("Read line: id=%s, idStore=%s", id, idStore)
			} else {
				log.Printf("Invalid line format: %s", line)
			}
		} else {
			log.Println("Empty line encountered.")
		}
	}

	if len(ids) == 0 {
		log.Println("No IDs found in the file.")
		return
	}

	log.Println("IDs and ID Stores:", ids, idStores)

	for i := range ids {
		ch <- 1
		wg.Add(1)
		go crawler(ids[i], idStores[i])
	}

	wg.Wait()
	xlsx := excelize.NewFile()
	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "商品码类型", "商品码值", "品牌", "标签", "标题", "评分", "评论数量", "价格", "卖家", "配送", "变体1", "变体2", "变体id", "到达时间", "库存", "类目", "跟卖数量", "跟卖最低价格", "库存数量", "划线价", "自发货运费"}); err != nil {
		log.Println(err)
	}
	for _, sv := range ids {
		for _, v := range res {
			if v.id == sv {
				other := ""
				for i := range v.otherIds {
					if i == 0 {
						other = v.otherIds[i]
						continue
					}
					other = other + "," + v.otherIds[i]
				}
				if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v.id, v.typez, v.value, v.brand, v.query, v.title, v.score, v.review, v.price, v.seller, v.delivery, v.variant1, v.variant2, other, v.deliveryDate, v.stock, v.category, v.moreSellerOptions, v.startingFrom, v.availableQuantity, v.crossedPrice, v.freeFreight}); err != nil {
					log.Println(err)
				}
				num++

			}
		}
	}
	fileName := "out.xlsx"
	for fileNum := 1; exists(fileName); fileNum++ {
		fileName = "out(" + strconv.Itoa(fileNum) + ").xlsx"
	}
	xlsx.SaveAs(fileName)

	log.Println("完成")
}

// 文件是否存在
func exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	return !os.IsNotExist(err)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var IsC = false
var IsC2 = true

func crawler(id string, id_store string) {
	defer func() {
		wg.Done()
		<-ch
	}()
	for i := 0; i < 16; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}

		proxyStr := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxyStr)

		client := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		id_store = strings.TrimSpace(id_store)
		url := "https://www.walmart.com/ip/" + id
		if id_store != "" {
			url += "?&selectedSellerId=" + id_store
		}
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("Failed to create request for id %s: %v", id, err)
			continue
		}

		request.Header.Set("User-Agent", getRandomUserAgent())
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")
		request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		request.Header.Set("Cache-Control", "max-age=0")
		request.Header.Set("Downlink", getRandomDownlink())
		request.Header.Set("Dpr", getRandomDpr())
		request.Header.Set("Priority", "u=0, i")
		request.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="129", "Not=A?Brand";v="8", "Chromium";v="129"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", `document`)
		request.Header.Set("Sec-Fetch-Mode", `navigate`)
		request.Header.Set("Sec-Fetch-Site", `same-origin`)
		request.Header.Set("Sec-Fetch-User", `?1`)
		request.Header.Set("Upgrade-Insecure-Requests", `1`)

		// 初始化一个计数器用于追踪连续出现的次数
		consecutiveElseCount := 0
		for {
			response, err := client.Do(request)
			if err != nil {
				if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
					log.Println("错误代码打印：" + err.Error())
					log.Println("等待请求头超时，重新开始当前ID：" + id)
					consecutiveElseCount = 0 // 重置计数器，因为不是 else
					continue
				} else if strings.Contains(err.Error(), "441") {
					log.Println("代理超频！暂停10秒后继续...")
					time.Sleep(time.Second * 10)
					consecutiveElseCount = 0 // 重置计数器，因为不是 else
					continue
				} else if strings.Contains(err.Error(), "440") {
					log.Println("代理宽带超频！暂停5秒后继续...")
					time.Sleep(time.Second * 5)
					consecutiveElseCount = 0 // 重置计数器，因为不是 else
					continue
				} else if strings.Contains(err.Error(), "Request Rate Over Limit") {
					// 新增对 "Request Rate Over Limit" 错误的处理
					log.Println("超频警告：" + err.Error())
					log.Println("超频，暂停5秒后继续...")
					time.Sleep(time.Second * 5)
					consecutiveElseCount = 0 // 重置计数器
					continue
				} else {
					log.Println("错误信息：" + err.Error())
					log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + id)
					// 增加连续出现的错误计数器
					consecutiveElseCount++

					// 检查是否已连续出现10次
					if consecutiveElseCount >= 6 {
						log.Println("已连续出现6次错误，切换请求头...")

						// 检查当前请求头是否包含 "Upgrade-Insecure-Requests"
						if request.Header.Get("Upgrade-Insecure-Requests") == "" {
							request.Header.Set("Upgrade-Insecure-Requests", "1")
						} else {
							request.Header.Del("Upgrade-Insecure-Requests")
							request.Header.Set("User-Agent", getRandomUserAgent())
						}

						// 重置连续错误计数器
						consecutiveElseCount = 0
					}
					continue

				}
			}
			// 请求成功时，重置连续错误计数器
			consecutiveElseCount = 0
			defer response.Body.Close()

			result := ""
			if response.Header.Get("Content-Encoding") == "gzip" {
				reader, err := gzip.NewReader(response.Body)
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
				result = string(dataBytes)
			}

			wal := Wal{}
			wal.id = id
			if strings.Contains(result, "This page could not be found.") {
				wal.typez = "该商品不存在"
				res = append(res, wal)
				log.Println("id:" + id + "商品不存在")
				return
			}

			upc := regexp.MustCompile("upc\":\"(.{4,30}?)\"").FindAllStringSubmatch(result, -1)
			gtin := regexp.MustCompile("gtin13\":\"(.{4,30}?)\"").FindAllStringSubmatch(result, -1)
			fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
			if len(upc) > 0 {
				wal.value = upc[0][1]
				wal.typez = "upc"
			} else if len(gtin) > 0 {
				wal.value = gtin[0][1]
				wal.typez = "gtin"
			} else if len(fk) > 0 {
				log.Println("id:" + id + " 被风控,更换IP继续")
				IsC = !IsC
				continue
			} else {
				wal.value = ""
				wal.typez = "ean"
			}

			doc1, err := htmlquery.Parse(strings.NewReader(result))
			if err != nil {
				log.Printf("Failed to parse HTML for id %s: %v", id, err)
				continue
			}

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
			if err != nil {
				log.Printf("Failed to create goquery document for id %s: %v", id, err)
				continue
			}

			brand := regexp.MustCompile("\"brand\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
			if len(brand) > 0 {
				wal.brand = brand[0][1]
			}
			//标签
			query, err := htmlquery.QueryAll(doc1, "//div[@class='flex items-center mv2 flex-wrap']//span")
			if err != nil {
				log.Println("无标签,错误代码:", err)
			} else {
				queryStr := ""
				for _, v := range query {
					text := strings.TrimSpace(htmlquery.InnerText(v)) // 去掉左右空白字符
					if !strings.Contains(queryStr, text) {
						if queryStr != "" { // 如果 queryStr 不为空，则添加竖线
							queryStr += "|"
						}
						queryStr += text
					}
				}
				wal.query = queryStr
			}
			/////queryStr := ""
			/////doc.Find("#maincontent > section > main > div.flex.undefined.flex-column.h-100 > div:nth-child(2) > div > div.w_aoqv.w_wRee.w_fdPt > div > div:nth-child(2) > div > div > div:nth-child(1) > div.flex.items-center.mv2.flex-wrap > div > div > span").Each(func(i int, s *goquery.Selection) {
			/////text := s.Text()
			/////if !strings.Contains(queryStr, text) {
			/////	queryStr += text + " "
			/////}
			/////})
			/////wal.query = queryStr

			title := regexp.MustCompile("\"productName\":\"(.+?)\",").FindAllStringSubmatch(result, -1)
			if len(title) > 0 {
				wal.title = strings.Replace(title[0][1], "\\u0026", "&", -1)
			} else {
				log.Printf("Failed to get title for id %s", id)
				continue
			}

			stock := regexp.MustCompile(`("message":"Currently out of stock")`).FindAllStringSubmatch(result, -1)
			if len(stock) == 0 {
				wal.stock = "有库存"
			} else {
				wal.stock = "无库存"
			}

			score := regexp.MustCompile("[(]([\\d][.][\\d])[)]").FindAllStringSubmatch(result, -1)
			if len(score) > 0 {
				wal.score = score[0][1]
			}

			review := regexp.MustCompile("\"totalReviewCount\":(\\d+)").FindAllStringSubmatch(result, -1)
			if len(review) > 0 {
				wal.review = review[0][1]
			}

			price := regexp.MustCompile(`"best[^{]+?,"priceDisplay":"([^"]+)"`)
			price1 := price.FindAllString(result, -1)
			if len(price1) > 0 {
				if strings.Contains(price1[0], `"priceDisplay":"`) {
					parts := strings.Split(price1[0], `"priceDisplay":"`)
					if len(parts) > 1 {
						valueParts := strings.Split(parts[1], `"`)
						if len(valueParts) > 0 {
							reg := regexp.MustCompile(`[^\d.]`)
							numericValue := reg.ReplaceAllString(valueParts[0], "")
							wal.price = numericValue
						}
					}
				}
			}

			category := regexp.MustCompile(`categoryName":"(.+?)",`).FindAllStringSubmatch(result, -1)
			if len(category) > 0 {
				wal.category = strings.Replace(category[0][1], `\u0026`, "&", -1)
			}

			all, err := htmlquery.QueryAll(doc1, "//div/div/span[@class=\"lh-title\"]//text()")
			if err != nil {
				log.Printf("Failed to get seller and delivery info for id %s", id)
			} else {
				for i, v := range all {
					sv := htmlquery.InnerText(v)
					if strings.Contains(sv, "Sold by") {
						wal.seller = htmlquery.InnerText(all[i+1])
						continue
					}
					if strings.Contains(sv, "Fulfilled by") {
						wal.delivery = strings.Replace(sv, "Fulfilled by ", "", -1)
						if len(wal.delivery) < 3 && len(all) > i+1 {
							wal.delivery = htmlquery.InnerText(all[i+1])
						}
						continue
					}
					if strings.Contains(sv, "Sold and shipped by") {
						wal.seller = htmlquery.InnerText(all[i+1])
						wal.delivery = wal.seller
						break
					}
				}
			}

			if wal.seller == "" {
				seller := regexp.MustCompile("\"sellerDisplayName\":\"(.*?)\"").FindAllStringSubmatch(result, -1)
				if len(seller) > 0 {
					wal.seller = seller[0][1]
				}
			}

			//到达时间
			delivery_date := doc.Find(".f7.mt1.ws-normal.ttn").Text()
			if delivery_date == "" {
				delivery_date1 := doc.Find(".ma1.f7").Text()
				wal.deliveryDate = delivery_date1
				log.Println("到达时间", delivery_date1)
			} else {
				wal.deliveryDate = delivery_date
				log.Println("到达时间", delivery_date)
			}

			/////nodeList, err := htmlquery.QueryAll(doc1, "//*[@id=\"maincontent\"]/section/main/div[2]/div[2]/div/div[1]/div/div[2]/div/div/div[8]/section/div/div/div/div/div[1]/button/label/div[3]")
			/////if err != nil {
			////log.Printf("Failed to get delivery date for id %s: %v", id, err)
			////} else {
			////	var deliveryDate string
			////	for _, node := range nodeList {
			////		deliveryDate += htmlquery.InnerText(node) + " "
			////	}
			////	wal.deliveryDate = deliveryDate
			////}
			//划线价获取
			crossedPrice := `<span aria-hidden="true" data-seo-id="strike-through-price" class="mr2 f6 gray mr1 strike">(.*?)</span>`
			re := regexp.MustCompile(crossedPrice)
			// 在整个HTML内容中查找匹配的价格
			matches := re.FindStringSubmatch(string(result))
			if len(matches) > 1 { // 确保有匹配且有分组
				//fmt.Println("找到的划线价格是:", matches[1]) // $12.99，matches[1]是第一个括号内的匹配
				wal.crossedPrice = matches[1]
			} else {
				fmt.Println("没有找到匹配的划线价格")
			}

			freeFreightStr := ""
			doc.Find(".mt1.h1 .f7").Each(func(index int, s *goquery.Selection) {
				text := s.Text()
				if !strings.Contains(freeFreightStr, text) {
					freeFreightStr += text + " "
				}
			})
			wal.freeFreight = freeFreightStr

			variant := regexp.MustCompile(":</span><span class=\"ml1\">(.*?)</span>").FindAllStringSubmatch(result, -1)
			if len(variant) == 1 {
				wal.variant1 = variant[0][1]
			} else if len(variant) == 2 {
				wal.variant1 = variant[0][1]
				wal.variant2 = variant[1][1]
			}

			allString := regexp.MustCompile("\",\"usItemId\":\"([0-9]+?)\"").FindAllStringSubmatch(result, -1)
			for i := range allString {
				wal.otherIds = append(wal.otherIds, allString[i][1])
			}

			startingFrom := regexp.MustCompile(`"priceType":.{0,20},"priceString":"(\$[^<]+?)",`).FindAllStringSubmatch(result, -1)
			if len(startingFrom) > 0 {
				wal.startingFrom = startingFrom[0][1]
			}

			moreSellerOptions := regexp.MustCompile(`"additionalOfferCount":(\d+),`).FindAllStringSubmatch(result, -1)
			if len(moreSellerOptions) > 0 {
				wal.moreSellerOptions = moreSellerOptions[0][1]
			}

			availableQuantity := regexp.MustCompile("availableQuantity\":(\\d+),").FindAllStringSubmatch(result, -1)
			if len(availableQuantity) > 0 {
				wal.availableQuantity = availableQuantity[0][1]
			}

			log.Println("id:" + wal.id + "完成")
			res = append(res, wal)
			return
		}
	}
}
