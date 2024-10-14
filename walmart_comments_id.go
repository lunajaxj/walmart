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
	id string

	zp      int
	zpStart time.Time
	zpEnd   time.Time

	vp      int
	vpStart time.Time
	vpEnd   time.Time

	rating1 int
	rating2 int
	rating3 int
	rating4 int
	rating5 int
}

var res = make(map[string]Mode)
var ids []string
var lock sync.Mutex
var wg = sync.WaitGroup{}
var ch = make(chan int, 5)

func main() {
	log.Println("自动化脚本-walmart-itemid获取评论信息")
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("id.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 0 {
			ids = append(ids, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}

	}
	log.Println("共", len(ids), "个任务")

	for _, v := range ids {
		ch <- 1
		wg.Add(1)
		go crawler(v)
	}
	wg.Wait()

	log.Println("完成")

}

func save(keys string) {
	lock.Lock()
	defer func() {
		wg.Done()
		lock.Unlock()
	}()
	fileName := "out.xlsx"
	xlsx, err := excelize.OpenFile("out.xlsx")
	if err != nil {
		xlsx = excelize.NewFile()
	}
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "zp数量", "zp最早时间", "zp最晚时间", "vp数量", "vp最早时间", "vp最晚时间", "vp一星", "vp二星", "vp三星", "vp四星", "vp五星"}); err != nil {
		log.Println(err)
	}
	num := 0
	for i := range ids {
		if ids[i] == keys {
			num = i + 2
		}
	}
	if num == 0 {
		return
	}
	if res[keys].id != "" {
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{res[keys].id, res[keys].zp, res[keys].zpStart.Format("2006/1/2"), res[keys].zpEnd.Format("2006/1/2"), res[keys].vp, res[keys].vpStart.Format("2006/1/2"), res[keys].vpEnd.Format("2006/1/2"), res[keys].rating1, res[keys].rating2, res[keys].rating3, res[keys].rating4, res[keys].rating5}); err != nil {
			log.Println(err)
		}
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
	defer func() {
		<-ch

		wg.Add(1)
		go save(id)
		wg.Done()

	}()

	mode := Mode{}
	mode.id = id
	for i := 1; i <= 25; {
		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		log.Println("id:"+id, "第", i, "页")
		idd := ""
		if i != 1 {
			idd = id + "?page=" + strconv.Itoa(i)
		} else {
			idd = id
		}
		request, _ := http.NewRequest("GET", "https://www.walmart.com/reviews/product/"+idd, nil)

		request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36")
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
				log.Println("代理IP不可用，更换代理继续执行 id:" + id)
				time.Sleep(time.Second * 1)
			} else if strings.Contains(err.Error(), "441") {
				log.Println("代理超频！暂停10秒后继续...")
				time.Sleep(time.Second * 10)
			} else if strings.Contains(err.Error(), "440") {
				log.Println("代理宽带超频！暂停5秒后继续...")
				time.Sleep(time.Second * 5)
			} else if strings.Contains(err.Error(), "tls: first record does not look like a TLS handshake") {
				log.Println("代理IP不可用，更换代理继续执行 id:" + id)
			} else if strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("请求超时，重新开始：" + id)
				time.Sleep(time.Second * 1)
			} else {
				log.Println(err)
				log.Println("请求错误，重新开始：" + id)
				time.Sleep(time.Second * 1)
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
		//log.Println(result)
		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
		if len(fk) > 0 {
			log.Println("被风控，更换IP重新开始" + id)
			IsC = !isc
			time.Sleep(time.Second * 1)
			continue
		}

		zp := regexp.MustCompile(`"badges":null`).FindAllStringSubmatch(result, -1)
		mode.zp += len(zp)
		zptime := regexp.MustCompile(`"badges":null.+?"reviewSubmissionTime":"(.+?)",`).FindAllStringSubmatch(result, -1)
		for i2 := range zptime {
			t, _ := time.Parse("1/2/2006", zptime[i2][1])
			if mode.zpStart.IsZero() {
				mode.zpStart = t
				mode.zpEnd = t
			} else {
				if t.Before(mode.zpStart) {
					mode.zpStart = t
				} else if mode.zpEnd.Before(t) {
					mode.zpEnd = t
				}
			}
		}
		vp := regexp.MustCompile(`"badges":\[`).FindAllStringSubmatch(result, -1)

		mode.vp += len(vp)
		ration := regexp.MustCompile(`"badges":\[.*?rating":(\d),`).FindAllStringSubmatch(result, -1)
		for i2 := range ration {
			switch ration[i2][1] {
			case "1":
				mode.rating1 += 1
			case "2":
				mode.rating2 += 1
			case "3":
				mode.rating3 += 1
			case "4":
				mode.rating4 += 1
			case "5":
				mode.rating5 += 1
			}
		}

		vptime := regexp.MustCompile(`"badges":\[.+?"reviewSubmissionTime":"(.+?)",`).FindAllStringSubmatch(result, -1)
		for i2 := range vptime {
			t, _ := time.Parse("1/2/2006", vptime[i2][1])
			if mode.vpStart.IsZero() {
				mode.vpStart = t
				mode.vpEnd = t
			} else {
				if t.Before(mode.vpStart) {
					mode.vpStart = t
				} else if mode.vpEnd.Before(t) {
					mode.vpEnd = t
				}
			}
		}
		//最大分页
		maxPage := regexp.MustCompile(`{"num":(\d+)?,"url`).FindAllStringSubmatch(result, -1)
		atoi := 25
		if len(maxPage) != 0 {
			atoi, err = strconv.Atoi(maxPage[len(maxPage)-1][1])
		}
		if i >= atoi {
			res[id] = mode
			return
		} else {
			i++
		}
	}
	res[id] = mode

}
