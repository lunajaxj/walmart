package main

import (
	"bufio"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/xuri/excelize/v2"
	"golang.org/x/net/context"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var res = make(map[string][]map[string]string)

//type uspto struct {
//	Key                       string
//	WordMark                  string
//	GoodsAndServices          string
//	StandardCharactersClaimed string
//	MarkDrawingCode           string
//	SerialNumber              string
//	FilingDate                string
//	CurrentBasis              string
//	OriginalFilingBasis       string
//	PublishedforOpposition    string
//	Owner                     string
//	Disclaimer                string
//	TypeOfMark                string
//	Register                  string
//	LiveDeadIndicator         string
//	AbandonmentDate           string
//}

var cgds []string
var nodes []*cdp.Node

var mkey []interface{}

func main() {
	fi, err := os.Open("关键词.txt")
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

	log.Printf("开始执行，有%d个关键词需要获取", len(cgds))
	uspto()

	xlsx := excelize.NewFile()

	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &mkey); err != nil {
		fmt.Println(err)
	}
	for _, v := range cgds {

		for i := range res[v] {
			var ss []interface{}

			for i2 := range mkey {
				s := mkey[i2].(string)
				ss = append(ss, res[v][i][s])
			}
			if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &ss); err != nil {
				fmt.Println(err)
			} else {
				num++
			}
		}

	}

	xlsx.SaveAs("out.xlsx")

	log.Printf("全部完成")
}

func uspto() {
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
	ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	err := chromedp.Run(ctx,
		//设置webdriver检测反爬
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
		forr(),
		//停止网页加载
		chromedp.Stop(),
	)
	if err != nil {
		log.Println(11, err)
	}

}

func getSource(ctx context.Context) string {
	var source string
	if err := chromedp.Run(ctx, chromedp.EvaluateAsDevTools("document.documentElement.outerHTML", &source)); err != nil {
		return ""
	}
	return source
}
func UniqueArr(arr []string) []string {
	newArr := make([]string, 0)
	tempArr := make(map[string]bool, len(newArr))
	for _, v := range arr {
		if tempArr[v] == false {
			tempArr[v] = true
			newArr = append(newArr, v)
		}
	}
	return newArr
}
func forr() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		for i := range cgds {
			nodes = nil
			timeout, cancel := context.WithTimeout(ctx, 999*time.Second)
			defer cancel()
			// 启动监听新页面的上下文

			log.Println("开始获取", cgds[i])
			chromedp.Navigate("https://tmsearch.uspto.gov/search/search-information").Do(timeout)
			chromedp.Sleep(time.Second * 3).Do(timeout)
			chromedp.SendKeys(`#searchbar`, cgds[i]+kb.Enter).Do(timeout)
			chromedp.Sleep(time.Second * 2).Do(timeout)
			//timeout20, cancel20 := context.WithTimeout(ctx, 2*time.Second)
			//defer cancel20()
			chromedp.Sleep(time.Second * 4)
			// 要存储结果的变量
			var elementContent string
			// 等待搜索结果
			chromedp.Text(`div.col.ps-0.ps-sm-1 > span.clickable`, &elementContent, chromedp.NodeVisible).Do(timeout) // 获取元素内容
			elementContent = strings.TrimSpace(strings.Replace(elementContent, "wordmark", "", -1))
			// 输出获取到的元素内容
			fmt.Println("Wordmark的值为:", elementContent)
			// 如果 elementContent 与 cgds[i] 忽略大小写相等，点击第一个 class="col ps-0 ps-sm-1" 的元素
			if strings.EqualFold(elementContent, cgds[i]) {
				log.Printf("elementContent (%s) 忽略大小写等于 cgds[%d] (%s), 准备点击...", elementContent, i, cgds[i])

				// 执行点击操作
				err = chromedp.Run(timeout,
					chromedp.Click(`.w-100.text-center.ng-star-inserted`, chromedp.NodeVisible), // 点击第一个 class="col ps-0 ps-sm-1" 的元素
					// 点击第一个 class="col ps-0 ps-sm-1" 的元素
				)
				if err != nil {
					log.Fatal("点击元素时出错:", err)
				} else {
					log.Println("成功点击元素")
				}
			} else {
				log.Printf("elementContent (%s) 忽略大小写不等于 cgds[%d] (%s), 跳过点击", elementContent, i, cgds[i])
			}
			log.Println("准备点击document按钮检查页面按钮是否可见")
			// 适当增加等待时间以确保页面跳转完成（例如 5 秒）
			chromedp.Sleep(time.Second * 5).Do(timeout)
			// 创建监听新页面打开的上下文
			// 获取所有的页面 targets
			targets, err1 := target.GetTargets().Do(ctx)
			if err1 != nil {
				log.Fatal("获取页面 targets 时出错:", err)
			}

			// 查找最近打开的页面目标
			var newTarget *target.Info
			for _, t := range targets {
				// 通过检查 t.Type == "page" 确认是页面，并排除当前页面
				if t.Type == "page" && t.URL != "https://tmsearch.uspto.gov/search/search-results" { // 确保跳过当前页面
					newTarget = t
					break
				}
			}

			if newTarget == nil {
				log.Fatal("未能找到新打开的页面")
			}

			// 输出新页面的 URL
			log.Printf("新页面的 URL: %s", newTarget.URL)

			// 切换到新页面的上下文
			// 检查当前页面的 URL，确保页面已经跳转
			var currentURL string
			err = chromedp.Run(timeout,
				chromedp.Location(&currentURL), // 获取当前页面的 URL
			)
			if err != nil {
				log.Fatal("获取当前页面 URL 时出错:", err)
			}
			log.Printf("当前页面的 URL: %s", currentURL)
			// 切换到新页面的上下文
			newCtx, cancelNewTab := chromedp.NewContext(ctx, chromedp.WithTargetID(newTarget.TargetID))
			defer cancelNewTab()
			// 确保页面已经跳转并加载新页面元素
			err = chromedp.Run(newCtx,
				chromedp.WaitVisible(`#documentsTabBtn>a`, chromedp.ByID), // 等待新页面的 #documentsTabBtn 元素可见
			)
			if err != nil {
				log.Fatal("等待新页面加载时出错:", err)
			}

			log.Println("新页面加载完成，准备点击 #documentsTabBtn")

			// 使用 JavaScript 点击 #documentsTabBtn
			err = chromedp.Run(newCtx,
				chromedp.Evaluate(`document.querySelector('#documentsTabBtn>a').click();`, nil),
			)
			if err != nil {
				log.Fatal("使用 JS 点击 documentsTabBtn 出错:", err)
			} else {
				log.Println("成功使用 JS 点击 #documentsTabBtn")
			}

			chromedp.Sleep(time.Second * 1).Do(timeout)
			// 模拟获取页面并操作
			// 存储 JavaScript 执行结果
			var jsResult string
			err = chromedp.Run(timeout,
				chromedp.Evaluate(`document.querySelector("#webwidget_tab") ? document.querySelector("#webwidget_tab").innerText : "元素不存在"`, &jsResult),
			)
			if err != nil {
				log.Fatal("页面加载失败:", err)
			}
			log.Println("JavaScript 执行结果: ", jsResult)
			// 存储所有<tr>节点
			//var nodes []*cdp.Node

			//// 执行Chromedp任务，获取所有的<tr>元素
			//err = chromedp.Run(timeout,
			//	chromedp.Nodes(`#docResultsTbody tr`, &nodes), // 获取<tbody>中所有<tr>元素
			//)
			//if err != nil {
			//	log.Fatal("获取<tr>元素时出错:", err)
			//}
			//
			//// 打印 nodes
			//for i, node := range nodes {
			//	log.Printf("Node %d: ID: %s, NodeName: %s, NodeValue: %s", i, node.AttributeValue("id"), node.NodeName, node.NodeValue)
			//	// 打印 attributes
			//	for _, attr := range node.Attributes {
			//		log.Printf("  Attribute: %s", attr)
			//	}
			//}
			//
			//// 遍历所有的<tr>元素，检查第三个<td>中的<a>标签的文本
			//for _, node := range nodes {
			//	var aText string
			//	// 获取第三个 <td> 中 <a> 标签的文本
			//	err := chromedp.Run(timeout,
			//		chromedp.Text(fmt.Sprintf("#%s td:nth-child(3) a", node.AttributeValue("id")), &aText), // 获取第三个<td>中的<a>标签的文本
			//	)
			//	if err != nil {
			//		log.Fatal("获取<a>标签文本时出错:", err)
			//		continue
			//	}
			//
			//	aText = strings.TrimSpace(aText) // 去除多余的空白
			//
			//	// 如果 <a> 标签的文本等于 "Registration Certificate"，点击它
			//	if aText == "Registration Certificate" {
			//		log.Printf("找到目标: %s", aText)
			//
			//		// 点击该 <a> 元素
			//		err = chromedp.Run(timeout,
			//			chromedp.Click(fmt.Sprintf("#%s td:nth-child(3) a", node.AttributeValue("id")), chromedp.NodeVisible), // 点击<a>元素
			//		)
			//
			//		if err != nil {
			//			log.Fatal("点击链接时出错:", err)
			//		} else {
			//			log.Println("成功点击链接")
			//		}
			//	}
			//}

			//chromedp.SendKeys(`button[id="fsrFocusFirst"]`, cgds[i]+kb.Enter).Do(timeout20)
			//chromedp.Sleep(time.Second * 1)
			//chromedp.SendKeys(`td[id="querytext"] input`, cgds[i]+kb.Enter).Do(timeout)
			////chromedp.Click(`input[value="Submit"]`).Do(timeout)
			//chromedp.Sleep(time.Second * 2)
			//timeout11, cancel11 := context.WithTimeout(ctx, 6*time.Second)
			//defer cancel11()
			//chromedp.Nodes(`table[id="searchResultTable"] tr td:nth-child(2) a`, &nodes).Do(timeout11) //缓一缓

			if len(nodes) == 0 {
				chromedp.Sleep(100).Do(timeout)
				source := getSource(timeout)
				key := regexp.MustCompile(`<td align="LEFT[\s\S]*?<b>(.*?)</b>[\s\S]*?</td>[\s\S]*?<td>[\s\S]*?</td>`).FindAllStringSubmatch(source, -1)
				value := regexp.MustCompile(`<td align="LEFT[\s\S]*?<b>.*?</b>[\s\S]*?</td>[\s\S]*?<td>([\s\S]*?)</td>`).FindAllStringSubmatch(source, -1)
				if len(key) == 0 {
					log.Println(cgds[i] + " 没有结果")
					continue
				}
				log.Println(fmt.Sprintf(`获取 %s 第 %d 个`, cgds[i], 1))
				m := make(map[string]string)
				var mkk []interface{}
				mkk = append(mkk, "关键词")
				m["关键词"] = cgds[i]
				for i3 := range key {
					mkk = append(mkk, key[i3][1])
					m[key[i3][1]] = strings.Replace(strings.Replace(value[i3][1], "<b>", "", -1), "</b>", "", -1)
				}
				if len(mkey) < len(key) {
					mkey = mkk
				}
				res[cgds[i]] = append(res[cgds[i]], m)
				continue
			}
			var url = nodes[0].Attributes[1][0 : len(nodes[0].Attributes[1])-1]
			for i2 := 1; i2 <= 50; i2++ {
				chromedp.Sleep(time.Second * 1).Do(timeout)
				err = chromedp.Run(timeout, chromedp.Tasks{
					//chromedp.SendKeys(fmt.Sprintf(`td:nth-child(2) a[href="%s"]`, nodes[i2].Attributes[1]), kb.Enter),
					//chromedp.Click(fmt.Sprintf(`td:nth-child(2) a[href="%s"]`, nodes[i2].Attributes[1])),
					chromedp.Navigate("https://tmsearch.uspto.gov" + url + strconv.Itoa(i2)),
				}) //执行爬虫任务

				if err != nil {
					log.Println(err)
					continue
				}
				//chromedp.SendKeys(fmt.Sprintf(`td:nth-child(2) a[href="%s"]`, nodes[i2].Attributes[1]), kb.Enter).Do(timeout)
				chromedp.Sleep(100).Do(timeout)
				source := getSource(timeout)
				key := regexp.MustCompile(`<td align="LEFT[\s\S]*?<b>(.*?)</b>[\s\S]*?</td>[\s\S]*?<td>[\s\S]*?</td>`).FindAllStringSubmatch(source, -1)
				value := regexp.MustCompile(`<td align="LEFT[\s\S]*?<b>.*?</b>[\s\S]*?</td>[\s\S]*?<td>([\s\S]*?)</td>`).FindAllStringSubmatch(source, -1)
				if len(key) == 0 {
					break
				}
				log.Println(fmt.Sprintf(`获取 %s 第 %d 个`, cgds[i], i2))
				m := make(map[string]string)
				var mkk []interface{}
				mkk = append(mkk, "关键词")
				m["关键词"] = cgds[i]
				for i3 := range key {
					mkk = append(mkk, key[i3][1])
					m[key[i3][1]] = strings.Replace(strings.Replace(value[i3][1], "<b>", "", -1), "</b>", "", -1)
				}
				if len(mkey) < len(key) {
					mkey = mkk
				}
				res[cgds[i]] = append(res[cgds[i]], m)
				//timeout20, cancel20 := context.WithTimeout(timeout, 2*time.Second)
				//defer cancel20()
				//chromedp.NavigateBack().Do(timeout20)

			}
			log.Println(cgds[i] + " 完成")
		}

		return
	}
}
