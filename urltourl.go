package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"sync"
)

func main() {
	// 读取跳转链接列表文件
	file, err := os.Open("img.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 解析跳转链接列表文件
	var redirectLinks []string
	scanner := csv.NewReader(file)
	for {
		record, err := scanner.Read()
		if err != nil {
			break
		}
		redirectLinks = append(redirectLinks, record[0])
	}

	// 创建结果文件
	resultFile, err := os.Create("result.csv")
	if err != nil {
		fmt.Println("Error creating result file:", err)
		return
	}
	defer resultFile.Close()

	// 写入结果文件表头
	resultWriter := csv.NewWriter(resultFile)
	resultWriter.Write([]string{"跳转前", "跳转后"})

	// 创建 WaitGroup 以便协调 goroutine
	var wg sync.WaitGroup
	wg.Add(len(redirectLinks))

	// 使用 goroutine 执行转换跳转链接
	for i := 0; i < len(redirectLinks); i++ {
		go func(index int) {
			// 发送 GET 请求获取跳转后的链接
			resp, err := http.Get(redirectLinks[index])
			if err != nil {
				fmt.Println("Error fetching URL:", err)
				wg.Done()
				return
			}

			// 关闭响应体
			defer resp.Body.Close()

			// 获取跳转后的链接
			finalLink := resp.Request.URL.String()

			// 写入结果文件
			resultWriter.Write([]string{redirectLinks[index], finalLink})

			// 标记 WaitGroup 计数器减 1
			wg.Done()
		}(i)
	}

	// 等待所有 goroutine 执行完成
	wg.Wait()

	// 刷新结果文件
	resultWriter.Flush()

	// 输出结果文件名
	fmt.Println("完成，结果文件: result.csv")
}
