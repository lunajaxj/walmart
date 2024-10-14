package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
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

var cgds []string
var username string
var passowrd string

func main() {
	fi, err := os.Open("采购单.txt")
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
	fii, err := os.Open("账号密码.txt")
	if err != nil {
		panic(err)
	}
	rr := bufio.NewReader(fii) // 创建 Reader
	var users []string
	for {
		lineB, err := rr.ReadBytes('\n')
		if len(lineB) > 0 {
			users = append(users, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}
	}
	username = users[0]
	passowrd = users[1]

	log.Printf("开始执行，有%d个任务需要获取", len(cgds))
	eccang()

	xlsx := excelize.NewFile()

	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"采购单号", "截图"}); err != nil {
		fmt.Println(err)
	}
	for _, v := range cgds {
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v, nil}); err != nil {
			fmt.Println(err)
		}
		if err := xlsx.AddPicture("Sheet1", "B"+strconv.Itoa(num), ".\\img\\"+v+".jpg", nil); err != nil {
			fmt.Println(err)
		}
		num++
	}

	xlsx.SaveAs("out.xlsx")

	log.Printf("全部完成")
}

var token string

func eccang() {
	url := "https://home.eccang.com/login"
	//配置
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
		//chromedp.Flag("headless", true),
		chromedp.Flag("headless", true),
		chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
		chromedp.Flag("ignore-certificate-errors", true),      //忽略错误
		chromedp.Flag("disable-web-security", true),           //禁用网络安全标志
		chromedp.Flag("disable-extensions", true),             //开启插件支持
		chromedp.Flag("disable-default-apps", true),
		//chromedp.Flag("disable-gpu", true), //开启gpu渲染
		chromedp.WindowSize(1920, 1080), // 设置浏览器分辨率（窗口大小）
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
	ctx, cancel = context.WithTimeout(ctx, 60000*time.Second)
	defer cancel()
	//var t1, t2, t3, t4, t6, t7 []*cdp.Node
	log.Printf("开始登录")
	err := chromedp.Run(ctx,
		//设置webdriver检测反爬
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
		//打开链接
		chromedp.Navigate(url),
		chromedp.Sleep(time.Second*1),
		chromedp.SendKeys("input#userName", username),
		chromedp.SendKeys("input#password", passowrd+kb.Enter),
		chromedp.Sleep(time.Second*3),
		chromedp.Navigate("https://home.eccang.com/entry/EZUKPA/ERP/iframe#/m_78382"),
		chromedp.Sleep(time.Second*5),
		chromedp.Navigate("https://ntg5xpz.eccang.com//purchase/orders/list?quick=-1"),
		info1(),
		chromedp.Sleep(time.Second*5),
		chromedp.Click(`a[class="selectUser homeFilter"]`),
		chromedp.Sleep(time.Second*5),
		forr(),

		//停止网页加载
		chromedp.Stop(),
	)
	if err != nil {
		log.Println(err)
	}

}

func info1() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		//刷新当前页面
		log.Println("登录成功，进入查询页面做准备工作")
		return
	}
}

//func forr() chromedp.ActionFunc {
//	return func(ctx context.Context) (err error) {
//		log.Println("准备就绪开始获取")
//		timeout, cancel := context.WithTimeout(ctx, 50*time.Second)
//		defer cancel()
//		for i := range cgds {
//			//time.Sleep(99999)
//			chromedp.SendKeys("div#module-container #searchCode", cgds[i]+kb.Enter).Do(timeout)
//			chromedp.Sleep(time.Second * 3)
//			screenShot(".\\img\\"+cgds[i]+".jpg", `#showCopyDiv + div#module-container #module-table`).Do(timeout)
//			chromedp.Sleep(time.Second * 3)
//			chromedp.SendKeys("div#module-container #searchCode", kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace).Do(timeout)
//			chromedp.Sleep(time.Second * 3)
//			log.Println(cgds[i] + " 完成")
//		}
//		return
//	}
//}

func forr() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		log.Println("准备就绪开始获取")
		for _, cgd := range cgds {
			actions := []chromedp.Action{
				chromedp.SendKeys("div#module-container #searchCode", cgd+kb.Enter),
				chromedp.Sleep(1 * time.Second), // 等待搜索结果
				chromedp.ActionFunc(func(ctx context.Context) error {
					// 这里执行屏幕截图的具体动作，需要确保screenShot函数返回一个chromedp.Action
					return screenShot(".\\img\\"+cgd+".jpg", "#showCopyDiv + div#module-container #module-table")(ctx)
				}),
				chromedp.Sleep(1 * time.Second), // 等待截图操作
				chromedp.SendKeys("div#module-container #searchCode", kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.ArrowRight+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace+kb.Backspace),
				chromedp.Sleep(1 * time.Second), // 等待搜索结果
			}

			// 使用context.Background()或其他适当的上下文
			if err := chromedp.Run(ctx, actions...); err != nil {
				log.Printf("%s 获取失败: %v", cgd, err)
				continue // 如果出错，跳过当前循环，继续下一个
			}

			log.Printf("%s 完成", cgd)
		}
		return
	}
}

func screenShot(name string, path string) chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {

		timeout, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		var code []byte
		if err = chromedp.Screenshot(path, &code).Do(timeout); err != nil {
			fmt.Println(err)
			return
		}
		return os.WriteFile(name, code, 0755)
	}
}
