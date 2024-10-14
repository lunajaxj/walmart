package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strings"
	"time"
)

var cgds []string

func main() {
	fi, err := os.Open("week_counts.txt")
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
	log.Printf("week个数为:%d", len(cgds))
	patentsgazette()
	//xlsx := excelize.NewFile()
	//num := 2
	//if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"专利号", "专利信息", "图片", "专利名字"}); err != nil {
	//	fmt.Println(err)
	//}
	//for _, v := range cgds {
	//	if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v, nil}); err != nil {
	//		fmt.Println(err)
	//	}
	//	if err := xlsx.AddPicture("Sheet1", "B"+strconv.Itoa(num), ".\\img\\"+v+".jpg", nil); err != nil {
	//		fmt.Println(err)
	//	}
	//	num++
	//}
	//xlsx.SaveAs("out.xlsx")
	log.Printf("全部完成")
}

func patentsgazette() {
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
	//log.Printf("开始登录")
	for _, cgd := range cgds {
		log.Println("处理:", cgd)
		url1 := fmt.Sprintf("https://patentsgazette.uspto.gov/week%s/OG/issue_home.html", cgd)
		err := chromedp.Run(ctx,
			fetchAndClick(ctx, url1), // 假设fetchAndClick已经设计为返回chromedp.Action类型
		)
		if err != nil {
			log.Println("执行失败:", err)
			continue
		}
	}
}

func fetchAndClick(ctx context.Context, url string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		// 导航到指定的URL
		if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
			return err
		}
		if err := chromedp.Run(ctx, chromedp.Sleep(2*time.Second)); err != nil {
			return err
		}
		if err := chromedp.Run(ctx, chromedp.Click(`/html/body/div/ul/li[2]/a`)); err != nil {
			return err
		}
		if err := chromedp.Run(ctx, chromedp.Sleep(2*time.Second)); err != nil {
			return err
		}
		log.Println("休息两秒")
		// 抓取所有链接
		var nodes []*cdp.Node
		log.Println("1")
		if err := chromedp.Run(ctx,
			//chromedp.WaitVisible(`html body table tbody`, chromedp.ByQuery), // 确保 tbody 已加载
			chromedp.Nodes(`html body`, &nodes, chromedp.ByQuery, chromedp.AtLeast(0)),
		); err != nil {
			return err
		}
		log.Println(nodes)
		// 提取href属性和文本
		for _, node := range nodes {
			log.Println(node)
			// 获取节点的所有属性
			attributes := make(map[string]string)
			if err := chromedp.Run(ctx, chromedp.Attributes(node.FullXPath(), &attributes, chromedp.ByQuery)); err != nil {
				log.Println("Failed to get attributes:", err)
				continue
			}
			log.Println("3")
			// 从属性中提取href
			href, ok := attributes["href"]
			if !ok {
				continue // 如果属性不存在，跳过
			}
			log.Println("4")
			// 提取链接文本
			var text string
			if err := chromedp.Run(ctx, chromedp.Text(node.FullXPath(), &text, chromedp.ByQuery)); err != nil {
				log.Printf("Failed to get text for node: %v", err)
				continue
			}
			log.Println("5")
			// 打印文本和链接
			fmt.Printf("Text: %s, Href: %s\n", text, href)
		}
		log.Println("6")
		return nil
	})
}

//func screenShot(name string, path string) chromedp.ActionFunc {
//	return func(ctx context.Context) (err error) {
//		timeout, cancel := context.WithTimeout(ctx, 30*time.Second)
//		defer cancel()
//		var code []byte
//		if err = chromedp.Screenshot(path, &code).Do(timeout); err != nil {
//			fmt.Println(err)
//			return
//		}
//		return os.WriteFile(name, code, 0755)
//	}
//}
