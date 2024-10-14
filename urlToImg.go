package main

import (
	"bufio"
	"github.com/xuri/excelize/v2"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}
var ch = make(chan int, 30)
var urls []string

func main() {

	log.Println("自动化脚本-url转图片")
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("url.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader
	var s int
	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 5 {
			urls = append(urls, strings.TrimSpace(string(lineB)))
			ch <- 1
			wg.Add(1)
			go imgxz(strconv.Itoa(s), strings.TrimSpace(string(lineB)))
			s++
		}
		if err != nil {
			break
		}
	}
	wg.Wait()

	xlsx := excelize.NewFile()
	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"url", "图片"}); err != nil {
		log.Println(err)
	}
	for i := range urls {
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{urls[i]}); err != nil {
			log.Println(err)
		}
		if err := xlsx.AddPicture("Sheet1", "B"+strconv.Itoa(num), ".\\img\\"+strconv.Itoa(i)+".jpeg", nil); err != nil {
			log.Println(err)
		}
		num++

	}
	fileName := "out.xlsx"
	xlsx.SaveAs(fileName)

	log.Println("完成")

}

func imgxz(id, img string) {
	defer func() {
		wg.Done()
		<-ch
		log.Println("完成：", id)
	}()
	for i := 0; i < 3; i++ {
		imgPath := ".\\img\\"
		img = strings.Replace(img, "\\u0026", "&", -1)
		//log.Println(img)
		ress, err := http.Get(img + "?odnHeight=290&odnWidth=290&odnBg=FFFFFF")
		if err != nil {
			log.Println("图片下载失败!"+err.Error(), "稍后重新开始下载")
			time.Sleep(3 * time.Second)
			continue
		}
		defer ress.Body.Close()
		// 获得get请求响应的reader对象
		reader := bufio.NewReaderSize(ress.Body, 32*1024)
		file, err := os.Create(imgPath + id + ".jpeg")
		if err != nil {
			log.Println("图片下载失败!")
			continue
		}
		// 获得文件的writer对象
		writer := bufio.NewWriter(file)
		io.Copy(writer, reader)
		return
	}

}
