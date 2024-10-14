package main

import (
	"bufio"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"golang.org/x/net/context"
	"log"
	"os"
	"strings"
	"time"
)

type User struct {
	username string
	password string
}

var users []User

func main() {
	log.Println("沃尔玛重复登录开始...")
	fi, err := os.Open("账号密码.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 0 {
			split := strings.Split(strings.TrimSpace(string(lineB)), ",")
			users = append(users, User{
				username: split[0],
				password: split[1],
			})
		}
		if err != nil {
			break
		}
	}

	chHoppingup()
	log.Println("完成")

}

// 操作浏览器
func chHoppingup() {

	//配置
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		//chromedp.ProxyServer("socks5://"+proxy), //代理
		chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
		chromedp.Flag("headless", false),
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
	ctx, cancel = context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	err := chromedp.Run(ctx,
		//设置webdriver检测反爬
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
		loginz(),
		//停止网页加载
		chromedp.Stop(),
	)
	if err != nil {
		fmt.Println(err)
	}

}

// 检测是否登录，未登录就登录
func loginz() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		timeout, cancel := context.WithTimeout(ctx, 180*time.Second)
		defer cancel()
		chromedp.Navigate("https://login.account.wal-mart.com/authorize?responseType=code&clientId=66620dfd-1f3f-479b-8b9c-e11f36c5438b&scope=openId&redirectUri=https://seller.walmart.com/resource/login/sso/torbit&nonce=BZJBV2OYDJ&state=RQ6BFZZD47&clientType=seller").Do(timeout)
		for i := 0; i < 2; i++ {
			timeout02, cancel02 := context.WithTimeout(ctx, 20*time.Second)
			defer cancel02()
			err := chromedp.WaitVisible(`input[data-automation-id="uname"]`).Do(timeout02)
			if err != nil {
				log.Println("页面加载失败，重新开始加载")
				timeout01, cancel01 := context.WithTimeout(ctx, 10*time.Second)
				defer cancel01()
				chromedp.Stop().Do(timeout01)
				chromedp.Sleep(time.Second * 1).Do(timeout)
				chromedp.Navigate("https://login.account.wal-mart.com/authorize?responseType=code&clientId=66620dfd-1f3f-479b-8b9c-e11f36c5438b&scope=openId&redirectUri=https://seller.walmart.com/resource/login/sso/torbit&nonce=BZJBV2OYDJ&state=RQ6BFZZD47&clientType=seller").Do(timeout01)
			} else {
				log.Println("未登录状态")
				break
			}
		}
		log.Println("开始登录")
		for i := 0; i < len(users); i++ {

			chromedp.Sleep(time.Second * 2).Do(timeout)
			chromedp.SendKeys(`input[data-automation-id="uname"]`, users[i].username).Do(timeout)
			chromedp.Sleep(time.Second * 2).Do(timeout)
			chromedp.SendKeys(`input[data-automation-id="pwd"]`, users[i].password+kb.Enter).Do(timeout)
			chromedp.Sleep(time.Second * 2).Do(timeout)
			for ii := 0; ii <= 10; ii++ {
				log.Printf("账号: %s 第%d 次登录", users[i].username, ii)
				timeout01, cancel01 := context.WithTimeout(ctx, 2*time.Second)
				defer cancel01()
				chromedp.SendKeys(`input[data-automation-id="pwd"]`, kb.Enter).Do(timeout01)
				chromedp.Sleep(time.Second * 3).Do(timeout)
				timeout03, cancel03 := context.WithTimeout(ctx, 5*time.Second)
				defer cancel03()
				err = chromedp.WaitVisible(`.alert-box`).Do(timeout03)
			}
			chromedp.Navigate("https://login.account.wal-mart.com/authorize?responseType=code&clientId=66620dfd-1f3f-479b-8b9c-e11f36c5438b&scope=openId&redirectUri=https://seller.walmart.com/resource/login/sso/torbit&nonce=BZJBV2OYDJ&state=RQ6BFZZD47&clientType=seller").Do(timeout)
			chromedp.Sleep(time.Second * 2).Do(timeout)
		}
		return err
	}
}
