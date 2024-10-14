package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/xuri/excelize/v2"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var mkd_id []string

// 定义要抓取的数据的变量
var imageUrl, sales, title, rating, reviewCount, beforePrice, afterPrice, discount, free_or_not, shippingMethod, pngUrl, fullResult, fraction, cents, countryName string
var i string
var Talla []string
var Color []string
var category string
var pageSource string

func main() {
	fi, err := os.Open("mkd_id.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 0 {
			mkd_id = append(mkd_id, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}
	}

	log.Printf("开始执行，有%d个任务需要获取", len(mkd_id))
	xlsx := excelize.NewFile()
	// 从 "A2" 开始嵌入图片
	rowNum := 2

	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"ID", "主图", "类目", "销量", "标题", "评分", "评论数量", "折前售价", "折后售价", "折扣", "是否包邮", "Color", "Talla", "发货方式", "国家地区"}); err != nil {
		log.Fatal(err)
	}
	// 遍历每个ID，并调用mkd()函数
	for _, i := range mkd_id {
		err := mkd(i)
		if err != nil {
			log.Printf("Error processing ID %s: %v", i, err)
			continue
		}

		if strings.HasSuffix(imageUrl, ".webp") {
			pngUrl = strings.TrimSuffix(imageUrl, ".webp") + ".png"
			if err != nil {
				log.Printf("Failed to process image URL for ID %s: %v", i, err)
				continue
			}
		}

		imgDir := "img"
		if err := os.MkdirAll(imgDir, os.ModePerm); err != nil {
			log.Printf("Failed to create directory for ID %s: %v", i, err)
			continue
		}

		fileName1 := filepath.Base(pngUrl)
		localPath := filepath.Join(imgDir, fileName1)

		if err := downloadImage(pngUrl, localPath); err != nil {
			log.Printf("Failed to download image for ID %s: %v", i, err)
			continue
		}

		cell, _ := excelize.CoordinatesToCellName(2, rowNum)
		if err := xlsx.AddPicture("Sheet1", cell, localPath, nil); err != nil {
			log.Printf("Failed to add picture to Excel for ID %s: %v", i, err)
			continue
		}

		TallaStr := strings.Join(Talla, ", ")
		ColorStr := strings.Join(Color, ", ")

		if strings.Contains(shippingMethod, "full") {
			fullResult = "full"
		} else {
			fullResult = ""
		}

		parts := strings.SplitN(i, "/", 2)
		if len(parts) != 2 {
			log.Printf("Invalid format for ID %s", i)
			continue
		}

		row := []interface{}{parts[1], localPath, category, sales, title, rating, reviewCount, beforePrice, afterPrice, discount, free_or_not, ColorStr, TallaStr, fullResult, countryName}
		for colNum, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colNum+1, rowNum)
			if err := xlsx.SetCellValue("Sheet1", cell, val); err != nil {
				log.Printf("Failed to set cell value for ID %s: %v", i, err)
				continue
			}
		}

		rowNum++
	}
	// 保存Excel文件
	fileName := "out.xlsx"
	for fileNum := 1; exists(fileName); fileNum++ {
		fileName = "out(" + strconv.Itoa(fileNum) + ").xlsx"
	}
	xlsx.SaveAs(fileName)

	log.Println("完成")
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

func mkd(i string) error {
	skipAttempts := false

	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", false),
		chromedp.Flag("blink-settings", "imagesEnabled=true"),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.WindowSize(600, 600),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.NoFirstRun,
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	ctx, taskCancel := context.WithTimeout(ctx, 60*time.Second)
	defer taskCancel()

	log.Println("读取id，当前id为:", i)
	parts := strings.SplitN(i, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format: %s", i)
	}
	countryCode := parts[0]
	countryName = getCountryName(countryCode)
	log.Printf("Country code %s corresponds to %s", countryCode, countryName)

	url := buildURL(countryCode, i)
	log.Println(url)

	var pageSource string
	tasks := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.OuterHTML("html", &pageSource),
		chromedp.Sleep(time.Second * 5),
	}

	err := chromedp.Run(ctx, tasks,
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
	)

	if err != nil {
		return fmt.Errorf("执行初始化任务时出错: %v", err)
	}

	if strings.Contains(pageSource, "Parece que esta página no existe") {
		log.Println("id=", i, "页面无数据,网址有误")
		skipAttempts = true
	}
	if strings.Contains(pageSource, "Publicación pausada") {
		log.Println("id=", i, "尚未发布")
		skipAttempts = true
	}

	log.Println("任务id", i, "执行完成")

	if skipAttempts {
		log.Println("跳过尝试执行可能会失败的任务")
		return nil
	}

	attempts := []chromedp.Action{
		chromedp.AttributeValue(`#gallery > div > div > span:nth-child(3) > figure > img`, "src", &imageUrl, nil),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('ol[class*="andes-breadcrumb"] li a')).map(a => a.getAttribute('title')).join('>')`, &category),
		chromedp.Text(`//*[@id="header"]/div/div[1]/span`, &sales, chromedp.NodeVisible),
		chromedp.Text(`//*[@id="header"]/div/div[2]/h1`, &title, chromedp.NodeVisible),
		chromedp.Text(`//*[@id="header"]/div/div[3]/a/span[1]`, &rating, chromedp.NodeVisible),
		chromedp.Text(`//*[@id="header"]/div/div[3]/a/span[4]`, &reviewCount, chromedp.NodeVisible),
		chromedp.Text(`//*[@id="price"]/div/div[1]/span/s/span[2]`, &beforePrice, chromedp.NodeVisible),
		chromedp.Text(`//*[@id="price"]/div/div[1]/div/span[1]/span/span[2]`, &afterPrice, chromedp.NodeVisible),
		chromedp.Text(`//*[@id="price"]/div/div[1]/div[1]/span[2]/span`, &discount, chromedp.NodeVisible),
		chromedp.Evaluate(`(() => { const el = document.querySelector('#shipping_summary > div > div > p:nth-child(1)'); return el ? el.textContent : ''; })()`, &free_or_not),
		chromedp.Evaluate(`(() => { try { const elements = document.querySelectorAll('#:R9j1d9hm:-menu-popper > div ul li > div > div > span > span > span'); if (elements.length > 0) { return Array.from(elements).map(span => span.innerText); } else { return []; } } catch (error) { return []; } })()`, &Color),
		chromedp.Evaluate(`(() => { const elements = document.querySelectorAll('#variations > div > div > div:nth-child(2) a'); return elements.length > 0 ? Array.from(elements).map(a => a.title).join(", ") : ''; })()`, &Talla),
		chromedp.Evaluate(`document.querySelector('#fulfillment_information div div div div figure svg use').getAttribute('href');`, &shippingMethod),
	}

	for _, attempt := range attempts {
		if err := chromedp.Run(ctx, attempt); err != nil {
			log.Printf("执行任务时出错1: %v", err)
			continue
		}
	}

	// 释放资源
	taskCancel()
	cancel()
	allocCancel()

	return nil
}

func getCountryName(code string) string {
	// 根据国家代码返回国家名称
	switch code {
	case "br":
		return "巴西"
	case "mx":
		return "墨西哥"
	case "cl":
		return "智利"
	case "co":
		return "哥伦比亚"
	default:
		return "未知"
	}
}

// 加载Cookies
func loadCookies() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// 如果cookies临时文件不存在则直接跳过
		if _, _err := os.Stat("cookies.txt"); os.IsNotExist(_err) {
			return
		}

		// 如果存在则读取cookies的数据
		cookiesData, err := os.ReadFile("cookies.txt")
		replace := strings.Replace(string(cookiesData), "no_restriction", "None", -1)
		replace = strings.Replace(replace, " ", "", -1)
		replace = strings.Replace(replace, "\n", "", -1)
		if err != nil {
			return
		}
		// 反序列化
		cookiesParams := network.SetCookiesParams{}
		if err = cookiesParams.UnmarshalJSON([]byte("{\"cookies\":" + replace + "}")); err != nil {
			return
		}
		// 设置cookies
		return network.SetCookies(cookiesParams.Cookies).Do(ctx)
	}
}

func downloadImage(url, savePath string) error {
	if url == "" {
		return fmt.Errorf("URL is empty")
	}

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	// 确保保存路径所在目录存在
	os.MkdirAll(filepath.Dir(savePath), os.ModePerm)

	// 创建文件
	file, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将响应内容写入文件
	_, err = io.Copy(file, response.Body)
	return err
}

func buildURL(countryCode, id string) string {
	switch countryCode {
	case "br":
		return "https://produto.mercadolivre.com." + id
	case "mx", "co":
		return "https://articulo.mercadolibre.com." + id
	case "cl":
		return "https://articulo.mercadolibre." + id
	default:
		log.Fatalf("Unsupported country code: %s", countryCode)
		return ""
	}
}
