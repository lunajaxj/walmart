package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
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

var res []Wal

type Wal struct {
	id                 string
	sellerId           string
	delivery           string //配送
	sellerDisplayName  string
	sellerName         string
	address            string
	sellerEmail        string
	brand              string
	deactivationStatus string //存活状态
	publishedDate      string //公开日期
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
		if len(lineB) > 5 {
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
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "配送", "店铺", "店铺", "公司", "地址", "邮箱", "deactivationStatus", "publishedDate"}); err != nil {
		log.Println(err)
	}
	for _, sv := range ids {
		for _, v := range res {
			if v.id == sv {
				if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v.id, v.delivery, v.sellerId, v.sellerDisplayName, v.sellerName, v.address, v.sellerEmail, v.deactivationStatus, v.publishedDate}); err != nil {
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
func crawler(id string) {

	//配置代理
	defer func() {
		wg.Done()
		<-ch
	}()

	for i := 0; i < 16; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}
		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		request, _ := http.NewRequest("GET", "https://www.walmart.com/ip/"+id, nil)

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
				log.Println("连续出现代理IP无效请联系我，重新开始：" + id)
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
		wal := Wal{}
		wal.id = id
		if strings.Contains(result, "This page could not be found.") {
			wal.sellerDisplayName = "该商品不存在"

			log.Println("id:" + id + "商品不存在")
			return
		}

		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)

		if len(fk) > 0 {
			log.Println("id:" + id + " 被风控,更换IP继续")
			IsC = !isc
			continue
		}

		doc, err := htmlquery.Parse(strings.NewReader(result))
		if err != nil {
			log.Println("错误信息：" + err.Error())
			return
		}
		all, err := htmlquery.QueryAll(doc, "//div/div/span[@class=\"lh-title\"]//text()")
		//log.Println(result)
		if err != nil {
			log.Println("卖家与配送获取失败")
		} else {
			for i, v := range all {
				sv := htmlquery.InnerText(v)
				if strings.Contains(sv, "Sold by") {
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
					wal.delivery = htmlquery.InnerText(all[i+1])
					break
				}
			}
		}

		//url
		price := regexp.MustCompile(`href="(/seller/.+?)\?`).FindAllStringSubmatch(result, -1)
		if len(price) > 0 {
			priceId := regexp.MustCompile(`href="/seller/(.+?)\?`).FindAllStringSubmatch(result, -1)
			wal.sellerId = priceId[0][1]
			crawler2(price[0][1], &wal)
		} else {
			wal.sellerDisplayName = "失败"
		}

		log.Println("id:" + wal.id + "完成")
		res = append(res, wal)
		return
	}
}

func crawler2(u string, wal *Wal) {

	for i := 0; i < 16; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}
		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		request, _ := http.NewRequest("GET", "https://www.walmart.com/"+u, nil)

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
		response, err := client.Do(request)

		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始：" + u)
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
				log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + u)
			}
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
				log.Println("gzip解压错误，重新开始：")
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
		if strings.Contains(result, "This page could not be found.") {
			log.Println("id:" + u + "商品不存在")
			return
		}

		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)

		if len(fk) > 0 {
			log.Println(u + " 被风控,更换IP继续")
			continue
		}
		sellerDisplayName := regexp.MustCompile(`"sellerDisplayName":"(.+?)"`).FindAllStringSubmatch(result, -1)
		sellerName := regexp.MustCompile(`"sellerName":"(.+?)"`).FindAllStringSubmatch(result, -1)
		if len(sellerDisplayName) > 0 {
			wal.sellerDisplayName = sellerDisplayName[0][1]
		}
		if len(sellerName) > 0 {
			wal.sellerName = sellerName[0][1]
		}

		var add string
		address1 := regexp.MustCompile(`"address1":"(.+?)"`).FindAllStringSubmatch(result, -1)
		address2 := regexp.MustCompile(`"address2":"(.+?)"`).FindAllStringSubmatch(result, -1)
		city := regexp.MustCompile(`"city":"(.+?)"`).FindAllStringSubmatch(result, -1)
		state := regexp.MustCompile(`"state":"(.+?)"`).FindAllStringSubmatch(result, -1)
		postalCode := regexp.MustCompile(`"postalCode":"(.+?)"`).FindAllStringSubmatch(result, -1)
		country := regexp.MustCompile(`"country":"(.+?)"`).FindAllStringSubmatch(result, -1)
		if len(address1) > 0 {
			add += address1[0][1] + ","
		}
		if len(address2) > 0 {
			add += address2[0][1] + ","
		}
		if len(city) > 0 {
			add += city[0][1] + ","
		}
		if len(state) > 0 {
			add += state[0][1] + " "
		}
		if len(postalCode) > 0 {
			add += postalCode[0][1] + ","
		}
		if len(country) > 0 {
			add += country[0][1]
		}
		wal.address = add
		//邮箱
		sellerEmail := regexp.MustCompile(`"sellerEmail":"(.+?)"`).FindAllStringSubmatch(result, -1)
		if len(sellerEmail) > 0 {
			wal.sellerEmail = sellerEmail[0][1]
		}
		//deactivationStatus
		deactivationStatus := regexp.MustCompile(`"deactivationStatus":"(.+?)"`).FindAllStringSubmatch(result, -1)
		if len(deactivationStatus) > 0 {
			wal.deactivationStatus = deactivationStatus[0][1]
		}
		//publishedDate
		publishedDate := regexp.MustCompile(`"publishedDate":(\d+)`).FindAllStringSubmatch(result, -1)
		if len(publishedDate) > 0 {
			// 将匹配到的字符串转换为整数
			milliseconds, err := strconv.ParseInt(publishedDate[0][1], 10, 64)
			if err != nil {
				log.Printf("Failed to parse published date for id %s: %v", u, err)
				return
			}

			// 将时间戳转换为可读的时间
			t := time.Unix(0, milliseconds*int64(time.Millisecond))

			// 格式化输出
			fmt.Println("可读日期时间:", t.Format("2006-01-02 15:04:05"))
			wal.publishedDate = t.Format("2006-01-02 15:04:05")
		}
		log.Println("id:" + wal.id + "卖家获取完成")
		return
	}
}
