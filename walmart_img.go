package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/xuri/excelize/v2"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var isPi = false

var res []Wal

type Wal struct {
	img      string
	id       string
	stock    string
	typez    string
	value    string
	brand    string //"name":"VicTsing"},"offers":{  品牌
	query    string //*[@id="maincontent"]//div[@data-testid="sticky-buy-box"]//div/p//span//text()  标签，多行相加
	title    string //"name":"Gymax 5 Piece Dining Set Glass Top Table & 4 Upholstered Chairs Kitchen Room Furniture","sku":  标题
	score    string //(4.5)  评分
	review   string //"totalReviewCount":1187}  评论数量
	price    string //aria-hidden="false">$22.98<	价格
	category string

	////div/div/span[@class="lh-title"]//text()  卖家+配送
	seller       string   //卖家
	delivery     string   //配送
	deliveryDate string   //配送时间
	variant1     string   //变体1 :</span><span aria-hidden="true" class="ml1">(.*?)</span>
	variant2     string   //变体1 :</span><span aria-hidden="true" class="ml1">(*?)</span>
	otherIds     []string //变体id
}

var ids []string
var wg = sync.WaitGroup{}
var wgImg = sync.WaitGroup{}

var ch = make(chan int, 10)
var chi = make(chan int, 6)

func main() {
	log.Println("自动化脚本-walmart-信息采集_带图片")
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("id.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

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
	log.Println("等待图片下载结束...")
	wgImg.Wait()
	log.Println("图片下载结束")

	xlsx := excelize.NewFile()
	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "图片", "商品码类型", "商品码值", "品牌", "标签", "标题", "评分", "评论数量", "价格", "卖家", "配送", "变体1", "变体2", "变体id", "到达时间", "库存", "类目"}); err != nil {
		log.Println(err)
	}
	for _, sv := range ids {
		for _, v := range res {
			if v.id == sv {
				other := ""
				for i := range v.otherIds {
					if i == 0 {
						other = v.otherIds[i]
						continue
					}
					other = other + "," + v.otherIds[i]
				}

				if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v.id, nil, v.typez, v.value, v.brand, v.query, v.title, v.score, v.review, v.price, v.seller, v.delivery, v.variant1, v.variant2, other, v.deliveryDate, v.stock, v.category}); err != nil {
					log.Println(err)
				}
				if err := xlsx.AddPicture("Sheet1", "B"+strconv.Itoa(num), v.img, nil); err != nil {
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

	log.Println("完成")

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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var IsC = false
var IsC2 = true

func init() {
	rand.Seed(time.Now().UnixNano()) // 初始化随机数生成器
}

func generateRandomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
func crawler(id string) {

	//配置代理
	defer func() {
		wg.Done()
		<-ch

	}()

	for i := 0; i < 16; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}
		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		request, _ := http.NewRequest("PUT", "https://www.walmart.com/ip/"+id, nil)

		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")
		request.Header.Set("Accept-Language", "zh")
		request.Header.Set("Sec-Ch-Ua", `"Not.A/Brand";v="8", "Chromium";v="114", "Google Chrome";v="114"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", `document`)
		request.Header.Set("Sec-Fetch-Mode", `navigate`)
		request.Header.Set("Sec-Fetch-Site", `none`)
		request.Header.Set("Sec-Fetch-User", `?1`)
		request.Header.Set("Upgrade-Insecure-Requests", `1`)
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")
		var isc = IsC
		if IsC {
			request.Header.Set("Cookie", generateRandomString(10))
		}
		response, err := client.Do(request)

		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始：" + id)
				continue
			} else if strings.Contains(err.Error(), "441") {
				log.Println("代理超频！暂停10秒后继续...")
				time.Sleep(time.Second * 10)
				continue
			} else if strings.Contains(err.Error(), "440") {
				log.Println("代理宽带超频！暂停5秒后继续...")
				time.Sleep(time.Second * 5)
				continue
			} else {
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + id)
				continue
			}
		}
		result := ""
		if response.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(response.Body) // gzip解压缩
			if err != nil {
				log.Println("解析body错误，重新开始：" + id)
				continue
			}
			defer reader.Close()
			con, err := io.ReadAll(reader)
			if err != nil {
				log.Println("gzip解压错误，重新开始：" + id)
				continue
			}
			result = string(con)
		} else {
			dataBytes, err := io.ReadAll(response.Body)
			if err != nil {
				if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "Service Unavailable") {
					log.Println("代理IP无效，自动切换中")
					log.Println("连续出现代理IP无效请联系我，重新开始：" + id)
				} else {
					log.Println("错误信息：" + err.Error())
					log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + id)
				}
				continue
			}
			defer response.Body.Close()
			result = string(dataBytes)
		}
		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
				log.Println("连续出现代理IP无效请联系我，重新开始：" + id)
				time.Sleep(time.Second * 1)
				continue
			} else {
				log.Println("错误信息：" + err.Error())
				log.Println("出现错误，如果同id连续出现请联系我，重新开始：" + id)
				time.Sleep(time.Second * 1)
				continue
			}

		}
		wal := Wal{}
		wal.id = id
		if strings.Contains(result, "This page could not be found.") {
			wal.typez = "该商品不存在"
			res = append(res, wal)
			log.Println("id:" + id + "商品不存在")
			return
		}

		//upc与upc类型
		upc := regexp.MustCompile("upc\":\"(.{4,30}?)\"").FindAllStringSubmatch(result, -1)
		gtin := regexp.MustCompile("gtin13\":\"(.{4,30}?)\"").FindAllStringSubmatch(result, -1)
		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
		if len(upc) > 0 {
			wal.value = upc[0][1]
			wal.typez = "upc"
		} else if len(gtin) > 0 {
			wal.value = gtin[0][1]
			wal.typez = "gtin"
		} else if len(fk) > 0 {
			log.Println("id:" + id + " 被风控,更换IP继续")
			IsC = !isc
			time.Sleep(time.Second * 1)
			continue
		} else {
			wal.value = ""
			wal.typez = "ean"
			//log.Println("id:"+id+" 获取为空，默认为ean")
		}

		doc, err := htmlquery.Parse(strings.NewReader(result))
		if err != nil {
			log.Println("错误信息：" + err.Error())
			return
		}

		doc1, err := goquery.NewDocumentFromReader(strings.NewReader(result))
		if err != nil {
			log.Println("解析HTML错误：", err)
			return
		}

		//品牌
		brand := regexp.MustCompile("\"brand\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		if len(brand) == 0 {
			//log.Println("品牌获取失败id："+id)
		} else {
			wal.brand = brand[0][1]
		}

		//标签
		queryStr := ""
		doc1.Find("#maincontent > section > main > div.flex.flex-column.h-100 > div:nth-child(2) > div > div.w_aoqv.w_wRee.w_fdPt > div > div:nth-child(2) > div > div > div:nth-child(1) > div.flex.items-center.mv2.flex-wrap > div:nth-child(1) > div > span").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			if !strings.Contains(queryStr, text) {
				queryStr += text + " "
			}
		})
		wal.query = queryStr

		//标题
		title := regexp.MustCompile("\"name\":\"(.+?)\",").FindAllStringSubmatch(result, -1)
		if len(title) == 0 {
			log.Println("获取失败id："+id, "重新请求")

		} else {
			wal.title = strings.Replace(title[0][1], "\\u0026", "&", -1)
		}

		//评分
		score := regexp.MustCompile("[(]([\\d][.][\\d])[)]").FindAllStringSubmatch(result, -1)
		if len(score) == 0 {
			//log.Println("评分获取失败id："+id)
		} else {
			wal.score = score[0][1]
		}

		//评论数量
		review := regexp.MustCompile("\"totalReviewCount\":(\\d+)").FindAllStringSubmatch(result, -1)
		if len(review) == 0 {
			//log.Println("评论数量获取失败id："+id)
		} else {
			wal.review = review[0][1]
		}

		// 使用具体的XPath查找目标span标签
		price := regexp.MustCompile(`"best[^{]+?,"priceDisplay":"([^"]+)"`)
		price1 := price.FindAllString(result, -1)
		if len(price1) > 0 {
			//log.Println(result)
			// Check if the matched string contains "priceDisplay"
			if strings.Contains(price1[0], `"priceDisplay":"`) {
				// Split the string to isolate the part after "priceDisplay":"
				parts := strings.Split(price1[0], `"priceDisplay":"`)
				if len(parts) > 1 {
					// Further split to get just the value before the closing quote
					valueParts := strings.Split(parts[1], `"`)
					if len(valueParts) > 0 {
						//fmt.Println("Extracted Value:", valueParts[0]) // Should print "Now $16.99"
						reg := regexp.MustCompile(`[^\d.]`)
						numericValue := reg.ReplaceAllString(valueParts[0], "")
						fmt.Println("Numeric Value:", numericValue) // Should print "16.99"
						wal.price = numericValue
					} else {
						fmt.Println("No value extracted after priceDisplay")
					}
				} else {
					fmt.Println("No priceDisplay part found in string")
				}
			} else {
				fmt.Println("String does not contain priceDisplay")
			}
		} else {
			fmt.Println("No matches found or result is empty")
			wal.price = "" // 如果是空的，赋值空字符串
		}

		//类目
		category := regexp.MustCompile(`categoryName":"(.+?)",`).FindAllStringSubmatch(result, -1)
		if len(category) == 0 {
			//log.Println("价格获取失败id："+id)
		} else {
			wal.category = strings.Replace(category[0][1], `\u0026`, "&", -1)
		}
		//卖家与配送
		all, err := htmlquery.QueryAll(doc, "//div/div/span[@class=\"lh-title\"]//text()")
		//log.Println(result)
		if err != nil {
			//log.Println("卖家与配送获取失败")
		} else {
			for i, v := range all {
				sv := htmlquery.InnerText(v)
				if strings.Contains(sv, "Sold by") {
					wal.seller = htmlquery.InnerText(all[i+1])
					continue
				}
				if strings.Contains(sv, "Fulfilled by") {
					wal.delivery = strings.Replace(sv, "Fulfilled by ", "", -1)
					if len(wal.delivery) < 3 && len(all) > i+1 {
						wal.delivery = htmlquery.InnerText(all[i+1])
					}
					continue
				}
				if strings.Contains(sv, "Sold and shipped by") {
					wal.seller = htmlquery.InnerText(all[i+1])
					wal.delivery = wal.seller
					break
				}
			}
		}
		if wal.seller == "" {
			seller := regexp.MustCompile("\"sellerDisplayName\":\"(.*?)\"").FindAllStringSubmatch(result, -1)
			if len(seller) > 0 {
				wal.seller = seller[0][1]
			}
		}

		//库存
		stock := regexp.MustCompile(`("message":"Currently out of stock")`).FindAllStringSubmatch(result, -1)
		if len(stock) == 0 {
			wal.stock = "有库存"
		} else {
			wal.stock = "无库存"
		}

		//配送时间
		delivery_date := doc1.Find(".f7.mt1.ws-normal.ttn").Text()
		if delivery_date == "" {
			delivery_date1 := doc1.Find(".ma1.f7").Text()
			wal.deliveryDate = delivery_date1
			log.Println("到达时间", delivery_date1)
		} else {
			wal.deliveryDate = delivery_date
			log.Println("到达时间", delivery_date)
		}
		//变体
		variant := regexp.MustCompile(":</span><span class=\"ml1\">(.*?)</span>").FindAllStringSubmatch(result, -1)
		if len(variant) == 0 {
			//log.Println("评论数量获取失败id："+id)
			log.Println("未找到变体,切换查找方式" + id)
			doc1.Find("div[role='listitem'] label span.w_iUH7").Each(func(i int, s *goquery.Selection) {
				//	直接提取<span class="w_iUH7">内的文本
				spanText := s.Text()
				// 检查文本中是否包含"selected"
				//fmt.Printf(spanText)
				if strings.Contains(spanText, "selected") {
					// 提取并处理包含"selected"的<span>元素的文本
					// 假设"selected"之后是尺寸和价格信息，格式为："selected, S, $3.99"
					parts := strings.Split(spanText, ",")
					log.Println("parts=", parts)
					if len(parts) >= 3 { // 确保分割后的部分至少有3个
						// 通常第二部分是尺寸，第三部分是价格
						size := strings.TrimSpace(parts[1])
						//price := strings.TrimSpace(parts[2])
						fmt.Printf("Selected Size: %s\n", size)
						// 既然已找到含有"selected"的项，可以结束循环
						wal.variant1 = size
						return
					}
					if len(parts) == 2 { // 确保分割后的部分至少有3个
						// 通常第二部分是尺寸，第三部分是价格
						size1 := strings.TrimSpace(parts[1])
						//price := strings.TrimSpace(parts[2])
						fmt.Printf("Selected Size: %s\n", size1)
						// 既然已找到含有"selected"的项，可以结束循环
						wal.variant2 = size1
						return
					}
				}
			})
		} else if len(variant) == 1 {
			wal.variant1 = variant[0][1]
		} else if len(variant) == 2 {
			wal.variant1 = variant[0][1]
			wal.variant2 = variant[1][1]
		}

		allString := regexp.MustCompile("\",\"usItemId\":\"([0-9]+?)\"").FindAllStringSubmatch(result, -1)
		for i := range allString {
			//if v[1]!= vv{
			wal.otherIds = append(wal.otherIds, allString[i][1])
			//}
		}

		img := regexp.MustCompile("<meta property=\"og:image\" content=\"(.*?)\"/>").FindAllStringSubmatch(result, -1)
		if len(img) != 0 {
			wgImg.Add(1)
			if isPi {
				chi <- 1
				go imgxzz(id, img[0][1])
			} else {
				go imgxz(id, img[0][1])
			}

			wal.img = ".\\img\\" + id + ".jpeg"
		}

		log.Println("id:" + wal.id + "完成")
		res = append(res, wal)
		return
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

func imgxzz(id, img string) {
	defer func() {
		<-chi
		wgImg.Done()
	}()
	if len(img) < 10 {
		return
	}
	for i := 0; i < 5; i++ {
		imgPath := ".\\img\\"
		img = strings.Replace(img, "\\u0026", "&", -1)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		proxyUrl, _ := url.Parse("http://l752.kdltps.com:15818")
		tr.Proxy = http.ProxyURL(proxyUrl)
		basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("t19932187800946:wsad123456"))
		tr.ProxyConnectHeader = http.Header{}
		tr.ProxyConnectHeader.Add("Proxy-Authorization", basicAuth)

		client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
		request, err := http.NewRequest("PUT", img, nil)
		if err != nil {
			log.Println("图片下载失败!"+err.Error(), "稍后重新开始下载")
			time.Sleep(3 * time.Second)
			continue
		}
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")
		request.Header.Set("Accept-Language", "zh")
		request.Header.Set("Sec-Ch-Ua", `"Not.A/Brand";v="8", "Chromium";v="114", "Google Chrome";v="114"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", `document`)
		request.Header.Set("Sec-Fetch-Mode", `navigate`)
		request.Header.Set("Sec-Fetch-Site", `none`)
		request.Header.Set("Sec-Fetch-User", `?1`)
		request.Header.Set("Upgrade-Insecure-Requests", `1`)
		request.Header.Set("Accept-Encoding", "gzip, deflate, br")
		response, err := client.Do(request)

		if err != nil {
			log.Println("图片下载失败!"+err.Error(), "稍后重新开始下载")
			time.Sleep(3 * time.Second)
			continue
		}
		dataBytes, err := io.ReadAll(response.Body)
		if err != nil {
			log.Println("图片下载失败!"+err.Error(), "稍后重新开始下载")
			time.Sleep(3 * time.Second)
			continue
		}

		defer response.Body.Close()
		file, err := os.Create(imgPath + id + ".jpeg")
		if err != nil {
			log.Println("图片下载失败!")
			continue
		}
		file.Write(dataBytes)
		log.Println(id, "图片下载成功")
		return
	}
}
