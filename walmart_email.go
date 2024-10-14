package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"github.com/chromedp/chromedp"
	"golang.org/x/net/context"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// 定义一个结构体用来存储电子邮件和密码
type Account struct {
	Email    string
	Password string
}

var ids []string
var wg = sync.WaitGroup{}
var ch = make(chan int, 10)

func main() {
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("邮箱密码.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader
	var accounts []Account
	for {
		line, err := r.ReadString('\n')
		if len(line) > 3 { //确保行里有内容
			parts := strings.Split(strings.TrimSpace(line), "|")
			if len(parts) == 2 { //确保可以分为邮箱和密码
				accounts = append(accounts, Account{Email: parts[0], Password: parts[1]})
			}
		}
		if err != nil {
			break
		}

	}

	for _, account := range accounts {
		ch <- 1
		wg.Add(1)
		go func(acc Account) {
			defer wg.Done()
			crawler(account.Email)
			<-ch
		}(account)
	}
	wg.Wait()
	log.Println("完成")

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

func crawler(email string) {
	//tr := &http.Transport{
	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//}
	//配置代理
	defer func() {
		wg.Done()
		<-ch
	}()
	for i := 0; i < 3; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}
		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
		//client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
		request, _ := http.NewRequest("GET", "https://www.walmart.com/account/login", nil)
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
		if IsC {
			request.Header.Set("Cookie", generateRandomString(10))
		}
		response, err := client.Do(request)
		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始：" + email)
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
				log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + email)
				continue
			}
		}
		result := ""
		if response.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(response.Body) // gzip解压缩
			if err != nil {
				log.Println("解析body错误，重新开始：" + email)
				continue
			}
			defer reader.Close()
			con, err := io.ReadAll(reader)
			if err != nil {
				log.Println("gzip解压错误，重新开始：" + email)
				continue
			}
			result = string(con)
		} else {
			dataBytes, err := io.ReadAll(response.Body)
			if err != nil {
				if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "Service Unavailable") {
					log.Println("代理IP无效，自动切换中")
					log.Println("连续出现代理IP无效请联系我，重新开始：" + email)
				} else {
					log.Println("错误信息：" + err.Error())
					log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + email)
				}
				continue
			}
			defer response.Body.Close()
			result = string(dataBytes)
		}
		if strings.Contains(result, "Enter your email and we’ll check for you") {
			log.Println("已登录")
			// 创建一个新的chromedp上下文
			ctx, cancel := chromedp.NewContext(context.Background())
			defer cancel()

			// 设置超时，防止操作卡住
			ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			// 执行任务：打开页面并输入电子邮件
			var pageHTML string
			err := chromedp.Run(ctx,
				chromedp.Navigate(`https://www.walmart.com/account/login`), // 替换为您的页面URL
				chromedp.WaitVisible(`#react-aria-1`, chromedp.ByID),       // 等待输入框可见
				chromedp.SendKeys(`#react-aria-1`, email, chromedp.ByID),   // 在输入框中输入电子邮件
				chromedp.Click(`#login-continue-button`, chromedp.NodeVisible),
				// 等待足够的时间确保页面已经刷新或跳转完成
				chromedp.Sleep(2*time.Second), // 可能需要根据页面实际加载时间调整
				// 获取当前页面的HTML
				chromedp.OuterHTML("html", &pageHTML, chromedp.ByQuery),
			)
			if err != nil {
				log.Fatalf("Failed to execute chromedp tasks: %v", err)
			}
			// 打印当前页面的HTML
			fmt.Println("Current page HTML:", pageHTML)
			return
		} else {
			fmt.Println("111111", result)
		}

		log.Println(Account{Email: email}, "完成")
		return
	}
}
