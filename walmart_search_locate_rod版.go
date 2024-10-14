package main

import (
	"bufio"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/xuri/excelize/v2"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Mode struct {
	keyword string
	id      string //div[@data-testid="list-view"]/preceding-sibling::a[@link-identifier]/@link-identifier
	page    string //div[@data-testid="list-view"]//div[@class="relative"]//img/@alt
	locate  string //div[@data-testid="list-view"]/div/div[1]/div[1]/font/font

}

var res = make(map[string]Mode)
var keywords []string
var lock sync.Mutex
var lockk sync.Mutex
var lock1 sync.Mutex
var wg = sync.WaitGroup{}
var wgg = sync.WaitGroup{}
var wggg = sync.WaitGroup{}
var ch = make(chan int, 1)
var chh = make(chan int, 10)
var bro *rod.Browser

func main() {

	log.Println("自动化脚本-walmart-关键词获取itemid位置")
	log.Println("开始执行...")
	// 创建句柄
	fi, err := os.Open("关键词与id.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 0 {
			keywords = append(keywords, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}

	}
	log.Println("共", len(keywords), "个任务")
	iniRod()
	for _, v := range keywords {
		ch <- 1
		wg.Add(1)
		split := strings.Split(v, "|")
		go crawler(split[0], split[1], split[2], split[2], true)
	}
	wg.Wait()

	log.Println("完成")

}

func save(keys string) {
	defer func() {
		wg.Done()
		lock1.Unlock()
	}()
	fileName := "out.xlsx"
	xlsx, err := excelize.OpenFile("out.xlsx")
	if err != nil {
		xlsx = excelize.NewFile()
	}
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"关键词", "id", "页", "位"}); err != nil {
		log.Println(err)
	}
	var num int
	for i := range keywords {
		if keywords[i] == keys {
			num = i + 2
		}
	}
	if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{res[keys].keyword, res[keys].id, res[keys].page, res[keys].locate}); err != nil {
		log.Println(err)
	}

	xlsx.Save()
	if err != nil {
		xlsx.SaveAs(fileName)
	}
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

func iniRod() {
	// 获取调试链接
	if bro != nil {
		bro.MustClose()
	}
	browser := rod.New()
	//启动无痕
	browser = browser.Trace(true)
	options := launcher.New().Devtools(true)
	//  NoSandbox fix linux下root运行报错的问题
	//options := launcher.New().NoSandbox(true).Headless(false).Set(flags.ProxyServer, "a749.kdltps.com:15818")
	//options := launcher.New().NoSandbox(true).Headless(false).Set(flags.ProxyServer, "127.0.0.1:7890")
	// 禁用所有提示防止阻塞 浏览器
	options = options.Append("disable-infobars", "")
	options = options.Append("disable-extensions", "")

	//if conf.GlobalConfig.BrowserConf.UnHeadless || conf.GlobalConfig.Dev {
	//	options = options.Delete("--headless")
	//	browser = browser.SlowMotion(time.Duration(conf.GlobalConfig.AutoConf.Slow) * time.Second)
	//}

	//options.Proxy("http://127.0.0.1:7890")
	path, _ := launcher.LookPath()
	//fmt.Println(path)
	u := launcher.New().NoSandbox(true).Headless(false).Set(flags.ProxyServer, "127.0.0.1:7890").Bin(path).MustLaunch()
	//container := launcher.MustResolveURL("127.0.0.1:9233")
	bro = rod.New().ControlURL(u).MustConnect()

	//go browser.MustHandleAuth("t16545052065610", "bxancsry")()
	//page := bro.MustPage("https://www.wikipedia.org/")
	//page.MustWaitLoad().MustScreenshot("a.png")
}

func crawler(keyword string, itemid, next string, pages string, t bool) {
	defer func() {
		if t {
			<-ch
			wg.Done()
		}
		wg.Add(1)
		lock1.Lock()
		go save(keyword + "|" + itemid + "|" + next)

	}()
	//page := stealth.MustPage(bro)
	page := bro.MustPage("https://www.walmart.com/search?q=iphone+charger")
	page.WaitLoad()
	//if err != nil {
	//	log.Println(111)
	//	//time.Sleep(10000 * time.Second)
	//	//iniRod()
	//}
	//time.Sleep(10000 * time.Second)

	//page.MustElement(`[type="search"]`).MustSelectAllText().MustInput("")
	//page.MustElement(`[type="search"]`).MustInput(keyword)
	//page.KeyActions().Type(input.Enter).MustDo()
	url := page.MustInfo().URL
	for i := 0; i < 25; {
		//if i != 0 {
		//	time.Sleep(time.Second * 1)
		//}
		page.WaitLoad()

		html, err2 := page.HTML()
		if err2 != nil {
			log.Println("获取内容错误，跳过该标题：" + keyword)
			res[keyword+"|"+itemid+"|"+next] = Mode{keyword, itemid, "获取内容错误", "获取内容错误"}
			return
		}
		cw1 := regexp.MustCompile("(is not valid JSON)").FindAllStringSubmatch(html, -1)
		cw2 := regexp.MustCompile("(The requested URL was rejected. Please consult with your administrator)").FindAllStringSubmatch(html, -1)
		if len(cw1) > 0 || len(cw2) > 0 {
			log.Println("搜索内容错误，跳过该标题：" + keyword)
			res[keyword+"|"+itemid+"|"+next] = Mode{keyword, itemid, "搜索内容错误", "跳过该标题"}
			return
		}
		fk := regexp.MustCompile("(Robot or human?)").FindAllStringSubmatch(html, -1)
		if len(fk) > 0 {
			log.Println("被风控，更换IP重新开始" + keyword)
			log.Println(html)
			return
			time.Sleep(100000)
			iniRod()
			continue
		}
		//doc, err := htmlquery.Parse(strings.NewReader(html))
		//if err != nil {
		//	log.Println("错误信息：" + err.Error())
		//	res[keyword+"|"+itemid+"|"+next] = Mode{keyword, itemid, "获取失败", err.Error()}
		//	return
		//}
		//log.Println(keyword+page)
		//log.Println(html)
		htmlStr := ""
		htmlS := regexp.MustCompile("(items\":\\[.+?\\].+?layoutEnum)").FindAllStringSubmatch(html, -1)
		allString := regexp.MustCompile("(There were no search htmls for)").FindAllString(html, -1)
		if len(allString) > 0 {
			log.Println("关键词:" + keyword + " 第" + pages + "页 无搜索结果")
			res[keyword+"|"+itemid+"|"+next] = Mode{keyword, itemid, "无", "无"}
			return
		}
		if len(htmlS) == 0 {
			log.Println("被风控，更换IP重新开始" + keyword)
			log.Println(html)
			return
			time.Sleep(100000)
			continue
		} else {
			htmlStr = htmlS[0][1]
		}
		//id
		id := regexp.MustCompile("usItemId\":\"([0-9]+?)\",\"[^c]").FindAllStringSubmatch(htmlStr, -1)
		//id,err  := htmlquery.QueryAll(doc, "//div[@data-testid=\"list-view\"]/preceding-sibling::a[@link-identifier]/@link-identifier")
		//log.Println(id[0][])
		split := strings.Split(itemid, ",")
		for i := range id {
			for i2 := range split {
				if id[i][1] == split[i2] {
					res[keyword+"|"+itemid+"|"+next] = Mode{keyword, split[i2], pages, strconv.Itoa(i + 1)}
					log.Println("关键词:" + keyword + " -> 第" + pages + "页  第" + strconv.Itoa(i+1) + "个")
					return
				}
			}
		}
		//最大分页
		maxPage := regexp.MustCompile("\"maxPage\":([0-9]+?),").FindAllStringSubmatch(html, -1)
		atoi := 25
		if len(maxPage) != 0 {
			atoi, _ = strconv.Atoi(maxPage[0][1])
		}
		ne, _ := strconv.Atoi(next)
		var pa, er = strconv.Atoi(pages)
		if er != nil {
			return
		}
		if pa < atoi && pa < 25 && pa < ne && pa > 1 {
			url += `&page=` + strconv.Itoa(pa-1) + `&affinityOverride=default`
			page = bro.MustPage(url)
			continue
		} else if pa < atoi && pa < 25 && pa >= ne {
			url += `&page=` + strconv.Itoa(pa+1) + `&affinityOverride=default`
			page = bro.MustPage(url)
			continue
		} else if pa == atoi && pa <= 25 {
			url += `&page=` + strconv.Itoa(ne-1) + `&affinityOverride=default`
			page = bro.MustPage(url)
			continue
		} else {
			res[keyword+"|"+itemid+"|"+next] = Mode{keyword, itemid, "无", "无"}
			return
		}
	}
	res[keyword+"|"+itemid+"|"+next] = Mode{keyword, itemid, "获取失败", "获取失败"}

}
