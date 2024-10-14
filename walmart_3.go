package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/xuri/excelize/v2"
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

var res []Wal

type Wal struct {
	id       string
	typez    string
	stock    string //库存
	value    string //商品码值
	brand    string //"name":"VicTsing"},"offers":{  品牌
	query    string //*[@id="maincontent"]//div[@data-testid="sticky-buy-box"]//div/p//span//text()  标签，多行相加
	title    string //"name":"Gymax 5 Piece Dining Set Glass Top Table & 4 Upholstered Chairs Kitchen Room Furniture","sku":  标题
	score    string //(4.5)  评分
	review   string //"totalReviewCount":1187}  评论数量
	price    string //aria-hidden="false">$22.98<	价格
	category string
	////div/div/span[@class="lh-title"]//text()  卖家+配送
	seller            string   //卖家
	delivery          string   //配送
	deliveryDate      string   //配送时间
	variant1          string   //变体1 :</span><span aria-hidden="true" class="ml1">(.*?)</span>
	variant2          string   //变体1 :</span><span aria-hidden="true" class="ml1">(*?)</span>
	otherIds          []string //变体id
	startingFrom      string   //>Starting from \$([^<]+)<
	moreSellerOptions string   //More seller options \((\d+)\)
	availableQuantity string
	crossedPrice      string //划线价
	freeFreight       string //自发货运费
}

var idStores []string
var ids []string
var wg = sync.WaitGroup{}
var ch = make(chan int, 10)

func main() {
	log.Println("自动化脚本-walmart-信息采集-跟卖库存信息")
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("id_storeid.txt")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer fi.Close()

	r := bufio.NewReader(fi) // 创建 Reader

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Read error: %v", err)
			}
			break
		}
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			parts := strings.Split(line, "|")
			if len(parts) == 2 {
				id := strings.TrimSpace(parts[0])
				idStore := strings.TrimSpace(parts[1])
				ids = append(ids, id)
				idStores = append(idStores, idStore)
				log.Printf("Read line: id=%s, idStore=%s", id, idStore)
			} else {
				log.Printf("Invalid line format: %s", line)
			}
		} else {
			log.Println("Empty line encountered.")
		}
	}

	if len(ids) == 0 {
		log.Println("No IDs found in the file.")
		return
	}

	log.Println("IDs and ID Stores:", ids, idStores)

	for i := range ids {
		ch <- 1
		wg.Add(1)
		go crawler(ids[i], idStores[i])
	}

	wg.Wait()

	xlsx := excelize.NewFile()
	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "卖家", "配送", "库存数量"}); err != nil {
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
				if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v.id, v.seller, v.delivery, v.availableQuantity}); err != nil {
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

func crawler(id string, id_store string) {
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
		id_store = strings.TrimSpace(id_store)
		if id_store == "" {
			request, _ := http.NewRequest("PUT", "https://www.walmart.com/ip/"+id+"?&selectedSellerId=", nil)
			//request, _ := http.NewRequest("GET", "https://www.walmart.com", nil)
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
			var isc = IsC
			if IsC {
				//request.Header.Set("Cookie", generateRandomString(10))
			}
			response, err := client.Do(request)
			if err != nil {
				if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
					log.Println("错误代码打印：" + err.Error())
					log.Println("等待请求头超时，重新开始当前ID：" + id)
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

			//获取result
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
			if strings.Contains(result, "This page could not be found.") {
				wal.typez = "该商品不存在"
				res = append(res, wal)
				log.Println("id:" + id + "商品不存在")
				return
			}

			fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)

			if len(fk) > 0 {
				log.Println("id:" + id + " 被风控,更换IP继续")
				IsC = !isc
				continue
			}

			doc1, err := htmlquery.Parse(strings.NewReader(result))
			if err != nil {
				log.Println("错误信息：" + err.Error())
				return
			}

			all, err := htmlquery.QueryAll(doc1, "//div/div/span[@class=\"lh-title\"]//text()")
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

			availableQuantity := regexp.MustCompile("availableQuantity\":(\\d+),").FindAllStringSubmatch(result, -1)
			if len(availableQuantity) > 0 {
				wal.availableQuantity = availableQuantity[0][1]
			}
			log.Println("id:" + wal.id + "完成")

			res = append(res, wal)
			return

		} else {
			request, _ := http.NewRequest("PUT", "https://www.walmart.com/ip/"+id+"?&selectedSellerId="+id_store, nil)
			//request, _ := http.NewRequest("GET", "https://www.walmart.com", nil)
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
			var isc = IsC
			if IsC {
				//request.Header.Set("Cookie", generateRandomString(10))
			}
			response, err := client.Do(request)
			if err != nil {
				if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
					log.Println("错误代码打印：" + err.Error())
					log.Println("等待请求头超时，重新开始当前ID：" + id)
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

			//获取result
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
			fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)

			if len(fk) > 0 {
				log.Println("id:" + id + " 被风控,更换IP继续")
				IsC = !isc
				continue
			}
			if strings.Contains(result, "This page could not be found.") {
				wal.typez = "该商品不存在"
				res = append(res, wal)
				log.Println("id:" + id + "商品不存在")
				return
			}

			doc1, err := htmlquery.Parse(strings.NewReader(result))
			if err != nil {
				log.Printf("Failed to parse HTML for id %s: %v", id, err)
				continue
			}

			all, err := htmlquery.QueryAll(doc1, "//div/div/span[@class=\"lh-title\"]//text()")
			if err != nil {
				log.Printf("Failed to get seller and delivery info for id %s", id)
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

			availableQuantity := regexp.MustCompile("availableQuantity\":(\\d+),").FindAllStringSubmatch(result, -1)
			if len(availableQuantity) > 0 {
				wal.availableQuantity = availableQuantity[0][1]
			}

			log.Println("id:" + wal.id + "完成")

			res = append(res, wal)
			return
		}

	}

}

// insertDecimalPoint 添加适当的小数点位置
func insertDecimalPoint(price string) string {
	pos := strings.Index(price, "$") + 1
	return price[:pos+2] + "." + price[pos+2:]
}
