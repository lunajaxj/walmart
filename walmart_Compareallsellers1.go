package main

import (
	"bufio"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
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
	// Seed the random number generator for generating random strings
	rand.Seed(time.Now().UnixNano())
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

		//cookieString := strings.Join(cookies, "; ")
		//fmt.Println(cookieString)
		//从文件读取 cookie
		//cookie, err := readCookieFromFile("cookie_Compareallsellers.txt")
		//if err != nil {
		//	fmt.Println("Error reading cookie from file:", err)
		//	return
		//}
		traceparent := generateTraceparent()
		correlationID := generateCorrelationID()
		cookie := generateCookie()
		headers := map[string]string{
			"accept":                  "application/json",
			"accept-encoding":         "gzip, deflate, br, zstd",
			"accept-language":         "en-US",
			"baggage":                 fmt.Sprintf("trafficType=customer,deviceType=desktop,renderScope=SSR,webRequestSource=Browser,pageName=itemPage,isomorphicSessionId=%s", generateSessionID()),
			"content-type":            "application/json",
			"cookie":                  cookie,
			"device_profile_ref_id":   "9qxhmmujuwprwslpsrptltmc6-4crzcjbri-",
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
			"traceparent":             traceparent,
			"user-agent":              "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
			"wm_mp":                   "true",
			"wm_page_url":             fmt.Sprintf("https://www.walmart.com/ip/%s?classType=VARIANT&adsRedirect=true", id),
			"wm_qos.correlation_id":   "XwkAEoWblpc3UsP8jzAfcsquhBqFz_md3twq",
			"x-apollo-operation-name": "GetAllSellerOffers",
			"x-enable-server-timing":  "1",
			"x-latency-trace":         "1",
			"x-o-bu":                  "WALMART-US",
			"x-o-ccm":                 "server",
			"x-o-correlation-id":      correlationID,
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
			continue
		}

		// 解析 JSON 响应（如果不是风控）
		var result Response
		if err := json.Unmarshal(body, &result); err != nil {
			if strings.HasPrefix(bodyStr, "<") {
				log.Printf("ID: %s 返回 HTML 内容，可能触发风控，尝试更换IP并重试...", id)
				retries++
				time.Sleep(time.Second * 5)
				continue
			}
			log.Printf("解析 JSON 失败, ID: %s 错误: %v", id, err)
			return
		}

		// 遍历所有 offer，并写入 Excel 文件
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
		}
		break
		if retries >= maxRetries {
			log.Printf("ID: %s 重试次数过多, 放弃处理", id)
		}
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

// Generate a random traceparent header in the correct format
func generateTraceparent() string {
	traceID := uuid.New().String() // Generate a unique trace ID
	spanID := make([]byte, 8)
	rand.Read(spanID) // Fill spanID with random bytes
	spanIDHex := hex.EncodeToString(spanID)
	return fmt.Sprintf("00-%s-%s-00", strings.ReplaceAll(traceID, "-", ""), spanIDHex)
}

// Generate a random correlation ID for tracking
func generateCorrelationID() string {
	uuidWithHyphen := uuid.New().String()
	return strings.ReplaceAll(uuidWithHyphen, "-", "")
}

// Generate a random session ID
func generateSessionID() string {
	sessionIDBytes := make([]byte, 12)
	rand.Read(sessionIDBytes)
	return hex.EncodeToString(sessionIDBytes)
}

// Generate a random cookie value based on given format
func generateCookie() string {
	userID := uuid.New().String()     // Simulate a user ID
	distinctID := uuid.New().String() // Simulate a distinct ID
	deviceID := uuid.New().String()   // Simulate a device ID
	//sessionID := generateSessionID()  // Simulate a session ID
	timestamp := time.Now().Unix() // Use current timestamp

	cookieTemplate := `QuantumMetricUserID=%s; mp_fdf49de7a057c741e9f953436f190f81_mixpanel={"distinct_id":"$device:%s","$device_id":"%s","$initial_referrer":"$direct","$initial_referring_domain":"$direct"}; vtc=SCyT2c7Dn84INZj0O7bXJs; isSeller=true; hubspotutk=7c34be9e831f1731b988550eaf43b105; __hssrc=1; _gcl_au=1.1.1374205395.%d; __hs_do_not_track=yes; TS0194e2a6=01c5a4e2f9696fff91072389b869cc9aa1743ba5ec173a3025500573b486a78961cda3b48a39fd475535162a66802f8532051387d7; TS011714b6=01c5a4e2f9696fff91072389b869cc9aa1743ba5ec173a3025500573b486a78961cda3b48a39fd475535162a66802f8532051387d7; OptanonConsent=isGpcEnabled=0&datestamp=Wed+Oct+09+2024+09:18:16+GMT+0800+(%E4%B8%AD%E5%9B%BD%E6%A0%87%E5%87%86%E6%97%B6%E9%97%B4)&version=202308.1.0&browserGpcFlag=0&isIABGlobal=false&hosts=&consentId=9589238c-8f8a-46e3-9489-cc2f66d96adc&interactionCount=0&landingPath=NotLandingPage&groups=C0007:1,C0008:1,C0009:1,C0010:1&AwaitingReconsent=false&geolocation=CN:GD; OptanonAlertBoxClosed=2024-10-09T01:18:16.278Z; mp_706c95f0b1efdbcfcce0f666821c2237_mixpanel={"distinct_id":"guibinwalmart@163.com","$device_id":"%s","$user_id":"guibinwalmart@163.com","Organization Name":"GUBIN","Partner Type":"SELLER","Is Internal":false,"Role":"Admin","MP V":"aurora","mart":"US","onboardingStatus":4,"Partner Id":"10001295723","Seller Id":"101276363","Is WFS Seller":"true","accountStatus":"ACTIVE","goLiveDate":"1669716967527","internationalSeller":false,"userLocale":"en-US","isGSE":true,"New Navigation":true,"$initial_referrer":"https://login.account.wal-mart.com/","$initial_referring_domain":"login.account.wal-mart.com"}; __hstc=195562739.7c34be9e831f1731b988550eaf43b105.%d.%d.1728436699776.2; abqme=true; bstc=Z06W2zP3EP9aROV_HDV140; mobileweb=0; xpth=x-o-mverified+false~x-o-mart+B2C; xpa=-G6VK|0Nhkx|0_oLh|5_Bo3|65NJH|6asrt|99IVe|AS0gq|BAGZ2|DWZZl|DZ6Ns|Do9sW|E4hiH|Fv5Fn|Gf1aH|H09KR|LHiaD|M6YU6|NGjHI|NbX17|OIg9i|OL8-a|OQe8n|UUz4s|V5H9l|XqSAL|a9FUi|cl8-a|fdm-7|hUYIs|jHZ2T|juw-t|kuozS|mjpKB|qeCl3|sWtW6|tDgxK|uvnI5; exp-ck=0Nhkx15_Bo3165NJH16asrt299IVe2BAGZ21DWZZl1DZ6Ns1E4hiH2Fv5Fn1Gf1aH1LHiaD1M6YU61NGjHI1NbX172OIg9i1OL8-a1OQe8n2UUz4s1V5H9l1XqSAL1cl8-a1fdm-71jHZ2T5juw-t1kuozS1mjpKB1qeCl31tDgxK1uvnI52; _pxhd=8e07c6094827e7b47fca02939e3ae2db31522f94860b7295c5ae5ab02338a888:59ba0547-8607-11ef-b239-b79c6eaf72b0; ACID=9ce4dc4a-3b11-46ff-8849-1839231897e3; _intlbu=false; _m=9; _shcc=US; assortmentStoreId=3081; auth=MTAyOTYyMDE4PTAxq8Xili63oXgOMA5CUy59qhsuEmsxAGQ6Mn9loSYi74MPZmSx0iP8v8maRt0unTIQ7GrpRGRXOOH6sopmgD1z7c049M5BlhAz4n/Y5eUeehVecaXkafqv5etgy/Qx767wuZloTfhm7Wk2KcjygkeeSCv4Chv5IarMOQ7pqjdzmU2dfp1jS8D9y/jupN1K7h6zqxJOy0cObUJU4yZn3r2mT1tbeOXUpjV0pSZ3W40UMk70P8glgOEpLOprhDfMJ0tmvH1FCaN9tZDh4SCrHWBMTt6C41vfcObGrkqrNyl3A5Lqzo8k2Zd1Bf9XdJOgO21MBX/vBdrc/3/VXAAv19nLBEq6Yub6EeGf/wxKvOi58Y6V26h4erCHd6J39SWxuG9NlAe2WXnWhmeWIDxC5EjyrOXbKKhH072NS/W0j/U=; hasACID=true; hasLocData=1; locDataV3=eyJpc0RlZmF1bHRlZCI6dHJ1ZSwiaXNFeHBsaWNpdCI6ZmFsc2UsImludGVudCI6IlNISVBQSU5HIiwicGlja3VwIjpbeyJidUlkIjoiMCIsIm5vZGVJZCI6IjMwODEiLCJkaXNwbGF5TmFtZSI6IlNhY3JhbWVudG8gU3VwZXJjZW50ZXIiLCJub2RlVHlwZSI6IlNUT1JFIiwiYWRkcmVzcyI6eyJwb3N0YWxDb2RlIjoiOTU4MjkiLCJhZGRyZXNzTGluZTEiOiI4OTE1IEdFUkJFUiBST0FEIiwiY2l0eSI6IlNhY3JhbWVudG8iLCJzdGF0ZSI6IkNBIiwiY291bnRyeSI6IlVTIiwicG9zdGFsQ29kZTkiOiI5NTgyOS0wMDAwIn0sImdlb1BvaW50Ijp7ImxhdGl0dWRlIjozOC40ODI2NzcsImxvbmdpdHVkZSI6LTEyMS4zNjkwMjZ9LCJpc0dsYXNzRW5hYmxlZCI6dHJ1ZSwic2NoZWR1bGVkRW5hYmxlZCI6dHJ1ZSwidW5TY2hlZHVsZWRFbmFibGVkIjp0cnVlLCJzdG9yZUhycyI6IjA2OjAwLTIzOjAwIiwiYWxsb3dlZFdJQ0FnZW5jaWVzIjpbIkNBIl0sInN1cHBvcnRlZEFjY2Vzc1R5cGVzIjpbIlBJQ0tVUF9TUEVDSUFMX0VWRU5UIiwiUElDS1VQX0lOU1RPUkUiLCJQSUNLVVBfQ1VSQlNJREUiXSwidGltZVpvbmUiOiJQU1QiLCJzdG9yZUJyYW5kRm9ybWF0IjoiV2FsbWFydCBTdXBlcmNlbnRlciIsImRpc3RhbmNlIjowLCJzZWxlY3Rpb25UeXBlIjoiREVGQVVMVEVEIn1dLCJzaGlwcGluZ0FkZHJlc3MiOnsibGF0aXR1ZGUiOjM4LjQ4MjY3NywibG9uZ2l0dWRlIjotMTIxLjM2OTAyNiwiY2l0eSI6IlNhY3JhbWVudG8iLCJzdGF0ZSI6IkNBIiwiY291bnRyeUNvZGUiOiJVUyIsImxvY2F0aW9uQWNjdXJhY3kiOiJsb3ciLCJnaWZ0QWRkcmVzcyI6ZmFsc2UsImFsbG93ZWRXSUNBZ2VuY2llcyI6WyJDQSJdfSwiYXNzb3J0bWVudCI6eyJub2RlSWQiOiIzMDgxIiwiZGlzcGxheU5hbWUiOiJTYWNyYW1lbnRvIFN1cGVyY2VudGVyIiwiZGlzdGFuY2UiOjAsImludGVudCI6IlBJQ0tVUCJ9LCJpbnN0b3JlIjpmYWxzZSwiZGVsaXZlcnkiOnsiYnVJZCI6IjAiLCJub2RlSWQiOiIzMDgxIiwiZGlzcGxheU5hbWUiOiJTYWNyYW1lbnRvIFN1cGVyY2VudGVyIiwibm9kZVR5cGUiOiJTVE9SRSIsImFkZHJlc3MiOnsicG9zdGFsQ29kZSI6Ijk1ODI5IiwiYWRkcmVzc0xpbmUxIjoiODkxNSBHRVJCRVIgUk9BRCIsImNpdHkiOiJTYWNyYW1lbnRvIiwic3RhdGUiOiJDQSIsImNvdW50cnkiOiJVUyIsInBvc3RhbENvZGU5IjoiOTU4MjktMDAwMCJ9LCJnZW9Qb2ludCI6eyJsYXRpdHVkZSI6MzguNDgyNjc3LCJsb25naXR1ZGUiOi0xMjEuMzY5MDI2fSwiaXNHbGFzc0VuYWJsZWQiOnRydWUsInNjaGVkdWxlZEVuYWJsZWQiOmZhbHNlLCJ1blNjaGVkdWxlZEVuYWJsZWQiOmZhbHNlLCJhY2Nlc3NQb2ludHMiOlt7ImFjY2Vzc1R5cGUiOiJERUxJVkVSWV9BRERSRVNTIn1dLCJpc0V4cHJlc3NEZWxpdmVyeU9ubHkiOmZhbHNlLCJhbGxvd2VkV0lDQWdlbmNpZXMiOlsiQ0EiXSwic3VwcG9ydGVkQWNjZXNzVHlwZXMiOlsiREVMSVZFUllfQUREUkVTUyJdLC`

	return fmt.Sprintf(cookieTemplate, userID, distinctID, deviceID, timestamp, deviceID, timestamp, timestamp)
}
