package main

import (
	"fmt"
	"github.com/go-vgo/robotgo"
	"github.com/otiai10/gosseract"
	"image/png"
	"os"
)

func main() {
	// 捕获屏幕区域：起始点(x, y)，宽度，高度
	bitmap := robotgo.CaptureScreen(10, 10, 100, 100)
	defer robotgo.FreeBitmap(bitmap)

	// 将捕获的区域保存为PNG文件
	fileName := "screenshot.png"
	file, _ := os.Create(fileName)
	defer file.Close()
	png.Encode(file, robotgo.ToImage(bitmap))

	// 使用gosseract进行OCR识别
	client := gosseract.NewClient()
	defer client.Close()
	client.SetImage(fileName)
	text, _ := client.Text()
	fmt.Println("识别的文字内容：", text)
}
