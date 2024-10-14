package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"strings"
	"time"
)

var email string
var password string

func main() {
	// 打开文件
	file, err := os.Open("邮箱密码.txt")
	if err != nil {
		fmt.Println("打开文件时出错:", err)
		return
	}
	defer file.Close()

	// 创建文件的缓冲读取器
	scanner := bufio.NewScanner(file)

	// 逐行读取文件内容
	for scanner.Scan() {
		// 获取当前行的内容
		line := scanner.Text()
		// 使用|符号分割字符串，分别获取邮箱和密码
		parts := strings.Split(line, "|")
		if len(parts) == 2 {
			email = parts[0]
			password = parts[1]
			fmt.Printf("邮箱: %s, 密码: %s\n", email, password)
		} else {
			fmt.Println("格式不正确:", line)
		}
	}

	// 检查扫描过程中是否有错误
	if err := scanner.Err(); err != nil {
		fmt.Println("读取文件时出错:", err)
	}
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
	url := "https://www.walmart.com/account/login"
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
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 执行点击
			script := `(function() {
		Element.prototype._attachShadow = Element.prototype.attachShadow;
		Element.prototype.attachShadow = function () {
			return this._attachShadow({mode:'open'});
		};
	})();
    `
			return chromedp.Evaluate(script, nil).Do(ctx)
		}),

		chromedp.Location(&currentURL), // 获取当前URL
		//向下滚动一段距离使元素可见
		chromedp.Sleep(time.Second*3),
		chromedp.WaitVisible(`#react-aria-1`, chromedp.ByID),           // 等待元素变为可见
		chromedp.SendKeys(`#react-aria-1`, email, chromedp.ByID),       // 输入电子邮件地址
		chromedp.Click(`#login-continue-button`, chromedp.NodeVisible), //点击注册
		chromedp.Sleep(time.Second*15),
		// 等待页面元素加载
		chromedp.Evaluate(`const hostElement = document.querySelector('#px-captcha');
const shadowRoot = hostElement.shadowRoot;
const iframesInShadow = shadowRoot.querySelectorAll('iframe');

Array.from(iframesInShadow).forEach((iframe) => {
    const displayStyle = window.getComputedStyle(iframe).display;
    console.log(displayStyle); // 打印每个 iframe 的 display 属性
    if (displayStyle === 'block') {
        console.log("已选择block的iframe");
        // 确保 iframe 完全加载后执行操作
        
    }
});`, &result),

		//chromedp.WaitVisible(`svg > g > path:nth-child(4)`, chromedp.ByQuery),
		// 点击指定元素，这里需要根据实际情况调整选择器
		chromedp.Click(`svg>g`, chromedp.ByQuery),
		// 等待特定元素确保页面加载完成
		chromedp.Sleep(time.Second*10),
		chromedp.WaitVisible(`#sign-in-widget > div.sign-in-widget > div`, chromedp.ByQuery),
		// 点击指定元素
		chromedp.Click(`#px-captcha`, chromedp.ByQuery),
		chromedp.Sleep(time.Second*600),
		//chromedp.Text(`body > div > div:nth-child(8) > div.col-xs-12.col-md-2 > div.mobile_hide`, &phoneNum, chromedp.NodeVisible),
		//chromedp.Text(`body > div > div:nth-child(8) > div.col-xs-0.col-md-2.mobile_hide`, &currentTime, chromedp.NodeVisible),
		//chromedp.Text(`body > div > div:nth-child(8) > div.col-xs-12.col-md-8`, &result, chromedp.NodeVisible),
	)
	if err != nil {
		return fmt.Errorf("导航失败: %v", err)
	}
	//根据URL是否包含特定字段执行不同的逻辑
	fmt.Println("当前手机号:", phoneNum)
	fmt.Println("当前时间:", currentTime)
	fmt.Println("提取验证码:", result)
	return nil
}
