package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/xuri/excelize/v2"
	"golang.org/x/net/context"
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
	starsCss         = `div[class="mr4"] section button`
	starsSubmitCss   = `div[data-focus-lock-disabled="false"]  [role="dialog"] > div:nth-child(3) > button`
	shoppingCss      = `div[data-testid="add-to-cart-section"] div[class="relative dib"] button`
	commentCss       = `a.hover-white`
	commentStarsCss  = `label[for="star-5"]`
	commentSubmitCss = `button[aria-describedby="tcs"]`
	ZH               []User
)

var userAgent = []string{
	"Mozilla/5.0 (Windows NT 6.3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.5249.103 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.71 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.121 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.5249.119 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.68 Safari/537.36",
}

var py string

type User struct {
	Username []string
	Password []string
	Keyword  []string
	Id       []string
	Page     []string
	ZPage    []int
	State    [][]string
}

func main() {
	fi, err := os.Open("账号密码.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader
	//file, err := os.ReadFile("代理.txt")
	//if err != nil {
	//	panic(err)
	//}
	//prxoy = string(file)
	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 0 {
			space := strings.TrimSpace(string(lineB))
			split := strings.Split(space, ",")
			usernames := strings.Split(split[0], "|")
			passwds := strings.Split(split[1], "|")
			keys := strings.Split(split[2], "|")
			idss := strings.Split(split[3], "|")
			pages := strings.Split(split[4], "|")
			if len(split) == 6 {
				states := strings.Split(split[5], "|")
				state := make([][]string, len(usernames))
				for i := range state {
					state[i] = make([]string, len(keys))
					for ii := range states {
						i2 := strings.Split(states[ii], "-")
						for i3 := range i2 {
							state[i] = append(state[i], i2[i3])
						}
					}
				}
				ZH = append(ZH, User{usernames, passwds, keys, idss, pages, make([]int, len(pages)), state})
			} else {
				state := make([][]string, len(usernames))
				for i := range state {
					state[i] = make([]string, len(keys))
				}
				ZH = append(ZH, User{usernames, passwds, keys, idss, pages, make([]int, len(pages)), state})
			}
		}
		if err != nil {
			break
		}

	}
	log.Println("任务数：", len(ZH))
	for {
		var ip []string
		if IO < len(ZH) {
			ip = getIp(1)
		} else {
			break
		}
		if len(ip) > 0 {
			log.Println("验证代理是否可用")
			if proxyIS(ip[0]) {
				log.Println("验证代理通过")
				u, io := getV()
				if u == nil {
					break
				}
				//ch <- 1
				//wg.Add(1)
				py = ip[0]
				CHROMEDP(u, ip[0], io)
			} else {
				time.Sleep(3 * time.Second)
				continue
			}
		} else {
			time.Sleep(3 * time.Second)
			continue
		}
	}
	//wg.Wait()

	log.Println("完成")

}

func save(u *User, i int) {
	lockd.Lock()
	defer lockd.Unlock()
	fileName := "out.xlsx"
	xlsx, err := excelize.OpenFile("out.xlsx")
	if err != nil {
		xlsx = excelize.NewFile()
	}

	num := i + 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"账号", "密码", "关键词", "id", "页数", "真实页数", "状态"}); err != nil {
		log.Println(err)
	}
	var user string
	var pass string
	var key string
	var id string
	var page string
	var zpage string
	var state string
	for ii := range u.Keyword {
		if ii == 0 {
			key = u.Keyword[ii]
			id = u.Id[ii]
			page = u.Page[ii]
			zpage = strconv.Itoa(u.ZPage[ii])
			continue
		}
		key = key + "," + u.Keyword[ii]
		id = id + "," + u.Id[ii]
		page = page + "," + u.Page[ii]
		zpage = zpage + "," + strconv.Itoa(u.ZPage[ii])
	}
	for ii := range u.Username {
		if ii != 0 {
			user += ","
			pass += ","
			state += ","
		}
		pass += u.Password[ii]
		user += u.Username[ii]
		for i2 := range u.State[ii] {
			if i2 == 0 {
				state += u.State[ii][i2]
				continue
			}
			state += "-" + u.State[ii][i2]
		}
	}
	if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{u.Username, u.Password, key, id, page, zpage, state}); err != nil {
		log.Println(err)
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
func CHROMEDP(u *User, proxy string, num int) {
	//defer func() {
	//	<-ch
	//	wg.Done()
	//}()
	var is bool
	for i := range u.State {
		for ii := range u.State[i] {
			if u.State[i][ii] != "完成" && u.State[i][ii] != "没有找到商品" {
				is = true
				break
			}
		}
	}

	if !is {
		return
	}

	//配置
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer("socks5://"+proxy), //代理
		chromedp.NoDefaultBrowserCheck,          //不检查默认浏览器
		chromedp.Flag("headless", false),
		chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
		chromedp.Flag("ignore-certificate-errors", true),      //忽略错误
		chromedp.Flag("disable-web-security", true),           //禁用网络安全标志
		chromedp.Flag("disable-extensions", true),             //开启插件支持
		chromedp.Flag("disable-default-apps", true),
		//chromedp.Flag("disable-gpu", true), //开启gpu渲染
		chromedp.WindowSize(1920, 1080), // 设置浏览器分辨率（窗口大小）
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.NoFirstRun, //设置网站不是首次运行
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"), //设置UserAgent

		chromedp.UserAgent(userAgent[rand.Int31n(5)]), //设置UserAgent
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	// 初始化chromedp上下文，后续这个页面都使用这个上下文进行操作
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	err := chromedp.Run(ctx,
		//设置webdriver检测反爬
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
		start(u, num, proxy),
		//停止网页加载
		chromedp.Stop(),
	)
	if err != nil {
		fmt.Println(err)
	}
}

func getV() (*User, int) {
	lock.Lock()
	defer lock.Unlock()
	var u *User
	if IO < len(ZH) {
		u = &ZH[IO]
		IO++
	}
	return u, IO - 1

}

func start(u *User, num int, proxy string) chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		defer save(u, num)
		for i := range u.Username {
			if !proxyIS(proxy) {
				u.State[i][0] = "代理失效"
				return
			}
			//if !login(ctx, u, i) {
			//	if u.State[i][0] == "" {
			//		u.State[i][0] = "登录失败"
			//	}
			//	if !proxyIS(proxy) {
			//		u.State[i][0] = "代理失效"
			//		return
			//	}
			//	return
			//}
			for ii := range u.Id {
				if u.State[i][ii] != "完成" && u.State[i][ii] != "没有找到商品" {
					if !proxyIS(proxy) {
						u.State[i][ii] = "代理失效"
						return
					}
					if !search(ctx, u, ii, i) {
						if u.State[i][ii] == "" {
							u.State[i][ii] = "搜索页失败"
						}
						if !proxyIS(proxy) {
							u.State[i][ii] = "代理失效"
							return
						}
						continue
					}
					if !proxyIS(proxy) {
						u.State[i][ii] = "代理失效"
						return
					}
					if !collecting(ctx, u, ii, i) {
						if u.State[i][ii] == "" {
							u.State[i][ii] = "产品页失败"
						}
						if !proxyIS(proxy) {
							u.State[i][ii] = "代理失效"
							return
						}
						continue
					}
					if !proxyIS(proxy) {
						u.State[i][ii] = "代理失效"
						return
					}
				}
			}
		}
		return
	}

}

func login(ctx context.Context, u *User, in2 int) bool {
	timeout0, cancel0 := context.WithTimeout(ctx, 10*time.Second)
	defer cancel0()
	log.Println(u.Username[in2], "开始登录")
	for i := 0; i < 6; i++ {
		//登录
		chromedp.Navigate(LoginUrl).Do(timeout0)
		//等待
		chromedp.Sleep(time.Second * 3).Do(ctx)
		//输入邮箱
		timeout1, cancel1 := context.WithTimeout(ctx, 30*time.Second)
		defer cancel1()
		if Fk(getSource(timeout1), u, 0, in2) {
			return false
		}
		err := chromedp.SendKeys(emailCss, u.Username[in2]+kb.Enter).Do(timeout1)
		if err != nil {
			u.State[in2][0] = "无法输入账号"
			log.Println(u.Username[in2], "无法输入账号")
			return false
		}

		//等待
		chromedp.Sleep(time.Second * 3).Do(timeout1)
		err = chromedp.WaitNotPresent(emailCss).Do(timeout1)
		if err != nil {
			log.Println(u.Username[in2], "登录页面跳转失败，重新输入账号")
			u.State[in2][0] = "登录页面跳转失败，重新输入账号"
			continue
		}
		chromedp.Sleep(time.Second * 10).Do(timeout1)

		//输入密码
		timeout2, cancel2 := context.WithTimeout(ctx, 20*time.Second)
		defer cancel2()
		if !isEx(ctx, passwordCss) {
			log.Println(u.Username[in2], "未加载到密码页面")
			u.State[in2][0] = "未加载到密码页面"
			return false
		}

		err = chromedp.SendKeys(passwordCss, u.Password[in2]+kb.Enter).Do(timeout2)
		if err != nil {
			log.Println(u.Username[in2], "无法输入密码")
			u.State[in2][0] = "无法输入密码"
			return false
		}
		//等待
		chromedp.Sleep(time.Second * 6).Do(ctx)

		if Fk(getSource(ctx), u, 0, in2) {
			return false
		}
		//判断是否登录成功
		timeout3, cancel3 := context.WithTimeout(ctx, 30*time.Second)
		defer cancel3()
		err = chromedp.WaitVisible(searchCss).Do(timeout3)
		if err != nil {
			u.State[in2][0] = "登录失败"
			log.Println(u.Username[in2], "登录失败")
			return false
		} else {
			log.Println(u.Username[in2], "登录成功")
			return true

		}
		return true
	}
	return false

}

func search(ctx context.Context, u *User, in int, in2 int) bool {
	var url string
	var next bool
	timeout0, cancel0 := context.WithTimeout(ctx, 6*time.Second)
	defer cancel0()
	//到首页
	chromedp.Navigate("https://www.walmart.com/").Do(timeout0)
	timeout1, cancel1 := context.WithTimeout(ctx, 60*time.Second)
	defer cancel1()
	log.Println(u.Username[in2], "开始搜索", u.Keyword[in])
	//搜索
	chromedp.SendKeys(searchCss, u.Keyword[in]+kb.Enter).Do(timeout1)
	//等待
	chromedp.Sleep(time.Second * 5).Do(ctx)
	if Fk(getSource(ctx), u, in, in2) {
		return false
	}
	url = getUrl(ctx)
	page, _ := strconv.Atoi(u.Page[in])
	chromedp.Navigate(url + "&affinityOverride=default").Do(timeout1)
	var zpage = 1
	//最大分页
	max := 25
	maxPage := regexp.MustCompile("\"maxPage\":([0-9]+?),").FindAllStringSubmatch(getSource(timeout1), -1)
	if len(maxPage) != 0 {
		max, _ = strconv.Atoi(maxPage[0][1])
	}
	for {
		chromedp.Sleep(time.Second * 2).Do(ctx)
		for i := 0; i < 60; i++ {
			chromedp.EvaluateAsDevTools(`window.scrollBy(0, 300)`, nil).Do(ctx)
		}
		timeout2, cancel2 := context.WithTimeout(ctx, 10*time.Second)
		defer cancel2()
		//等待
		var is bool
		is = isEx(ctx, `a[link-identifier="`+u.Id[in]+`"]`)
		if is {
			chromedp.Sleep(time.Second * 2).Do(ctx)
			u.ZPage[in] = zpage
			chromedp.SendKeys("a[link-identifier=\""+u.Id[in]+"\"]", kb.Enter).Do(timeout2)
			chromedp.Sleep(time.Second * 2).Do(ctx)
			log.Println(u.Username[in2], u.Id[in], "进入产品")
			return true
		}
		//chromedp.EvaluateAsDevTools(`document.querySelector('a[aria-label="Next Page"]') !== null`, &iss).Do(timeout2)
		//chromedp.EvaluateAsDevTools(`document.querySelector('a[aria-label="Previous Page"]') !== null`, &iss2).Do(timeout2)
		timeout3, cancel3 := context.WithTimeout(ctx, 60*time.Second)
		defer cancel3()
		if Fk(getSource(timeout3), u, in, in2) {
			return false
		}
		if page <= max && !next {
			chromedp.Navigate(url + "&page=" + strconv.Itoa(page) + "&affinityOverride=default").Do(timeout3)
			log.Println(u.Username[in2], u.Id[in], "往下翻页:", page)
			page++
		} else if page > 0 {
			if !next {
				page, _ = strconv.Atoi(u.Page[in])
				page -= 1
				next = true
			}
			log.Println(u.Username[in2], u.Id[in], "往上翻页:", page)
			chromedp.Navigate(url + "&page=" + strconv.Itoa(page) + "&affinityOverride=default").Do(timeout3)
			page -= 1
		} else {
			log.Println(u.Username[in2], u.Id[in], "没有找到商品")
			u.State[in2][in] = "没有找到商品"

			return false
		}
		zpage = page
	}

	return false
}
func collecting(ctx context.Context, u *User, in int, in2 int) bool {
	chromedp.Sleep(time.Second * 2).Do(ctx)
	chromedp.Evaluate("location.replace(location.href);", nil).Do(ctx)
	chromedp.Sleep(time.Second * 10).Do(ctx)
	if Fk(getSource(ctx), u, in, in2) {
		return false
	}
	////点击心愿
	//timeout1, cancel1 := context.WithTimeout(ctx, 30*time.Second)
	//defer cancel1()
	//if !isEx(ctx, starsCss) {
	//	u.State[in2][in] = "加入星愿失败"
	//	log.Println(u.Username[in2], u.Id[in], "加入星愿失败")
	//}
	//log.Println(u.Username[in2], u.Id[in], "加入星愿")
	//chromedp.SendKeys(starsCss, kb.Enter).Do(timeout1)
	////等待
	//chromedp.Sleep(time.Second * 3).Do(ctx)
	//if Fk(getSource(ctx), u, in, in2) {
	//	return false
	//}
	//log.Println(u.Username[in2], u.Id[in], "保存星愿")
	//chromedp.SendKeys(starsSubmitCss, kb.Enter).Do(timeout1)

	//加购物车
	//var url1 string
	//url1 = getUrl(ctx)
	//等待
	chromedp.Sleep(time.Second * 3).Do(ctx)
	if Fk(getSource(ctx), u, in, in2) {
		return false
	}

	timeout2, cancel2 := context.WithTimeout(ctx, 30*time.Second)
	var is bool
	defer cancel2()
	log.Println(u.Username[in2], u.Id[in], "加入购物车")
	err := chromedp.SendKeys(shoppingCss, kb.Enter).Do(timeout2)
	if err != nil {
		u.State[in2][in] = "加入购物车失败"
		log.Println(u.Username[in2], u.Id[in], "加入购物车失败")
		return false
	}
	//等待

	chromedp.Sleep(time.Second * 3).Do(ctx)
	if Fk(getSource(ctx), u, in, in2) {
		return false
	}
	//var url2 string
	//url2 = getUrl(ctx)

	//chromedp.EvaluateAsDevTools(`document.querySelector('`+shoppingCss+`') !== null`, &is).Do(timeout2)
	if !isEx(ctx, shoppingCss) {
		log.Println(u.Username[in2], u.Id[in], "购物车返回上一步")
		chromedp.NavigateBack().Do(timeout2)
		chromedp.Sleep(time.Second * 3).Do(ctx)
		if Fk(getSource(ctx), u, in, in2) {
			return false
		}
	} else {
		chromedp.EvaluateAsDevTools(`document.querySelector('button[aria-label="Close dialog"]') !== null`, &is).Do(timeout2)
		if isEx(ctx, `button[aria-label="Close dialog"]`) {
			log.Println(u.Username[in2], u.Id[in], "关闭购物车弹窗")
			chromedp.SendKeys(`button[aria-label="Close dialog"]`, kb.Enter).Do(timeout2)
		}
	}
	//n := rand.Int31n(30-15) + 15
	//log.Println(u.Username[in2], u.Id[in], "开始滑轮", n, "下")
	////等待
	//chromedp.Sleep(time.Second * 3).Do(ctx)
	//
	//for i := 0; i < int(n); i++ {
	//	chromedp.EvaluateAsDevTools(`window.scrollBy(0, 100)`, nil).Do(ctx)
	//	chromedp.Sleep(300).Do(ctx)
	//	time.Sleep(300)
	//}
	//
	////等待
	//chromedp.Sleep(time.Second * 3).Do(ctx)
	//
	////留评论
	//timeout3, cancel3 := context.WithTimeout(ctx, 60*time.Second)
	//defer cancel3()
	//log.Println(u.Username[in2], u.Id[in], "开始留评论")
	//chromedp.SendKeys(commentCss, kb.Enter).Do(timeout3)
	////等待
	//chromedp.Sleep(time.Second * 6).Do(ctx)
	//if Fk(getSource(ctx), u, in, in2) {
	//	return false
	//}
	//
	//log.Println(u.Username[in2], u.Id[in], "点星")
	//chromedp.Click(commentStarsCss).Do(timeout3)
	//
	////等待
	//chromedp.Sleep(time.Second * 3).Do(ctx)
	//if Fk(getSource(ctx), u, in, in2) {
	//	return false
	//}
	//var iss int
	//for {
	//	var is bool
	//	chromedp.EvaluateAsDevTools(`document.querySelector('`+commentStarsCss+`') !== null`, &is).Do(timeout3)
	//	if is {
	//		iss++
	//		log.Println(u.Username[in2], u.Id[in], "保存")
	//		chromedp.SendKeys(commentSubmitCss, kb.Enter).Do(timeout3)
	//		//等待
	//		chromedp.Sleep(1 * time.Second).Do(ctx)
	//		time.Sleep(1 * time.Second)
	//		if Fk(getSource(ctx), u, in, in2)` {
	//			return false
	//		}
	//	} else {
	//		if iss == 0 {
	//			u.State[in2][in] = "产品页失败"
	//			log.Println(u.Username[in2], u.Id[in], "失败，未知原因")
	//		}
	//		break
	//	}
	//}
	u.State[in2][in] = "完成"
	log.Println(u.Username[in2], u.Id[in], "完成")
	chromedp.Sleep(3 * time.Second).Do(ctx)
	return true
}

func getSource(ctx context.Context) string {
	var source string
	if err := chromedp.Run(ctx, chromedp.EvaluateAsDevTools("document.documentElement.outerHTML", &source)); err != nil {
		return ""
	}
	return source
}

func getUrl(ctx context.Context) string {
	var source string
	if err := chromedp.Run(ctx, chromedp.Location(&source)); err != nil {
		return ""
	}
	return source
}
func Fk(str string, u *User, in int, in2 int) bool {
	if !proxyIS(py) {
		log.Println(u.Username[in2], u.Id[in], "代理失效或网络过慢")
		u.State[in2][in] = "代理失效或网络过慢"
		return true
	}
	fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(str, -1)
	if len(fk) > 0 {
		log.Println(u.Username[in2], u.Id[in], "风控")
		u.State[in2][in] = "风控"
		return true
	}
	return false
}

func IsProxy(proxy string) bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	proxyUrl, _ := url.Parse(proxy)
	tr.Proxy = http.ProxyURL(proxyUrl)

	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	request, _ := http.NewRequest("GET", "http://myip.top/", nil)

	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
	//request.Header.Set("Accept", "*/*")
	_, err := client.Do(request)
	if err != nil {
		if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
			log.Println("代理失效")
			return true
		}
	}
	return false
}

func getIp(num int) []string {
	log.Println("获取代理")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	//5分钟
	request, _ := http.NewRequest("GET", fmt.Sprintf("https://mobile.huashengdaili.com/servers.php?session=U79cabbaf0110141141--0a24c7f5e6a70842c415989de21171ef&time=5&count=%d&type=text&only=1&province=310000&city=310100&pw=no&protocol=s5&separator=5&ip_type=tunnel", num), nil)
	//10分钟
	//request, _ := http.NewRequest("GET", fmt.Sprintf("https://mobile.huashengdaili.com/servers.php?session=U79cabbaf0110141141--0a24c7f5e6a70842c415989de21171ef&time=10&count=%d&type=text&pw=no&protocol=s5&separator=5&ip_type=direct", num), nil)
	response, err := client.Do(request)
	if err != nil {
		if strings.Contains(err.Error(), "441") {
			log.Println("代理超频！暂停10秒后继续...")
			time.Sleep(time.Second * 10)
		}
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

func proxyIS(proxy string) bool {
	for i := 1; i < 5; i++ {
		time.Sleep(1000)
		proxyUrl, _ := url.Parse("socks5://" + proxy)
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
		client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
		//request, _ := http.NewRequest("GET", "https://www.walmart.com/search?q=123", nil)
		request, _ := http.NewRequest("GET", "http://myip.top", nil)
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
				IO = len(ZH) + 1
				log.Println(proxy, "代理无法使用，可能不在代理白名单：", err)
				return false
			} else if strings.Contains(err.Error(), "socks connect tcp ") {
				log.Println(proxy, "代理验证错误：", err)
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
					log.Println("连续出现代理IP无效请联系我，重新开始：")
				} else {
					log.Println("错误信息：" + err.Error())
					log.Println("出现错误，如果同id连续出现请联系我，重新开始：")
				}
				continue
			}
			defer response.Body.Close()
			result = string(dataBytes)
		}
		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
		if len(fk) > 0 {
			log.Println(proxy, "代理被风控")
			return false
		}
		return true
	}
	return true

}

func isEx(ctx context.Context, ex string) bool {
	timeout2, cancel2 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel2()
	err := chromedp.WaitVisible(ex).Do(timeout2)
	if err != nil {
		return false
	} else {
		return true
	}
}
