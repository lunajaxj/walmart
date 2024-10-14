package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
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

var res = make(map[string]Wal)

type Wal struct {
	id                string
	transitStatusName string
	shippingDay       string
	lastTrackingTime  string
}

type Track struct {
	LogisticsCode string   `json:"logisticsCode"`
	TrackNos      []string `json:"trackNos"`
}

var ids []string
var wg = sync.WaitGroup{}
var ch = make(chan int, 5)

func splitArray(arr []string, chunkSize int) [][]string {
	var result [][]string
	for i := 0; i < len(arr); i += chunkSize {
		end := i + chunkSize
		if end > len(arr) {
			end = len(arr)
		}
		result = append(result, arr[i:end])
	}
	return result
}
func main() {
	log.Println("自动化脚本-跟踪订单号")
	log.Println("开始执行...")
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
	array := splitArray(ids, 40)
	log.Println("共", len(array), "批需要抓取（每批40个）")
	for i := range array {
		track := Track{TrackNos: array[i]}
		ch <- 1
		wg.Add(1)
		go crawler(i, track)
	}
	wg.Wait()

	xlsx := excelize.NewFile()
	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "状态", "天数", "最后跟踪时间"}); err != nil {
		log.Println(err)
	}
	for _, sv := range ids {
		var v Wal
		v = res[sv]
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{v.id, v.transitStatusName, v.shippingDay, v.lastTrackingTime}); err != nil {
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

func crawler(num int, track Track) {

	//配置代理
	defer func() {
		wg.Done()
		<-ch
	}()
fo:
	for i := 0; i < 30; i++ {
		if i != 0 {
			time.Sleep(time.Second * 1)
		}
		proxy_str := fmt.Sprintf("http://%s:%s@%s", "t19932187800946", "wsad123456", "l752.kdltps.com:15818")
		proxy, _ := url.Parse(proxy_str)

		client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}

		marshal, err := json.Marshal(track)
		request, _ := http.NewRequest("POST", "https://www.track123.com/endApi/tk/api/v2/anonymous/track/query-track-nos", bytes.NewBuffer(marshal))

		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Origin", "https://www.track123.com")
		request.Header.Set("Track-Key", "+JNNPsLqFA6Irw0+wQmIPga3f334f9289743a9cecd16a4e87a5a8ff8934d3ec2ea140e88af0d3ec109883e")

		response, err := client.Do(request)

		if err != nil {
			if strings.Contains(err.Error(), "Proxy Bad Serve") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Println("代理IP无效，自动切换中")
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
		//id
		fmt.Println(result)
		var wals []Wal
		trackingNo := regexp.MustCompile(`"shippingDay":".+?","trackingNo":"(.+?)"`).FindAllStringSubmatch(result, -1)
		for i2 := range trackingNo {
			wals = append(wals, Wal{id: trackingNo[i2][1]})
		}
		//状态
		transitStatusName := regexp.MustCompile(`"transitStatusName":"(.+?)"`).FindAllStringSubmatch(result, -1)
		for i2 := range transitStatusName {
			if transitStatusName[i2][1] == "待查询" {
				time.Sleep(15 * time.Second)
				log.Println("第", num+1, "批待查询，等待结果")
				continue fo
			}
			wals[i2].transitStatusName = transitStatusName[i2][1]
		}
		//天数
		shippingDay := regexp.MustCompile(`"shippingDay":"(.+?)"`).FindAllStringSubmatch(result, -1)
		for i2 := range transitStatusName {
			wals[i2].shippingDay = shippingDay[i2][1]
		}
		//最近时间
		lastTrackingTime := regexp.MustCompile(`"lastTrackingTime":"(.+?)"`).FindAllStringSubmatch(result, -1)
		for i2 := range lastTrackingTime {
			wals[i2].lastTrackingTime = lastTrackingTime[i2][1]
		}
		for i2 := range wals {
			res[wals[i2].id] = wals[i2]
		}
		log.Println("第", num+1, "批完成")
		time.Sleep(1 * time.Second)
		return
	}
}
