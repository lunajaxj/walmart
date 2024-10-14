package main

import (
	"bufio"
	"context"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	keywords []string
	wg       = sync.WaitGroup{}
	ch       = make(chan int, 1)
	mu       = sync.Mutex{}
)

func main() {
	log.Println("开始执行...")
	// 读取输入文件
	fi, err := os.Open("email_20.txt")
	if err != nil {
		panic(err)
	}
	defer fi.Close()

	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 5 {
			keywords = append(keywords, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}
	}
	log.Println("有", len(keywords), "个链接任务")
	for _, v := range keywords {
		ch <- 1
		wg.Add(1)
		go chHoppingup1(v)
	}
	wg.Wait()
	log.Println("完成")
}

// 操作浏览器
func chHoppingup1(email string) {
	defer func() {
		wg.Done()
		<-ch
	}()

	// 创建一个新的浏览器实例
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false), // 若希望浏览器无头运行，改为true
	)

	allocator, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// 创建一个新的浏览器上下文
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()

	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	err := chromedp.Run(ctx, startup(email))
	if err != nil {
		log.Printf("任务执行失败: %v", err)
	}
}

// 控制器
func startup(email string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		if loginup(ctx, email, "abc123456") {
			log.Println("登录成功，开始上传任务")
			// 调用你的uploadup函数
		} else {
			log.Println("登录失败")
		}
		return nil
	}
}

func loginup(ctx context.Context, username, password string) bool {
	timeout, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	chromedp.Navigate("https://seller.walmart.com/items-and-inventory/bulk-updates?returnUrl=%2Fitems-and-inventory%2Fmanage-items").Do(timeout)
	for i := 0; i < 2; i++ {
		timeout0, cancel0 := context.WithTimeout(ctx, 30*time.Second)
		defer cancel0()

		err := chromedp.WaitVisible(`input[id="radioFulfillmentSF"]`).Do(timeout0)
		if err != nil {
			timeout02, cancel02 := context.WithTimeout(ctx, 20*time.Second)
			defer cancel02()
			err := chromedp.WaitVisible(`input[data-automation-id="uname"]`).Do(timeout02)
			if err != nil {
				log.Println("页面加载失败，重新开始加载")
				timeout01, cancel01 := context.WithTimeout(ctx, 10*time.Second)
				defer cancel01()
				chromedp.Stop().Do(timeout01)
				chromedp.Sleep(time.Second * 1).Do(timeout)
				chromedp.Navigate("https://seller.walmart.com/items-and-inventory/bulk-updates?returnUrl=%2Fitems-and-inventory%2Fmanage-items").Do(timeout01)
			} else {
				log.Println("未登录状态")
				break
			}
		} else {
			log.Println("已是登录状态")
			return true
		}
	}
	log.Println("开始登录")
	chromedp.Sleep(time.Second * 2).Do(timeout)
	chromedp.SendKeys(`input[data-automation-id="uname"]`, username).Do(timeout)
	chromedp.Sleep(time.Second * 2).Do(timeout)
	chromedp.SendKeys(`input[data-automation-id="pwd"]`, password+kb.Enter).Do(timeout)
	chromedp.Sleep(time.Second * 2).Do(timeout)
	timeout01, cancel01 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel01()
	chromedp.SendKeys(`input[data-automation-id="pwd"]`, kb.Enter).Do(timeout01)
	chromedp.Sleep(time.Second * 30).Do(timeout)
	chromedp.Navigate("https://seller.walmart.com/items-and-inventory/bulk-updates?returnUrl=%2Fitems-and-inventory%2Fmanage-items").Do(timeout)
	chromedp.Sleep(time.Second * 10).Do(timeout)
	timeout03, cancel03 := context.WithTimeout(ctx, 30*time.Second)
	defer cancel03()
	err := chromedp.WaitVisible(`input[id="radioFulfillmentSF"]`).Do(timeout03)
	if err != nil {
		log.Println("登录失败")
		chromedp.Stop().Do(ctx)
		return false
	}
	log.Println("登录成功")
	return true
}
