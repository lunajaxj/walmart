package main

import (
	"bufio"
	"github.com/xuri/excelize/v2"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	var name []string
	fi, err := os.Open("img.txt")
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(fi) // 创建 Reader

	for {
		lineB, err := r.ReadBytes('\n')
		if len(lineB) > 0 {
			name = append(name, strings.TrimSpace(string(lineB)))
		}
		if err != nil {
			break
		}
	}

	xlsx := excelize.NewFile()
	num := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"id", "图片"}); err != nil {
		log.Println(err)
	}
	for _, na := range name {
		log.Println(na)
		if err := xlsx.SetSheetRow("Sheet1", "A"+strconv.Itoa(num), &[]interface{}{na}); err != nil {
			log.Println(err)
		}
		if err := xlsx.AddPicture("Sheet1", "B"+strconv.Itoa(num), "img/"+na+".jpeg", ""); err != nil {
			if err := xlsx.AddPicture("Sheet1", "B"+strconv.Itoa(num), "./img/"+na+".jpg", ""); err != nil {
				if err := xlsx.AddPicture("Sheet1", "B"+strconv.Itoa(num), "./img/"+na+".png", ""); err != nil {
				}
			}

		}
		num++
	}

	fileName := "out.xlsx"
	xlsx.SaveAs(fileName)

	log.Println("完成")
}
