package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"net/http"
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
		if err := file.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "到达时间"}); err != nil {
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

func crawler(id string, id_store string) {
	defer func() {
		wg.Done()
		<-ch
	}()
	for i := 0; i < 16; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}

		//proxyStr := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		//proxy, _ := url.Parse(proxyStr)

		//client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		id_store = strings.TrimSpace(id_store)
		url := "https://www.walmart.com/ip/" + id
		if id_store != "" {
			url += "?&selectedSellerId=" + id_store
		}
		request, err := http.NewRequest("PUT", url, nil)
		if err != nil {
			log.Printf("Failed to create request for id %s: %v", id, err)
			continue
		}

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

		response, err := client.Do(request)
		if err != nil {
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
			} else {
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + id)
				continue
			}
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

		//doc1, err := htmlquery.Parse(strings.NewReader(result))
		//if err != nil {
		//	log.Printf("Failed to parse HTML for id %s: %v", id, err)
		//	continue
		//}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
		if err != nil {
			log.Printf("Failed to create goquery document for id %s: %v", id, err)
			continue
		}

		// 正则表达式匹配日期
		dateRegex := regexp.MustCompile(`(\d+)`)

		// 处理匹配的元素
		deliveryDateStr := ""
		doc.Find(".f7.mt1.ws-normal.ttn").Each(func(index int, s *goquery.Selection) {
			text := s.Text()
			// 查找日期部分
			matches := dateRegex.FindStringSubmatch(text)
			if len(matches) > 1 {
				day, err := strconv.Atoi(matches[1])
				if err == nil {
					// 日期加 1
					newDay := day + 1
					// 替换原来的日期
					newText := strings.Replace(text, matches[1], strconv.Itoa(newDay), 1)
					if !strings.Contains(deliveryDateStr, newText) {
						deliveryDateStr += newText + " "
					}
				}
			}
		})

		fmt.Println("Updated Delivery Dates:", strings.TrimSpace(deliveryDateStr))
		if deliveryDateStr == "" {
			// 选择特定class的label标签，并获取其内部的第一个div标签的值
			firstLabel := doc.Find(".flex.flex-column.dark-gray.relative.overflow-hidden.items-center.flex-auto.mv1").First()
			firstLabel.Find(".ma1.f7").Each(func(i int, div *goquery.Selection) {
				text := div.Text()
				fmt.Println("First label div value:", text)
				deliveryDateStr = text
				// 只取第一个div的值，退出循环
				return
			})
		}

		wal.deliveryDate = deliveryDateStr
		log.Println("id:" + wal.id + "完成")
		appendToExcel(wal)
		return
	}
}

func appendToExcel(wal Wal) {
	mu.Lock()
	defer mu.Unlock()

	//other := strings.Join(wal.otherIds, ",")
	row := []interface{}{wal.id, wal.deliveryDate}

	if err := file.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &row); err != nil {
		log.Println("Failed to set sheet row:", err)
	}
	num++

	fileName := "out.xlsx"
	if err := file.SaveAs(fileName); err != nil {
		log.Println("Failed to save Excel file:", err)
	}
}
