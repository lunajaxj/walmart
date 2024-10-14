package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var iaa = 0

var demos []string

var hc = make(map[string]int)

var brs = make(map[string][]Browse)

type Browse struct {
	id    string //div[@data-testid="list-view"]/preceding-sibling::a[@link-identifier]/@link-identifier
	name  string //div[@data-testid="list-view"]//div[@class="relative"]//img/@alt
	price string //div[@data-testid="list-view"]/div/div[1]/div[1]/font/font
}

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
			br(space)
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

	file.WriteString("id,title,name") // 写入UTF-8 BOM，防止中文乱码
	writer := bufio.NewWriter(file)
	for i := range demos {
		writer.WriteString(demos[i] + "\n")
	}

	writer.Flush() //内容是先写到缓存对，所以需要调用flush将缓存对数据真正写到文件中
	fmt.Println("完成")
}

func br(urll string) {
	var browses []Browse
	if hc[urll] == 1 {
		return
	}
	hc[urll] = 1
	if !strings.Contains(urll, "http") {
		urll = "https://www.walmart.com" + urll
	}
	max := 1
	h1Title := ""
	for i := 1; i <= max; i++ {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

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

		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始")
				time.Sleep(1 * time.Second)
				br(urll)
			} else {
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同id连续出现请联系我，重新开始")
				time.Sleep(1 * time.Second)
				br(urll)
			}
			return
		}
		resultStr := ""
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
			resultStr = string(con)
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
			resultStr = string(dataBytes)
		}
		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(resultStr, -1)
		if len(fk) > 0 {
			log.Println("被风控,更换IP继续")
			time.Sleep(1 * time.Second)
			br(urll)
			return
		}

		//分类标题
		h1 := regexp.MustCompile("\"h1\":\"(.*?)\"").FindAllStringSubmatch(resultStr, -1)
		if h1Title != "" && len(h1) != 0 {
			h1Title = h1[0][1]
		}

		//id
		id := regexp.MustCompile("usItemId\":\"([0-9]+?)\",\"[^c]").FindAllStringSubmatch(resultStr, -1)

		//价格
		price := regexp.MustCompile("\"price\":(.+?),").FindAllStringSubmatch(resultStr, -1)

		//标题
		name := regexp.MustCompile("\"usItemId\":\"[0-9]+?\",\"fitmentLabel\":.{0,100}?\"name\":\"(.*?)\"").FindAllStringSubmatch(resultStr, -1)

		//最大分页
		maxPage := regexp.MustCompile("\"maxPage\":([0-9]+?),").FindAllStringSubmatch(resultStr, -1)
		atoi, err := strconv.Atoi(maxPage[0][1])
		max = atoi

		for i := range id {
			browses = append(browses, Browse{
				id:    id[i][1],
				name:  name[i][1],
				price: price[i][1],
			})
		}
	}
	brs[h1Title] = browses
	iaa++
	log.Println("完成:", iaa, "br:", len(browses))

}
