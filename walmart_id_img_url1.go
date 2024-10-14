package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
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

	"github.com/PuerkitoBio/goquery"

	"github.com/xuri/excelize/v2"
)

var res []Wal

type Wal struct {
	img            []string
	id             string
	desc           string
	rules          string
	organized_desc string
}

var ids []string
var wg = sync.WaitGroup{}
var wgImg = sync.WaitGroup{}

// cleanText 清理文本，去除 Unicode 转义序列、HTML 标签和特殊字符
func cleanText(s string) string {
	// 去除 Unicode 转义序列
	s = unescapeUnicode(s)

	// 去除 HTML 标签
	s = removeHTMLTags(s)

	// 去除特殊字符
	s = removeSpecialCharacters(s)

	// 去除换行符和&nbsp;
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "&nbsp;", "")

	return s
}

// unescapeUnicode 将 Unicode 转义序列转换为对应的字符
func unescapeUnicode(s string) string {
	re := regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		var r rune
		fmt.Sscanf(match, "\\u%04x", &r)
		return string(r)
	})
}

// removeHTMLTags 去除字符串中的 HTML 标签
func removeHTMLTags(s string) string {
	re := regexp.MustCompile(`<[^>]+>`)
	return re.ReplaceAllString(s, "")
}

// removeSpecialCharacters 去除字符串中的特殊字符
func removeSpecialCharacters(s string) string {
	// 定义一个包含所有特殊字符的字符集
	specialCharacters := "🐕✅"

	// 使用 strings.Map 过滤掉特殊字符
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(specialCharacters, r) {
			return -1
		}
		return r
	}, s)
}

// downloadImage 下载图片并返回image.Image对象
//func downloadImage(url string) (image.Image, error) {
//	resp, err := http.Get(url)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	img, _, err := image.Decode(resp.Body)
//	return img, err
//}

// resizeAndSaveImage 用于调整图片大小并保存
//func resizeAndSaveImage(img image.Image, width, height uint, savePath string) error {
//	newImg := resize.Resize(width, height, img, resize.Lanczos3)
//	out, err := os.Create(savePath)
//	if err != nil {
//		return err
//	}
//	defer out.Close()
//	return jpeg.Encode(out, newImg, nil)
//}

// findJpegUrlsInText 直接从文本中提取以https开头并以.jpeg结尾的URL
func findJpegUrlsInText(jsText string) []string {
	// 正则表达式同时考虑了https协议和.jpeg文件扩展名
	urlPattern := `https?://(?:[a-zA-Z]|[0-9]|[$-_@.&+]|[!*\\(\\),]|(?:%[0-9a-fA-F][0-9a-fA-F]))+\.jpeg\b(?:\?[^\s]*)?`
	re := regexp.MustCompile(urlPattern)

	// 在文本中查找所有匹配项
	urls := re.FindAllString(jsText, -1)
	return urls
}

// normalizeUrl 用于标准化URL，移除查询参数，只保留协议、域名和路径。
func normalizeUrl(rawUrl string) (string, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return "", err
	}

	// 重构URL，只包含协议、主机和路径，忽略查询参数
	normalizedUrl := fmt.Sprintf("%s://%s%s", parsedUrl.Scheme, parsedUrl.Host, parsedUrl.Path)
	return normalizedUrl, nil
}

func removeDuplicates(urls []string) []string {
	uniqueUrls := make(map[string]bool)
	var result []string

	for _, url := range urls {
		normalizedUrl, err := normalizeUrl(url)
		if err == nil {
			if _, found := uniqueUrls[normalizedUrl]; !found {
				uniqueUrls[normalizedUrl] = true
				result = append(result, url) // 存储原始URL
			}
		}
	}
	return result
}

// findJpegUrls 提取以.jpeg结尾的URL，可能包含查询参数
// findJpegUrls 提取以.jpeg结尾的URL，可能包含查询参数或不包含任何附加参数
func findJpegUrls(jsTexts []string) []string {
	// 正则表达式匹配直接以.jpeg结尾的URL，可能后面跟查询参数或不带任何额外参数
	urlPattern := `https://i5\.walmartimages\.com/asr/[^\s]+?\.jpeg(\?[^\s]*)?`
	re := regexp.MustCompile(urlPattern)

	var urls []string
	// 遍历字符串数组，对每个元素使用正则表达式
	for _, jsText := range jsTexts {
		matches := re.FindAllString(jsText, -1)
		urls = append(urls, matches...)
	}
	return urls
}

var ch = make(chan int, 8)

func main() {
	log.Println("自动化脚本-walmart-id采集图片")
	log.Println("开始执行...")

	// // 读取 gpt_order.txt 文件
	// gpt_content, err := os.ReadFile("gpt_order_context.txt")
	// if err != nil {
	// 	fmt.Printf("Error reading file: %v\n", err)
	// 	return
	// }
	// description1 := string(gpt_content)

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
	log.Println("共有:", len(ids), "个id")
	for _, v := range ids {
		ch <- 1
		wg.Add(1)
		go crawler(v, "")
		break
	}
	wg.Wait()

	xlsx := excelize.NewFile()
	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "图片1", "图片2", "图片3", "图片4", "图片5", "图片6", "图片7", "图片8", "图片9", "图片10", "图片11", "图片12", "商品描述", "需求", "输出结果"}); err != nil {
		log.Println(err)
	}
	//log.Println("res=", res)
	for _, w := range res {
		rowData := []interface{}{w.id}
		for i2 := range w.img {
			rowData = append(rowData, w.img[i2])
		}
		//log.Println("w.img=", w.img)
		//log.Println("imgpath=", imgPath)
		//if _, err := os.Stat(imgPath); err != nil {
		//	log.Printf("File does not exist: %s\n", imgPath)
		//	continue
		//}
		rowData = append(rowData, "")
		// 在这里添加图片，确保使用正确的方法调用
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &rowData); err != nil {
			log.Println(err)
		}

		fmt.Println("len", len(rowData))
		for len(rowData) < 16 {
			rowData = append(rowData, "")
		}

		rowData[13] = w.desc           // 第10列，因为切片索引从0开始
		rowData[14] = w.rules          // 第11列
		rowData[15] = w.organized_desc // 第12列

		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &rowData); err != nil {
			log.Println(err)
		}
		num++
	}
	fileName := "out.xlsx"
	for fileNum := 1; exists(fileName); fileNum++ {
		fileName = "out(" + strconv.Itoa(fileNum) + ").xlsx"
	}
	xlsx.SaveAs(fileName)

	log.Println("全部完成")

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

//func newDescriptionByGpt(description string, rules string) (string, error) {
//	timeout := time.Minute * 3
//	ctx, cancel := context.WithTimeout(context.Background(), timeout)
//	defer cancel()
//	config := openai.DefaultConfig("sk-proj-THEHntGFqKYnt7EQu222T3BlbkFJvOq4uTZCV2tfAcGUbLIO")
//	// 配置代理
//	proxyStr := fmt.Sprintf("http://%s:%s@%s:%d", "a15079807913@foxmail.com", "dszzhernandezjudy720", "127.0.0.1", 51599)
//	proxyURL, err := url.Parse(proxyStr)
//	if err != nil {
//		return "", fmt.Errorf("failed to parse proxy URL: %v", err)
//	}
//	transport := &http.Transport{
//		Proxy: http.ProxyURL(proxyURL),
//	}
//	config.HTTPClient = &http.Client{
//		Transport: transport,
//	}
//	client := openai.NewClientWithConfig(config)
//	resp, err := client.CreateChatCompletion(
//		ctx,
//		openai.ChatCompletionRequest{
//			Model: openai.GPT3Dot5Turbo,
//			Messages: []openai.ChatCompletionMessage{
//				{
//					Role:    openai.ChatMessageRoleUser,
//					Content: rules + "\"\"\"" + description + "\"\"\"",
//				},
//			},
//		},
//	)
//
//	if err != nil {
//		return "", err
//	}
//	return string(resp.Choices[0].Message.Content), err
//}

func crawler(id string, description1 string) {

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

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
		if err != nil {
			log.Println("解析HTML错误：", err)
			return
		}

		queryStr := ""
		doc.Find("#maincontent > section > main > div.flex.undefined.flex-column.h-100 > div:nth-child(2) > div > div.w_aoqv.w_wRee.w_fdPt > div > div:nth-child(2) > div > div > section > div > div > a").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			if !strings.Contains(queryStr, text) {
				queryStr += text + " "
			}
		})
		log.Println("queryStr=", queryStr)

		// log.Println(result)
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

		re := regexp.MustCompile(`"imageInfo":\{"allImages":(\[[\w\W]*?\])`)

		matches := re.FindStringSubmatch(result)

		if len(matches) > 1 {
			listContent := matches[1]

			type Image struct {
				ID       string `json:"id"`
				URL      string `json:"url"`
				Zoomable bool   `json:"zoomable"`
			}
			var images []Image

			err := json.Unmarshal([]byte(listContent), &images)
			if err != nil {
				fmt.Println("Error unmarshalling JSON:", err)
				return
			}

			var newUrl []string

			for _, image := range images {
				//newUrl = append(newUrl, image.URL)
				wal.img = append(wal.img, image.URL)
			}

			fmt.Println("New URLs:", newUrl)

		} else {
			fmt.Println("No matches found")
		}

		log.Println(id, "完成")
		res = append(res, wal)
		return
	}

}
