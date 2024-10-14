package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"time"
)

func main() {
	ctx, cancel := initializeChromedp()
	defer cancel() //确保程序结束时关闭浏览器
	if err := register(ctx); err != nil {
		log.Fatal(err)
	}
}
func initializeChromedp() (context.Context, context.CancelFunc) {
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
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	return ctx, cancel
}

func register(ctx context.Context) error {
	url := "https://yunduanxin.net/US-Phone-Number/"
	//设置超时时间
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	var phoneNum, currentTime, result, currentURL string
	log.Printf("开始登录")
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
		chromedp.Navigate(url),
		// 增加等待页面加载的逻辑
		chromedp.WaitVisible(`body`, chromedp.ByQuery), // 确保主体内容已加载
		//chromedp.Click(`body > nav > div > div > ul.nav.navbar-nav.mr-auto > li:nth-child(3) > a`),
		chromedp.Location(&currentURL), // 获取当前URL
		//chromedp.WaitVisible(`#dismiss-button > div > svg`, chromedp.ByQuery), // 确保SVG元素可见
		//chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(`window.scrollBy(0, window.innerHeight * 0.5);`, nil),
		//向下滚动一段距离使元素可见
		chromedp.Sleep(time.Second*3),
		chromedp.Click(`#content > div > div.row > div:nth-child(2) > div.row.mt-auto > div > a`, chromedp.NodeVisible), //点击查看短信
		chromedp.Sleep(time.Second*3),
		chromedp.Click(`body > div > div:nth-child(4) > button > b`, chromedp.NodeVisible), //点击查看短信
		chromedp.Sleep(time.Second*3),
		chromedp.Text(`body > div > div:nth-child(8) > div.col-xs-12.col-md-2 > div.mobile_hide`, &phoneNum, chromedp.NodeVisible),
		chromedp.Text(`body > div > div:nth-child(8) > div.col-xs-0.col-md-2.mobile_hide`, &currentTime, chromedp.NodeVisible),
		chromedp.Text(`body > div > div:nth-child(8) > div.col-xs-12.col-md-8`, &result, chromedp.NodeVisible),
	)
	if err != nil {
		return fmt.Errorf("导航失败: %v", err)
	}
	//根据URL是否包含特定字段执行不同的逻辑
	// 打开文件，如果文件不存在则创建，如果存在则在文件末尾添加内容（追加模式）
	file, err := os.OpenFile("验证码.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("打开文件时出错:", err)
		return nil
	}
	defer file.Close()

	// 写入信息到文件
	_, err = file.WriteString(fmt.Sprintf("当前手机号: %s\n当前时间: %s\n提取验证码: %s\n", phoneNum, currentTime, result))
	if err != nil {
		fmt.Println("写入文件时出错:", err)
		return nil
	}

	fmt.Println("信息已成功写入文件")
	return nil
}
