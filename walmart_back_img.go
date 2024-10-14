package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v2"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type config struct {
	Headers map[string]string `yaml:"headers"`
}

// var conf config
var ids []string
var wg = sync.WaitGroup{}
var wgImg = sync.WaitGroup{}
var ch = make(chan int, 8)
var res []Mode
var conf config

type Mode struct {
	id          string
	color       string
	typex       string
	image       string
	imageUrl    string
	value       string
	shipPrice   string
	itemId      string
	name        string
	offerCount  string
	buyBoxPrice string
	category    string
	productName string
	productType string
	brand       string
}

func main() {
	fmt.Println("自动化脚本-walmart-后台取码")
	fmt.Println("开始执行...")
	GetConfig("config.yml")
	// 创建句柄
	fi, err := os.Open("id.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 2 {
			ids = append(ids, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}

	}
	fmt.Println(ids)
	for _, v := range ids {
		wg.Add(1)
		ch <- 1
		go crawler(v)
	}
	time.Sleep(2 * time.Second)
	wg.Wait()
	wgImg.Wait()

	xlsx := excelize.NewFile()
	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "tiemId", "img", "商品码类型", "商品码值", "标题", "跟卖数量", "运费", "购物车价格", "category", "productName", "productType", "imageUrl", "brand"}); err != nil {
		log.Println(err)
	}
	for _, sv := range ids {
		for _, v := range res {
			if v.id == sv {
				if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v.id, v.itemId, nil, v.typex, v.value, v.name, v.offerCount, v.shipPrice, v.buyBoxPrice, v.category, v.productName, v.productType, v.imageUrl, v.brand}); err != nil {
					log.Println(err)
				}
				if err := xlsx.AddPicture("Sheet1", "C"+strconv.Itoa(num), v.image, nil); err != nil {
					log.Println(err)
				}
				num++

			}
		}
	}
	fileName := "out.xlsx"
	for fileNum := 1; exists(fileName); fileNum++ {
		fileName = "out(" + strconv.Itoa(fileNum) + ").xlsx"
	}
	xlsx.SaveAs(fileName)

}
func exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
func crawler(id string) {

	defer func() {
		wg.Done()
		<-ch

	}()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	for i := 0; i < 16; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}
		client := &http.Client{Timeout: 30 * time.Second, Transport: tr}
		request, err := http.NewRequest("GET", "https://seller.walmart.com/resource/item/item-search?search="+id, nil)
		log.Println("request=", request)
		//request.Header.Add("Accept-Encoding", "gzip, deflate, br") //使用gzip压缩传输数据让访问更快
		if err != nil {
			fmt.Println("请求超时，重新开始：" + id)
			continue
		}
		for k, v := range conf.Headers {
			request.Header.Add(k, v)
		}
		response, err := client.Do(request)
		if err != nil {
			fmt.Println("请求超时，重新开始：" + id)
			continue
		}
		result := ""
		reader, err := gzip.NewReader(response.Body) // gzip解压缩
		if err != nil {
			log.Println("response.StatusCode=", response.StatusCode)
			log.Println(gzip.NewReader(response.Body))

			if err == io.EOF {
				log.Println("已经读取到文件末尾")
				// 这是正常情况
			} else {
				log.Println("读取数据错误:", err)
				continue
			}

		}
		defer reader.Close()
		con, err := io.ReadAll(reader)
		if err != nil {
			log.Println(err)
			continue
		}
		result = string(con)
		log.Println(string(con))
		//upc与upc类型
		ean := regexp.MustCompile("ean\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		upc := regexp.MustCompile("upc\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		gtin := regexp.MustCompile("gtin\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		shipPrice := regexp.MustCompile("shipPrice\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		itemId := regexp.MustCompile("itemId\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		name := regexp.MustCompile("productName\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		offerCount := regexp.MustCompile("offerCount\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		buyBoxPrice := regexp.MustCompile("buyBoxPrice\":(.+?),").FindAllStringSubmatch(result, -1)
		color := regexp.MustCompile("Actual Color\":(.+?),").FindAllStringSubmatch(result, -1)
		image := regexp.MustCompile("image\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		category1 := regexp.MustCompile("category\":(.+?),").FindAllStringSubmatch(result, -1)
		product_name := regexp.MustCompile("productName\":(.+?),").FindAllStringSubmatch(result, -1)
		product_type := regexp.MustCompile("productType\":(.+?),").FindAllStringSubmatch(result, -1)
		brand1 := regexp.MustCompile("brand\":(.+?),").FindAllStringSubmatch(result, -1)
		mode := Mode{}
		mode.id = id
		if len(gtin) > 0 {
			mode.value = gtin[0][1]
			mode.typex = "gtin"
		} else if len(ean) > 0 {
			mode.typex = "ean"
			mode.value = ean[0][1]
		} else if len(upc) > 0 {
			mode.value = upc[0][1]
			mode.typex = "upc"
		}

		if len(shipPrice) > 0 {
			mode.shipPrice = shipPrice[0][1]
		}
		if len(buyBoxPrice) > 0 {
			mode.buyBoxPrice = buyBoxPrice[0][1]
		}
		if len(color) > 0 {
			mode.color = color[0][1]
		}
		if len(itemId) > 0 {
			mode.itemId = itemId[0][1]
		}
		if len(name) > 0 {
			mode.name = name[0][1]
		}
		if len(offerCount) > 0 {
			mode.offerCount = offerCount[0][1]
		}
		if len(image) > 0 {
			wgImg.Add(1)
			go imgxz(id, image[0][1])
			mode.image = ".\\img\\" + id + ".jpeg"
			mode.imageUrl = image[0][1]
		}
		if len(category1) > 0 {
			mode.category = category1[0][1]
		}
		if len(product_name) > 0 {
			mode.productName = product_name[0][1]
		}
		if len(product_type) > 0 {
			mode.productType = product_type[0][1]
		}
		if len(brand1) > 0 {
			mode.brand = brand1[0][1]
		}

		res = append(res, mode)
		fmt.Println("id:" + id + "完成")
		return
	}

}

// 读取配置文件
func GetConfig(path string) {
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

func imgxz(id, img string) {
	defer func() {
		wgImg.Done()
	}()
	if len(img) < 10 {
		return
	}
	for i := 0; i < 5; i++ {
		imgPath := ".\\img\\"
		img = strings.Replace(img, "\\u0026", "&", -1)
		//log.Println(img)
		ress, err := http.Get(img)
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
			time.Sleep(time.Second * 1)
			continue
		}
		// 获得文件的writer对象
		writer := bufio.NewWriter(file)
		io.Copy(writer, reader)
		log.Println(id, "图片下载完成")
		return
	}

}
