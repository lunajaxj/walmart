package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/xuri/excelize/v2"
	"golang.org/x/net/context"
	"golang.org/x/net/html"
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

var lock sync.Mutex
var wg = sync.WaitGroup{}
var lockd sync.Mutex
var ch = make(chan int, 1)

var (
	IO               = 0
	LoginUrl         = `https://www.walmart.com/account/login`
	emailCss         = `input[type=email]`
	passwordCss      = `input[type=password]`
	submitCss        = `button[type=submit]`
	searchCss        = `#__next > div:nth-child(1) > div > span > header > form > div > input`
	starsCss         = `.mh0-l .f6:nth-child(1`
	starsSubmitCss   = `div[data-focus-lock-disabled="false"]  [role="dialog"] > div:nth-child(3) > button`
	shoppingCss      = `div[data-testid="add-to-cart-section"] div[class="relative dib"] button`
	commentCss       = `a.hover-white`
	commentStarsCss  = `label[for="star-5"]`
	commentSubmitCss = `button[aria-describedby="tcs"]`
)

var userAgent = []string{
	"Mozilla/5.0 (Windows NT 6.3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.5249.103 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.71 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.121 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.5249.119 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.68 Safari/537.36",
}

type Commodity struct {
	Id        string
	Seller    string
	Price     string
	Delivery  string
	SellWiths []SellWith
}

type SellWith struct {
	Seller   string
	Price    string
	Delivery string
}

var ids []string

var count = 1

var idNum int

var py string

func main() {
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

	start()

	//wg.Wait()

	log.Println("完成")

}

func save(u *Commodity) {
	lockd.Lock()
	defer lockd.Unlock()
	fileName := "out.xlsx"
	xlsx, err := excelize.OpenFile("out.xlsx")
	if err != nil {
		xlsx = excelize.NewFile()
	}

	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"item id", "购物车卖家名", "购物车价格", "购物车运输方式", "跟卖卖家数量", "跟卖卖家名", "跟卖价格", "跟卖运输方式"}); err != nil {
		log.Println(err)
	}
	if len(u.SellWiths) > 0 {
		for i2 := range u.SellWiths {
			count++
			if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(count), &[]interface{}{u.Id, u.Seller, u.Price, u.Delivery, len(u.SellWiths), u.SellWiths[i2].Seller, u.SellWiths[i2].Price, u.SellWiths[i2].Delivery}); err != nil {
				log.Println(err)
			}
		}
	} else {
		count++
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(count), &[]interface{}{u.Id, u.Seller, u.Price, u.Delivery, 0, nil, nil, nil}); err != nil {
			log.Println(err)
		}

	}

	xlsx.SaveAs(fileName)
}

func CHROMEDP() {
	//配置
	py = getIp(1)[0]
	for i := 0; i < 3 && !proxyIS(); i++ {
		py = getIp(1)[0]
	}
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer("socks5://"+py), //代理
		chromedp.NoDefaultBrowserCheck,       //不检查默认浏览器
		chromedp.Flag("headless", false),
		chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
		chromedp.Flag("ignore-certificate-errors", true),      //忽略错误
		chromedp.Flag("disable-web-security", true),           //禁用网络安全标志
		chromedp.Flag("disable-extensions", true),             //开启插件支持
		chromedp.Flag("disable-default-apps", true),
		//chromedp.Flag("disable-gpu", true), //开启gpu渲染
		//chromedp.WindowSize(1920, 1080), // 设置浏览器分辨率（窗口大小）
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.NoFirstRun,                           //设置网站不是首次运行
		chromedp.UserAgent(userAgent[rand.Int31n(5)]), //设置UserAgent
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	// 初始化chromedp上下文，后续这个页面都使用这个上下文进行操作
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 3000*time.Minute)
	defer cancel()
	err := chromedp.Run(ctx,
		//设置webdriver检测反爬
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
		seller(),
		//停止网页加载
		chromedp.Stop(),
	)
	if err != nil {
		fmt.Println(err)
	}
}

func start() {
	for idNum < len(ids) {
		CHROMEDP()
	}
	return

}

func seller() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		for idNum < len(ids) {
			//打开产品页面
			timeout0, cancel0 := context.WithTimeout(ctx, 20*time.Second)
			defer cancel0()
			chromedp.Navigate("https://www.walmart.com/ip/" + ids[idNum]).Do(timeout0)
			//点击跟卖
			timeout1, cancel1 := context.WithTimeout(ctx, 60*time.Second)
			defer cancel1()
			var gm bool
			err := chromedp.SendKeys(`button[aria-label="Compare all sellers"]`, kb.Enter).Do(timeout1)
			if err != nil {
				gm = false
				log.Println(ids[idNum], "无跟卖")
			} else {
				gm = true
			}
			chromedp.Sleep(5 * time.Second).Do(ctx)

			source := getSource(ctx)
			doc, err := htmlquery.Parse(strings.NewReader(source))
			if err != nil {
				log.Println(err)
				return err
			}

			fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(source, -1)
			if len(fk) > 0 {
				log.Println("代理被风控")
				return err
			}
			//跟卖
			var all []*html.Node
			var com Commodity
			com.Id = ids[idNum]
			stock := regexp.MustCompile("(Out of stock)").FindAllStringSubmatch(source, -1)
			if len(stock) > 0 {
				com.Seller = "无库存"
				log.Println(ids[idNum], com)
				save(&com)
				idNum++
				continue
			}

			if !gm {
				//没有跟卖
				//价格
				price := regexp.MustCompile("<span itemprop=\"price\".*?.{0,20}(\\$[.,\\d]+).{0,20}?</span>").FindAllStringSubmatch(source, -1)
				if len(price) > 0 {
					com.Price = price[0][1]
				}
				//卖家与配送
				all, err = htmlquery.QueryAll(doc, "//div/div/span[@class=\"lh-title\"]//text()")
			} else {
				//有跟卖
				prices, err := htmlquery.QueryAll(doc, "//div[@data-testid=\"allSellersOfferLine\"]/div[@class=\"pb3\"]/div[1]/span[1]//text()")
				if err != nil {
					log.Println("价格获取失败")
				} else {
					for i, v := range prices {
						if i == 0 {
							com.Price = htmlquery.InnerText(v)
							continue
						}
						com.SellWiths = append(com.SellWiths, SellWith{Price: htmlquery.InnerText(v)})
					}
				}
				//卖家与配送
				all, err = htmlquery.QueryAll(doc, "//div[@data-testid=\"allSellersOfferLine\"]//span[@class=\"lh-title\"]//text()")

			}
			var in int
			if err != nil {
				log.Println("卖家与配送获取失败")
			} else {
				for i, v := range all {
					sv := htmlquery.InnerText(v)
					if strings.Contains(sv, "Sold by") {
						//log.Println("seller", htmlquery.InnerText(all[i+1]))
						if in == 0 {
							com.Seller = htmlquery.InnerText(all[i+1])
							continue
						}
						com.SellWiths[in-1].Seller = htmlquery.InnerText(all[i+1])
						//continue
					} else if strings.Contains(sv, "Fulfilled by") && strings.Contains(sv, "Walmart") {
						if in == 0 {
							com.Delivery = "Walmart.com"
							in++
							continue
						}
						com.SellWiths[in-1].Delivery = "Walmart.com"
						in++
					} else if strings.Contains(sv, "Fulfilled by") {
						if in == 0 {
							com.Delivery = htmlquery.InnerText(all[i+1])
							in++
							continue
						}
						com.SellWiths[in-1].Delivery = htmlquery.InnerText(all[i+1])
						in++
						//continue
					} else if strings.Contains(sv, "Sold and shipped by") && strings.Contains(sv, "Walmart") {
						if in == 0 {
							com.Seller = "Walmart.com"
							com.Delivery = "Walmart.com"
							in++
							continue
						}
						com.SellWiths[in-1].Seller = "Walmart.com"
						com.SellWiths[in-1].Delivery = "Walmart.com"
						in++
					} else if strings.Contains(sv, "Sold and shipped by") {
						//log.Println("seller", htmlquery.InnerText(all[i+1]))
						//log.Println("delivery", htmlquery.InnerText(all[i+1]))
						if in == 0 {
							com.Seller = htmlquery.InnerText(all[i+1])
							com.Delivery = htmlquery.InnerText(all[i+1])
							in++
							continue
						}
						com.SellWiths[in-1].Seller = htmlquery.InnerText(all[i+1])
						com.SellWiths[in-1].Delivery = htmlquery.InnerText(all[i+1])
						in++
						//break
					}
				}
			}
			if !proxyIS() {
				return err
			}

			//标题,判断是否存在
			title := regexp.MustCompile("\"productName\":\"(.+?)\",").FindAllStringSubmatch(source, -1)
			if com.Seller == "" && com.Price == "" && com.Delivery == "" && len(com.SellWiths) == 0 && len(title) == 0 {
				log.Println("代理可能失效，重新获取代理开始")
				return err
			} else if com.Seller == "" && com.Price == "" && com.Delivery == "" && len(com.SellWiths) == 0 {
				log.Println("重新获取")
				continue
			}
			if len(title) == 0 {
				save(&com)
				log.Println("该产品不存在")
				idNum++
				continue
			}
			save(&com)
			log.Println(ids[idNum], com)
			idNum++
		}
		return err
	}

}

func getSource(ctx context.Context) string {
	var source string
	if err := chromedp.Run(ctx, chromedp.EvaluateAsDevTools("document.documentElement.outerHTML", &source)); err != nil {
		return ""
	}
	return source
}

func getIp(num int) []string {
	log.Println("获取代理")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	//5分钟
	request, _ := http.NewRequest("GET", fmt.Sprintf("https://mobile.huashengdaili.com/servers.php?session=U79cabbaf0110141141--0a24c7f5e6a70842c415989de21171ef&time=5&count=%d&type=text&pw=no&protocol=s5&separator=5&ip_type=direct", num), nil)
	//10分钟
	//request, _ := http.NewRequest("GET", fmt.Sprintf("https://mobile.huashengdaili.com/servers.php?session=U79cabbaf0110141141--0a24c7f5e6a70842c415989de21171ef&time=10&count=%d&type=text&pw=no&protocol=s5&separator=5&ip_type=direct", num), nil)
	response, err := client.Do(request)
	if err != nil {
		log.Println("代理提取错误：", err)
		return nil
	}
	dataBytes, err := io.ReadAll(response.Body)
	result := string(dataBytes)
	if strings.Contains(result, "暂未添加白名单") {
		log.Println("不在代理白名单，无法获取使用")
		return nil
	}
	ips := strings.Split(result, " ")
	if len(ips) > 0 {
		log.Println("获取代理成功:", ips)
	}
	return ips
}

func proxyIS() bool {
	for i := 1; i < 5; i++ {
		time.Sleep(1000)
		proxyUrl, _ := url.Parse("socks5://" + py)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
		client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
		request, _ := http.NewRequest("PUT", "https://www.walmart.com/ip/205440965", nil)

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
			if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
				log.Println(py, "代理无法使用，可能不在代理白名单：", err)
				return false
			} else if strings.Contains(err.Error(), "socks connect tcp ") {
				log.Println(py, "代理验证错误：", err)
			}
			continue
		}
		result := ""
		if response.Header.Get("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(response.Body) // gzip解压缩
			if err != nil {
				log.Println("解析body错误，重新开始")
				continue
			}
			defer reader.Close()
			con, err := io.ReadAll(reader)
			if err != nil {
				log.Println("gzip解压错误，重新开始")
				continue
			}
			result = string(con)
		} else {
			dataBytes, err := io.ReadAll(response.Body)
			if err != nil {
				if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "Service Unavailable") {
					log.Println("代理IP无效，自动切换中")
					log.Println("连续出现代理IP无效请联系我，重新开始")
				} else {
					log.Println("错误信息：" + err.Error())
					log.Println("出现错误，如果同id连续出现请联系我，重新开始")
				}
				continue
			}
			defer response.Body.Close()
			result = string(dataBytes)
		}
		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
		if len(fk) > 0 {
			log.Println(py, "代理被风控")
			return false
		}
		return true
	}
	return true
}
