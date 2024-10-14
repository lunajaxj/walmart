package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/xuri/excelize/v2"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var cgds []string
var nodes []*cdp.Node

var m = make(map[string]string)

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

	log.Printf("开始执行，有%d个任务需要获取", len(cgds))
	eccang()

	xlsx := excelize.NewFile()

	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"采购单号", "截图", "商品链接"}); err != nil {
		fmt.Println(err)
	}
	for _, v := range cgds {
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v, nil, m[v]}); err != nil {
			fmt.Println(err)
		}
		if err := xlsx.AddPicture("Sheet1", "B"+strconv.Itoa(num), ".\\img\\"+v+".jpg", ""); err != nil {
			fmt.Println(err)
		}

		num++
	}

	xlsx.SaveAs("out.xlsx")

	log.Printf("全部完成")
}

var token string

func eccang() {
	//配置
	options := append(chromedp.DefaultExecAllocatorOptions[:],
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
	ctx, cancel = context.WithTimeout(ctx, 6000000*time.Second)
	defer cancel()
	log.Println("开始登录")
	err := chromedp.Run(ctx,
		//设置webdriver检测反爬
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
		loadCookies(),
		chromedp.Navigate("https://trade.1688.com/order/buyer_order_list.htm?tracelog=work_2_m_buyList"),
		chromedp.Sleep(time.Second*5),
		forr(),
		//停止网页加载
		chromedp.Stop(),
	)
	if err != nil {
		log.Println(err)
	}

}

func forr() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		log.Println("准备就绪开始获取")
		for i, email := range emails {
			// 使用随机等待时间来模拟人类行为
			waitTime := time.Duration(4+rand.Intn(3)) * time.Second
			err := chromedp.Run(ctx,
				chromedp.SendKeys("#react-aria-1", email+kb.Enter),
				chromedp.Sleep(waitTime), // 等待页面加载
			)
			if err != nil {
				log.Println(err)
				return err
			}

			// 点击特定ID的<a>标签按钮
			err = chromedp.Run(ctx,
				chromedp.Click("#jePrfxOxkCIsErN", chromedp.NodeVisible), // 确保元素可见后再点击
				chromedp.Sleep(10*time.Second),                           // 等待10秒钟
			)
			if err != nil {
				log.Println("点击操作出错:", err)
				return err
			}

			// 获取页面内容、处理节点、清除输入框等其他操作...

			// 随机等待
			chromedp.Sleep(waitTime)
			log.Printf("%d 完成\n", i)
		}

		return nil
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
		cookies, err := network.GetAllCookies().Do(ctx)
		if err != nil {
			return
		}
		// 2. 序列化
		cookiesData, err := network.GetAllCookiesReturns{Cookies: cookies}.MarshalJSON()
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
