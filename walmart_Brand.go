package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
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

var res []Wal
var mu sync.Mutex // 定义互斥锁

type Wal struct {
	id                string
	sellerDisplayName string
	brand             string
	departments       string
}

var ids []string
var wg = sync.WaitGroup{}
var ch = make(chan int, 6)

func main() {
	log.Println("自动化脚本-walmart-卖家信息获取")
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("id.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 2 {
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

	xlsx := excelize.NewFile()
	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "卖家名", "品牌", "系别"}); err != nil {
		log.Println(err)
	}
	for _, v := range res {
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v.id, v.sellerDisplayName, v.brand, v.departments}); err != nil {
			log.Println(err)
		}
		num++
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
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var IsC = true
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

func crawler(u string) {
	defer func() {
		<-ch
		wg.Done()
	}()

	const maxRetries = 3 // 最大重试次数
	var success bool
	var wal Wal
	wal.id = u

	for i := 0; i < maxRetries; i++ {

		if i != 0 {
			time.Sleep(time.Second * 2)
		}
		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		request, _ := http.NewRequest("GET", "https://www.walmart.com/global/seller/"+u, nil)

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
			log.Println("请求错误:", err)
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
				log.Println("读取body错误，重新开始:", err)
				continue
			}
			defer response.Body.Close()
			result = string(dataBytes)
		}

		if strings.Contains(result, "This page could not be found.") {
			log.Println("id:" + u + "商品不存在")
			break
		}

		if strings.Contains(result, "Robot or human?") {
			log.Println(u + " 被风控,更换IP继续")
			IsC = !isc
			continue
		}

		result = strings.Replace(result, `\u0026`, `&`, -1)
		sellerDisplayName := regexp.MustCompile(`"sellerDisplayName":"(.+?)"`).FindAllStringSubmatch(result, -1)
		if len(sellerDisplayName) > 0 {
			wal.sellerDisplayName = sellerDisplayName[0][1]
		}

		// Match Brand
		Brand := regexp.MustCompile(`{"title":"Brand",(.+?){"title":"Speed"`).FindAllStringSubmatch(result, -1)
		if len(Brand) > 0 {
			matchedBlock := Brand[0][0]
			nameRegex := regexp.MustCompile(`"name":"(.*?)"`)
			nameMatches := nameRegex.FindAllStringSubmatch(matchedBlock, -1)
			itemCountRegex := regexp.MustCompile(`"itemCount":(\d+)`)
			itemCountMatches := itemCountRegex.FindAllStringSubmatch(matchedBlock, -1)
			if len(nameMatches) > 0 && len(itemCountMatches) > 0 {
				var combined []string
				if len(nameMatches)-1 == len(itemCountMatches) {
					for j := 1; j < len(nameMatches); j++ {
						combined = append(combined, fmt.Sprintf("%s(%s)", nameMatches[j][1], itemCountMatches[j-1][1]))
					}
					wal.brand = strings.Join(combined, "|")
					success = true
				} else {
					log.Println("name 和 itemCount 的数量不一致")
				}
			} else {
				log.Println("未找到任何 'name' 或 'itemCount' 字段")
			}
		}

		// Match Departments
		Departments := regexp.MustCompile(`{"title":"Departments",(.+?){"title":"Price"`).FindAllStringSubmatch(result, -1)
		if len(Departments) > 0 {
			matchedBlock := Departments[0][0]
			nameRegex := regexp.MustCompile(`"name":"(.*?)"`)
			nameMatches := nameRegex.FindAllStringSubmatch(matchedBlock, -1)
			itemCountRegex := regexp.MustCompile(`"itemCount":(\d+)`)
			itemCountMatches := itemCountRegex.FindAllStringSubmatch(matchedBlock, -1)
			if len(nameMatches) > 0 && len(itemCountMatches) > 0 {
				var combined []string
				if len(nameMatches)-1 == len(itemCountMatches) {
					for j := 1; j < len(nameMatches); j++ {
						combined = append(combined, fmt.Sprintf("%s(%s)", nameMatches[j][1], itemCountMatches[j-1][1]))
					}
					wal.departments = strings.Join(combined, "|")
					success = true
				} else {
					log.Println("name 和 itemCount 的数量不一致")
				}
			} else {
				log.Println("未找到任何 'name' 或 'itemCount' 字段")
			}
		}

		if success {
			break
		}
	}

	if !success {
		log.Println("未找到匹配的 Brand 或 Departments，id:", u)
		wal.sellerDisplayName = "none"
		wal.brand = "none"
		wal.departments = "none"
	}

	// 使用互斥锁保证并发安全
	mu.Lock()
	res = append(res, wal)
	mu.Unlock()

	log.Println("id:" + u + "卖家获取完成")
}
