package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zlib"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mu sync.Mutex
var maxRetries = 2 // 设置最大重试次数

type Response struct {
	Data struct {
		Product struct {
			AllOffers []Offer `json:"allOffers"`
		} `json:"product"`
	} `json:"data"`
}
type Image struct {
	ID  *string `json:"id"`  // 用指针类型来处理可能为 null 的 ID
	URL string  `json:"url"` // URL 一般是必填的 string 类型
}
type Offer struct {
	UsItemID         string   `json:"usItemId"`
	OfferLevelBadges []string `json:"offerLevelBadges"`
	ImageInfo        struct {
		ThumbnailUrl string  `json:"thumbnailUrl"`
		AllImages    []Image `json:"allImages"`
	} `json:"imageInfo"`
	OfferID            string `json:"offerId"`
	OfferType          string `json:"offerType"`
	AvailabilityStatus string `json:"availabilityStatus"`
	FulfillmentType    string `json:"fulfillmentType"`
	FulfillmentOptions []struct {
		Type              string `json:"type"`
		AvailableQuantity int    `json:"availableQuantity"`
		DeliveryDate      string `json:"deliveryDate"`
	} `json:"fulfillmentOptions"`
	SellerID          string `json:"sellerId"`
	CatalogSellerID   int    `json:"catalogSellerId"`
	LmpEligible       bool   `json:"lmpEligible"`
	SellerName        string `json:"sellerName"`
	SellerDisplayName string `json:"sellerDisplayName"`
	SellerType        string `json:"sellerType"`
	PriceInfo         struct {
		CurrentPrice struct {
			Price       float64 `json:"price"`
			PriceString string  `json:"priceString"`
		} `json:"currentPrice"`
		WasPrice struct {
			Price       float64 `json:"price"`
			PriceString string  `json:"priceString"`
		} `json:"wasPrice"`
	} `json:"priceInfo"`
	ReturnPolicy struct {
		ReturnPolicyText string `json:"returnPolicyText"`
		Returnable       bool   `json:"returnable"`
		FreeReturns      bool   `json:"freeReturns"`
		ReturnWindow     struct {
			Value    int    `json:"value"`
			UnitType string `json:"unitType"`
		} `json:"returnWindow"`
	} `json:"returnPolicy"`
	ShippingOption struct {
		AvailabilityStatus string `json:"availabilityStatus"`
		DeliveryDate       string `json:"deliveryDate"`
		ShipPrice          struct {
			Price       float64 `json:"price"`
			PriceString string  `json:"priceString"`
		} `json:"shipPrice"`
	} `json:"shippingOption"`
}

// 从文件中读取 Cookie
func readCookieFromFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 假设文件中至少有一行 cookie
	reader := bufio.NewReader(file)
	for {
		cookie, err := reader.ReadString('\n')
		if err == io.EOF {
			if len(cookie) == 0 {
				return "", fmt.Errorf("file is empty")
			}
			cookie = strings.TrimSpace(cookie)
			return cookie, nil
		}
		if err != nil {
			return "", err
		}
		cookie = strings.TrimSpace(cookie)
		if len(cookie) > 0 {
			return cookie, nil
		}
	}
}

var ids []string
var wg = sync.WaitGroup{}
var ch = make(chan int, 1) // 控制并发数量，降低请求频率

func main() {
	// 创建日志文件
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer logFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	log.Println("walmart_Compareallsellers")
	log.Println("开始执行...")

	// 打开 Excel 文件或创建新文件
	var fileName = "out.xlsx"
	var file *excelize.File
	if exists(fileName) {
		file, err = excelize.OpenFile(fileName)
		if err != nil {
			log.Fatalf("Failed to open existing Excel file: %v", err)
		}
	} else {
		file = excelize.NewFile()
		if err := file.SetSheetRow("Sheet1", "A1", &[]interface{}{
			"UsItemID", "OfferLevelBadges", "ThumbnailUrl", "OfferID", "OfferType", "AvailabilityStatus", "FulfillmentType",
			"AvailableQuantity", "DeliveryDate", "SellerID", "CatalogSellerID", "LmpEligible", "SellerName", "SellerDisplayName",
			"SellerType", "CurrentPrice", "WasPrice", "ReturnPolicyText", "Returnable", "FreeReturns", "ReturnWindow", "ShippingPrice", "ShippingAvailability", "DeliveryDate",
		}); err != nil {
			log.Println(err)
		}
	}

	// 读取ID列表
	fi, err := os.Open("id.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi)

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 3 {
			ids = append(ids, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}
	}

	if len(ids) == 0 {
		log.Println("No IDs found in the file.")
		return
	}

	// 遍历ID
	for _, v := range ids {
		ch <- 1
		wg.Add(1)
		go crawler(v, file, fileName)
		time.Sleep(1000)
	}

	wg.Wait()

	log.Println("完成")
}

// 文件是否存在
func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func crawler(id string, file *excelize.File, fileName string) {
	defer func() {
		wg.Done()
		<-ch
	}()

	// 在请求发送之前增加随机延时
	delay := time.Duration(rand.Intn(5)+1) * time.Second
	log.Printf("等待 %v 秒...", delay)
	time.Sleep(delay)
	retries := 0
	for retries < maxRetries {
		proxyStr := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxyStr)

		client := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy), DisableKeepAlives: true, TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

		// 构造请求的 URL
		url := "https://www.walmart.com/orchestra/home/graphql/GetAllSellerOffers/e0f55408c56e247738f13b11ac88a92b4c9c433f55f4666ddf8124ed4127b51d"

		// 构造请求参数
		params := fmt.Sprintf(`{"itemId":"%s","isSubscriptionEligible":true}`, id)

		// 创建请求
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("Error creating request, ID: %s 错误: %v", id, err)
			return
		}

		// 设置请求头信息
		q := req.URL.Query()
		q.Add("variables", params)
		req.URL.RawQuery = q.Encode()
		cookies := []string{
			fmt.Sprintf("abqme=%t", rand.Intn(2) == 0),
			fmt.Sprintf("vtc=%s", randomString(22)),
			fmt.Sprintf("_pxhd=%s:%s", randomString(64), randomString(36)),
			fmt.Sprintf("ACID=%s", randomString(36)),
			fmt.Sprintf("_m=%d", rand.Intn(10)),
			"hasACID=true",
			fmt.Sprintf("_pxvid=%s", randomString(36)),
			fmt.Sprintf("AID=wmlspartner%%3D0%%3Areflectorid%%3D0000000000000000000000%%3Alastupd%%3D%d", time.Now().Unix()),
			fmt.Sprintf("_sp_id.ad94=%s.%d.%d.%d.%d.%s", randomString(36), time.Now().Unix(), rand.Intn(10), time.Now().Unix(), time.Now().Unix(), randomString(36)),
			fmt.Sprintf("mdLogger=%t", rand.Intn(2) == 0),
			fmt.Sprintf("kampyle_userid=%s", randomString(36)),
			fmt.Sprintf("kampyleUserSession=%d", time.Now().UnixNano()),
			fmt.Sprintf("kampyleUserSessionsCount=%d", rand.Intn(5)+1),
			fmt.Sprintf("kampyleSessionPageCounter=%d", rand.Intn(10)+1),
			fmt.Sprintf("_gcl_au=1.1.%d.%d", rand.Int63(), time.Now().Unix()),
			fmt.Sprintf("_uetvid=%s", randomString(36)),
			fmt.Sprintf("locGuestData=%s", randomString(128)),
		}
		cookieString := strings.Join(cookies, "; ")
		fmt.Println(cookieString)
		//从文件读取 cookie
		//cookie, err := readCookieFromFile("cookie_Compareallsellers.txt")
		//if err != nil {
		//	fmt.Println("Error reading cookie from file:", err)
		//	return
		//}
		deviceProfileRefID := generateDeviceProfileRefID()
		headers := map[string]string{
			"accept":                  "application/json",
			"accept-encoding":         "gzip, deflate, br, zstd",
			"accept-language":         "en-US",
			"baggage":                 "trafficType=customer,deviceType=desktop,renderScope=SSR,webRequestSource=Browser,pageName=itemPage,isomorphicSessionId=qwyUBqLjLC-pvHA9_ADKb",
			"content-type":            "application/json",
			"cookie":                  cookieString,
			"device_profile_ref_id":   deviceProfileRefID,
			"downlink":                "1.45",
			"dpr":                     "1.5",
			"priority":                "u=1, i",
			"referer":                 fmt.Sprintf("https://www.walmart.com/ip/%s?classType=VARIANT&adsRedirect=true", id),
			"sec-ch-ua":               "\"Google Chrome\";v=\"129\", \"Not=A?Brand\";v=\"8\", \"Chromium\";v=\"129\"",
			"sec-ch-ua-mobile":        "?0",
			"sec-ch-ua-platform":      "\"Windows\"",
			"sec-fetch-dest":          "empty",
			"sec-fetch-mode":          "cors",
			"sec-fetch-site":          "same-origin",
			"tenant-id":               "elh9ie",
			"traceparent":             "00-17fc215b77992e5b092790910c441298-c61f7967d4c6af86-00",
			"user-agent":              "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
			"wm_mp":                   "true",
			"wm_page_url":             fmt.Sprintf("https://www.walmart.com/ip/%s?classType=VARIANT&adsRedirect=true", id),
			"wm_qos.correlation_id":   "XwkAEoWblpc3UsP8jzAfcsquhBqFz_md3twq",
			"x-apollo-operation-name": "GetAllSellerOffers",
			"x-enable-server-timing":  "1",
			"x-latency-trace":         "1",
			"x-o-bu":                  "WALMART-US",
			"x-o-ccm":                 "server",
			"x-o-correlation-id":      "XwkAEoWblpc3UsP8jzAfcsquhBqFz_md3twq",
			"x-o-gql-query":           "query GetAllSellerOffers",
			"x-o-mart":                "B2C",
			"x-o-platform":            "rweb",
			"x-o-platform-version":    "us-web-1.166.0-edb00ff3c95e4c8122d22b595a8b3d39d6e3b177-100519",
			"x-o-segment":             "oaoh",
		}

		// 设置请求头
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		// 发起请求
		response, err := client.Do(req)
		if err != nil {
			if strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") {
				log.Printf("请求 ID: %s 超时, 尝试更换IP并重试...", id)
				retries++
				time.Sleep(time.Second * 5)
				continue
			} else {
				log.Printf("请求 ID: %s 失败, 错误: %v", id, err)
				return
			}
		}
		defer response.Body.Close()

		// 检查是否触发风控
		if response.StatusCode == http.StatusTooManyRequests {
			log.Printf("ID: %s 被风控, 尝试更换IP并重试...", id)
			retries++
			time.Sleep(time.Second * 5)
			continue
		}

		// 检查响应内容编码并解压
		var reader io.ReadCloser
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(response.Body)
			if err != nil {
				log.Printf("解压 GZIP 响应失败, ID: %s 错误: %v", id, err)
				return
			}
			defer reader.Close()
		case "deflate":
			reader, err = zlib.NewReader(response.Body)
			if err != nil {
				log.Printf("解压 DEFLATE 响应失败, ID: %s 错误: %v", id, err)
				return
			}
			defer reader.Close()
		default:
			reader = response.Body
		}

		// 读取响应内容
		body, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("读取响应失败, ID: %s 错误: %v", id, err)
			return
		}
		// 转换为字符串格式的响应内容
		bodyStr := string(body)
		log.Printf("响应内容: %s", bodyStr)

		// 检查响应内容是否包含 "Robot or human?"，如果包含，则算作风控
		if strings.Contains(bodyStr, "blocked?") {
			log.Printf("ID: %s 被风控, 响应包含 'Robot or human?'，尝试更换IP并重试...", id)
			retries++
			time.Sleep(time.Second * 5)
			row := []interface{}{
				id, "被风控，仅写入id", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
			}
			appendToExcel(file, fileName, row)
			log.Printf("ID: %s 被风控，仅写入 ID", id)
			break
		}

		// 解析 JSON 响应（如果不是风控）
		var result Response
		if err := json.Unmarshal(body, &result); err != nil {
			if strings.HasPrefix(bodyStr, "<") {
				log.Printf("ID: %s 返回 HTML 内容，可能触发风控，尝试更换IP并重试...", id)
				retries++
				time.Sleep(time.Second * 5)
				row := []interface{}{
					id, "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
				}
				appendToExcel(file, fileName, row)
				log.Printf("ID: %s 返回 HTML 内容，可能触发风控，尝试更换IP并重试，仅写入 ID", id)
				break
			}
			log.Printf("解析 JSON 失败, ID: %s 错误: %v", id, err)
			return
		}
		// 添加判断逻辑，如果 message=404，出现 UNPUBLISHED，或者 data 是空的，则记录 id 并跳过
		if result.Data.Product.AllOffers == nil || len(result.Data.Product.AllOffers) == 0 || strings.Contains(bodyStr, "UNPUBLISHED") || strings.Contains(bodyStr, "\"message\":\"404\"") {
			// 记录 ID 但其他列留空，只写入一次
			row := []interface{}{
				id, "空", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
			}
			appendToExcel(file, fileName, row)
			log.Printf("ID: %s 返回的数据无效，仅写入 ID", id)
			break
		}

		// 遍历所有 offer，并写入 Excel 文件
		offersWritten := false // 标志是否已经写入了数据行
		for _, offer := range result.Data.Product.AllOffers {
			row := []interface{}{
				offer.UsItemID, offer.OfferLevelBadges, offer.ImageInfo.ThumbnailUrl, offer.OfferID, offer.OfferType, offer.AvailabilityStatus, offer.FulfillmentType,
				offer.FulfillmentOptions[0].AvailableQuantity, offer.FulfillmentOptions[0].DeliveryDate, offer.SellerID, offer.CatalogSellerID, offer.LmpEligible, offer.SellerName, offer.SellerDisplayName,
				offer.SellerType, offer.PriceInfo.CurrentPrice.Price, offer.PriceInfo.WasPrice.Price, offer.ReturnPolicy.ReturnPolicyText, offer.ReturnPolicy.Returnable, offer.ReturnPolicy.FreeReturns, offer.ReturnPolicy.ReturnWindow.Value,
				offer.ShippingOption.ShipPrice.Price, offer.ShippingOption.AvailabilityStatus, offer.ShippingOption.DeliveryDate,
			}

			// 写入 Excel
			appendToExcel(file, fileName, row)
			log.Printf("ID: %s 的数据写入成功", id)
			offersWritten = true // 标记为 true，表明已经有数据写入
		}
		// 如果没有写入任何 offer 数据
		if !offersWritten {
			row := []interface{}{id, "空", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""}
			appendToExcel(file, fileName, row)
			log.Printf("ID: %s 无有效 offer，仅写入 ID", id)
			break
		}
		break
		//if retries >= maxRetries {
		//	log.Printf("ID: %s 重试次数过多, 放弃处理", id)
		//}
	}
}

func appendToExcel(file *excelize.File, fileName string, row []interface{}) {
	mu.Lock()
	defer mu.Unlock()

	sheetName := "Sheet1"
	rows, err := file.GetRows(sheetName)
	if err != nil {
		log.Println("Failed to get rows from Excel file:", err)
		return
	}
	num := len(rows) + 1

	if err := file.SetSheetRow(sheetName, "A"+strconv.Itoa(num), &row); err != nil {
		log.Println("Failed to set sheet row:", err)
	}

	if err := file.SaveAs(fileName); err != nil {
		log.Println("Failed to save Excel file:", err)
	}
}
func randomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// 生成随机的伪装 device_profile_ref_id
func generateDeviceProfileRefID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 48
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
