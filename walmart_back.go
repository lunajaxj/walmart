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
	"math/rand"
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

var conf config
var ids []string
var wg = sync.WaitGroup{}
var ch = make(chan int, 8)
var res []Mode

type Mode struct {
	id          string
	color       string
	typex       string
	image       string
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

	xlsx := excelize.NewFile()
	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "itemId", "商品码类型", "商品码值", "标题", "跟卖数量", "运费", "购物车价格", "category", "productName", "productType", "imageUrl", "brand"}); err != nil {
		log.Println(err)
	}
	for _, sv := range ids {
		for _, v := range res {
			if v.id == sv {
				if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v.id, v.itemId, v.typex, v.value, v.name, v.offerCount, v.shipPrice, v.buyBoxPrice, v.category, v.productName, v.productType, v.image, v.brand}); err != nil {
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
		category1 := regexp.MustCompile("category\":(.+?),").FindAllStringSubmatch(result, -1)
		product_name := regexp.MustCompile("productName\":(.+?),").FindAllStringSubmatch(result, -1)
		product_type := regexp.MustCompile("productType\":(.+?),").FindAllStringSubmatch(result, -1)
		img := regexp.MustCompile("image\":(.+?),").FindAllStringSubmatch(result, -1)
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
		if len(category1) > 0 {
			mode.category = category1[0][1]
		}
		if len(product_name) > 0 {
			mode.productName = product_name[0][1]
		}
		if len(product_type) > 0 {
			mode.productType = product_type[0][1]
		}
		if len(img) > 0 {
			mode.image = img[0][1]
		}
		if len(brand1) > 0 {
			mode.brand = brand1[0][1]
		}

		res = append(res, mode)
		fmt.Println("id:" + id + "完成")
		return
	}

}

func generateCookie() string {
	cookieFields := map[string]string{
		"vtc":                             "YjSrVyYa28flCbB5jEci5w",
		"isSeller":                        "true",
		"hubspotutk":                      "53a191650f7a4787018a5abd49e9a11c",
		"_fbp":                            "fb.1.1716889836463.1995548808",
		"QuantumMetricUserID":             "e6f05b84eb1aeffe07fa545bedbf2d5b",
		"_ga":                             "GA1.2.717795621.1717152253",
		"_ga_LBH66B4XCL":                  "GS1.2.1717152253.1.0.1717152253.0.0.0",
		"ACID":                            "388212af-8e82-4a3e-9234-1ac67e9a7e62",
		"hasACID":                         "true",
		"_m":                              "9",
		"_pxvid":                          "f96bf538-2cbc-11ef-904f-92037280f6b6",
		"AID":                             "wmlspartner%3D0%3Areflectorid%3D0000000000000000000000%3Alastupd%3D1718637927505",
		"_ga_QJYGYYQ7NE":                  "GS1.2.1719302775.1.1.1719302785.0.0.0",
		"_sp_id.ad94":                     "b880ef91-58c4-4fc2-9f60-16af868f692b.1718643303.4.1719997224.1718866760.d40388da-35cb-4115-8595-9cadf2bfe920",
		"_gcl_au":                         "1.1.2096921581.1724814257",
		"io_id":                           "cbd7d748-8c9a-4209-a9e0-4944a53885c6",
		"xptwj":                           "uz:35025ce48d044086d5c2:pGRzqa9jgcaJ/76K+t7OTd1TjAYWyF0WUAb3oMdFJftJSHsl2d+KMsJVP6T4kdo9SQgg8jBfpB+Bteiq9OWPfnsYkPtmi8EK/Fe8ZtNpinvZsYO0CqGeVc2Ty5TuVlywyPTSEQQ5GYzkPERS07d5Rhy3umCzJvkN7b+euMmXJctBtUUJNZElz2a7AYTm4RWMmHsMyfpBr93B0elX7NbDQA==",
		"com.wm.reflector":                "\"reflectorid:0000000000000000000000@lastupd:1727094779000@firstcreate:1718637562356\"",
		"if_id":                           "FMEZARSFPGF7byVOdT5W/MFI5COARFAZ1czqw8YQWDepCsPe+4G38FVY3LKWSXryh+GfOc/8cHrEG1PiygEBfHH6mY0lnurU8zOvTlRTnIXrDXOn9KgLflXHr7GnHAtZYF34Jp7HYe7gMhTCCBZhwwZyzKO3d/TiIW4wPuH5hq8G1/wm3KIX3dQAOlrFCdOiHixIfv+EC+2l9N9+",
		"bstc":                            "ZH98cKvlBfDqpz5IfDDPpA",
		"XSRF-T2_MART":                    "US",
		"_auth":                           "MTAyOTYyMDE4o1X006u0+hwPfZ2/BrLkfc3NzkM3EJgqsaD09VoIdVxmNCSov8Rbvj10q4ChUYpoTKHOYQcowzAI8YbRQEdCdp//Fpq7Vk4vGjfUPRKdkKA3AZ8OybnQuZByn/aQ8MKveBBEowoIjNekYh1LFtC6A0Tmgve7fjaeci8VxMduLSkPjAPh+ztfMV+d5RRdQmHJDNbYaWja8Zr5+/cANmndB8NSfmh5GZCXiW64cvzJmpulvaDPvi56q7H4c2JyKogBEyiUsQ+GrHr3SmlqOXPsGLDxrNRryDWFm4qAeCwfUCxqbPQyninOIq6p4/uhAFlnCISylNz3eh7EkArNTgosTmMR1gvrs5DxJH957CPHPYWi3PevtNzTH/RPmgllaNohqQfC9gYloF5MjI/hE6dyMQixD00rrOk74KnRToBqkg4=",
		"TS01817376":                      "01c5a4e2f99dc32459849f16c0987bc06c62d202d3e31ebadae63c480067ce9ec9b251163c27b1a796f91d890f2dab8f033b73a520",
		"JSESSIONID":                      "1b210dc7-e1e3-4733-9270-74f997635724",
		"XSRF-TOKEN":                      "1e2db4e494a6841e213b5e0252aad57ef5975e107dbfd75f8fff6d7c47061106",
		"SC_EXP_FLAVOUR":                  "aurora",
		"SC_SELLER_FLAVOUR":               "allen",
		"SC_GLOBAL_NAV_ENABLED":           "true",
		"SC_MULTIBOX":                     "true",
		"SC_TAX_PROFILE_FLAVOUR":          "aurora",
		"SC_INSURANCE_FLAVOUR":            "aurora",
		"SC_PAYMENT_INFO_FLAVOUR":         "aurora",
		"appstore.seller":                 "true",
		"SC_SELLER_SETTINGS":              "aurora",
		"isbm.enabled.seller":             "true",
		"SC_WFS_FLAVOUR":                  "aurora",
		"IS_MLMQ":                         "aurora",
		"ORDERSV2_REVAMP_SELLERS":         "aurora",
		"sc.manage.items.rollout":         "true",
		"sc.activity.feeds.rollout":       "true",
		"sc.ioh.rollout":                  "true",
		"sc.wfs.multiBox.rollout":         "true",
		"sc.items.aurora.migration":       "true",
		"SC_RATINGS_REVIEWS":              "aurora",
		"SC_ANALYTICS_OVERVIEW_WIDGET":    "true",
		"SHOW_SALES_BY_DEPARTMENT_WIDGET": "true",
		"TS011714b6":                      "01c5a4e2f99dc32459849f16c0987bc06c62d202d3e31ebadae63c480067ce9ec9b251163c27b1a796f91d890f2dab8f033b73a520",
		"ak_bmsc":                         "F7C2AFD887D506129B81943F25DE99F8~000000000000000000000000000000~YAAQxN81Fye6VxGSAQAA34h5LBl4qooHgJHIWQbtkrDdK7yM/KckyqKpb3ywMSv6NXQY66LOwWzH0vjlDq0QYNG8SkQMfV8REEZ3xZ7aRWtOvZLLGi4uZBPpQkw1kSlw2jAa8xDyx7O0LkyzQ4oF7wlbkLsALEuZVJ/LdKdWsowMvdLoAy9OCyXQ2l59h0O1vOu21DpyBe0tAY4lD8oCLWaJN+s4czuROREG6SZ/dz8fOphC0/olurVhKbI88ZZpXvnXJh18Htp+qXRkqPYgL83P0Rt+/4DXNqCtdBvib9fgkTV/9dLhkAYOn2kC5ky6sbq4fW5+nOfErYJ49pzv8g==",
		"bm_sv":                           "2263DA3F2000CB4C844D429CA7B83EB2~YAAQxN81Fye6VxGSAQAA34h5LBl4qooHgJHIWQbtkrDdK7yM/KckyqKpb3ywMSv6NXQY66LOwWzH0vjlDq0QYNG8SkQMfV8REEZ3xZ7aRWtOvZLLGi4uZBPpQkw1kSlw2jAa8xDyx7O0LkyzQ4oF7wlbkLsALEuZVJ/LdKdWsowMvdLoAy9OCyXQ2l59h0O1vOu21DpyBe0tAY4lD8oCLWaJN+s4czuROREG6SZ/dz8fOphC0/olurVhKbI88ZZpXvnXJh18Htp+qXRkqPYgL83P0Rt+/4DXNqCtdBvib9fgkTV/9dLhkAYOn2kC5ky6sbq4fW5+nOfErYJ49pzv8g==",
	}

	cookie := buildCookieString(cookieFields)
	return cookie
}

func buildCookieString(fields map[string]string) string {
	var cookie strings.Builder
	for k, v := range fields {
		cookie.WriteString(fmt.Sprintf("%s=%s; ", k, v))
	}
	return cookie.String()
}

func randomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomBool() string {
	if rand.Intn(2) == 0 {
		return "true"
	}
	return "false"
}

func randomUUID() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", randomString(8), randomString(4), randomString(4), randomString(4), randomString(12))
}

func randomGA() string {
	return fmt.Sprintf("GS1.2.%d.%d.1.0.%d.0.0.0", rand.Int63(), rand.Int63(), rand.Int63())
}

func randomAID() string {
	return fmt.Sprintf("wmlspartner%%3D0%%3Areflectorid%%3D%s%%3Alastupd%%3D%d", randomString(16), rand.Int63())
}

func randomGCL() string {
	return fmt.Sprintf("1.1.%d.%d", rand.Int63(), rand.Int63())
}

func randomAuth() string {
	return randomString(150)
}

// 读取配置文件
func GetConfig(path string) {
	a := generateCookie()
	log.Println("cookie=", a)
	con := &config{}
	if f, err := os.Open(path); err != nil {
		if strings.Contains(err.Error(), "The system cannot find the file specified") || strings.Contains(err.Error(), "no such file or directory") {
			con.Headers = map[string]string{"Host": "seller.walmart.com", "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:105.0) Gecko/20100101 Firefox/105.0", "x-xsrf-token": "", "Cookie": "vtc=YjSrVyYa28flCbB5jEci5w; isSeller=true; _pxvid=ba785ae0-1cd7-11ef-b36f-9ca94e4f80cb; hubspotutk=53a191650f7a4787018a5abd49e9a11c; _fbp=fb.1.1716889836463.1995548808; QuantumMetricUserID=e6f05b84eb1aeffe07fa545bedbf2d5b; _ga=GA1.2.717795621.1717152253; _ga_LBH66B4XCL=GS1.2.1717152253.1.0.1717152253.0.0.0; ACID=388212af-8e82-4a3e-9234-1ac67e9a7e62; hasACID=true; _m=9; _pxvid=f96bf538-2cbc-11ef-904f-92037280f6b6; AID=wmlspartner%3D0%3Areflectorid%3D0000000000000000000000%3Alastupd%3D1718637927505; _ga_QJYGYYQ7NE=GS1.2.1719302775.1.1.1719302785.0.0.0; _sp_id.ad94=b880ef91-58c4-4fc2-9f60-16af868f692b.1718643303.4.1719997224.1718866760.d40388da-35cb-4115-8595-9cadf2bfe920; _gcl_au=1.1.2096921581.1724814257; io_id=cbd7d748-8c9a-4209-a9e0-4944a53885c6; xptwj=uz:35025ce48d044086d5c2:pGRzqa9jgcaJ/76K+t7OTd1TjAYWyF0WUAb3oMdFJftJSHsl2d+KMsJVP6T4kdo9SQgg8jBfpB+Bteiq9OWPfnsYkPtmi8EK/Fe8ZtNpinvZsYO0CqGeVc2Ty5TuVlywyPTSEQQ5GYzkPERS07d5Rhy3umCzJvkN7b+euMmXJctBtUUJNZElz2a7AYTm4RWMmHsMyfpBr93B0elX7NbDQA==; com.wm.reflector=\"reflectorid:0000000000000000000000@lastupd:1727094779000@firstcreate:1718637562356\"; if_id=FMEZARSFPGF7byVOdT5W/MFI5COARFAZ1czqw8YQWDepCsPe+4G38FVY3LKWSXryh+GfOc/8cHrEG1PiygEBfHH6mY0lnurU8zOvTlRTnIXrDXOn9KgLflXHr7GnHAtZYF34Jp7HYe7gMhTCCBZhwwZyzKO3d/TiIW4wPuH5hq8G1/wm3KIX3dQAOlrFCdOiHixIfv+EC+2l9N9+; bstc=ZH98cKvlBfDqpz5IfDDPpA; XSRF-T2_MART=US; _auth=MTAyOTYyMDE4o1X006u0+hwPfZ2/BrLkfc3NzkM3EJgqsaD09VoIdVxmNCSov8Rbvj10q4ChUYpoTKHOYQcowzAI8YbRQEdCdp//Fpq7Vk4vGjfUPRKdkKA3AZ8OybnQuZByn/aQ8MKveBBEowoIjNekYh1LFtC6A0Tmgve7fjaeci8VxMduLSkPjAPh+ztfMV+d5RRdQmHJDNbYaWja8Zr5+/cANmndB8NSfmh5GZCXiW64cvzJmpulvaDPvi56q7H4c2JyKogBEyiUsQ+GrHr3SmlqOXPsGLDxrNRryDWFm4qAeCwfUCxqbPQyninOIq6p4/uhAFlnCISylNz3eh7EkArNTgosTmMR1gvrs5DxJH957CPHPYWi3PevtNzTH/RPmgllaNohqQfC9gYloF5MjI/hE6dyMQixD00rrOk74KnRToBqkg4=; TS01817376=01c5a4e2f99dc32459849f16c0987bc06c62d202d3e31ebadae63c480067ce9ec9b251163c27b1a796f91d890f2dab8f033b73a520; JSESSIONID=1b210dc7-e1e3-4733-9270-74f997635724; XSRF-TOKEN=1e2db4e494a6841e213b5e0252aad57ef5975e107dbfd75f8fff6d7c47061106; SC_EXP_FLAVOUR=aurora; SC_SELLER_FLAVOUR=allen; SC_GLOBAL_NAV_ENABLED=true; SC_MULTIBOX=true; SC_TAX_PROFILE_FLAVOUR=aurora; SC_INSURANCE_FLAVOUR=aurora; SC_PAYMENT_INFO_FLAVOUR=aurora; appstore.seller=true; SC_SELLER_SETTINGS=aurora; isbm.enabled.seller=true; SC_WFS_FLAVOUR=aurora; IS_MLMQ=aurora; ORDERSV2_REVAMP_SELLERS=aurora; sc.manage.items.rollout=true; sc.activity.feeds.rollout=true; sc.ioh.rollout=true; sc.wfs.multiBox.rollout=true; sc.items.aurora.migration=true; SC_RATINGS_REVIEWS=aurora; SC_ANALYTICS_OVERVIEW_WIDGET=true; SHOW_SALES_BY_DEPARTMENT_WIDGET=true; TS011714b6=01c5a4e2f99dc32459849f16c0987bc06c62d202d3e31ebadae63c480067ce9ec9b251163c27b1a796f91d890f2dab8f033b73a520; ak_bmsc=F7C2AFD887D506129B81943F25DE99F8~000000000000000000000000000000~YAAQxN81Fye6VxGSAQAA34h5LBl4qooHgJHIWQbtkrDdK7yM/KckyqKpb3ywMSv6NXQY66LOwWzH0vjlDq0QYNG8SkQMfV8REEZ3xZ7aRWtOvZLLGi4uZBPpQkw1kSlw2jAa8xDyx7O0LkyzQ4oF7wlbkLsALEuZVJ/LdKdWsowMvdLoAy9OCyXQ2l59h0O1vOu21DpyBe0tAY4lD8oCLWaJN+s4czuROREG6SZ/dz8fOphC0/olurVhKbI88ZZpXvnXJh18Htp+qXRkqPYgL83P0Rt+/4DXNqCtdBvib9fgkTV/9dLhkAYOn2kC5ky6sbq4fW5+nOfErYJMsQW1ksjgt0ypVFh8vflavJsBz+LRPpZt5JFvyJQP6/fueQ==; TS0194e2a6=01c5a4e2f99dc32459849f16c0987bc06c62d202d3e31ebadae63c480067ce9ec9b251163c27b1a796f91d890f2dab8f033b73a520; mp_706c95f0b1efdbcfcce0f666821c2237_mixpanel=%7B%22distinct_id%22%3A%20%22xinshenwm%40163.com%22%2C%22%24device_id%22%3A%20%2218fbe9bd2ecfea-0c7808c65e1ead-26001d51-1fa400-18fbe9bd2ecfea%22%2C%22%24user_id%22%3A%20%22xinshenwm%40163.com%22%2C%22Organization%20Name%22%3A%20%22Best%20Choice%22%2C%22Partner%20Type%22%3A%20%22SELLER%22%2C%22Is%20Internal%22%3A%20false%2C%22Role%22%3A%20%22Admin%22%2C%22MP%20V%22%3A%20%22aurora%22%2C%22mart%22%3A%20%22US%22%2C%22onboardingStatus%22%3A%204%2C%22Partner%20Id%22%3A%20%2210001258034%22%2C%22Seller%20Id%22%3A%20%22101238674%22%2C%22Is%20WFS%20Seller%22%3A%20%22true%22%2C%22accountStatus%22%3A%20%22TERMINATED%22%2C%22goLiveDate%22%3A%20%221665830558629%22%2C%22internationalSeller%22%3A%20true%2C%22userLocale%22%3A%20%22en-US%22%2C%22%24initial_referrer%22%3A%20%22https%3A%2F%2Flogin.account.wal-mart.com%2F%22%2C%22%24initial_referring_domain%22%3A%20%22login.account.wal-mart.com%22%2C%22__mps%22%3A%20%7B%7D%2C%22__mpso%22%3A%20%7B%7D%2C%22__mpus%22%3A%20%7B%7D%2C%22__mpa%22%3A%20%7B%7D%2C%22__mpu%22%3A%20%7B%7D%2C%22__mpr%22%3A%20%5B%5D%2C%22__mpap%22%3A%20%5B%5D%2C%22New%20Navigation%22%3A%20true%2C%22isGSE%22%3A%20true%2C%22source%22%3A%20%22wml-analytics%22%7D; pxcts=60c15b97-7bbb-11ef-b5ff-75bffbcce7bc; OptanonConsent=isGpcEnabled=0&datestamp=Thu+Sep+26+2024+11%3A56%3A59+GMT%2B0800+(%E4%B8%AD%E5%9B%BD%E6%A0%87%E5%87%86%E6%97%B6%E9%97%B4)&version=202308.1.0&browserGpcFlag=0&isIABGlobal=false&hosts=&consentId=0cb7e45a-0c65-45a3-b6c9-17ecfcbcbec1&interactionCount=1&landingPath=NotLandingPage&groups=C0007%3A1%2CC0008%3A1%2CC0009%3A1%2CC0010%3A1&AwaitingReconsent=false&geolocation=HK%3B; OptanonAlertBoxClosed=2024-09-26T03:56:59.159Z; TSe809d73e027=08cb8c7367ab20001b700562aa72fc7f269b169905426235c6f97ac9f7b508a4f973709f71967143086ab618b3113000d87fa1389e0e6e54eadafbd53a20c471d14c5fca58907aca26af39ae1d94fdc0df77c38869163ba8a32aa12af67e2cfe; _px3=a59bd6b171b75371e2efb7c0e801784fdd3fafa58421933a205389dfebdc4978:O2kjfkfutiYfj4AFLMK/JG+vKy5Y6peGDN+FbXU7Zp9UisqQIKQ3OEWIYP+zCOUX6R8VxcH3xTQNnaQMC3tJMg==:1000:vj5YngkyqnZ7evk7LxpVMJKXFZiNkvkS9yVOJRa2gW4rlx1drComAQQ7mU9XiAeFzQ/dGyqcAnn9EnbATXIHiRc7eR9A36KgTluRdmNrNIvrFRh7HWMq3eEZK0at1fyEYDjvbDIG2LqS5vkY9uFmreIpYbBmG0bOqvEHHd1R8gljweVA+VWZ5/vH8S/An+0rg/uPOTcG73B3KlxZUFCLMQ5VEqmMxwuDZOI14ReaHKA=; __hs_do_not_track=yes; __hstc=195562739.53a191650f7a4787018a5abd49e9a11c.1716889835731.1727057664206.1727323021677.70; __hssrc=1; __hssc=195562739.1.1727323021677; QuantumMetricSessionID=13b4a922c82e57e2924c4743761b1227; TS24c05192027=0800b316f6ab20002321269cd8f4b40413aa421b74e1bec0dd81e1be5d6ce97662778d9e8490f7ee08b3355a2711300047545e04cf74a88ca1be4edfad87ef4a5e47e43ecbf9338684bfe3d3b81a7a60c8fee8ef1c4f5719aa05e71a9023033c; bm_sv=ED8FF1A12C5EBB59DC7D02000973046B~YAAQxN81F+nJVxGSAQAANr15LBnhFEWA7QlNg9FNjRElNYebk1qjNv5LGyKTktpqBRRTLRO7HCsfrmFAeyZzwqBu2SlIFjB/HOIKjN9jmPthd5+a5Hi+ujtDxhMFOoZ+NEpof7KzCN034m2972Inl79V5R+MNSg61bIzqp8dfYQ+U3WvGozdLan+aC+eVd3qwjRxei9B5K6fQ5uzGB92D40eZ0czPBkGGQ2uWr0a90coaR2loECgduPrx4DD3G8UGg==~1"}
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
