package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/transport"
	"github.com/wangluozhe/requests/url"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"math/rand"
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
var ch = make(chan int, 5)

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

	log.Println("自动化脚本-walmart-信息采集-跟卖库存信息")
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
		if err := file.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "商品码类型", "商品码值", "品牌", "标签", "标题", "评分", "评论数量", "价格", "卖家", "配送", "变体1", "变体2", "变体id", "到达时间", "库存", "类目", "跟卖数量", "跟卖最低价格", "库存数量", "划线价", "自发货运费"}); err != nil {
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
		ch <- 1
		wg.Add(1)
		go crawler(ids[i], idStores[i])
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var IsC = false
var IsC2 = true

// 动态生成指纹（包括User-Agent、sec-ch-ua等）
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
}

var secChUa = []string{
	`"Chromium";v="128", "Not;A=Brand";v="24", "Google Chrome";v="128"`,
	`"Chromium";v="114", "Not;A=Brand";v="8", "Google Chrome";v="114"`,
}

// 随机生成User-Agent和Sec-Ch-Ua指纹信息
func getRandomFingerprint() (string, string) {
	rand.Seed(time.Now().UnixNano())

	randomUserAgent := userAgents[rand.Intn(len(userAgents))]
	randomSecChUa := secChUa[rand.Intn(len(secChUa))]

	return randomUserAgent, randomSecChUa
}

// 生成随机的16字节长字符串，用于动态生成cookie中的某些字段
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(bytes)
}

// 随机生成cookie字段
//func generateRandomCookie() string {
//	pxhd := generateRandomString(16)  // 随机生成16字节字符串
//	pxvid := generateRandomString(16) // 随机生成16字节字符串
//
//	// 随机生成其他部分
//	abqmeValues := []string{"false", "true"}
//	vtcValues := []string{"SZOP6XX-elbFxV3dtNEggI", "QWERTY-12345-ABCDE"}
//
//	// 随机选择值
//	abqme := abqmeValues[randInt(0, len(abqmeValues))]
//	vtc := vtcValues[randInt(0, len(vtcValues))]
//
//	// 组合生成随机的cookie
//	cookie := fmt.Sprintf("_pxhd=%s:%s; _pxvid=%s; abqme=%s; vtc=%s", pxhd, generateRandomString(10), pxvid, abqme, vtc)
//	return cookie
//}

// 使用 math/rand 随机生成整数
func randInt(min, max int) int {
	rand.Seed(time.Now().UnixNano()) // 需要为 math/rand 设定种子
	return rand.Intn(max-min) + min
}

func randomJA3(ja3 string) string {
	// Split the JA3 string into its components
	parts := strings.Split(ja3, ",")
	if len(parts) != 5 {
		return ""
	}

	sslVersion := parts[0]
	ciphers := parts[1]
	extensions := parts[2]
	curves := parts[3]
	orders := parts[4]

	// Split the extensions into a slice
	extensionsList := strings.Split(extensions, "-")

	// Check if "41" is in the extensions list
	is41 := false
	for _, ext := range extensionsList {
		if ext == "41" {
			is41 = true
			break
		}
	}

	// Randomly shuffle the extensions list
	rand.Seed(time.Now().UnixNano())
	if !is41 {
		rand.Shuffle(len(extensionsList), func(i, j int) {
			extensionsList[i], extensionsList[j] = extensionsList[j], extensionsList[i]
		})
	} else {
		// Remove "41" from the list
		for i := 0; i < len(extensionsList); i++ {
			if extensionsList[i] == "41" {
				extensionsList = append(extensionsList[:i], extensionsList[i+1:]...)
				break
			}
		}
		// Shuffle the remaining extensions
		rand.Shuffle(len(extensionsList), func(i, j int) {
			extensionsList[i], extensionsList[j] = extensionsList[j], extensionsList[i]
		})
		// Add "41" back to the end of the list
		extensionsList = append(extensionsList, "41")
	}

	// Join the extensions back into a single string
	extensions = strings.Join(extensionsList, "-")

	// Reassemble and return the JA3 string
	return strings.Join([]string{sslVersion, ciphers, extensions, curves, orders}, ",")
}

func crawler(id string, id_store string) {
	defer func() {
		wg.Done()
		<-ch
	}()

	for i := 0; i < 16; i++ {
		if i != 0 {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(200)+100)) // 添加100ms到300ms的随机延迟
		}

		// 设置代理
		req := url.NewRequest()
		req.Timeout = 10 * time.Second
		req.Proxies = "http://t19932187800946:wsad123456@l752.kdltps.com:15818"
		//proxyStr := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		//proxy, _ := url.Parse(proxyStr)

		// 创建HTTP客户端
		//client := &http.Client{
		//	Timeout: 10 * time.Second,
		//	Transport: &http.Transport{
		//		Proxy:             http.ProxyURL(proxy),
		//		DisableKeepAlives: true,
		//		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		//	},
		//}

		id_store = strings.TrimSpace(id_store)
		url_walmart := "https://www.walmart.com/ip/" + id
		if id_store != "" {
			url_walmart += "?&selectedSellerId=" + id_store
		}
		log.Println(url_walmart)

		headers := url.NewHeaders()
		headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		headers.Set("cache-control", "no-cache")
		headers.Set("downlink", "10")
		headers.Set("dpr", "1")
		headers.Set("pragma", "no-cache")
		headers.Set("priority", "u=0, i")
		headers.Set("sec-ch-ua", "\"Not)A;Brand\";v=\"99\", \"Google Chrome\";v=\"127\", \"Chromium\";v=\"127\"")
		headers.Set("sec-ch-ua-mobile", "?0")
		headers.Set("sec-ch-ua-platform", "\"Windows\"")
		headers.Set("sec-fetch-dest", "document")
		headers.Set("sec-fetch-mode", "navigate")
		headers.Set("sec-fetch-site", "same-origin")
		headers.Set("sec-fetch-user", "?1")
		headers.Set("upgrade-insecure-requests", "1")
		headers.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")
		(*headers)["Header-Order:"] = []string{ // 请求头排序, 值必须为小写
			"accept",
			"cache-control",
			"downlink",
			"dpr",
			"pragma",
			"priority",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"upgrade-insecure-requests",
			"user-agent",
		}
		req.Headers = headers

		//original_ja3 := "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-21,29-23-24,0"
		//newJA3 := randomJA3(original_ja3)
		//req.Ja3 = newJA3
		// fmt.Println("Original JA3:", original_ja3)
		// fmt.Println("Modified JA3:", newJA3)

		es := &transport.Extensions{
			SupportedSignatureAlgorithms: []string{
				"ECDSAWithP256AndSHA256",
				"ECDSAWithP384AndSHA384",
				"ECDSAWithP521AndSHA512",
				"PSSWithSHA256",
				"PSSWithSHA384",
				"PSSWithSHA512",
				"PKCS1WithSHA256",
				"PKCS1WithSHA384",
				"PKCS1WithSHA512",
				"ECDSAWithSHA1",
				"PKCS1WithSHA1",
			},
			RecordSizeLimit: 4001,
			DelegatedCredentials: []string{
				"ECDSAWithP256AndSHA256",
				"ECDSAWithP384AndSHA384",
				"ECDSAWithP521AndSHA512",
				"ECDSAWithSHA1",
			},
			SupportedVersions: []string{
				"1.3",
				"1.2",
			},
			PSKKeyExchangeModes: []string{
				"PskModeDHE",
			},
			KeyShareCurves: []string{
				"X25519",
				"P256",
			},
		}
		tes := transport.ToTLSExtensions(es)
		req.TLSExtensions = tes

		h2s := &transport.H2Settings{
			Settings: map[string]int{
				"HEADER_TABLE_SIZE":   65536,
				"INITIAL_WINDOW_SIZE": 131072,
				"MAX_FRAME_SIZE":      16384,
			},
			SettingsOrder: []string{
				"HEADER_TABLE_SIZE",
				"INITIAL_WINDOW_SIZE",
				"MAX_FRAME_SIZE",
			},
			ConnectionFlow: 12517377,
			HeaderPriority: map[string]interface{}{
				"weight":    42,
				"streamDep": 13,
				"exclusive": false,
			},
			PriorityFrames: []map[string]interface{}{
				{
					"streamID": 3,
					"priorityParam": map[string]interface{}{
						"weight":    201,
						"streamDep": 0,
						"exclusive": false,
					},
				},
				{
					"streamID": 5,
					"priorityParam": map[string]interface{}{
						"weight":    101,
						"streamDep": 0,
						"exclusive": false,
					},
				},
				{
					"streamID": 7,
					"priorityParam": map[string]interface{}{
						"weight":    1,
						"streamDep": 0,
						"exclusive": false,
					},
				},
				{
					"streamID": 9,
					"priorityParam": map[string]interface{}{
						"weight":    1,
						"streamDep": 7,
						"exclusive": false,
					},
				},
				{
					"streamID": 11,
					"priorityParam": map[string]interface{}{
						"weight":    1,
						"streamDep": 3,
						"exclusive": false,
					},
				},
				{
					"streamID": 13,
					"priorityParam": map[string]interface{}{
						"weight":    241,
						"streamDep": 0,
						"exclusive": false,
					},
				},
			},
		}
		h2ss := transport.ToHTTP2Settings(h2s)
		req.HTTP2Settings = h2ss

		// 创建请求
		r, err := requests.Get(url_walmart, req)
		if err != nil {
			log.Printf("Failed to create request for id %s: %v", id, err)
			return
		}
		//fmt.Println(r.Text)

		//request, err := http.NewRequest("PUT", url, nil)
		//if err != nil {
		//	log.Printf("Failed to create request for id %s: %v", id, err)
		//	return
		//}

		// 获取随机指纹信息
		//randomUserAgent, randomSecChUa := getRandomFingerprint()

		// 设置所有请求头
		//request.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		//request.Header.Set("accept-encoding", "gzip, deflate, br, zstd")
		//request.Header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
		//request.Header.Set("cache-control", "max-age=0")

		// 设置动态生成的cookie
		// request.Header.Set("cookie", generateRandomCookie())

		//request.Header.Set("downlink", "10")
		//request.Header.Set("dpr", "1.5")
		//request.Header.Set("priority", "u=0, i")
		//request.Header.Set("sec-ch-ua", randomSecChUa)
		//request.Header.Set("sec-ch-ua-mobile", "?0")
		//request.Header.Set("sec-ch-ua-platform", `"Windows"`)
		//request.Header.Set("sec-fetch-dest", "document")
		//request.Header.Set("sec-fetch-mode", "navigate")
		//request.Header.Set("sec-fetch-site", "same-origin")
		//request.Header.Set("sec-fetch-user", "?1")
		//request.Header.Set("upgrade-insecure-requests", "1")
		//request.Header.Set("user-agent", randomUserAgent)
		//request.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")

		//response, err := client.Do(request)
		//if err != nil {
		//	if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
		//		log.Println("错误代码打印：" + err.Error())
		//		log.Println("等待请求头超时，重新开始当前ID：" + id)
		//		continue
		//	} else if strings.Contains(err.Error(), "441") {
		//		log.Println("代理超频！暂停10秒后继续...")
		//		time.Sleep(time.Second * 10)
		//		continue
		//	} else if strings.Contains(err.Error(), "440") {
		//		log.Println("代理宽带超频！暂停5秒后继续...")
		//		time.Sleep(time.Second * 5)
		//		continue
		//	} else {
		//		log.Println("错误信息：" + err.Error())
		//		log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + id)
		//		continue
		//	}
		//}
		//defer response.Body.Close()
		//// 检查响应
		//log.Printf("Response status for id %s: %s", id, response.Status)
		//result := ""
		//if response.Header.Get("Content-Encoding") == "gzip" {
		//	reader, err := gzip.NewReader(response.Body)
		//	if err != nil {
		//		log.Println("解析body错误，重新开始：" + id)
		//		continue
		//	}
		//	defer reader.Close()
		//	con, err := io.ReadAll(reader)
		//	if err != nil {
		//		log.Println("gzip解压错误，重新开始：" + id)
		//		continue
		//	}
		//	result = string(con)
		//} else {
		//	dataBytes, err := io.ReadAll(response.Body)
		//	if err != nil {
		//		if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "Service Unavailable") {
		//			log.Println("代理IP无效，自动切换中")
		//			log.Println("连续出现代理IP无效请联系我，重新开始：" + id)
		//		} else {
		//			log.Println("错误信息：" + err.Error())
		//			log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + id)
		//		}
		//		continue
		//	}
		//	result = string(dataBytes)
		//}

		result := ""
		result = r.Text

		wal := Wal{}
		wal.id = id
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

		queryStr := ""
		doc.Find("#maincontent > section > main > div.flex.flex-column.h-100 > div:nth-child(2) > div > div.w_aoqv.w_wRee.w_fdPt > div > div:nth-child(2) > div > div > div:nth-child(1) > div.flex.items-center.mv2.flex-wrap > div:nth-child(1) > div > span").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			if !strings.Contains(queryStr, text) {
				queryStr += text + " "
			}
		})
		wal.query = queryStr

		title := regexp.MustCompile("\"name\":\"(.+?)\",").FindAllStringSubmatch(result, -1)
		if len(title) > 0 {
			wal.title = strings.Replace(title[0][1], "\\u0026", "&", -1)
		} else {
			log.Printf("Failed to get title for id %s", id)
			//continue
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

		crossedPrice := `<span aria-hidden="true" class="mr2 f6 gray strike">(.*?)</span>`
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

		log.Println("id:" + wal.id + "完成")
		appendToExcel(wal)
		return
	}
}

func appendToExcel(wal Wal) {
	mu.Lock()
	defer mu.Unlock()

	other := strings.Join(wal.otherIds, ",")
	row := []interface{}{
		wal.id, wal.typez, wal.value, wal.brand, wal.query, wal.title, wal.score, wal.review,
		wal.price, wal.seller, wal.delivery, wal.variant1, wal.variant2, other, wal.deliveryDate,
		wal.stock, wal.category, wal.moreSellerOptions, wal.startingFrom, wal.availableQuantity,
		wal.crossedPrice, wal.freeFreight,
	}

	if err := file.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &row); err != nil {
		log.Println("Failed to set sheet row:", err)
	}
	num++

	fileName := "out.xlsx"

	for {
		if err := file.SaveAs(fileName); err != nil {
			log.Println("Failed to save Excel file, retrying:", err)
			time.Sleep(1 * time.Second) // 每秒钟重试一次
		} else {
			log.Println("Successfully saved Excel file.")
			break // 保存成功后退出循环
		}
	}
}
