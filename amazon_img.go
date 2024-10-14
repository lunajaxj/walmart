package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/nfnt/resize"
	"github.com/xuri/excelize/v2"
	_ "golang.org/x/image/tiff" // 支持TIFF
	"image"
	_ "image/gif" // 支持GIF
	"image/jpeg"
	_ "image/jpeg" // 支持JPEG
	_ "image/png"  // 支持PNG
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

var ch = make(chan int, 8)

// downloadImage 用于从指定的URL下载图片并保存到指定的本地路径
//
//	func downloadImage(url, savePath string) error {
//		// 发起HTTP GET请求
//		resp, err := http.Get(url)
//		if err != nil {
//			return err
//		}
//		defer resp.Body.Close()
//
//		// 确保目标目录存在
//		if err := os.MkdirAll("img", os.ModePerm); err != nil {
//			return err
//		}
//
//		// 创建文件用于保存下载的数据
//		out, err := os.Create(savePath)
//		if err != nil {
//			return err
//		}
//		defer out.Close()
//
//		// 将数据从HTTP响应复制到文件
//		_, err = io.Copy(out, resp.Body)
//		return err
//	}
//
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

// resizeAndSaveImage 调整图片大小并保存
func resizeAndSaveImage(img image.Image, width, height uint, savePath string) error {
	// 调整图片大小
	newImg := resize.Resize(width, height, img, resize.Lanczos3)

	// 创建文件
	file, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 保存文件
	return jpeg.Encode(file, newImg, nil)
}

func main() {
	log.Println("amazon采集图片")
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("amazon_id.txt")
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

type Image struct {
	colorImages string `json:"colorImages"` // json标签指定了JSON键与结构体字段的映射
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

		request, _ := http.NewRequest("GET", "https://www.amazon.com/dp/"+id, nil)
		//log.Println(request.URL)
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
		// 打开文件，如果文件不存在则创建，如果存在则清空内容后写入（os.O_TRUNC）
		//file, err := os.OpenFile("amazon_result_page.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		//if err != nil {
		//	log.Fatalf("打开文件失败: %v", err)
		//}
		//defer file.Close()
		//
		//// 写入内容到文件
		//_, err = file.WriteString(result)
		//if err != nil {
		//	log.Fatalf("写入文件失败: %v", err)
		//}
		//
		//log.Println("内容已成功写入到文件")
		// 打印出结果的一部分进行检查
		//fmt.Println(result[:1000]) // 打印前1000个字符，检查内容

		wal := Wal{}
		wal.id = id
		if strings.Contains(result, "This page could not be found.") {
			res = append(res, wal)
			log.Println("id:" + id + "商品不存在")
			return
		}

		fk := regexp.MustCompile("Enter the characters you see below").FindAllStringSubmatch(result, -1)

		if len(fk) > 0 {
			log.Println("id:" + id + " 被风控,更换IP继续")
			IsC = !isc
			continue
		}

		//log.Println(findJpgLinks(result))
		//log.Println(result)

		//// 使用正则表达式匹配所需的JS对象

		doc, err := htmlquery.Parse(strings.NewReader(result))
		if err != nil {
			fmt.Println("Error parsing HTML:", err)
			return
		}

		// 使用XPath查找所有<p>标签
		nodes, err := htmlquery.QueryAll(doc, `//*[@id="imageBlock_feature_div"]/script[1]`)
		if err != nil {
			fmt.Println("Error finding nodes:", err)
			return
		}

		// 遍历所有找到的<p>节点并打印内容
		for _, node := range nodes {
			//fmt.Printf("Paragraph %d: %s\n", i+1, htmlquery.InnerText(node))
			reColorImages := regexp.MustCompile(`'colorImages':\s*(\{[\s\S]*?\})\s*,\s*'`)
			img1 := reColorImages.FindAllStringSubmatch(htmlquery.InnerText(node), -1)
			log.Println("searched result")
			//log.Println(img1)
			//var image Image
			jsonStr := fmt.Sprintf(`{%s}`, img1[0][0])
			standardJSON := strings.Replace(jsonStr, `'`, `"`, -1)
			standardJSON = strings.Replace(standardJSON, `\"`, `\\"`, -1) // 转义数据中的真实双引号
			standardJSON = strings.Replace(standardJSON, `\\"'`, `\'`, -1)
			//log.Println(standardJSON)
			re := regexp.MustCompile(`"hiRes":"(https?://[^"]+)"`)
			matches := re.FindAllStringSubmatch(standardJSON, -1)
			if err := os.MkdirAll("img", os.ModePerm); err != nil {
				log.Fatal(err)
			}
			// 打印所有找到的URL
			for x, match := range matches {
				img, err := downloadImage(match[1])
				if err != nil {
					log.Printf("Failed to download %s: %v\n", match[1], err)
					continue
				}
				//fmt.Println("match[1]=", match[1]) // 第一个子匹配是URL
				filePath := fmt.Sprintf("img/%s(%d).png", id, x)
				err = resizeAndSaveImage(img, 1000, 1000, filePath)
				if err != nil {
					log.Printf("Failed to resize or save %s: %v\n", match[1], err)
				} else {
					log.Printf("Downloaded and resized %s to %s\n", match[1], filePath)
				}
				wal.img = append(wal.img, filePath)
			}

		}

		log.Println(id, "完成")
		res = append(res, wal)
		return
	}

}
