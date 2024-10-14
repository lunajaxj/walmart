package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
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
	url        string
	name       string
	sellerName string
}

var ids []string
var wg = sync.WaitGroup{}
var ch = make(chan int, 10)

func main() {
	log.Println("自动化脚本-1688-信息获取")
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("url.txt")
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
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"产品链接", "产品名称", "供货商全称"}); err != nil {
		log.Println(err)
	}
	for _, sv := range ids {
		for _, v := range res {
			if v.url == sv {
				if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v.url, v.name, v.sellerName}); err != nil {
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

func crawler(ur string) {

	//配置代理
	defer func() {
		wg.Done()
		<-ch
	}()

	for i := 0; i < 10; i++ {
		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		request, _ := http.NewRequest("OPTIONS", ur, nil)

		request.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 11; Pixel 5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.91 Mobile Safari/537.36")
		request.Header.Set("sec-ch-ua", `"Not_A Brand";v="99", "Google Chrome";v="109", "Chromium";v="109"`)
		request.Header.Set("accept-language", `zh,zh-CN;q=0.9`)
		request.Header.Set("cache-control", `no-cache`)
		request.Header.Set("pragma", `no-cache`)
		request.Header.Set("accept-encoding", "gzip, deflate, br")

		request.Header.Set("cookie", "_med=dw:393&dh:851&pw:1080.75&ph:2340.25&ist:0; ali_ab=116.30.162.56.1666840558517.0; ali_apache_track=c_mid=b2b-2212813495591f7a87|c_lid=%E6%B7%B1%E5%9C%B3%E5%B8%82%E5%8F%AF%E5%A6%AE%E6%9C%8D%E8%A3%85%E6%9C%89%E9%99%90%E5%85%AC%E5%8F%B8|c_ms=1|c_mt=1; cna=byXeG75j3zkCAXQeojjPpuwB; taklid=dedcf924a31446c89d4f8e228b03e99e; cookie2=1339e2942f70344bc6e86d4611837f99; sgcookie=E100btODgiuaQ5W4ixYRmixPy3w9F%2FxFLTrVT6AfbzZtn6a%2Bapixdza7fAAKWxdM1raDcY%2BySse3S8X20N9L3Vyeo9kcllIBB3V%2FMz6Wo7qbWSo%3D; hng=SG%7Czh-CN%7CSGD%7C702; t=878a75bdc32b53c30fb9aaca00368453; _tb_token_=573e14839ee5; lid=%E5%AE%9D%E5%AE%9D%E5%BF%83%E9%87%8C%E8%8B%A6%E5%93%A6%E4%B8%B6; uc4=nk4=0%400FJ7kSS1jybMSGShpoGhmHIyHyTiJS27Ig%3D%3D&id4=0%40U2OT6jTUTbKf9Gy%2BBKPpDy42env6; __cn_logon__=false; __mwb_logon_id__=undefined; mwb=ng; _csrf_token=1679447788307; alicnweb=touch_tb_at%3D1679450497303%7Clastlogonid%3D%25E6%25B7%25B1%25E5%259C%25B3%25E5%25B8%2582%25E5%258F%25AF%25E5%25A6%25AE%25E6%259C%258D%25E8%25A3%2585%25E6%259C%2589%25E9%2599%2590%25E5%2585%25AC%25E5%258F%25B8%7Cshow_inter_tips%3Dfalse; XSRF-TOKEN=a486dee1-5adf-4031-b0a3-d81c7c410688; _bl_uid=UClU5fa7j5q1I3sF6cIadppzv3Ov; _m_h5_tk=ca921383307a75d24e1ba98d755535b9_1679459831592; _m_h5_tk_enc=caf3d293ef66fd77ec53bd5803938923; csrfToken=4fc57b27-a100-4ac7-90fa-458c791bd566; isg=BKurZF-NS-hA2ZDeQRAvFZu6Os-VwL9CSVmzXR0oh-pTvMsepZBPkkldEnTShxc6; tfstk=cBlVBNGSFIdV2Ig9v7VZ4muDIRzfaX2TL64bnh38sCsGxee0YsbPXzm7ey4dBCec.; l=fBjG7uvHTwSH6y2JBOfaPurza779IIRbYuPzaNbMi9fP9w66oxqcW1Mao-KBCn1NF67DR3lCMSApBeYBqgI-nxvtCIXJgIkmnmOk-Wf..")

		//request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		response, err := client.Do(request)

		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始：" + ur)
			} else {
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + ur)
			}
			time.Sleep(1000)
			continue
		}

		wal := Wal{}
		wal.url = ur
		result := ""
		reader, err := gzip.NewReader(response.Body) // gzip解压缩
		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始：" + ur)
			} else {
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + ur)
			}
			time.Sleep(1000)
			continue
		}
		defer reader.Close()
		con, err := io.ReadAll(reader)
		if err != nil {
			log.Println(err)
			time.Sleep(1000)
			continue
		}
		result = string(con)
		//fk := regexp.MustCompile(`(tracker.install)`).FindAllStringSubmatch(result, -1)
		//if len(fk) > 0 {
		//	log.Println(ur, "风控重试")
		//	time.Sleep(1000)
		//	continue
		//}
		log.Println(result)

		name := regexp.MustCompile(`],"title":"(.+?)",`).FindAllStringSubmatch(result, -1)
		if len(name) > 0 {
			wal.name = name[0][1]
		}

		sellerName := regexp.MustCompile(`{"companyName":"(.*?)"`).FindAllStringSubmatch(result, -1)
		if len(sellerName) > 0 {
			wal.sellerName = sellerName[0][1]
			log.Println(sellerName[0][1])
		}

		log.Println("url:" + wal.url + "完成")
		res = append(res, wal)
		return
	}
}
