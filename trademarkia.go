package main

import (
	"bufio"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/xuri/excelize/v2"
	"golang.org/x/net/context"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var res = make(map[string][]map[string]string)
var mu sync.Mutex
var file *excelize.File
var num int

type Trademark struct {
	keyword                    string //关键词
	Last_Applicant_Owned_by    string //左侧上方地址Last Applicant/ Owned by
	Last_Applicant_Owned_by2   string //左侧上方地址Last Applicant/ Owned by
	Serial_Number              string //Serial Number
	Registration_Number        string //Registration Number
	Correspondent_Address      string //Correspondent Address
	Filing_Basis               string //Filing_Basis
	category                   string //category
	Classification_Information string //Classification Information

}

var cgds []string
var mkey []interface{}

func main() {
	// 创建日志文件
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// 日志同时输出到文件和控制台
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	// 打开或创建Excel文件
	var fileName = "out.xlsx"
	var fileErr error
	if exists(fileName) {
		file, fileErr = excelize.OpenFile(fileName)
		if fileErr != nil {
			log.Fatalf("Failed to open existing Excel file: %v", fileErr)
		}
		sheetName := file.GetSheetName(0)
		rows, err := file.GetRows(sheetName)
		if err != nil {
			log.Fatalf("Failed to get rows from existing Excel file: %v", err)
		}
		num = len(rows) + 1
	} else {
		file = excelize.NewFile()
		if err := file.SetSheetRow("Sheet1", "A1", &[]interface{}{"关键词", "Last_Applicant_Owned_by", "Last_Applicant_Owned_by2", "Serial_Number", "Registration_Number", "Correspondent_Address", "Filing_Basis", "category", "Classification_Information"}); err != nil {
			log.Println("Failed to set header row:", err)
		}
		num = 2
	}
	fi, err := os.Open("关键词.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 0 {
			cgds = append(cgds, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}

	}

	log.Printf("开始执行，有%d个关键词需要获取", len(cgds))
	uspto()

	log.Printf("全部完成")
}

// 文件是否存在
func exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	return !os.IsNotExist(err)
}
func uspto() {
	//配置
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
		chromedp.Flag("headless", true),
		chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
		chromedp.Flag("ignore-certificate-errors", true),      //忽略错误
		chromedp.Flag("disable-web-security", true),           //禁用网络安全标志
		chromedp.Flag("disable-extensions", true),             //开启插件支持
		chromedp.Flag("disable-default-apps", true),
		//chromedp.Flag("disable-gpu", true), //开启gpu渲染
		chromedp.WindowSize(700, 700), // 设置浏览器分辨率（窗口大小）
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.NoFirstRun, //设置网站不是首次运行
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"), //设置UserAgent
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	// 初始化chromedp上下文，后续这个页面都使用这个上下文进行操作
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 99999*time.Minute)
	defer cancel()
	err := chromedp.Run(ctx,
		//设置webdriver检测反爬
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
		forr(),
		//停止网页加载
		chromedp.Stop(),
	)
	if err != nil {
		log.Println(11, err)
	}

}

func forr() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		for i := range cgds {
			timeout, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()
			// 启动监听新页面的上下文

			log.Println("开始获取", cgds[i])
			chromedp.Navigate(fmt.Sprintf("https://www.trademarkia.com/search/trademarks?query=%s&reset_page=true&country=us", cgds[i])).Do(timeout)
			chromedp.Sleep(time.Second * 2).Do(timeout)
			//timeout20, cancel20 := context.WithTimeout(ctx, 2*time.Second)
			//defer cancel20()
			var elementContent string
			var elementContent_number string
			var elementContent_status string
			// 等待搜索结果
			// 等待搜索结果并获取第一个元素内容
			err := chromedp.Text(`tbody a td:nth-of-type(2) p:nth-of-type(1)`, &elementContent).Do(timeout)
			if err != nil || elementContent == "" {
				log.Printf("没有获取到数据，跳过当前关键词: %s", cgds[i])
				continue // 跳过此关键词，继续处理下一个
			}
			chromedp.Text(`tbody a td:nth-of-type(2) p:nth-of-type(3)`, &elementContent_number).Do(timeout)    // 获取元素内容
			chromedp.Text(`tbody a td:nth-of-type(3) span:nth-of-type(1)`, &elementContent_status).Do(timeout) // 获取元素内容
			// 输出获取到的元素内容
			fmt.Println("mark的值为:", elementContent)
			fmt.Println("mark的number值为:", elementContent_number)
			fmt.Println("mark的status值为:", elementContent_status)
			// 如果 elementContent 与 cgds[i] 忽略大小写相等，点击第一个 class="col ps-0 ps-sm-1" 的元素
			if strings.EqualFold(elementContent, cgds[i]) && elementContent_status != "Dead/Cancelled" {
				log.Printf("elementContent (%s) 忽略大小写等于 cgds[%d] (%s), 准备跳转到%s", elementContent, i, cgds[i], fmt.Sprintf("https://www.trademarkia.com/%s-%s", elementContent, elementContent_number))
				chromedp.Navigate(fmt.Sprintf("https://www.trademarkia.com/%s-%s", elementContent, elementContent_number)).Do(timeout)
				// 执行点击操作
			} else {
				log.Printf("elementContent (%s) 忽略大小写不等于 cgds[%d] (%s)或者elementContent_status(%s)不等于Dead/Cancelled, 不再执行后续操作", elementContent, i, cgds[i])
				continue
			}
			log.Println("获取页面元素...")
			// 适当增加等待时间以确保页面跳转完成（例如 5 秒）
			chromedp.Sleep(time.Second * 2).Do(timeout)

			// 检查当前页面的 URL，确保页面已经跳转
			var currentURL string
			err = chromedp.Run(timeout,
				chromedp.Location(&currentURL), // 获取当前页面的 URL
			)
			if err != nil {
				// 如果错误是 context deadline exceeded，记录日志并跳过当前处理
				if strings.Contains(err.Error(), "context deadline exceeded") {
					log.Printf("获取当前页面 URL 时出错: %v, 跳过当前关键词", err)
					continue // 跳过当前关键词，继续处理下一个
				}

				// 如果是其他类型的错误，记录并继续
				log.Printf("获取当前页面 URL 时出错: %v", err)
				continue
			}
			log.Printf("当前页面的 URL: %s", currentURL)

			trademark := Trademark{}
			trademark.keyword = cgds[i]
			//Last_Applicant_Owned_by
			var Last_Applicant_Owned_by1 string
			chromedp.Text(`div.flex.p-5>div>div>div>div:nth-child(2)>div>div>div>a>p`, &Last_Applicant_Owned_by1, chromedp.NodeVisible).Do(timeout)
			Last_Applicant_Owned_by1 = strings.TrimSpace(strings.Replace(Last_Applicant_Owned_by1, "Last Applicant/ Owned by", "", -1))
			trademark.Last_Applicant_Owned_by = Last_Applicant_Owned_by1
			//Last_Applicant_Owned_by
			var Last_Applicant_Owned_by2 string
			chromedp.Text(`div.flex.p-5>div>div>div>div:nth-child(2)>div>div>div:nth-child(2)`, &Last_Applicant_Owned_by2, chromedp.NodeVisible).Do(timeout)
			Last_Applicant_Owned_by2 = strings.TrimSpace(strings.Replace(Last_Applicant_Owned_by2, "Last Applicant/ Owned by", "", -1))
			trademark.Last_Applicant_Owned_by2 = Last_Applicant_Owned_by2

			//Serial_Number
			var Serial_Number1 string
			chromedp.Text(`div.flex.p-5>div>div>div>div:nth-child(2)>div:nth-child(2)`, &Serial_Number1, chromedp.NodeVisible).Do(timeout)
			Serial_Number1 = strings.TrimSpace(strings.Replace(Serial_Number1, "Serial Number", "", -1))

			trademark.Serial_Number = Serial_Number1

			//Registration_Number
			var Registration_Number1 string
			chromedp.Text(`div.flex.p-5>div>div>div>div:nth-child(2)>div:nth-child(3)`, &Registration_Number1, chromedp.NodeVisible).Do(timeout)
			Registration_Number1 = strings.TrimSpace(strings.Replace(Registration_Number1, "Registration Number", "", -1))

			trademark.Registration_Number = Registration_Number1

			//Correspondent Address
			var Correspondent_Address1 string
			chromedp.Text(`div.flex.p-5>div>div>div>div:nth-child(2)>div:nth-child(4)`, &Correspondent_Address1, chromedp.NodeVisible).Do(timeout)
			Correspondent_Address1 = strings.TrimSpace(strings.Replace(Correspondent_Address1, "Correspondent Address", "", -1))

			trademark.Correspondent_Address = Correspondent_Address1

			//Filing Basis
			var Filing_Basis1 string
			chromedp.Text(`div.flex.p-5>div>div>div>div:nth-child(2)>div:nth-child(5)`, &Filing_Basis1, chromedp.NodeVisible).Do(timeout)
			Filing_Basis1 = strings.TrimSpace(strings.Replace(Filing_Basis1, "Filing Basis", "", -1))
			trademark.Filing_Basis = Filing_Basis1

			//品牌category
			var category1 string
			chromedp.Text(`div.flex.flex-col.space-y-5>div`, &category1, chromedp.NodeVisible).Do(timeout)
			trademark.category = category1

			//Classification Information
			var Classification_Information1 string
			chromedp.Text(`div.flex.flex-col.space-y-5>div:nth-child(2)>div`, &Classification_Information1, chromedp.NodeVisible).Do(timeout)
			trademark.Classification_Information = Classification_Information1

			appendToExcel(trademark)
			log.Println(cgds[i] + " 完成")
		}

		return
	}
}
func appendToExcel(trademark Trademark) {
	mu.Lock()
	defer mu.Unlock()
	row := []interface{}{trademark.keyword, trademark.Last_Applicant_Owned_by, trademark.Last_Applicant_Owned_by2, trademark.Serial_Number, trademark.Registration_Number, trademark.Correspondent_Address, trademark.Filing_Basis, trademark.category, trademark.Classification_Information}

	if err := file.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &row); err != nil {
		log.Println("Failed to set sheet row:", err)
	}
	num++

	fileName := "out.xlsx"
	if err := file.SaveAs(fileName); err != nil {
		log.Println("Failed to save Excel file:", err)
	}
}
