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

// 伪装userAgents
func getRandomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
		//"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0.2 Safari/605.1.15",
		//"Mozilla/5.0 (Linux; Android 10; Pixel 3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.127 Mobile Safari/537.36",

	}

	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}

// 伪装下载速度
func getRandomDownlink() string {
	return fmt.Sprintf("%.1f", rand.Float64()*2+1) // 1-3 MB
}

// 伪装物理设备像素比
func getRandomDpr() string {
	return fmt.Sprintf("%.1f", rand.Float64()*1+1) // 1-2
}

// 结构体
type Wal struct {
	id    string
	brand string
}

var ids []string

// 并发同步锁，goroutine组
var wg = sync.WaitGroup{}

// 通道数10
var ch = make(chan int, 10)

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

	log.Println("自动化脚本-walmart-brand")
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
		if err := file.SetSheetRow("Sheet1", "A1", &[]interface{}{"brand id", "brand"}); err != nil {
			log.Println(err)
		}
		num = 2
	}

	// 创建句柄
	fi, err := os.Open("brand_id.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 3 {
			ids = append(ids, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}

	}
	log.Println(ids)
	for _, v := range ids {
		ch <- 1
		wg.Add(1)
		go crawler(v)
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

func crawler(id string) {
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

		url := "https://www.walmart.com/brand/" + id

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
			//if strings.Contains(result, "This page could not be found.") {
			//	wal.typez = "该商品不存在"
			//	appendToExcel(wal)
			//	log.Println("id:" + id + "商品不存在")
			//	return
			//}

			fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
			if len(fk) > 0 {
				log.Println("id:" + id + " 被风控,更换IP继续")
				IsC = !IsC
				continue
			}
			//支持Xpath选择器
			//doc1, err := htmlquery.Parse(strings.NewReader(result))
			//if err != nil {
			//	log.Printf("Failed to parse HTML for id %s: %v", id, err)
			//	continue
			//}
			//支持select及正则查询
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
			if err != nil {
				log.Printf("Failed to create goquery document for id %s: %v", id, err)
				continue
			}

			// 使用 CSS 选择器获取 h1 元素
			h1 := doc.Find("#__next > div > div > div:nth-child(2) > div > main > h1").Text()
			if h1 == "" {
				log.Printf("未找到 h1 元素，id %s", id)
			} else {
				wal.brand = strings.TrimSpace(h1) // 提取并去除空格
			}
			log.Println("id:" + wal.id + "完成")
			appendToExcel(wal)
			return
		}
	}
}

// 单独封装实时存储excel方法
func appendToExcel(wal Wal) {
	mu.Lock()
	defer mu.Unlock()

	row := []interface{}{wal.id, wal.brand}

	if err := file.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &row); err != nil {
		log.Println("Failed to set sheet row:", err)
	}
	num++

	fileName := "out.xlsx"
	if err := file.SaveAs(fileName); err != nil {
		log.Println("Failed to save Excel file:", err)
	}
}
