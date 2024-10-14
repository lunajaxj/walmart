package main

import (
	"crypto/tls"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}

func main() {
	// 打开原始 xlsx 文件
	xlsx, err := excelize.OpenFile("input.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	// 获取第一个工作表
	sheet := "Sheet1"

	// 获取数据行数
	rows, err := xlsx.GetRows(sheet)
	if err != nil {
		log.Fatal(1, err)
	}

	// 初始化 HTTP 客户端
	proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
	proxy, _ := url.Parse(proxy_str)

	client := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	// 循环处理每一行数据
	log.Printf("共%d行需要翻译\n", len(rows))
	for rowIdx := 1; rowIdx <= len(rows); rowIdx++ {
		var english1, english2, english3, english4, english5 string
		row := rows[rowIdx-1]
		if rowIdx == 1 {
			for colIdx, value := range row {
				cellName, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
				xlsx.SetCellValue(sheet, cellName, value)
			}
			continue
		}
		// 获取英文列的内容
		if len(row) >= 1 {
			english1 = row[0]
		}
		if len(row) >= 2 {
			english2 = row[1]
		}
		if len(row) >= 3 {
			english3 = row[2]
		}
		if len(row) >= 4 {
			english4 = row[3]
		}
		if len(row) >= 5 {
			english5 = row[4]
		}

		// 初始化翻译结果切片
		translated := make([]string, 10)
		copy(translated, row)
		// 调用翻译 API 进行翻译
		if english1 != "" {
			wg.Add(1)
			go func() { translated[5] = translateText(client, english1) }()
		}
		if english2 != "" {
			wg.Add(1)
			go func() { translated[6] = translateText(client, english2) }()
		}
		if english3 != "" {
			wg.Add(1)
			go func() { translated[7] = translateText(client, english3) }()
		}
		if english4 != "" {
			wg.Add(1)
			go func() { translated[8] = translateText(client, english4) }()
		}
		if english5 != "" {
			wg.Add(1)
			go func() { translated[9] = translateText(client, english5) }()
		}
		wg.Wait()
		log.Printf("第%d行翻译结束\n", rowIdx-1)
		// 将翻译结果写入表格
		for colIdx, value := range translated {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
			xlsx.SetCellValue(sheet, cellName, value)
		}
	}
	// 保存结果到新的 xlsx 文件
	err = xlsx.SaveAs("output.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("全部完成")
}

func translateText(client *http.Client, text string) string {
	defer func() {
		wg.Done()
	}()
	var str string
	for i := 1; i < 10; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}
		apiURL := "https://fanyi.youdao.com/translate?&doctype=json&type=EN2ZH_CN&i=" + url.QueryEscape(text)
		response, err := client.Get(apiURL)
		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始")
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
				log.Println("出现错误，如果同id连续出现请联系我，重新开始")
				continue
			}
		}
		defer response.Body.Close()

		dataBytes, err := io.ReadAll(response.Body)
		if err != nil {
			log.Println("Failed to request translation API:", err)
			str = "翻译失败,解析错误"
			continue
		}
		defer response.Body.Close()
		result := string(dataBytes)
		tgt := regexp.MustCompile(`"tgt":"(.*?)"}`).FindAllStringSubmatch(result, -1)
		for i := range tgt {
			str += " " + tgt[i][1]
		}
		return str
	}
	return str
}
