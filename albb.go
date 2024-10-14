package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	_ "github.com/tebeka/selenium"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

var cgds []string
var nodes []*cdp.Node
var username string
var password string

var status string            //发货状态
var logistics_company string //物流公司
var tracking_number string   //运单号码
var latest_progress string   //最新进度

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
	//fii, err := os.Open("账号密码albb.txt")
	//if err != nil {
	//	panic(err)
	//}
	//rr := bufio.NewReader(fii) // 创建 Reader
	//var users []string
	//for {
	//	lineB, err := rr.ReadBytes('\n')
	//	if len(lineB) > 0 {
	//		users = append(users, strings.TrimSpace(string(lineB)))
	//	}
	//	if err != nil {
	//		break
	//	}
	//}
	//username = users[0]
	//password = users[1]
	log.Printf("开始执行，有%d个任务需要获取", len(cgds))
	eccang()
	log.Printf("全部完成")
}

//var token string

func eccang() {
	//const (
	//	//seleniumPath    = "path/to/selenium-server-standalone.jar" // 如果使用Selenium Standalone Server
	//	chromeDriverPath = "path/to/chromedriver" // ChromeDriver的路径
	//	port             = 1222                   // 你想要控制的端口
	//)
	//opts := []selenium.ServiceOption{
	//	selenium.ChromeDriver(chromeDriverPath), // 指定ChromeDriver位置
	//}
	//
	//// 启动chromedriver，指定其端口号
	//service, err := selenium.NewChromeDriverService(chromeDriverPath, port, opts...)
	//if err != nil {
	//	panic(err) // 替换为更优雅的错误处理
	//}
	//defer service.Stop()
	//
	//caps := selenium.Capabilities{"browserName": "chrome"}
	//wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	//if err != nil {
	//	panic(err) // 替换为更优雅的错误处理
	//}
	//defer wd.Quit()
	//配置
	//options := append(chromedp.DefaultExecAllocatorOptions[:],
	//	chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
	//	chromedp.Flag("headless", false),
	//	chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
	//	chromedp.Flag("ignore-certificate-errors", true),      //忽略错误
	//	chromedp.Flag("disable-web-security", true),           //禁用网络安全标志
	//	chromedp.Flag("disable-extensions", true),             //开启插件支持
	//	chromedp.Flag("disable-default-apps", true),
	//	//chromedp.Flag("disable-gpu", true), //开启gpu渲染
	//	chromedp.WindowSize(1920, 1080), // 设置浏览器分辨率（窗口大小）
	//	chromedp.Flag("hide-scrollbars", true),
	//	chromedp.Flag("mute-audio", true),
	//	chromedp.Flag("no-sandbox", true),
	//	//chromedp.Flag("incognito", true),
	//	chromedp.Flag("no-default-browser-check", true),
	//	chromedp.NoFirstRun, //设置网站不是首次运行
	//	chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"), //设置UserAgent
	//	//chromedp.Flag("remote-debugging-port", "1222"),
	//)
	// 创建远程分配器
	ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), "http://localhost:1222")
	defer cancel()
	// 初始化chromedp上下文
	ctx, cancel = chromedp.NewContext(ctx)
	//defer cancel()
	//ctx, cancel = chromedp.NewContext(ctx, chromedp.WithTargetID(1))
	//defer cancel()

	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	//defer cancel()

	//allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	//defer cancel()
	//// 初始化chromedp上下文，后续这个页面都使用这个上下文进行操作
	//ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	//defer cancel()
	//// 设置超时时间
	//ctx, cancel = context.WithTimeout(ctx, 6000000*time.Second)
	//defer cancel()
	log.Println("开始登录")
	err := chromedp.Run(ctx,
		//设置webdriver检测反爬
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
		//loadCookies(),
		chromedp.Navigate("https://trade.1688.com/order/buyer_order_list.htm?tracelog=work_2_m_buyList"),
		//chromedp.Sleep(time.Second*5),
		//chromedp.SendKeys("#fm-login-id", username),
		//chromedp.SendKeys("#fm-login-password", password+kb.Enter),
		forr(),
		//停止网页加载
		chromedp.Stop(),
	)
	if err != nil {
		log.Println(err)
	}

}

func takeScreenshot(url string, buf *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 获取页面布局指标
			_, _, contentSize, _, _, _, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := contentSize.Width, contentSize.Height

			// 设置浏览器窗口大小
			err = emulation.SetDeviceMetricsOverride(int64(math.Ceil(width)), int64(math.Ceil(height)), 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// 捕获整个视口的截图
			*buf, err = page.CaptureScreenshot().
				WithFormat(page.CaptureScreenshotFormatPng).
				WithClip(&page.Viewport{
					X:      0,
					Y:      0,
					Width:  width,
					Height: height,
					Scale:  1,
				}).Do(ctx)
			return err
		}),
	}
}
func htmlToScreenshot(html string, buf *[]byte) chromedp.Tasks {
	// 使用 json.Marshal 对 HTML 进行转义，确保可以安全地注入到 JavaScript 字符串中
	escapedHTML, _ := json.Marshal(html)
	js := fmt.Sprintf(`document.documentElement.outerHTML = %s;`, string(escapedHTML))

	return chromedp.Tasks{
		chromedp.Navigate(`about:blank`), // 导航到空白页
		chromedp.WaitReady("html"),       // 等待 HTML 元素准备就绪
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(js, nil).Do(ctx) // 执行 JavaScript 设置页面内容
		}),
		chromedp.Sleep(1 * time.Second), // 可选：等待页面 JavaScript 执行完成
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 获取页面的布局指标
			_, _, contentSize, _, _, _, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := contentSize.Width, contentSize.Height

			// 设置视口大小以确保整个页面被截图
			err = emulation.SetDeviceMetricsOverride(int64(math.Ceil(width)), int64(math.Ceil(height)), 1, false).Do(ctx)
			if err != nil {
				return err
			}

			// 捕获整个视口的截图
			*buf, err = page.CaptureScreenshot().
				WithFormat(page.CaptureScreenshotFormatPng).
				WithClip(&page.Viewport{
					X:      0,
					Y:      0,
					Width:  width,
					Height: height,
					Scale:  1,
				}).Do(ctx)
			return err
		}),
	}
}

func forr() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		log.Println("准备就绪开始获取")

		//var result string
		//var status1 string            //发货状态
		//var logistics_company1 string //物流公司
		//var tracking_number1 string   //运单号码
		//var latest_progress1 string //最新进度
		for i1 := range cgds {
			// 使用随机等待时间来模拟人类行为
			waitTime := time.Duration(4+rand.Intn(3)) * time.Second
			timeout, cancel := context.WithTimeout(ctx, 120*time.Second)
			defer cancel()
			chromedp.Sleep(time.Second * 3)
			//log.Println("1111111111")
			chromedp.SendKeys("#keywords", cgds[i1]+kb.Enter).Do(timeout)
			chromedp.Sleep(waitTime) // 等待页面加载
			err = chromedp.Run(ctx, chromedp.Tasks{
				chromedp.Nodes(`.//a[@class="productName"]`, &nodes), //缓一缓
			}) //执行爬虫任务
			if err != nil {
				log.Println(err)
				continue
			}
			//if len(nodes) > 0 {
			//	m[cgds[i1]] = nodes[0].Attributes[3]
			//}

			fmt.Println("get information success ", nodes)
			// 定义需要尝试点击的元素的索引数组
			// 定义需要尝试的索引数组
			//indexes := []int{3, 4, 5, 6, 7}
			//
			//// 成功标志
			//success := false
			//
			//for _, index := range indexes {
			//	// 为每次点击尝试设置2秒超时
			//	tryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			//	defer cancel()
			//
			//	selector := fmt.Sprintf("#listBox>ul>li>div.order-detail>div:nth-child(2)>table>tbody>tr>td.s7>div:nth-child(%d)>a", index)
			//	// 直接执行点击操作
			//	log.Println(selector)
			//	err := chromedp.Run(tryCtx, chromedp.Click(selector, chromedp.NodeVisible))
			//	if err != nil {
			//		log.Printf("Failed to click selector for index %d: %v", index, err)
			//	} else {
			//		log.Printf("Successfully clicked selector for index %d", index)
			//		success = true
			//		break // 成功后中断循环
			//	}
			//}
			//
			//if !success {
			//	log.Println("Failed to click any of the selectors.")
			//}
			// 选择器
			// 基础选择器路径
			baseSelector := "#listBox > ul > li > div.order-detail > div:nth-child(2) > table > tbody > tr > td.s7"

			// JavaScript 代码来模拟点击，判断第3个还是第4个子元素存在
			jsScript := fmt.Sprintf(`
		var selector3 = '%s > div:nth-child(3) > a';
		var selector4 = '%s > div:nth-child(4) > a';
		var element = document.querySelector(selector3);
		if (!element) {
			element = document.querySelector(selector4);
		}
		if (element) {
			element.click();
		}
	`, baseSelector, baseSelector)

			// 执行 JS 点击
			var result interface{}
			err = chromedp.Run(ctx, chromedp.Evaluate(jsScript, &result))
			if err != nil {
				log.Fatalf("Failed to click via JavaScript: %v", err)
			} else {
				log.Println("Successfully clicked via JavaScript.")
			}
			//if err != nil {
			//	log.Println(err)
			//	return err
			//}

			log.Println("点击查看物流")
			checkHTMLElement(ctx)
			log.Println("等待元素可见")
			chromedp.WaitVisible("#content > div > div > div:nth-child(10) > div > div > div.content.tab-content > div:nth-child(2) > div > div > div.logistics-item.SIGN > div.item-content > div.unit-logistics-flow > div > div.logistics-flow-container > div.trace-list > div > ul > li:nth-child(1)", chromedp.ByID) // 等待新页面的某个元素可见
			log.Println("元素可见")

			//截图
			log.Println("截图")
			var pageSource string

			// 获取所有标签
			// 获取所有标签页
			targets, _ := chromedp.Targets(ctx)
			if err != nil {
				log.Println("Failed to retrieve targets:", err)
				return
			}

			// 找到符合条件的新标签页，并激活
			var newCtx context.Context
			var found bool
			var newUrl string
			for _, t := range targets {
				if t.Type == "page" && strings.Contains(t.URL, "orderId=") {
					newUrl = t.URL
					newCtx, cancel = chromedp.NewContext(ctx, chromedp.WithTargetID(t.TargetID))
					//defer cancel() // 确保在不再需要时取消上下文
					newCtx, cancel = context.WithTimeout(newCtx, 120*time.Second)
					//defer cancel()
					found = true
					log.Println("New page context created")
					break
				}
			}

			if !found {
				log.Println("No new page matching the criteria was found.")
				return
			}

			extendedCtx, cancelExtended := context.WithTimeout(newCtx, 120*time.Second)
			defer cancelExtended()

			// https://trade.1688.com/order/new_step_order_detail.htm?orderId=2132612366027499155
			err = chromedp.Run(extendedCtx, chromedp.Location(&newUrl)) // 等待5秒
			if err != nil {
				fmt.Println("erro during sleep: %v", err)
			}
			//log.Println("body",)
			//chromedp.WaitVisible("body") // 等待页面body元素可见
			//chromedp.OuterHTML("html", &pageSource)
			err = chromedp.Run(extendedCtx,
				chromedp.OuterHTML("html", &pageSource, chromedp.ByQuery),
			)

			if err != nil {
				fmt.Println("erro OuterHTML: %v", err)
			}

			chromedp.ListenTarget(newCtx, func(ev interface{}) {
				if ev, ok := ev.(*network.EventLoadingFinished); ok {
					log.Println("Loading finished:", ev.Timestamp)
				}
			})
			//log.Println(pageSource)
			var buf []byte

			if newCtx.Err() != nil {
				// newCtx 已经被取消或者超时了
				if newCtx.Err() == context.Canceled {
					fmt.Println("newCtx 已经被取消")
				} else if newCtx.Err() == context.DeadlineExceeded {
					fmt.Println("newCtx 因为超时而被取消")
				}
			} else {
				// newCtx 仍然是活跃的
				fmt.Println("newCtx 仍然是活跃的")
			}

			err = chromedp.Run(extendedCtx,
				htmlToScreenshot(pageSource, &buf),
			)

			if newCtx.Err() != nil {
				// newCtx 已经被取消或者超时了
				if newCtx.Err() == context.Canceled {
					fmt.Println("newCtx 已经被取消")
				} else if newCtx.Err() == context.DeadlineExceeded {
					fmt.Println("newCtx 因为超时而被取消")
				}
			} else {
				// newCtx 仍然是活跃的
				fmt.Println("newCtx 仍然是活跃的")
			}

			log.Println(buf)
			if err != nil {
				fmt.Println("erro screenShot: %v", err)
			}
			err = ioutil.WriteFile(".\\img\\"+cgds[i1]+".png", buf, 0644)
			if err != nil {
				fmt.Println("err:%v", err)
			}

			//fmt.Println("source: ", pageSource)
			//err = chromedp.Run(newCtx,
			//	chromedp.Location(&url),
			//	chromedp.OuterHTML("html", &pageSource, chromedp.ByQuery),
			//	chromedp.Sleep(3*time.Second),
			//	chromedp.WaitVisible("body"),
			//	screenShot(".\\img\\"+cgds[i1]+".jpg", `html`),
			//)
			//if err != nil {
			//	fmt.Println("erro during sleep: %v", err)
			//}
			//log.Println(url)

			//screenShot(".\\img\\"+cgds[i1]+".jpg", `html`).Do(timeout)
			log.Println("截图完成")
			//关闭当前页
			//chromedp.SendKeys(`body`, `\x17w`, chromedp.ByQuery) // \x17 是 Ctrl 键的代码
			log.Println("打开新页面")
			chromedp.Navigate("https://trade.1688.com/order/buyer_order_list.htm?tracelog=work_2_m_buyList")
			chromedp.Sleep(waitTime)
			// 清除输入框内容
			//err = clearInputField("#keywords").Do(ctx)
			//if err != nil {
			//	log.Println(err)
			//	return err
			//}
			log.Println("清除输入框内容")
			// 随机等待
			chromedp.Sleep(waitTime)
			log.Println(cgds[i1] + " 完成")
			break
		}

		//chromedp.Evaluate(`document.querySelector("#contractTab > a").click()`, nil)

		return nil
	}

}

func screenShot(name string, path string) chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {

		timeout, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		var code []byte
		if err = chromedp.Screenshot(path, &code).Do(timeout); err != nil {
			fmt.Println(err)
			return
		}
		return os.WriteFile(name, code, 0755)
	}
}

// 检查网页源代码是否存在
func checkHTMLElement(ctx context.Context) error {
	var exists bool
	// 检查html元素是否存在
	err := chromedp.Run(ctx, chromedp.Evaluate(`document.querySelector('html') !== null`, &exists))
	if err != nil {
		return err
	}
	if !exists {
		fmt.Println("No 'html' element found!")
	} else {
		fmt.Println("'html' element is present.")
	}
	return nil
}
func clearInputField(selector string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		script := fmt.Sprintf(`document.querySelector('%s').value = '';`, selector)
		return chromedp.Evaluate(script, nil).Do(ctx)
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

// 保存Cookies
func saveCookies() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// cookies的获取对应是在devTools的network面板中
		// 1. 获取cookies
		cookies, err := network.GetCookies().Do(ctx)
		if err != nil {
			return
		}
		// 2. 序列化
		cookiesData, err := network.GetCookiesReturns{Cookies: cookies}.MarshalJSON()
		if err != nil {
			return
		}

		// 3. 存储到临时文件
		if err = os.WriteFile("cookies.txt", cookiesData, 0755); err != nil {
			return
		}
		return
	}
}
