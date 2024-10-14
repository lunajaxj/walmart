package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type config struct {
	Headers map[string]string `yaml:"headers"`
}

var conf config
var ids []string
var wg = sync.WaitGroup{}
var ch = make(chan int, 8)
var res []Mode

//type Mode struct {
//	id          string
//	color       string
//	typex       string
//	image       string
//	value       string
//	shipPrice   string
//	itemId      string
//	name        string
//	offerCount  string
//	buyBoxPrice string
//	category    string
//	productName string
//	productType string
//	brand       string
//}

func main() {
	//fmt.Println("自动化脚本-walmart-后台取码")
	fmt.Println("开始执行...")
	GetConfig("config.yml")

	request, _ := http.NewRequest("POST", "https://seller.walmart.com/aurora/v1/feeds/MP_INVENTORY?validateFeed=false&submitFeed=true", file="C:\Users\Administrator\Desktop\file\YLymm-价格-8点-全-下架.xlsx")
	log.Println("request=", request)
	//request.Header.Add("Accept-Encoding", "gzip, deflate, br") //使用gzip压缩传输数据让访问更快

	for k, v := range conf.Headers {
		request.Header.Add(k, v)
	}

	return
}

// 读取配置文件
func GetConfig(path string) {
	a := generateCookie()
	log.Println("cookie=", a)
	con := &config{}
	if f, err := os.Open(path); err != nil {
		if strings.Contains(err.Error(), "The system cannot find the file specified") || strings.Contains(err.Error(), "no such file or directory") {
			con.Headers = map[string]string{"Host": "seller.walmart.com", "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:105.0) Gecko/20100101 Firefox/105.0", "x-xsrf-token": "", "Cookie": ""}
			data, err2 := yaml.Marshal(con)
			err2 = os.WriteFile(path, data, 0644)
			if err2 != nil {
				fmt.Println(err)
			} else {
				fmt.Println("未找到配置文件,已在当面目录下创建配置文件: config.yaml")
			}
		} else {
			fmt.Println("配置文件错误,请尝试重新生成配置文件")
			fmt.Println(err)
		}
		os.Exit(1)
	} else {
		yaml.NewDecoder(f).Decode(con)
		conf = *con
	}
}
