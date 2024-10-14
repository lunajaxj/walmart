package main

import (
	"bufio"
	"context"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strings"
	"time"
)

var cgds []string
var username string
var password string

func main() {
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
	password = users[1]
	eccang()
}

func eccang() {
	url := "https://home.eccang.com/login"
	//配置
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
		//chromedp.Flag("headless", true),
		chromedp.Flag("headless", false),
		chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
		chromedp.Flag("ignore-certificate-errors", true),      //忽略错误
		chromedp.Flag("disable-web-security", true),           //禁用网络安全标志
		chromedp.Flag("disable-extensions", true),             //开启插件支持
		chromedp.Flag("disable-default-apps", true),
		//chromedp.Flag("disable-gpu", true), //开启gpu渲染
		chromedp.WindowSize(1920, 1080), // 设置浏览器分辨率（窗口大小）
		chromedp.Flag("hide-scrollbars", false),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.NoFirstRun, //设置网站不是首次运行
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"), //设置UserAgent
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	//defer cancel()
	// 初始化chromedp上下文，后续这个页面都使用这个上下文进行操作
	ctx, cancel := chromedp.NewContext(
		allocCtx,
		chromedp.WithLogf(log.Printf),
	)
	//defer cancel()
	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	//defer cancel()
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
		chromedp.SendKeys("input#password", password+kb.Enter),
		chromedp.Sleep(time.Second*3),
		chromedp.Navigate("https://home.eccang.com/entry/EZUKPA/ERP/iframe#/m_45668"),
		chromedp.Sleep(time.Second*5),
		info1(),
		chromedp.Click(`body > div:nth-child(59) > div > button`),
		chromedp.Click(`#search-module-baseSearch > div > input.baseBtn.submitToSearch`),

		chromedp.Sleep(time.Second*5),
		//点击搜索按钮
		chromedp.Click(`#module-table > div.opration_area > div:nth-child(4)`),
		chromedp.Sleep(time.Second*3),
		//chromedp.MouseClickXY(900, 600),
		//chromedp.Sleep(time.Second*5),
		//chromedp.KeyEvent(kb.ArrowDown),
		//chromedp.Sleep(time.Second*1),
		//chromedp.KeyEvent(kb.ArrowDown),
		chromedp.Sleep(time.Second*2),
		chromedp.Click(`body > div:nth-child(60) > div.ui-dialog-buttonpane.ui-widget-content.ui-helper-clearfix > div > button:nth-child(1) > span`),
		//chromedp.MouseClickXY(850, 670),
		//chromedp.MouseClickXY(1060, 750),
		chromedp.Sleep(time.Second*99999),
		//停止网页加载
		//chromedp.Stop(),
	)

	if err != nil {
		log.Fatal(err)
		cancel()
	} else {
		chromedp.Sleep(time.Second * 99999)
	}
	//err = chromedp.Cancel(ctx)
	//if err != nil {
	//	log.Fatal(err)
	//}
}
func info1() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		//刷新当前页面
		log.Println("登录成功，进入库存页面做准备工作")
		return
	}
}
