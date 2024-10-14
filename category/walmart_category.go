package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var iaa = 0

var demos []string

var hc = make(map[string]int)

func main() {
	fi, err := os.Open("cps.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader
	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 0 {
			space := strings.TrimSpace(string(lineB))
			cp(space)
		}
		if err != nil {
			break
		}

	}

	fileName := "out.txt"
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	file.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM，防止中文乱码
	writer := bufio.NewWriter(file)
	for i := range demos {
		writer.WriteString(demos[i] + "\n")
	}

	writer.Flush() //内容是先写到缓存对，所以需要调用flush将缓存对数据真正写到文件中
	fmt.Println("完成")
}

func cp(urll string) {
	if hc[urll] == 1 {
		return
	}
	hc[urll] = 1
	if !strings.Contains(urll, "http") {
		urll = "https://www.walmart.com" + urll
	}
	fmt.Println(urll)

	proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "m78z02hx", "l752.kdltps.com:15818")
	proxy, _ := url.Parse(proxy_str)

	client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	request, _ := http.NewRequest("GET", urll, nil)

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
	if response.StatusCode == 441 || strings.Contains(err.Error(), "441") {
		log.Println("代理超频！暂停10秒后继续...")
		time.Sleep(time.Second * 10)
	}
	if err != nil {
		if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
			log.Println("代理IP无效，自动切换中")
			log.Println("连续出现代理IP无效请联系我，重新开始")
			time.Sleep(1 * time.Second)
			cp(urll)
		} else {
			log.Println("错误信息：" + err.Error())
			log.Println("出现错误，如果同id连续出现请联系我，重新开始")
			time.Sleep(1 * time.Second)
			cp(urll)
		}
		return
	}
	result := ""
	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(response.Body) // gzip解压缩
		if err != nil {
			log.Println("解析body错误，重新开始")
			cp(urll)
			return
		}
		defer reader.Close()
		con, err := io.ReadAll(reader)
		if err != nil {
			log.Println("gzip解压错误，重新开始")
			cp(urll)
			return
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
			cp(urll)
			return
		}
		defer response.Body.Close()
		result = string(dataBytes)
	}
	fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
	if len(fk) > 0 {
		log.Println("被风控,更换IP继续")
		time.Sleep(1 * time.Second)
		cp(urll)
		return
	}
	result = strings.Replace(result, "\\u0026", "&", -1)
	metaCanon := regexp.MustCompile("\"metaCanon\":\"(.*?)\"").FindAllStringSubmatch(result, -1)
	iaa++
	if len(metaCanon) > 0 && strings.Contains(metaCanon[0][1], "/browse/") {
		demos = append(demos, metaCanon[0][1])
		log.Println("完成:", iaa, "br:", 1, "cp:", 0)
	} else if len(metaCanon) > 0 && strings.Contains(metaCanon[0][1], "/cp/") {
		browse := regexp.MustCompile("\"value\":\"(https://www.walmart.com/browse[^*]*?)\"|\"value\":\"(/browse[^*]*?)\"").FindAllStringSubmatch(result, -1)
		browse2 := regexp.MustCompile("\"value\":\"(https://www.walmart.com/browse[^*^\"]*?/0\\?[^\"]*?)\"|\"value\":\"(/browse[^*^\"]*?/0\\?[^\"]*?)\"").FindAllStringSubmatch(result, -1)
		copy(browse, browse2)
		for i := range browse {
			split := strings.Split(browse[i][1], "?")
			demos = append(demos, split[0])
		}
		cps := regexp.MustCompile("\"value\":\"(https://www.walmart.com/cp/[^\\[].*?)\"|\"value\":\"(/cp/[^\\[].*?)\"").FindAllStringSubmatch(result, -1)
		for i := range cps {
			split := strings.Split(cps[i][1], "?")
			cp(split[0])
			//fmt.Println(cpUrls[i][1])
		}
		log.Println("完成:", iaa, "br:", len(browse), "cp:", len(cps))
	} else {
		log.Println("不存在:", urll)
	}

}
