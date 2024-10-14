package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"log"
	"net/url"
	"sync"
)

var wg = sync.WaitGroup{}

func main() {
	xlsx, err := excelize.OpenFile("input.xlsx")
	if err != nil {
		log.Fatal(err)
	}
	sheet := "Sheet1"
	//获取数据行数
	rows, err := xlsx.GetRows(sheet)
	if err != nil {
		log.Fatal(1, err)
	}
	proxy_str := fmt.Sprint("http://%s:%s@%s", "t19932187800946", "m78z02hx")
	proxy, _ := url.Parse(proxy_str)

}
