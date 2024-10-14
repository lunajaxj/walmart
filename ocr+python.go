package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"golang.org/x/image/draw"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/chromedp/chromedp"
)

func main() {
	// 从文件中读取网址和坐标
	url, rect, err := readOCRConfig("ocr.txt")
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	// 配置
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", false),
		chromedp.Flag("blink-settings", "imagesEnabled=true"),
		//chromedp.WindowSize(1920, 1080),
		//chromedp.Flag("start-fullscreen", true), //全屏无最上方浏览器装饰
		chromedp.Flag("start-maximized", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-sync", true),
		chromedp.NoFirstRun,
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60000*time.Second)
	defer cancel()

	log.Println("开始登录")
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
		// 使用从文件读取的 URL 和坐标
		forr1(url, rect),
	)
	if err != nil {
		log.Println("导航过程中出现错误：", err)
	}
}

// 从文件中读取网址和坐标
func readOCRConfig(filename string) (string, image.Rectangle, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", image.Rectangle{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 读取第一行：网址
	scanner.Scan()
	url := scanner.Text()

	// 读取第二行：坐标
	scanner.Scan()
	coordStr := scanner.Text()
	coords := strings.Split(coordStr, ",")

	if len(coords) != 4 {
		return "", image.Rectangle{}, fmt.Errorf("坐标数量不正确")
	}

	// 将坐标转换为整数
	x1, _ := strconv.Atoi(coords[0])
	y1, _ := strconv.Atoi(coords[1])
	x2, _ := strconv.Atoi(coords[2])
	y2, _ := strconv.Atoi(coords[3])

	rect := image.Rect(x1, y1, x2, y2)
	return url, rect, nil
}

func forr1(url string, rect image.Rectangle) chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		log.Println("等待30秒后开始截图")
		err = chromedp.Run(ctx,
			chromedp.Navigate(url),         // 使用从文件中读取的 URL
			chromedp.Sleep(30*time.Second), // 确保页面加载完成
		)
		if err != nil {
			log.Fatal("导航过程中出现错误：", err)
		}

		// 获取整个页面的截图
		var buf []byte
		err = chromedp.Run(ctx, chromedp.CaptureScreenshot(&buf))
		if err != nil {
			log.Fatal("截取当前屏幕失败：", err)
		}

		// 打印截图大小，确保数据正确
		fmt.Printf("截图大小: %d bytes\n", len(buf))

		// 保存完整的页面截图
		err = saveScreenshot(buf, "./img/full_screenshot.png")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("完整页面截图已保存：./img/full_screenshot.png")

		// 剪裁指定坐标区域的截图
		croppedPath := "./img/cropped_screenshot.png"
		err = cropScreenshot(buf, rect, croppedPath) // 使用从文件读取的坐标
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("剪裁后的截图已保存：./img/cropped_screenshot.png")

		text, err := ocrWithPython(croppedPath)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("识别出的文本:", text, "10秒后关闭程序")
		time.Sleep(10 * time.Second)
		return
	}
}

func saveScreenshot(buf []byte, filename string) error {
	err := os.WriteFile(filename, buf, 0644)
	if err != nil {
		return fmt.Errorf("保存截图失败: %v", err)
	}
	return nil
}

func cropScreenshot(buf []byte, rect image.Rectangle, outputPath string) error {
	// 读取字节数组为图像
	img, _, err := image.Decode(bytes.NewReader(buf)) // 自动检测图像格式
	if err != nil {
		return fmt.Errorf("图像解码失败: %v", err)
	}

	// 获取图片的尺寸
	bounds := img.Bounds()
	fmt.Printf("图片大小: 宽 %d, 高 %d\n", bounds.Dx(), bounds.Dy())

	// 确保剪裁区域在图片范围内
	if rect.Min.X < 0 || rect.Min.Y < 0 || rect.Max.X > bounds.Dx() || rect.Max.Y > bounds.Dy() {
		return fmt.Errorf("剪裁区域超出了图片范围")
	}

	// 创建一个新图像用于存储裁剪后的内容
	subImg := image.NewRGBA(rect)
	draw.CatmullRom.Scale(subImg, rect, img, rect, draw.Over, nil)

	// 保存裁剪后的图片
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer outFile.Close()

	err = png.Encode(outFile, subImg)
	if err != nil {
		return fmt.Errorf("编码 PNG 失败: %v", err)
	}

	return nil
}

func ocrWithPython(imagePath string) (string, error) {
	pythonCode := `
import sys
from PIL import Image
import pytesseract

# 手动指定 Tesseract 可执行文件路径
pytesseract.pytesseract.tesseract_cmd = r'D:\Tesseract-OCR\tesseract.exe'

def perform_ocr(image_path):
    # 使用中文简体语言 'chi_sim' 进行 OCR 识别
    text = pytesseract.image_to_string(Image.open(image_path), lang='chi_sim')
    # 确保输出为 UTF-8 编码
    sys.stdout.buffer.write(text.encode('utf-8'))

if __name__ == "__main__":
    image_path = sys.argv[1]
    perform_ocr(image_path)
`
	// 创建一个临时 Python 文件
	tmpPyFile, err := os.CreateTemp("", "ocr_script*.py")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpPyFile.Name()) // 在使用完后删除临时文件

	_, err = tmpPyFile.WriteString(pythonCode)
	if err != nil {
		return "", err
	}
	tmpPyFile.Close()

	// 运行 Python 代码，执行 OCR 识别
	cmd := exec.Command("python", tmpPyFile.Name(), imagePath)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("OCR 脚本执行失败: %v, 错误信息: %s", err, stderr.String())
	}

	// 确保输出为 UTF-8 编码
	output := out.String()
	if !utf8.ValidString(output) {
		return "", fmt.Errorf("OCR 输出的文本不是有效的 UTF-8 字符串")
	}

	return output, nil
}
