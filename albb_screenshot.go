package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/xuri/excelize/v2"
	_ "image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}
var ch chan int

var ids []string

func main() {
	// 读取并发数
	concurrency := getConcurrency("chrome_counts.txt")
	ch = make(chan int, concurrency) // 动态设置并发数
	log.Println("自动化脚本-1688-信息获取,当前并发数:", concurrency)
	log.Println("开始执行...")

	// 读取 ID 文件
	fi, err := os.Open("1688info_id.txt")
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	r := bufio.NewReader(fi)

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 5 {
			ids = append(ids, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}
	}

	log.Println(ids)
	for _, v := range ids {
		ch <- 1
		wg.Add(1)
		go crawler(v)
	}
	wg.Wait()

	// 创建 Excel 文件
	xlsx := excelize.NewFile()
	sheet := "Sheet1"
	xlsx.SetSheetRow(sheet, "A1", &[]interface{}{"ID", "Screenshot"})

	for i, id := range ids {
		row := i + 2
		xlsx.SetCellValue(sheet, fmt.Sprintf("A%d", row), id)
		imgFile := fmt.Sprintf("img/%s.png", id)
		if _, err := os.Stat(imgFile); os.IsNotExist(err) {
			log.Printf("截图文件不存在: %s\n", imgFile)
			continue
		}
		// 设置图片插入选项
		options := &excelize.GraphicOptions{
			ScaleX: 0.5,
			ScaleY: 0.5,
		}
		if err := xlsx.AddPicture(sheet, fmt.Sprintf("B%d", row), imgFile, options); err != nil {
			log.Println("插入图片错误：", err)
		}
	}

	// 保存 Excel 文件
	fileName := "out.xlsx"
	if err := xlsx.SaveAs(fileName); err != nil {
		log.Println("保存 Excel 文件错误：", err)
	}

	log.Println("完成")
}

func crawler(ur string) {
	defer func() {
		wg.Done()
		<-ch
	}()

	for i := 0; i < 10; i++ {
		var buf []byte
		// 创建 Chrome 远程调试上下文
		ctx, cancel := chromedp.NewRemoteAllocator(context.Background(), "ws://localhost:9222")
		defer cancel()

		// 创建新的上下文
		ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
		defer cancel()
		err := chromedp.Run(ctx,
			chromedp.Navigate("https://trade.1688.com/order/new_step_order_detail.htm?orderId="+ur+"&tracelog=20120313bscentertologisticsbuyer#logisticsTabTitle"),
			chromedp.Sleep(3*time.Second),     // 等待页面加载
			chromedp.FullScreenshot(&buf, 90), // 截图，质量90%
		)
		if err != nil {
			log.Println("错误信息：" + err.Error())
			log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + ur)
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		// 保存截图
		fileName := fmt.Sprintf("img/%s.png", ur)
		err = os.MkdirAll(filepath.Dir(fileName), 0755)
		if err != nil {
			log.Println("创建文件夹错误：" + err.Error())
			continue
		}
		err = os.WriteFile(fileName, buf, 0644)
		if err != nil {
			log.Println("保存截图错误：" + err.Error())
			continue
		}

		log.Println("截图已保存:", fileName)
		return
	}
}

func getConcurrency(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("无法打开文件 %s: %v", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		concurrency, err := strconv.Atoi(scanner.Text())
		if err != nil {
			log.Fatalf("无法解析并发数: %v", err)
		}
		return concurrency
	}

	log.Fatalf("无法读取并发数")
	return 1 // 默认并发数
}
