package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/nfnt/resize"
	"github.com/xuri/excelize/v2"
	"image"
	"image/jpeg"
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

// downloadImage 下载图片并返回image.Image对象
func downloadImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	return img, err
}

// resizeAndSaveImage 用于调整图片大小并保存
func resizeAndSaveImage(img image.Image, width, height uint, savePath string) error {
	// 如果宽度或高度为0，则使用图像的原始尺寸
	if width == 0 {
		width = uint(img.Bounds().Dx())
	}
	if height == 0 {
		height = uint(img.Bounds().Dy())
	}

	newImg := resize.Resize(width, height, img, resize.Lanczos3)
	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()
	return jpeg.Encode(out, newImg, nil)
}

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

// findJpegUrls 提取以.jpeg结尾的URL，并通过检查过滤掉包含'px'的URL
func findJpegUrls1(jsTexts []string) []string {
	// 正则表达式匹配以.jpeg结尾的URL
	urlPattern := `https?://[^\s]+?\.jpeg\b(?:\?[^\s]*)?`
	re := regexp.MustCompile(urlPattern)

	var filteredUrls []string
	// 遍历字符串数组，对每个元素使用正则表达式
	for _, jsText := range jsTexts {
		matches := re.FindAllString(jsText, -1)
		for _, match := range matches {
			// 检查URL中是否包含'px'
			if !strings.Contains(match, "px") {
				filteredUrls = append(filteredUrls, match)
			}
		}
	}
	return filteredUrls
}

var ch = make(chan int, 8)

func main() {
	log.Println("自动化脚本-walmart-id采集图片")
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

	xlsx := excelize.NewFile()
	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "图片1", "图片2", "图片3", "图片4", "图片5", "图片6", "图片7", "图片8", "图片9", "图片10", "图片11", "图片12"}); err != nil {
		log.Println(err)
	}
	for _, w := range res {
		rowData := []interface{}{w.id}
		for i, imgPath := range w.img {
			//log.Println("w.img=", w.img)
			//log.Println("imgpath=", imgPath)
			if _, err := os.Stat(imgPath); err != nil {
				log.Printf("File does not exist: %s\n", imgPath)
				continue
			}
			cell, _ := excelize.CoordinatesToCellName(i+2, num)
			rowData = append(rowData, "")
			// 在这里添加图片，确保使用正确的方法调用
			if err := xlsx.AddPicture("Sheet1", cell, imgPath, nil); err != nil {
				log.Printf("Failed to add picture at %s: %v\n", cell, err)
				continue
			}
		}
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
		//img := regexp.MustCompile(`645px"><div style="line-height:0" class="tc b--white ba bw1 b--blue mb2 overflow-hidden br3"><button class="pa0 ma0 bn bg-white b--white pointer" data-testid="item-page-vertical-carousel-hero-image-button"><div class="relative" data-testid="media-thumbnail" style="line-height:0"><img loading="lazy" srcset="([^,^"^?]+)`).FindAllStringSubmatch(result, -1)
		//img1 := regexp.MustCompile(`data-testid="media-thumbnail" style="line-height:0"><img loading="lazy" srcset="([^,^"^?]+)(?:\?[^"]*)?`).FindAllStringSubmatch(result, -1)
		//img := regexp.MustCompile(`data-testid="media-thumbnail" style="line-height:0"><img loading="lazy" srcset="([^,^"^?]+)(?:\?[^"]*)?`).FindAllStringSubmatch(result, -1)
		//jpegUrls := findJpegUrlsInText(result)
		//uniqueUrls := removeDuplicates(jpegUrls)
		//urls3 := findJpegUrls(uniqueUrls)
		//urls4 := findJpegUrls1(urls3)
		// 打印结果和添加到图片列表
		//if len(img1) > 0 {
		//	firstMatch := img1[0][1] // 获取第一个匹配值
		//	// 下载和调整特殊图片
		//	img, err := downloadImage(firstMatch)
		//	if err != nil {
		//		log.Printf("Failed to download %s: %v\n", firstMatch, err)
		//	} else {
		//		specialFilePath := fmt.Sprintf("img/%s_special.png", id)
		//		err = resizeAndSaveImage(img, 1500, 1500, specialFilePath)
		//		if err != nil {
		//			log.Printf("Failed to resize or save %s: %v\n", firstMatch, err)
		//		} else {
		//			log.Printf("Downloaded and resized special image %s to %s\n", firstMatch, specialFilePath)
		//			// 在列表开始处插入特殊图片
		//			wal.img = append([]string{specialFilePath}, wal.img...)
		//		}
		//	}
		//}
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

			for x, image := range images {
				img, err := downloadImage(image.URL)
				if err != nil {
					log.Printf("Failed to download %s: %v\n", image, err)
					continue
				}

				filePath := fmt.Sprintf("img/%s(%d).png", id, x)
				err = resizeAndSaveImage(img, 0, 0, filePath)
				if err != nil {
					log.Printf("Failed to resize or save %s: %v\n", image, err)
					continue
				}
				log.Printf("Downloaded and resized %s to %s\n", image, filePath)

				wal.img = append(wal.img, filePath) // 添加到图片列表
			}
			//fmt.Println("New URLs:", newUrl)

		} else {
			fmt.Println("No matches found")
		}

		//log.Println(urls)
		//wal.img = append(wal.img, urls...)

		log.Println(id, "完成")
		res = append(res, wal)
		return
	}

}
