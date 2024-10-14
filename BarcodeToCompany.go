package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var barcode []string
var wg sync.WaitGroup
var ch = make(chan int, 5)
var resultWriter *csv.Writer

type bar struct {
	barcode string
	company string
}

var bars []bar

func main() {
	// 读取跳转链接列表文件
	file, err := os.Open("条码.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 解析跳转链接列表文件
	log.Println("条码获取公司自动化脚本")

	r := bufio.NewReader(file)
	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 5 {
			barcode = append(barcode, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}
	}

	// 创建结果文件
	resultFile, err := os.Create("result.csv")
	if err != nil {
		fmt.Println("Error creating result file:", err)
		return
	}
	defer resultFile.Close()

	// 写入结果文件表头
	resultWriter = csv.NewWriter(resultFile)
	resultWriter.Write([]string{"条码", "公司"})

	// 创建 WaitGroup 以便协调 goroutine

	// 使用 goroutine 执行转换跳转链接
	for i := range barcode {
		wg.Add(1)
		ch <- 1
		go to(i)
	}
	// 等待所有 goroutine 执行完成
	wg.Wait()
	for i := range barcode {
		for i2 := range bars {
			if bars[i2].barcode == barcode[i] {
				resultWriter.Write([]string{bars[i2].barcode, bars[i2].company})
				break
			}
		}
	}

	// 刷新结果文件
	resultWriter.Flush()

	// 输出结果文件名
	fmt.Println("完成，结果文件: result.csv")
}

func to(index int) {
	defer func() {
		wg.Done()
		<-ch
	}()
	for {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		proxyUrl, _ := url.Parse("http://l752.kdltps.com:15818")

		tr.Proxy = http.ProxyURL(proxyUrl)
		basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("t19932187800946:wsad123456"))
		tr.ProxyConnectHeader = http.Header{}
		tr.ProxyConnectHeader.Add("Proxy-Authorization", basicAuth)
		client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
		request, _ := http.NewRequest("GET", "https://bff.gds.org.cn/gds/searching-api/ImportProduct/GetImportProductDataForGtin?PageSize=30&PageIndex=1&AndOr=0&Gtin="+barcode[index], nil)

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
				log.Println("连续出现代理IP无效请联系我，重新开始：" + barcode[index])
				time.Sleep(time.Second * 1)
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
				log.Println("出现错误，如果同条码连续出现请联系我，重新开始：" + barcode[index])
				time.Sleep(time.Second * 1)
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

		//品牌
		b := bar{}
		realname := regexp.MustCompile(`"realname":"(.+)?",`).FindAllStringSubmatch(result, -1)
		if len(realname) == 0 {
			b.barcode = barcode[index]
			b.company = "无"
		} else {
			b.barcode = barcode[index]
			b.company = realname[0][1]
		}
		bars = append(bars, b)
		log.Println(b.barcode, "完成")
		return
	}

}
