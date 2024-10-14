package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/xuri/excelize/v2"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}
var ch = make(chan int, 8)

type DownloadedData struct {
	Link      string
	ImagePath string
	Title     string
	Price     string
}

func main() {

	//	// 创建一个新的Excel文件
	//	f := excelize.NewFile()
	//	// 定义图片的路径
	//	imagePath := "img/img1.PNG"
	//	// 图片将被插入的单元格
	//	cellLocation := "B1"
	//
	//	// 在Sheet1的B1位置插入图片
	//	if err := f.AddPicture("Sheet1", cellLocation, imagePath, nil); err != nil {
	//		log.Fatalf("插入图片失败: %v", err)
	//	}
	//
	//	// 设置文件保存的路径
	//	fileName := "output.xlsx"
	//	// 保存文件
	//	if err := f.SaveAs(fileName); err != nil {
	//		log.Fatalf("保存文件失败: %v", err)
	//	}
	//
	//	fmt.Println("Excel文件保存成功:", fileName)
	//}

	fi, err := os.Open("zouyueurl.txt")
	if err != nil {
		log.Fatalf("打开文件失败：%v", err)
	}
	defer fi.Close()

	r := bufio.NewReader(fi)
	xlsx := excelize.NewFile()

	sheetName := "Sheet1"
	xlsx.NewSheet(sheetName)
	// 插入标题行
	titleRow := []interface{}{"链接", "图片", "标题", "价格"}
	if err := xlsx.SetSheetRow(sheetName, "A1", &titleRow); err != nil {
		log.Fatalf("写入Excel标题行失败：%v", err)
	}

	var downloadedData []DownloadedData
	index := 2 // 从第二行开始写数据
	for {
		line, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		link := strings.TrimSpace(string(line))
		src, title, price := getCrawlerData(link)
		if src != "" && title != "" && price != "" {
			imagePath, err := downloadImage(src) // 下载图片
			if err != nil {
				log.Printf("下载图片失败: %v", err)
				continue
			}

			downloadedData = append(downloadedData, DownloadedData{
				Link:      link,
				ImagePath: imagePath,
				Title:     title,
				Price:     price,
			})

			// 立即处理Excel插入逻辑
			pictureCell, _ := excelize.CoordinatesToCellName(2, index)
			if err := xlsx.AddPicture("Sheet1", pictureCell, imagePath, nil); err != nil {
				log.Printf("插入图片到Excel失败: %v", err)
			} else {
				log.Printf("图片成功插入Excel: %s", imagePath)
			}

			row := []interface{}{link, imagePath, title, price}
			cell, _ := excelize.CoordinatesToCellName(1, index)
			if err := xlsx.SetSheetRow("Sheet1", cell, &row); err != nil {
				log.Printf("写入Excel失败：%v", err)
			}

			index++
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var IsC = false
var IsC2 = true

func init() {
	rand.Seed(time.Now().UnixNano()) // 初始化随机数生成器
}

// getCrawlerData 用于获取页面数据
func getCrawlerData(link string) (string, string, string) {
	// 在函数开头声明变量，确保它们在整个函数范围内都是可访问的
	var src, title, price string

	// 这个循环尝试最多3次获取数据
	for i := 0; i < 3; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http",
				Host:   "127.0.0.1:51599", // 使用你的代理地址和端口
			}),
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // 忽略 TLS 证书验证
		}

		client := &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second, // 设置请求超时时间
		}

		request, err := http.NewRequest("GET", link, nil)
		if err != nil {
			log.Printf("创建请求失败：%v, URL: %s", err, link)
			continue
		}
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")
		//request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		//request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		//request.Header.Set("Sec-Ch-Ua", `"Not)A;Brand";v="99", "Google Chrome";v="127", "Chromium";v="127"`)
		//request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		//request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		//request.Header.Set("Sec-Fetch-Dest", "document")
		//request.Header.Set("Sec-Fetch-Mode", "navigate")
		//request.Header.Set("Sec-Fetch-Site", "none")
		//request.Header.Set("Sec-Fetch-User", "?1")
		//request.Header.Set("Upgrade-Insecure-Requests", "1")
		//request.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
		//request.Header.Set("Cache-Control", "max-age=0")
		////request.Header.Set("Cookie", "_shopify_y=ceb48d4e-7ebb-4d10-a208-289388022913; secure_customer_sig=; localization=US; cart_currency=USD; _tracking_consent=%7B%22con%22%3A%7B%22CMP%22%3A%7B%22a%22%3A%22%22%2C%22m%22%3A%22%22%2C%22p%22%3A%22%22%2C%22s%22%3A%22%22%7D%7D%2C%22v%22%3A%222.1%22%2C%22region%22%3A%22HKKYT%22%2C%22reg%22%3A%22%22%7D; _cmp_a=%7B%22purposes%22%3A%7B%22a%22%3Atrue%2C%22p%22%3Atrue%2C%22m%22%3Atrue%2C%22t%22%3Atrue%7D%2C%22display_banner%22%3Afalse%2C%22sale_of_data_region%22%3Afalse%7D; _orig_referrer=; _landing_page=%2Fproducts%2F3031718573; receive-cookie-deprecation=1; _shopify_sa_p=; _shopify_s=3568014d-c775-410e-9488-93d767fe58cf; _shopify_sa_t=2024-08-29T02%3A16%3A24.519Z; keep_alive=69cc9141-e08b-4b8c-82f0-87de027cb138")
		//request.Header.Set("If-None-Match", `"cacheable:6e18f859ca9acaeed03211460f81d423"`)
		//request.Header.Set("Priority", "u=0, i")

		response, err := client.Do(request)
		if err != nil {
			log.Printf("请求失败：%v, URL: %s", err, link)
			continue
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			log.Printf("非预期的HTTP状态码：%d, URL: %s", response.StatusCode, link)
			continue
		}

		dataBytes, err := io.ReadAll(response.Body)
		if err != nil {
			log.Printf("读取响应体失败：%v, URL: %s", err, link)
			continue
		}

		result := string(dataBytes)
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
		if err != nil || doc == nil {
			log.Printf("解析HTML文档失败：%v, URL: %s", err, link)
			continue
		}

		result = string(dataBytes)
		doc, err = goquery.NewDocumentFromReader(strings.NewReader(result))
		if err != nil || doc == nil { // 确保doc不为nil
			log.Printf("创建文档失败或文档为空：%v", err)
			continue
		}
		//log.Println(result)
		//创建或打开文件
		//file, err := os.Create("111.txt")
		//if err != nil {
		//	fmt.Println("无法创建文件:", err)
		//}
		//// 确保在函数结束时关闭文件
		//defer file.Close()
		//
		//// 将文本写入文件
		//_, err = file.WriteString(result)
		//if err != nil {
		//	fmt.Println("无法写入文件:", err)
		//}
		//
		//fmt.Println("文本已成功写入文件")

		// 提取图片、标题和价格
		// 假设doc是已经加载了您的HTML文档的*goquery.Document实例
		found := false
		doc.Find("#Slider-Gallery-template--18348938756340__main li").EachWithBreak(func(i int, s *goquery.Selection) bool {
			if i == 0 { // 仅处理第一个li元素
				imgSrc, exists := s.Find("div.product__media.media.media--transparent img").Attr("src")
				if exists {
					src = imgSrc
					found = true
					return false // 停止遍历
				}
			}
			return true // 继续遍历
		})

		if !found {
			//log.Println(result)
			log.Printf("未找到图片src")
			continue // 尝试下一次循环
		}

		fmt.Println(src)

		title = doc.Find("#ProductInfo-template--18348938756340__main > div.product__title > h1").Text()
		price = doc.Find("#price-template--18348938756340__main > div > div > div.price__regular > span.price-item.price-item--regular").Text()

		if title != "" && price != "" {
			src = strings.TrimSpace(src)
			title = strings.TrimSpace(title)
			price = strings.TrimSpace(price)
			return src, title, price // 成功获取到所有数据后返回
		}

		// 如果本次循环未成功获取到所有数据，尝试下一次循环
	}
	// 如果3次尝试后仍未成功获取到所有数据，返回空字符串
	return "", "", ""
}

// downloadImage 下载图片并返回保存的本地文件路径

//	func downloadImage1(url string) (string, error) {
//		// 发起请求
//		resp, err := http.Get("https:" + url)
//		if err != nil {
//			return "", err
//		}
//		defer resp.Body.Close()
//
//		// 从URL获取图片文件名
//
//		dirPath := filepath.Join("img")
//		//确保img路径存在
//		if err := os.MkdirAll(dirPath, 0755); err != nil {
//			return "", err
//		}
//		imageCounter++
//		fileName := fmt.Sprintf("img%d.jpeg", imageCounter)
//		filePath := filepath.Join(dirPath, fileName)
//
//		// 创建文件
//		file, err := os.Create(filePath)
//		if err != nil {
//			return "", err
//		}
//		defer file.Close()
//
//		// 写入文件
//		_, err = io.Copy(file, resp.Body)
//		if err != nil {
//			return "", err
//		}
//
//		return filePath, nil
//	}

var imageCounter int

func downloadImage(imageUrl string) (string, error) {
	// 创建一个 HTTP 客户端，并配置代理、超时时间和其他设置
	transport := &http.Transport{
		Proxy: http.ProxyURL(&url.URL{
			Scheme: "http",
			Host:   "127.0.0.1:51599", // 使用你的代理地址和端口
		}),
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // 忽略 TLS 证书验证
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second, // 设置请求超时时间
	}

	// 发起请求
	resp, err := client.Get("https:" + imageUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 确保img路径存在
	dirPath := filepath.Join("img")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}

	// 根据 URL 生成文件名
	imageCounter++
	var fileName string
	if strings.Contains(imageUrl, "png") {
		fileName = fmt.Sprintf("img%d.png", imageCounter)
	} else {
		fileName = fmt.Sprintf("img%d.jpeg", imageCounter)
	}

	filePath := filepath.Join(dirPath, fileName)

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 写入文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
