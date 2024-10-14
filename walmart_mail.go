package main

import (
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/xuri/excelize/v2"
	"golang.org/x/net/context"
	"io"
	"log"
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

//var wg = sync.WaitGroup{}
//var ch = make(chan int, 4)

var (
	IO            = 0
	logoutUrl     = `https://www.walmart.com/account/logout`
	emailCss      = `input[type=email]`
	passwordCss   = `input[type=password]`
	submitCss     = `button[type=submit]`
	searchCss     = `#__next > div:nth-child(1) > div > span > header > form > div > input`
	mailCss       = `span[class="lh-title"] button[type="button"]`
	mailTextCss   = `textarea`
	mailSubmitCss = `div[role="dialog"] button[type="button"]`

	ZH []User
)

var userAgent = []string{
	"Mozilla/5.0 (Windows NT 6.3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.5249.103 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.71 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.121 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.5249.119 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.5304.68 Safari/537.36",
}

type User struct {
	Username []string
	Password []string
	MailUrl  []string
	Text     []string
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
			mails := strings.Split(split[2], "|")
			tests := strings.Split(split[3], "|")
			if len(split) == 5 {
				states := strings.Split(split[4], "|")
				state := make([][]string, len(usernames))
				for i := range state {
					state[i] = make([]string, len(mails))
					for ii := range states {
						i2 := strings.Split(states[ii], "-")
						for i3 := range i2 {
							state[i] = append(state[i], i2[i3])
						}
					}
				}

				ZH = append(ZH, User{usernames, passwds, mails, tests, state})
			} else {
				state := make([][]string, len(usernames))
				for i := range state {
					state[i] = make([]string, len(mails))
				}
				ZH = append(ZH, User{usernames, passwds, mails, tests, state})
			}
		}
		if err != nil {
			break
		}

	}

	for {
		var ip []string
		if IO < len(ZH) {
			ip = getIp(1)
		} else {
			break
		}
		if len(ip) > 0 {
			if proxyIS(ip[0]) {
				u := getV()
				if u == nil {
					break
				}
				//ch <- 1
				//wg.Add(1)
				CHROMEDP(u, ip[0])
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

	xlsx := excelize.NewFile()

	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"账号", "密码", "url", "内容", "状态"}); err != nil {
		log.Println(err)
	}
	for i := range ZH {
		var user string
		var pass string
		var mailUrl string
		var text string
		var state string
		for ii := range ZH[i].MailUrl {
			if ii == 0 {
				mailUrl = ZH[i].MailUrl[ii]
				text = ZH[i].Text[ii]
				continue
			}
			mailUrl = mailUrl + "," + ZH[i].MailUrl[ii]
			text = text + "," + ZH[i].Text[ii]
		}
		for ii := range ZH[i].Username {
			if ii != 0 {
				user += ","
				pass += ","
				state += ","
			}
			pass += ZH[i].Password[ii]
			user += ZH[i].Username[ii]
			for i2 := range ZH[i].State[ii] {
				if i2 == 0 {
					state += ZH[i].State[ii][i2]
					continue
				}
				state += "-" + ZH[i].State[ii][i2]
			}
		}
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{user, pass, mailUrl, text, state}); err != nil {
			log.Println(err)
		}
		num++
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
func CHROMEDP(u *User, proxy string) {
	//defer func() {
	//	<-ch
	//	wg.Done()
	//}()
	var is bool
	for i := range u.State {
		for ii := range u.State[i] {
			if u.State[i][ii] != "完成" {
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
		//chromedp.ProxyServer("socks5://"+proxy), //代理
		chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
		chromedp.Flag("headless", false),
		chromedp.Flag("blink-settings", "imagesEnabled=true"), //开启图像界面,重点是开启这个
		chromedp.Flag("ignore-certificate-errors", true),      //忽略错误
		chromedp.Flag("disable-web-security", true),           //禁用网络安全标志
		chromedp.Flag("disable-extensions", true),             //开启插件支持
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-gpu", true), //开启gpu渲染
		//chromedp.WindowSize(1920, 1080),    // 设置浏览器分辨率（窗口大小）
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("disable-application-cache", true),
		chromedp.NoFirstRun, //设置网站不是首次运行
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
		start(u),

		//停止网页加载
		chromedp.Stop(),
	)
	if err != nil {
		fmt.Println(err)
	}
}

func getV() *User {
	lock.Lock()
	defer lock.Unlock()
	var u *User
	if IO < len(ZH) {
		u = &ZH[IO]
		IO++
	}
	return u

}
func start(u *User) chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		for i := range u.Username {
			if !login(ctx, u, i) {
				if u.State[i][0] == "" {
					u.State[i][0] = "登录失败"
				} else if u.State[i][0] == "风控" {
					return
				} else if u.State[i][0] == "页面加载失败" {
					u.State[i][0] = "页面加载失败"
					return
				}
				continue
			}
			for ii := range u.MailUrl {
				if u.State[i][ii] != "完成" {
					if !mail(ctx, u, ii, i) {
						if u.State[i][ii] == "" {
							u.State[i][ii] = "失败"
						}
						if u.State[i][ii] == "风控" {
							return
						} else if u.State[i][ii] == "页面加载失败" || u.State[i][ii] == "点击发邮件失败" || u.State[i][ii] == "输入内容失败" || u.State[i][ii] == "点击发送失败" {
							if !mail(ctx, u, ii, i) {
								if u.State[i][ii] == "" {
									u.State[i][ii] = "失败"
								}
								if u.State[i][ii] == "风控" {
									return
								}
								continue
							}
						}
						continue
					}
				}
			}
		}
		return
	}

}

func login(ctx context.Context, u *User, in2 int) bool {
	for i := 0; i < 6; i++ {
		timeout0, cancel0 := context.WithTimeout(ctx, 20*time.Second)
		defer cancel0()
		//登录
		chromedp.Navigate(logoutUrl).Do(timeout0)
		//等待
		chromedp.Sleep(time.Second * 3).Do(ctx)
		//输入邮箱
		timeout01, cancel01 := context.WithTimeout(ctx, 5*time.Second)
		defer cancel01()
		err := chromedp.WaitVisible(emailCss).Do(timeout01)
		if err != nil {
			u.State[in2][0] = "页面加载失败"
			log.Println(u.Username[in2], "页面加载失败")
			i++
			continue
		}
		log.Println(u.Username[in2], "开始登录")
		//输入邮箱
		timeout1, cancel1 := context.WithTimeout(ctx, 30*time.Second)
		defer cancel1()
		if Fk(getSource(timeout1), u, 0, in2) {
			return false
		}
		var va string
		err = chromedp.Value(emailCss, &va).Do(timeout1)
		if err != nil {
			u.State[in2][0] = "获取v失败"
			log.Println(u.Username[in2], "获取v失败")
			continue
		}
		if len(va) < 1 {
			err = chromedp.SendKeys(emailCss, u.Username[in2]+kb.Enter).Do(timeout1)
			if err != nil {
				u.State[in2][0] = "登录失败"
				log.Println(u.Username[in2], "无法输入账号")
				continue
			}
			chromedp.Sleep(time.Second * 6).Do(ctx)
		} else {
			err = chromedp.SendKeys(emailCss, kb.Enter).Do(timeout1)
			if err != nil {
				u.State[in2][0] = "输入账号下一步失败"
				log.Println(u.Username[in2], "输入账号下一步失败")
				continue
			}
		}

		chromedp.Sleep(time.Second * 6).Do(ctx)

		//输入密码
		timeout2, cancel2 := context.WithTimeout(ctx, 30*time.Second)
		defer cancel2()
		if Fk(getSource(timeout1), u, 0, in2) {
			return false
		}
		err = chromedp.SendKeys(passwordCss, u.Password[in2]+kb.Enter).Do(timeout2)
		if err != nil {
			u.State[in2][0] = "登录失败"
			log.Println(u.Username[in2], "无法输入密码")
			continue
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

func mail(ctx context.Context, u *User, in int, in2 int) bool {
	if Fk(getSource(ctx), u, in, in2) {
		return false
	}
	timeout0, cancel0 := context.WithTimeout(ctx, 60*time.Second)
	defer cancel0()
	//登录
	err := chromedp.Navigate(u.MailUrl[in]).Do(timeout0)
	if err != nil {
		log.Println(u.Username[in2], u.MailUrl[in], "页面加载失败")
		u.State[in2][in] = "页面加载失败"
		return false
	}
	timeout1, cancel1 := context.WithTimeout(ctx, 40*time.Second)
	defer cancel1()
	log.Println(u.Username[in2], "开始发送邮件", u.MailUrl[in])
	//等待
	if Fk(getSource(ctx), u, in, in2) {
		return false
	}
	//点击发邮件
	err = chromedp.SendKeys(mailCss, kb.Enter).Do(timeout1)
	if err != nil {
		log.Println(u.Username[in2], u.MailUrl[in], "点击发邮件失败")
		u.State[in2][in] = "点击发邮件失败"
		return false
	}
	//等待
	chromedp.Sleep(time.Second * 2).Do(ctx)
	if Fk(getSource(ctx), u, in, in2) {
		return false
	}
	//输入内容
	err = chromedp.SendKeys(mailTextCss, u.Text[in]+kb.Enter).Do(timeout1)
	if err != nil {
		log.Println(u.Username[in2], u.MailUrl[in], "输入内容失败")
		u.State[in2][in] = "输入内容失败"
		return false
	}
	chromedp.Sleep(time.Second * 2).Do(ctx)
	if Fk(getSource(ctx), u, in, in2) {
		return false
	}
	//发送
	err = chromedp.SendKeys(mailSubmitCss, kb.Enter).Do(timeout1)
	if err != nil {
		log.Println(u.Username[in2], u.MailUrl[in], "点击发送失败")
		u.State[in2][in] = "点击发送失败"
		return false
	}
	chromedp.Sleep(time.Second * 4).Do(ctx)

	cg := regexp.MustCompile("(Your message has been sent!)").FindAllStringSubmatch(getSource(ctx), -1)
	if len(cg) == 0 {
		log.Println(u.Username[in2], u.MailUrl[in], "邮箱留言失败")
		u.State[in2][in] = "失败"
		return false
	}
	log.Println(u.Username[in2], u.MailUrl[in], "邮箱留言成功")
	u.State[in2][in] = "完成"
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
	if !strings.Contains(str, "walmart.com") {
		log.Println(u.Username[in2], "代理失效")
		u.State[in2][in] = "代理失效"
		return true
	}
	fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(str, -1)
	if len(fk) > 0 {
		log.Println(u.Username[in2], "风控")
		u.State[in2][in] = "风控"
		return true
	}
	return false
}

func getIp(num int) []string {
	log.Println("获取代理")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	//request, _ := http.NewRequest("GET", fmt.Sprintf("https://dps.kdlapi.com/api/getdps/?secret_id=ozy18llf5y0vhe2jjlzw&num=%d&signature=n50p46lgnzb99r3ctydj669rwzu89gvw&pt=2&dedup=1&sep=4", num), nil)
	request, _ := http.NewRequest("GET", fmt.Sprintf("https://mobile.huashengdaili.com/servers.php?session=U79cabbaf0110141141--0a24c7f5e6a70842c415989de21171ef&time=5&count=%d&type=text&pw=no&protocol=s5&separator=5&ip_type=direct", num), nil)
	response, err := client.Do(request)
	if response.StatusCode == 441 || strings.Contains(err.Error(), "441") {
		log.Println("代理超频！暂停10秒后继续...")
		time.Sleep(time.Second * 10)
	}
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

func proxyIS(proxy string) bool {
	log.Println("验证代理是否可用")
	proxyUrl, _ := url.Parse("socks5://" + proxy)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	tr.Proxy = http.ProxyURL(proxyUrl)
	client := &http.Client{Timeout: 10 * time.Second, Transport: tr}
	request, _ := http.NewRequest("GET", "https://www.walmart.com/", nil)

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
	if response.StatusCode == 441 || strings.Contains(err.Error(), "441") {
		log.Println("代理超频！暂停10秒后继续...")
		time.Sleep(time.Second * 10)
	}
	if err != nil {
		if strings.Contains(err.Error(), "An existing connection was forcibly closed by the remote host") {
			log.Println(proxy, "代理无法使用，可能不在代理白名单：", err)
			IO = len(ZH) + 1
			return false
		} else if strings.Contains(err.Error(), "socks connect tcp ") {
			log.Println(proxy, "代理无法使用，可能不在代理白名单：", err)
			//IO = len(ZH) + 1
			return false
		}
		log.Println(proxy, "代理无效：", err)
		return false
	}
	result := ""
	if response.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(response.Body) // gzip解压缩
		if err != nil {
			log.Println("解析body错误，重新开始")
			return false
		}
		defer reader.Close()
		con, err := io.ReadAll(reader)
		if err != nil {
			log.Println("gzip解压错误，重新开始")
			return false
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
			return false
		}
		defer response.Body.Close()
		result = string(dataBytes)
	}
	fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(result, -1)
	if len(fk) > 0 {
		log.Println(proxy, "代理无效：风控")
		return false
	}
	log.Println("代理验证成功")
	return true
}
