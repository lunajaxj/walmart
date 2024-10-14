package main

import (
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/tealeg/xlsx"
	"golang.org/x/net/context"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	// 打开 Excel 文件
	xlFile, err := xlsx.OpenFile("gpt_out.xlsx")
	if err != nil {
		log.Fatalf("无法打开 Excel 文件：%s\n", err)
	}

	// 遍历每个工作表
	for _, sheet := range xlFile.Sheets {
		fmt.Printf("工作表名称：%s\n", sheet.Name)

		// 遍历每一行
		for rowIndex, row := range sheet.Rows {
			if rowIndex == 0 {
				// 第一行是标题行，打印列标题
				fmt.Println("ID:", row.Cells[0].String())
				fmt.Println("输出结果:", row.Cells[15].String()) // P 列对应的索引是 15
			} else {
				// 从第二行开始，打印每一行的内容
				//id := row.Cells[0].String()
				outputResult := row.Cells[15].String() // P 列对应的索引是 15
				// 创建新的上下文和取消函数
				// 创建新的上下文和取消函数
				ctx, cancel := chromedp.NewContext(context.Background())
				defer cancel()

				// 创建可见的 Chrome 实例
				opts := append(chromedp.DefaultExecAllocatorOptions[:],
					chromedp.Flag("headless", false),
					chromedp.Flag("disable-gpu", false),
					chromedp.Flag("window-size", "1200,800"),
				)
				allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
				defer cancel()

				// 使用可见的 Chrome 实例创建新的上下文
				ctx, cancel = chromedp.NewContext(allocCtx)
				defer cancel()

				// 导航到指定网址
				if err := chromedp.Run(ctx,
					chromedp.Navigate("https://zaixianwangyebianji.bmcx.com/"),
				); err != nil {
					log.Fatal(err)
				}
				// 等待一定时间，确保页面加载完成
				time.Sleep(2 * time.Second)

				// 将 outputResult 变量的值添加到 iframe 的 body 元素中
				if err := chromedp.Run(ctx,
					chromedp.EvaluateAsDevTools(`
            // 查找指定的 iframe 元素
            const iframe = document.querySelector("#main_content > div.ke-container.ke-container-default > div.ke-edit > iframe");
            // 获取 iframe 中的 body 元素
            const body = iframe.contentDocument.body;
            // 将 outputResult 变量的值添加到 body 元素中
            body.innerText = "122221";
        `, nil),
				); err != nil {
					log.Fatal(err)
				}

				// 在这里你可以对 outputResult 进行处理，比如输出到控制台
				log.Println("P 列的输出结果:", outputResult)

				// 退出 iframe 框架
				if err := chromedp.Run(ctx,
					chromedp.EvaluateAsDevTools(`parent.location.href = location.href;`, nil),
				); err != nil {
					log.Fatal(err)
				}
				time.Sleep(2 * time.Second)

				// 点击按钮
				if err := chromedp.Run(ctx,
					chromedp.Click(`#main_content > div.ke-container.ke-container-default > div.ke-toolbar > span:nth-child(1)`, chromedp.NodeVisible),
				); err != nil {
					log.Fatal(err)
				}
				rowIndex += 1
				//time.Sleep(9999999 * time.Second)

				//fmt.Printf("第 %d 行 - ID: %s, 输出结果: %s\n", rowIndex+1, id, outputResult)
			}
		}
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
