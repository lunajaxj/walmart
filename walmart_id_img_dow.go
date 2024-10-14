package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"fmt"
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

var res []Wal

type Wal struct {
	img []string
	id  string
}

var ids []string
var wg = sync.WaitGroup{}
var wgImg = sync.WaitGroup{}

var ch = make(chan int, 3)

func main() {
	log.Println("自动化脚本-walmart-id图片下载")
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
	for _, v := range ids {
		ch <- 1
		wg.Add(1)
		go crawler(v)
	}
	wg.Wait()
	log.Println("开始下载图片")
	for i := range res {
		for i2 := range res[i].img {
			if len(res[i].img[i2]) >= 2 {
				ch <- 1
				wg.Add(1)
				go imgxzz(res[i].id, res[i].img[i2], strconv.Itoa(i2))
			}
		}
	}
	wg.Wait()
	log.Println("全部下载完成")

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
			res = append(res, wal)
			log.Println("id:" + id + "商品不存在")
			return
		}

		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)

		if len(fk) > 0 {
			log.Println("id:" + id + " 被风控,更换IP继续")
			IsC = !isc
			continue
		}
		img := regexp.MustCompile(`data-testid="media-thumbnail" style="line-height:0"><img loading="lazy" srcset="([^,^"^?]+)(?:\?[^"]*)?`).FindAllStringSubmatch(result, -1)
		for i := range img {
			wal.img = append(wal.img, img[i][1])
		}
		log.Println(id, "图片地址获取完成")
		res = append(res, wal)
		return
	}

}

func imgxzz(id, img, num string) {
	defer func() {
		wg.Done()
		<-ch
	}()
	if len(img) < 10 {
		return
	}
	for i := 0; i < 5; i++ {
		imgPath := ".\\img\\" + id + "\\"
		img = strings.Replace(img, "\\u0026", "&", -1)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		proxyUrl, _ := url.Parse("http://l752.kdltps.com:15818")
		tr.Proxy = http.ProxyURL(proxyUrl)
		basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("t19932187800946:wsad123456"))
		tr.ProxyConnectHeader = http.Header{}
		tr.ProxyConnectHeader.Add("Proxy-Authorization", basicAuth)

		client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
		request, err := http.NewRequest("GET", img, nil)
		if err != nil {
			log.Println("图片下载失败!"+err.Error(), "稍后重新开始下载")
			time.Sleep(3 * time.Second)
			continue
		}
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
		response, err := client.Do(request)
		var res []byte
		if err != nil {
			log.Println(id, num, "图片下载失败!"+err.Error(), "稍后重新开始下载")
			time.Sleep(3 * time.Second)
			continue
		}

		res, err = io.ReadAll(response.Body)
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

		// 判断目录是否存在
		if _, err := os.Stat(imgPath); os.IsNotExist(err) {
			// 目录不存在，创建目录
			err := os.MkdirAll(imgPath, os.ModePerm)
			if err != nil {
				fmt.Println("无法创建目录:", err)
			}
		}
		file, err := os.Create(imgPath + num + ".jpeg")
		if err != nil {
			log.Println(id, num, "图片下载失败!")
			continue
		}
		file.Write(res)
		log.Println(id, "第", num, "张图片下载成功")
		return
	}
}
