package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/xuri/excelize/v2"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var ids []string
var nodes []*cdp.Node
var res []Wal

type Wal struct {
	url        string
	name       string
	sellerName string
}

var m = make(map[string]string)

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
func main() {
	fi, err := os.Open("url.txt")
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

	log.Printf("开始执行，有%d个任务需要获取", len(ids))
	eccang()

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

var token string

func eccang() {
	//配置
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
		chromedp.Flag("headless", false),
		chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
		//chromedp.Flag("disable-gpu", true), //开启gpu渲染
		chromedp.WindowSize(1920, 1080), // 设置浏览器分辨率（窗口大小）

		chromedp.NoFirstRun, //设置网站不是首次运行
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	// 初始化chromedp上下文，后续这个页面都使用这个上下文进行操作
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 60000*time.Second)
	defer cancel()
	log.Println("开始登录")
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
		log.Println(err)
	}

}

func urllo(d string) chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		//nodes = []*cdp.Node{}
		err = chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Nodes(`.//a[@class="productName"]`, &nodes), //缓一缓
		}) //执行爬虫任务
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(nodes[0].Attributes[3])
		if len(nodes) > 0 {
			fmt.Println(nodes[0].Attributes[3])
			m[d] = nodes[0].Attributes[3]
		}
		return
	}
}
func getSource(ctx context.Context) string {
	var source string
	if err := chromedp.Run(ctx, chromedp.EvaluateAsDevTools("document.documentElement.outerHTML", &source)); err != nil {
		return ""
	}
	return source
}
func forr() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		log.Println("准备就绪开始获取")

		for i := range ids {
			timeout, cancel := context.WithTimeout(ctx, 50*time.Second)
			defer cancel()
			log.Println(ids[i])
			chromedp.Navigate(ids[i]).Do(timeout)
			chromedp.Sleep(time.Second * 1)
			source := getSource(timeout)
			log.Println(source)
			if err != nil {
				return
			}
			if len(nodes) > 0 {
				m[ids[i]] = nodes[0].Attributes[3]
			}
			chromedp.Sleep(time.Second * 2000).Do(ctx)

			log.Println(ids[i] + " 完成")
		}

		return
	}
}
