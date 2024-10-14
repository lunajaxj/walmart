package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"github.com/xuri/excelize/v2"
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

type Mode struct {
	keyword string
	id      string //div[@data-testid="list-view"]/preceding-sibling::a[@link-identifier]/@link-identifier
	page    string //div[@data-testid="list-view"]//div[@class="relative"]//img/@alt
	locate  string //div[@data-testid="list-view"]/div/div[1]/div[1]/font/font

}

var res = make(map[string]Mode)
var keywords []string
var lock sync.Mutex
var lockk sync.Mutex
var lock1 sync.Mutex
var wg = sync.WaitGroup{}
var wgg = sync.WaitGroup{}
var wggg = sync.WaitGroup{}
var ch = make(chan int, 4)
var chh = make(chan int, 10)

func main() {
	log.Println("自动化脚本-walmart-关键词获取itemid位置")
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("关键词与id.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 0 {
			keywords = append(keywords, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}

	}
	log.Println("共", len(keywords), "个任务")

	for _, v := range keywords {
		ch <- 1
		wg.Add(1)
		split := strings.Split(v, "|")
		go crawler(split[0], split[1], split[2], split[2], true)
	}
	wg.Wait()

	log.Println("完成")

}

func save(keys string) {
	defer func() {
		wg.Done()
		lock1.Unlock()
	}()
	fileName := "out.xlsx"
	xlsx, err := excelize.OpenFile("out.xlsx")
	if err != nil {
		xlsx = excelize.NewFile()
	}
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"关键词", "id", "页", "位"}); err != nil {
		log.Println(err)
	}
	var num int
	for i := range keywords {
		if keywords[i] == keys {
			num = i + 2
		}
	}
	if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{res[keys].keyword, res[keys].id, res[keys].page, res[keys].locate}); err != nil {
		log.Println(err)
	}

	xlsx.Save()
	if err != nil {
		xlsx.SaveAs(fileName)
	}
}

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

func crawler(keyword string, itemid, next string, page string, t bool) {
	defer func() {

		wg.Add(1)
		lock1.Lock()
		go save(keyword + "|" + itemid + "|" + next)
		if t {
			<-ch
			wg.Done()
		}
	}()

	for i := 0; i < 25; {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}
		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		var k = strings.Replace(url.QueryEscape(keyword), "%20", "+", -1)
		urll := ""
		var pa, er = strconv.Atoi(page)
		if er != nil {
			return
		}
		if pa > 25 || pa < 1 {
			res[keyword+"|"+itemid+"|"+next] = Mode{keyword, itemid, "无", "无"}
			return
		} else if page != "1" {
			urll = k + "&page=" + page + "&affinityOverride=default"
		} else {
			urll = k + "&affinityOverride=default"
		}
		//urll = strings.Replace(urll, " ", "+", -1)
		log.Println("关键词:"+keyword, "搜索第"+page+"页")
		request, _ := http.NewRequest("GET", "https://www.walmart.com/search?q="+urll, nil)

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
			if strings.Contains(err.Error(), "Proxy Bad Server") {
				log.Println("代理IP不可用，更换代理继续执行 关键词:" + keyword)
			} else if strings.Contains(err.Error(), "441") {
				log.Println("代理超频！暂停10秒后继续...")
				time.Sleep(time.Second * 10)
			} else if strings.Contains(err.Error(), "440") {
				log.Println("代理宽带超频！暂停5秒后继续...")
				time.Sleep(time.Second * 5)
			} else if strings.Contains(err.Error(), "tls: first record does not look like a TLS handshake") {
				log.Println("代理IP不可用，更换代理继续执行 关键词:" + keyword)
			} else if strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("请求超时，重新开始：" + keyword)
			} else {
				log.Println(err)
				log.Println("请求错误，重新开始：" + keyword)
			}
			i++
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
		//log.Println(result)

		cw1 := regexp.MustCompile("(is not valid JSON)").FindAllStringSubmatch(result, -1)
		cw2 := regexp.MustCompile("(The requested URL was rejected. Please consult with your administrator)").FindAllStringSubmatch(result, -1)
		if len(cw1) > 0 || len(cw2) > 0 {
			log.Println("搜索内容错误，跳过该标题：" + keyword)
			res[keyword+"|"+itemid+"|"+next] = Mode{keyword, itemid, "搜索内容错误", "跳过该标题"}
			return
		}
		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
		if len(fk) > 0 {
			log.Println("被风控，更换IP重新开始" + keyword)
			IsC = !isc
			i++
			continue
		}
		//doc, err := htmlquery.Parse(strings.NewReader(result))
		if err != nil {
			log.Println("错误信息：" + err.Error())
			res[keyword+"|"+itemid+"|"+next] = Mode{keyword, itemid, "获取失败", err.Error()}
			return
		}
		//log.Println(keyword+page)
		//log.Println(result)
		resultStr := ""
		resultS := regexp.MustCompile("(items\":\\[.+?\\].+?layoutEnum)").FindAllStringSubmatch(result, -1)
		allString := regexp.MustCompile("(There were no search results for)").FindAllString(result, -1)
		if len(allString) > 0 {
			log.Println("关键词:" + keyword + " 第" + page + "页 无搜索结果")
			res[keyword+"|"+itemid+"|"+next] = Mode{keyword, itemid, "无", "无"}
			return
		}
		if len(resultS) == 0 {
			log.Println("被风控，更换IP重新开始" + keyword)
			i++
			continue
		} else {
			resultStr = resultS[0][1]
		}
		//id
		id := regexp.MustCompile("\",\"usItemId\":\"([0-9]+?)\"").FindAllStringSubmatch(resultStr, -1)
		//id,err  := htmlquery.QueryAll(doc, "//div[@data-testid=\"list-view\"]/preceding-sibling::a[@link-identifier]/@link-identifier")
		//log.Println(id[0][])
		split := strings.Split(itemid, ",")
		for i := range id {
			for i2 := range split {
				if id[i][1] == split[i2] {
					res[keyword+"|"+itemid+"|"+next] = Mode{keyword, split[i2], page, strconv.Itoa(i + 1)}
					log.Println("关键词:" + keyword + " -> 第" + page + "页  第" + strconv.Itoa(i+1) + "个")
					return
				}
			}
		}
		//最大分页
		maxPage := regexp.MustCompile("\"maxPage\":([0-9]+?),").FindAllStringSubmatch(result, -1)
		atoi := 25
		if len(maxPage) != 0 {
			atoi, err = strconv.Atoi(maxPage[0][1])
		}
		ne, _ := strconv.Atoi(next)
		if pa < atoi && pa < 25 && pa < ne && pa > 1 {
			crawler(keyword, itemid, next, strconv.Itoa(pa-1), false)
		} else if pa < atoi && pa < 25 && pa >= ne {
			crawler(keyword, itemid, next, strconv.Itoa(pa+1), false)
		} else if pa == atoi && pa <= 25 {
			crawler(keyword, itemid, next, strconv.Itoa(ne-1), false)
		} else {
			res[keyword+"|"+itemid+"|"+next] = Mode{keyword, itemid, "无", "无"}
		}
		return
	}

	res[keyword+"|"+itemid+"|"+next] = Mode{keyword, itemid, "获取失败", "获取失败"}

}
