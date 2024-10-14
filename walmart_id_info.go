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

type Wal struct {
	id                string
	sellerDisplayName string
	sellerName        string
	address           string
	Pf                string
	PlNum             string
	sellerEmail       string
	brand             string
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
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"店铺id", "店铺", "公司", "地址", "评分", "评论数量", "邮箱"}); err != nil {
		log.Println(err)
	}
	for _, sv := range ids {
		for _, v := range res {
			if v.id == sv {
				if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v.id, v.sellerDisplayName, v.sellerName, v.address, v.Pf, v.PlNum, v.sellerEmail}); err != nil {
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

	for i := 0; i < 16; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}
		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		request, _ := http.NewRequest("GET", "https://www.walmart.com/seller/"+u, nil)

		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")
		request.Header.Set("Accept-Language", "en")
		request.Header.Set("Sec-Ch-Ua", `"Not.A/Brand";v="8", "Chromium";v="114", "Google Chrome";v="114"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", `document`)
		request.Header.Set("Sec-Fetch-Mode", `navigate`)
		request.Header.Set("Sec-Fetch-Site", `none`)
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
		//if strings.Contains(result, "This page could not be found.") {
		//	//log.Println(result)
		//	log.Println("id:" + u + "商品不存在")
		//	continue
		//}

		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
		if len(fk) > 0 {
			log.Println(u + " 被风控,更换IP继续")
			IsC = !isc
			continue
		}
		result = strings.Replace(result, `\u0026`, `&`, -1)
		sellerDisplayName := regexp.MustCompile(`"sellerDisplayName":"(.+?)"`).FindAllStringSubmatch(result, -1)
		sellerName := regexp.MustCompile(`"sellerName":"(.+?)"`).FindAllStringSubmatch(result, -1)
		wal := Wal{}
		wal.id = u
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

		plNum := regexp.MustCompile(`>\((\d+) reviews\)<`).FindAllStringSubmatch(result, -1)

		pf := regexp.MustCompile(`"averageOverallRating":(.*?),`).FindAllStringSubmatch(result, -1)
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
		if len(plNum) > 0 {
			wal.PlNum = plNum[0][1]
		}
		if len(pf) > 0 {
			wal.Pf = pf[0][1]
		}

		wal.address = add
		//邮箱
		sellerEmail := regexp.MustCompile(`"sellerEmail":"(.+?)"`).FindAllStringSubmatch(result, -1)
		if len(sellerEmail) > 0 {
			wal.sellerEmail = sellerEmail[0][1]
		}

		//添加所有结果
		res = append(res, wal)
		log.Println("id:" + u + "卖家获取完成")
		return
	}
}
