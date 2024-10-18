package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/xuri/excelize/v2"
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

var mu sync.Mutex
var file *excelize.File
var num int

type Wal struct {
	id                string
	ids_Wal           string
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

// var ch chan int
var sem chan struct{} // 信号量控制并发
// var ch = make(chan int, 3)
var currentRate int

func init() {
	// 初始化 currentRate 通过读取速率文件
	currentRate = readRateFromFile("speed_walmart_2.txt")
	if currentRate <= 0 {
		currentRate = 6 // 如果读取失败，设置默认值为5
	}
	fmt.Printf("当前速率为: %d\n", currentRate)
	//ch = make(chan int, currentRate) // 创建通道
	sem = make(chan struct{}, currentRate) // 创建信号量
}

func readRateFromFile(filename string) int {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("无法读取速率文件 %s，使用默认速率: %v", filename, err)
		return 6 // 默认速率
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		rate, err := strconv.Atoi(scanner.Text())
		if err != nil {
			log.Printf("速率文件中的值无效，使用默认速率: %v", err)
			return 6 // 默认速率
		}
		return rate
	}
	log.Printf("速率文件为空，使用默认速率")
	return 6 // 默认速率
}

func main() {
	log.Println("自动化脚本-walmart-信息采集")
	log.Println("开始执行...")

	// 打开或创建Excel文件
	var fileName = "out.xlsx"
	var fileErr error
	if exists(fileName) {
		file, fileErr = excelize.OpenFile(fileName)
		if fileErr != nil {
			log.Fatalf("Failed to open existing Excel file: %v", fileErr)
		}
		sheetName := file.GetSheetName(0)
		rows, err := file.GetRows(sheetName)
		if err != nil {
			log.Fatalf("Failed to get rows from existing Excel file: %v", err)
		}
		num = len(rows) + 1
	} else {
		file = excelize.NewFile()
		if err := file.SetSheetRow("Sheet1", "A1", &[]interface{}{"id_storeid", "id", "商品码类型", "商品码值", "品牌", "标签", "标题", "评分", "评论数量", "价格", "卖家", "配送", "变体1", "变体2", "变体id", "到达时间", "库存", "类目", "跟卖数量", "跟卖最低价格", "库存数量", "划线价", "自发货运费"}); err != nil {
			log.Println(err)
		}
		num = 2
	}

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
		wg.Add(1)
		sem <- struct{}{} // 获取信号量
		go func(id, idStore string) {
			defer func() {
				<-sem // 释放信号量
				wg.Done()
			}()
			crawler(id, idStore)
		}(ids[i], idStores[i])
	}

	wg.Wait()

	if err := file.SaveAs(fileName); err != nil {
		log.Println("Failed to save Excel file:", err)
	}

	log.Println("完成")
}

// 文件是否存在
func exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	return !os.IsNotExist(err)
}

// var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
// var IsC = false
// var IsC2 = true
var prevHeaders map[string]string

func crawler(id string, id_store string) {
	//defer func() {
	//	wg.Done()
	//	<-ch
	//}()
	rateLimitCounter := 0 // 风控计数器
	for i := 0; i < 3; i++ {
		// 在每次请求之间增加一个随机的延迟，延迟的时间根据 currentRate 动态调整
		time.Sleep(time.Second * time.Duration(11-currentRate))
		if i != 0 {
			time.Sleep(time.Second * 2)
		}

		proxyStr := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxyStr)

		client := &http.Client{Timeout: 60 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		id_store = strings.TrimSpace(id_store)
		url := "https://www.walmart.com/ip/" + id
		if id_store != "" {
			url += "?&selectedSellerId=" + id_store
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("Failed to create request for id %s: %v", id, err)
			continue
		}
		// 再次检查 req 是否为 nil，防止请求在使用过程中被错误操作
		if req == nil {
			log.Printf("Request became nil for id %s, skipping", id)
			break
		}

		setHeaders(req) // 设置初始请求头
		// 检查在请求发送之前，当前请求的所有请求头
		//log.Printf("准备发送请求的请求头 for id %s:\n%v", id, req.Header)
		response, err := client.Do(req)
		if err != nil {
			// 错误处理逻辑
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("错误代码打印：" + err.Error())
				log.Println("等待请求头超时，重新开始当前ID：" + id)

				continue
			} else if strings.Contains(err.Error(), "441") {
				log.Println("代理超频！暂停10秒后继续...")
				time.Sleep(time.Second * 10)

				continue
			} else if strings.Contains(err.Error(), "440") {
				log.Println("代理宽带超频！暂停5秒后继续...")
				time.Sleep(time.Second * 5)

				continue
			} else if strings.Contains(err.Error(), "Request Rate Over Limit") {
				log.Println("超频警告：" + err.Error())
				log.Println("超频，暂停5秒后继续...")
				time.Sleep(time.Second * 5)

				continue
			} else {
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + id)
			}
		}
		// 检查 response 是否为 nil
		if response == nil {
			log.Printf("Response is nil for id %s, skipping this iteration", id)
			continue
		}
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
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + id)
				continue
			}
			result = string(dataBytes)
		}

		wal := Wal{}
		wal.id = id
		if id_store != "" {
			wal.ids_Wal = id + "|" + id_store
		} else {
			wal.ids_Wal = id + "|"
		}
		if strings.Contains(result, "This page could not be found.") {
			wal.typez = "该商品不存在"
			appendToExcel(wal)
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
			log.Println("id:" + id + " 被风控，重新请求")
			//IsC = !IsC
			rateLimitCounter++
			log.Printf("当前出现风控次数统计: %d", rateLimitCounter)

			// 当风控连续出现2次时，调整请求头和速率
			if rateLimitCounter >= 2 {
				log.Println("连续2次被风控，调整请求头参数和速率")
				// 更换请求头，随机删除一个请求头
				removableHeaders := []string{"Upgrade-Insecure-Requests"}
				headerToRemove := removableHeaders[rand.Intn(len(removableHeaders))]
				req.Header.Del(headerToRemove)
				log.Printf("由于风控连续两次，请删除或增加requestHeaders_wal_2中请求头参数")
				log.Printf("当前请求头参数:%s", req.Header)

				// 随机等待 5 到 10 秒
				time.Sleep(time.Second * time.Duration(5+rand.Intn(6)))

				// 将请求速率降低
				if currentRate > 4 {
					currentRate--
					log.Printf("当前请求速率已降低至: %d", currentRate)
				} else if currentRate == 4 {
					log.Printf("当前请求速率已降低至4,无法降速,等待5秒后重新请求")
					time.Sleep(5 * time.Second)
				}

				rateLimitCounter = 0 // 重置计数器
			}

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

		nodeList, err := htmlquery.QueryAll(doc1, "//*[@id=\"maincontent\"]/section/main/div[2]/div[2]/div/div[1]/div/div[2]/div/div/div[8]/section/div/div/div/div/div[1]/button/label/div[3]")
		if err != nil {
			log.Printf("Failed to get delivery date for id %s: %v", id, err)
		} else {
			var deliveryDate string
			for _, node := range nodeList {
				deliveryDate += htmlquery.InnerText(node) + " "
			}
			wal.deliveryDate = deliveryDate
		}

		crossedPrice := `<span aria-hidden="true" data-seo-id="strike-through-price" class="mr2 f6 gray strike">(.*?)</span>`
		re := regexp.MustCompile(crossedPrice)
		matches := re.FindStringSubmatch(string(result))
		if len(matches) > 1 {
			wal.crossedPrice = matches[1]
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

		if id_store != "" {
			log.Println("id:" + id + "|" + id_store + " 完成")
		} else {
			log.Println("id:" + id + " 完成")
			//log.Printf("更换后的请求头信息 for id %s:\n%v", id, req.Header)
		}
		appendToExcel(wal)
		return
	}
}

func setHeaders(req *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	rand.Seed(time.Now().UnixNano())

	// 读取请求头参数从文件中
	file, err := os.Open("requestHeaders_wal_2.txt")
	if err != nil {
		log.Fatalf("无法打开请求头文件: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var headerBuffer strings.Builder

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		// 检查是否为请求头的值部分（即前一行没分割完的值）
		if !strings.Contains(line, ":") {
			headerBuffer.WriteString(line)
			headerBuffer.WriteString(" ")
			continue
		}

		// 如果读取到新的请求头字段，首先处理之前的缓冲区内容
		if headerBuffer.Len() > 0 {
			headerParts := strings.SplitN(headerBuffer.String(), ":", 2)
			headerBuffer.Reset()
			if len(headerParts) == 2 {
				headerName := strings.TrimSpace(headerParts[0])
				headerValue := strings.TrimSpace(headerParts[1])

				// 检查头字段名称是否合法
				if !isValidHeaderName(headerName) {
					//log.Printf("忽略非法请求头字段名称: %s", headerName)
					continue
				}

				req.Header.Set(headerName, headerValue)
			}
		}

		// 将当前行内容写入缓冲区
		headerBuffer.WriteString(line)
	}

	// 处理文件末尾的缓冲区内容
	if headerBuffer.Len() > 0 {
		headerParts := strings.SplitN(headerBuffer.String(), ":", 2)
		if len(headerParts) == 2 {
			headerName := strings.TrimSpace(headerParts[0])
			headerValue := strings.TrimSpace(headerParts[1])

			// 检查头字段名称是否合法
			if !isValidHeaderName(headerName) {
				//log.Printf("忽略非法请求头字段名称: %s", headerName)
				return
			}

			req.Header.Set(headerName, headerValue)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("读取请求头文件出错: %v", err)
	}
}

// 检查头字段名称是否合法
func isValidHeaderName(name string) bool {
	for _, r := range name {
		if !(r == '-' || r == '_' || ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z') || ('0' <= r && r <= '9')) {
			return false
		}
	}
	return true
}
func appendToExcel(wal Wal) {
	mu.Lock()
	defer mu.Unlock()

	other := strings.Join(wal.otherIds, ",")
	row := []interface{}{wal.ids_Wal, wal.id, wal.typez, wal.value, wal.brand, wal.query, wal.title, wal.score, wal.review, wal.price, wal.seller, wal.delivery, wal.variant1, wal.variant2, other, wal.deliveryDate, wal.stock, wal.category, wal.moreSellerOptions, wal.startingFrom, wal.availableQuantity, wal.crossedPrice, wal.freeFreight}

	if err := file.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &row); err != nil {
		log.Println("Failed to set sheet row:", err)
	}
	num++

	fileName := "out.xlsx"
	if err := file.SaveAs(fileName); err != nil {
		log.Println("Failed to save Excel file:", err)
	}
}
