package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
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

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/xuri/excelize/v2"
)

var res []Wal

type Wal struct {
	id       string
	typez    string
	stock    string
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
	crossedPrice string   //划线价
	freeFreight  string   //自发货运费
}

var ids []string
var wg = sync.WaitGroup{}
var ch = make(chan int, 5)

func extractBrand(input string) string {
	// 检查是否包含"/browse/0?facet=brand:"前缀
	if strings.Contains(input, "/browse/0?facet=brand:") {
		return strings.TrimPrefix(input, "/browse/0?facet=brand:")
	}
	// 检查是否包含"/c/brand/"前缀
	if strings.Contains(input, "/c/brand/") {
		return strings.TrimPrefix(input, "/c/brand/")
	}
	return ""
}
func main() {
	log.Println("自动化脚本-walmart-信息采集")
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("id.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 3 {
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

	xlsx := excelize.NewFile()
	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "商品码类型", "商品码值", "品牌", "标签", "标题", "评分", "评论数量", "价格", "卖家", "配送", "变体1", "变体2", "变体id", "到达时间", "库存", "类目", "划线价", "自发货运费"}); err != nil {
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
				if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v.id, v.typez, v.value, v.brand, v.query, v.title, v.score, v.review, v.price, v.seller, v.delivery, v.variant1, v.variant2, other, v.deliveryDate, v.stock, v.category, v.crossedPrice, v.freeFreight}); err != nil {
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

// 文件是否存在
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
	//tr := &http.Transport{
	//	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//}
	//配置代理
	defer func() {
		wg.Done()
		<-ch
	}()
	for i := 0; i < 16; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}
		//proxyUrl, _ := url.Parse("http://a749.kdltps.com:15818")
		//
		//tr.Proxy = http.ProxyURL(proxyUrl)
		//basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("t16545052065610:bxancsry"))
		//tr.ProxyConnectHeader = http.Header{}
		//tr.ProxyConnectHeader.Add("Proxy-Authorization", basicAuth)

		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
		//client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
		request, _ := http.NewRequest("PUT", "https://www.walmart.com/ip/"+id, nil)
		log.Println("https://www.walmart.com/ip/" + id)
		//request, _ := http.NewRequest("GET", "https://www.walmart.com", nil)
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36")
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		request.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
		request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
		request.Header.Set("Cache-Control", "max-age=0")
		request.Header.Set("Sec-Ch-Ua", `"Not)A;Brand";v="99", "Google Chrome";v="127", "Chromium";v="127"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", `document`)
		request.Header.Set("Sec-Fetch-Mode", `navigate`)
		request.Header.Set("Sec-Fetch-Site", `same-origin`)
		request.Header.Set("Sec-Fetch-User", `?1`)
		request.Header.Set("Upgrade-Insecure-Requests", `1`)
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
		wal := Wal{}
		wal.id = id
		if (strings.Contains(result, "This page could not be found.") || strings.Contains(result, "We couldn’t find this page")) && !strings.Contains(result, "h-100 relative") {
			wal.typez = "该商品不存在"
			res = append(res, wal)
			log.Println("id:" + id + "商品不存在")
			return
		}

		//log.Println(result)
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
			continue
		} else {
			wal.value = ""
			wal.typez = "ean"
			//log.Println("id:"+id+" 获取为空，默认为ean")
		}

		doc1, err := htmlquery.Parse(strings.NewReader(result))
		if err != nil {
			log.Println("错误信息：" + err.Error())
			return
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(result))
		if err != nil {
			log.Println("解析HTML错误：", err)
			return
		}

		//品牌
		brand := regexp.MustCompile("\"brandCanonical\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		if len(brand) == 0 {
			log.Println("品牌获取失败id：" + id)
		} else {
			index := extractBrand(brand[0][1])
			if index != "" {
				fmt.Println("Extracted brand:", brand)
			} else {
				fmt.Println("No brand extracted from:", index)
			}
			wal.brand = index
		}

		//标签
		query, err := htmlquery.QueryAll(doc1, "//div[@class='flex items-center mv2 flex-wrap']//span")
		if err != nil {
			log.Println("无标签,错误代码:", err)
		} else {
			queryStr := ""
			for _, v := range query {
				text := strings.TrimSpace(htmlquery.InnerText(v)) // 去掉左右空白字符
				if !strings.Contains(queryStr, text) {
					if queryStr != "" { // 如果 queryStr 不为空，则添加竖线
						queryStr += "|"
					}
					queryStr += text
				}
			}
			wal.query = queryStr
		}

		//if wal.query == "" {
		//	queryf := regexp.MustCompile("2\" aria-hidden=\"false\">(.*?)</span>").FindAllStringSubmatch(result, -1)
		//	queryStr := ""
		//	for _, v := range queryf {
		//		queryStr += v[1] + " "
		//	}
		//	wal.query = queryStr
		//}

		//queryStr := ""
		//
		//doc.Find("#maincontent > section > main > div.flex.flex-column.h-100 > div:nth-child(2) > div > div.w_aoqv.w_wRee.w_fdPt > div > div:nth-child(2) > div > div > div:nth-child(1) > div.flex.items-center.mv2.flex-wrap > div:nth-child(1) > div > span").Each(func(i int, s *goquery.Selection) {
		//	text := s.Text()
		//	if !strings.Contains(queryStr, text) {
		//		queryStr += text + " "
		//	}
		//})
		//wal.query = queryStr

		//标题
		title := regexp.MustCompile("\"name\":\"(.+?)\",").FindAllStringSubmatch(result, -1)
		if len(title) == 0 {
			log.Println("获取失败id："+id, "重新请求")
		} else {
			wal.title = strings.Replace(title[0][1], "\\u0026", "&", -1)
		}

		//库存
		//检查各种可能的库存状态消息
		//outOfStockPattern := regexp.MustCompile(`"message":"(Currently out of stock|Out of stock|Unavailable|Sold out)"`)
		//inStockPattern := regexp.MustCompile(`"message":"(In stock|Available)"`)
		//
		//if outOfStockPattern.MatchString(result) {
		//	wal.stock = "无库存"
		//} else if inStockPattern.MatchString(result) {
		//	wal.stock = "有库存"
		//} else {
		//	wal.stock = "库存状态未知" // 当没有明确的库存信息时
		//}

		stock := regexp.MustCompile(`("message":"Out of stock")`).FindAllStringSubmatch(result, -1)
		if len(stock) == 0 {
			wal.stock = "有库存"
		} else {
			wal.stock = "无库存"
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

		//价格
		//price := regexp.MustCompile("<span itemprop=\"price\".*?.{0,20}(\\$[.,\\d]+).{0,20}?</span>").FindAllStringSubmatch(result, -1)
		//if len(price) == 0 {
		//	//log.Println("价格获取失败id："+id)
		//} else {
		//	wal.price = price[0][1]
		//}
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
		//fulfilled := regexp.MustCompile(">Fulfilled by (.*?)</div>|>Fulfilled by .*?>(.*?)</a>?").FindAllStringSubmatch(result, -1)
		//sold := regexp.MustCompile(">Sold by ([^<]*?)</div>|>Sold by .*?>(.*?)</a>?").FindAllStringSubmatch(result, -1)
		//shipped := regexp.MustCompile("<div>Sold and shipped by ([^/]*?)</div>|<div>Sold and shipped by.*?>(.*?)</a>?").FindAllStringSubmatch(result, -1)
		//if len(fulfilled) != 0 && len(sold) != 0 {
		//	wal.seller = sold[0][1]
		//	wal.delivery = fulfilled[0][1]
		//} else if len(shipped) != 0 {
		//	wal.seller = shipped[0][1]
		//	wal.delivery = shipped[0][1]
		//} else {
		//卖家与配送
		all, err := htmlquery.QueryAll(doc1, "//div/div/span[@class=\"lh-title\"]//text()")
		//log.Println(result)
		if err != nil {
			log.Println("卖家与配送获取失败")
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

		//配送时间
		delivery_date := doc.Find(".f7.mt1.ws-normal.ttn").Text()
		if delivery_date == "" {
			delivery_date1 := doc.Find(".ma1.f7").Text()
			wal.deliveryDate = delivery_date1
			log.Println("到达时间", delivery_date1)
		} else {
			wal.deliveryDate = delivery_date
			log.Println("到达时间", delivery_date)
		}

		//deliveryDate := regexp.MustCompile("\"fulfillmentText\":\"(.+?)\"").FindAllStringSubmatch(result, -1)
		//if len(deliveryDate) == 0 {
		//	//log.Println("配送时间获取失败id："+id)
		//} else {
		//	wal.deliveryDate = deliveryDate[0][1]
		//}
		//deliveryDate, err := htmlquery.QueryAll(doc1, "//*[@id=\"maincontent\"]/section/main/div[2]/div[2]/div/div[1]/div/div[2]/div/div/div[9]/section/div/div/div/div/div[1]/button/label/div[3]")
		//wal.deliveryDate = deliveryDate
		//for _, index := range childIndices {
		// 为每个索引构建完整的选择器
		//fullSelector := fmt.Sprintf("%s%d%s", baseSelector, index, remainderSelector)
		//deliveryDate, err := htmlquery.QueryAll(doc1, "")
		// iconContainer := doc.Find("div[data-testid='shipping-tile']").First()
		// if iconContainer != nil {
		// 	thirdDivLabel := iconContainer.Find("div:nth-child(3) label").First()
		// 	if thirdDivLabel.Length() > 0 {
		// 		log.Println("thirdDivLabel: " + thirdDivLabel.Text())
		// 		wal.deliveryDate = thirdDivLabel.Text()
		// 	} else {
		// 		log.Println("deliveryDate not found")
		// 	}
		// }
		//devliveryMatches := regexp.MustCompile(`shippingText":"Shipping","fulfillmentText":"([\w\W]*?)"`).FindAllStringSubmatch(result, -1)
		//if len(devliveryMatches) > 0 { // 确保有匹配且有分组
		//	//fmt.Println("找到的划线价格是:", matches[1]) // $12.99，matches[1]是第一个括号内的匹配
		//	deliveryDate := devliveryMatches[0][1]
		//	seps := strings.Split(deliveryDate, ",")
		//	if len(seps) > 1 {
		//		devliveryStr := strings.TrimSpace(seps[len(seps)-1])
		//
		//		// devDate, err := time.Parse("Jan 2", devliveryStr)
		//		// if err != nil {
		//		// 	fmt.Println("Error parsing date:", err)
		//		// 	return
		//		// }
		//
		//		// 获取当前年份
		//		currentYear := time.Now().Year()
		//
		//		// 新增年份信息，组合成完整日期
		//		completeDateStr := fmt.Sprintf("%d %s", currentYear, devliveryStr)
		//		completeDate, err := time.Parse("2006 Jan 2", completeDateStr)
		//		if err != nil {
		//			fmt.Println("Error parsing complete date:", err)
		//			return
		//		}
		//
		//		// 向后新增一天
		//		newDate := completeDate.AddDate(0, 0, 1)
		//		fmt.Println("New date:", newDate.Format("Jan 2, 2006"))
		//		wal.deliveryDate = newDate.Format("Jan 2, 2006")
		//	} else {
		//		fmt.Println("sep error to ", deliveryDate)
		//	}
		//} else {
		//	fmt.Println("没有找到匹配的划线价格")
		//}
		//log.Println("deliveryDate" + wal.deliveryDate)
		// wal.deliveryDate = deliveryDate

		//划线价
		crossedPrice := `<span aria-hidden="true" data-seo-id="strike-through-price" class="mr2 f6 gray strike">(.*?)</span>`
		re := regexp.MustCompile(crossedPrice)

		// 在整个HTML内容中查找匹配的价格
		matches := re.FindStringSubmatch(string(result))

		if len(matches) > 1 { // 确保有匹配且有分组
			//fmt.Println("找到的划线价格是:", matches[1]) // $12.99，matches[1]是第一个括号内的匹配
			wal.crossedPrice = matches[1]
		} else {
			fmt.Println("没有找到匹配的划线价格")
		}

		//自发货运费
		//选择器的基础部分，直到可变的子索引
		//baseSelector := "#maincontent > section > main > div.flex.undefined.flex-column.h-100 > div:nth-child(2) > div > div.w_aoqv.w_wRee.w_fdPt > div > div:nth-child(2) > div > div > div:nth-child("
		//// 选择器的剩余部分，子索引之后
		//remainderSelector := ") > section > div > div > div > div > button:nth-child(1) > label > div.mt1.h1 > div"
		//
		//// 子索引列表
		//childIndices := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
		//
		//freeFreightStr := ""
		//for _, index := range childIndices {
		//	// 为每个索引构建完整的选择器
		//	fullSelector := fmt.Sprintf("%s%d%s", baseSelector, index, remainderSelector)
		//	doc.Find(fullSelector).Each(func(i int, s *goquery.Selection) {
		//		text := s.Text()
		//		if !strings.Contains(freeFreightStr, text) {
		//			freeFreightStr += text + " "
		//		}
		//	})
		//}

		// XPath选择器
		// 可能的索引数组
		//indices := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30}
		freeFreightStr := ""

		doc.Find(".mt1.h1 .f7").Each(func(index int, s *goquery.Selection) {
			text := s.Text()
			if !strings.Contains(freeFreightStr, text) {
				freeFreightStr += text + " "
			}
		})

		// 假设wal是你要赋值的结构体，这里将获取的字符串赋值给它的freeFreight字段
		wal.freeFreight = freeFreightStr
		//log.Println("提取的文本:", freeFreightStr)

		//变体
		variant := regexp.MustCompile(":</span><span class=\"ml1\">(.*?)</span>").FindAllStringSubmatch(result, -1)
		if len(variant) == 0 {
			//log.Println("评论数量获取失败id："+id)
			log.Println("未找到变体,切换查找方式" + id)
			doc.Find("div[role='listitem'] label span.w_iUH7").Each(func(i int, s *goquery.Selection) {
				//	直接提取<span class="w_iUH7">内的文本
				spanText := s.Text()
				// 检查文本中是否包含"selected"
				fmt.Printf(spanText)
				if strings.Contains(spanText, "selected") {
					// 提取并处理包含"selected"的<span>元素的文本
					// 假设"selected"之后是尺寸和价格信息，格式为："selected, S, $3.99"
					parts := strings.Split(spanText, ",")
					if len(parts) >= 3 { // 确保分割后的部分至少有3个
						// 通常第二部分是尺寸，第三部分是价格
						size := strings.TrimSpace(parts[1])
						//price := strings.TrimSpace(parts[2])
						fmt.Printf("Selected Size: %s\n", size)
						// 既然已找到含有"selected"的项，可以结束循环
						wal.variant1 = size
						return
					}
				}
			})
			//doc.Find("label span[aria-hidden='true']").Each(func(i int, s *goquery.Selection) {
			//	// 直接提取<span>内的文本
			//	spanText := s.Text()
			//	fmt.Printf("Found: %s\n", spanText)
			//	wal.variant1 = spanText
			//})

		} else if len(variant) == 1 {
			wal.variant1 = variant[0][1]
		} else if len(variant) == 2 {
			wal.variant1 = variant[0][1]
			wal.variant2 = variant[1][1]
		}

		// 查找所有符合条件的元素并处理

		allString := regexp.MustCompile("\",\"usItemId\":\"([0-9]+?)\"").FindAllStringSubmatch(result, -1)
		for i := range allString {
			//if v[1]!= vv{
			wal.otherIds = append(wal.otherIds, allString[i][1])
			//}
		}
		fmt.Println("https://www.walmart.com/ip/" + wal.id)
		log.Println("id:" + wal.id + "完成")

		res = append(res, wal)
		return

	}

}
