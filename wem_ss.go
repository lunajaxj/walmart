package main

import (
	"encoding/csv"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
	"strings"

	"log"
)

var db *sqlx.DB

type ProductDetails struct {
	Mark         string  `db:"mark"`          //标记
	Remark       string  `db:"remark"`        //备注
	Id           int     `db:"id"`            //id
	Img          string  `db:"img"`           //图片
	CodeType     string  `db:"code_type"`     //商品码类型
	Code         string  `db:"code"`          //商品码值
	Brands       string  `db:"brands"`        //品牌
	Tags         string  `db:"tags"`          //标签
	Title        string  `db:"title"`         //标题
	Rating       string  `db:"rating"`        //评分
	Comments     string  `db:"comments"`      //评论数量
	Price        float64 `db:"price"`         //价格
	Sellers      string  `db:"sellers"`       //卖家
	Distribution string  `db:"distribution"`  //配送
	Variants1    string  `db:"variants1"`     //变体1
	Variants2    string  `db:"variants2"`     //变体2
	VariantsId   string  `db:"variants_id"`   //变体id
	ArrivalTime  string  `db:"arrival_time"`  //到达时间
	Category1    string  `db:"category1"`     //类目1
	Category2    string  `db:"category2"`     //类目2
	Category3    string  `db:"category3"`     //类目3
	Category4    string  `db:"category4"`     //类目4
	Category5    string  `db:"category5"`     //类目5
	Category6    string  `db:"category6"`     //类目6
	Category7    string  `db:"category7"`     //类目7
	CategoryName string  `db:"category_name"` //类目id
	CreateTime   string  `db:"create_time"`   //创建时间
	UpdateTime   string  `db:"update_time"`   //更新时间
	Num          int     `db:"num"`           //count

}

type Temp struct {
	Sellers string `db:"sellers"`
	Count   int    `db:"count"`
}
type Tempe struct {
	Sellers string `db:"sellers"`
	Id      int    `db:"id"`
}

func main() {
	var err error
	db, err = sqlx.Open("mysql", "root:disen88888888@tcp(192.168.2.8:3316)/walmart?charset=utf8mb4")
	if err != nil {
		panic(err)
	}

	open()

}
func open() {

	// 读取跳转链接列表文件
	file, err := os.Open("链接数量.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 解析跳转链接列表文件
	var num string
	scanner := csv.NewReader(file)
	for {
		record, err := scanner.Read()
		if err != nil {
			break
		}
		num += record[0]
	}
	temps := select1()
	//fmt.Println(temps)
	atoi, err := strconv.Atoi(num)
	if err != nil {
		panic(err)
	}
	strings := select2(temps, atoi)
	te := select3(strings)
	//fmt.Println(te)
	//for i := range te {
	//	for i2 := range temps {
	//		if temps[i2].Sellers == "DECLUTTR" {
	//			fmt.Println(temps[i2])
	//		}
	//		if te[i].Sellers == temps[i2].Sellers {
	//			te[i].Count = temps[i2].Count
	//			break
	//		}
	//	}
	//}

	xlsx := excelize.NewFile()
	nume := 2
	if err := xlsx.SetSheetRow("Sheet1", "A1", &[]interface{}{"卖家", "id"}); err != nil {
		log.Println(err)
	}

	for _, sv := range te {
		itoa := strconv.Itoa(nume)
		if err := xlsx.SetSheetRow("Sheet1", "A"+itoa, &[]interface{}{sv.Sellers, sv.Id}); err != nil {
			log.Println(err)
		}
		nume++
	}
	fileName := "out.xlsx"
	for fileNum := 1; exists(fileName); fileNum++ {
		fileName = "out(" + strconv.Itoa(fileNum) + ").xlsx"
	}
	xlsx.SaveAs(fileName)

	log.Println("全部完成")

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

func select1() (temps []Temp) {
	err := db.Select(&temps, "SELECT sellers, COUNT(*) AS count FROM product_details  where sellers !='Walmart' and sellers !='Walmart.com'  GROUP BY sellers;")
	if err != nil {
		panic(err)
	}
	return
}

func select2(temps []Temp, count int) (res []string) {
	for i := range temps {
		if temps[i].Count > count {
			res = append(res, temps[i].Sellers)
		}
	}
	return
}

func select3(strs []string) (temps []Tempe) {
	join := strings.Join(strs, `","`)
	replace := strings.Replace(join, `"`, `\"`, -1)
	replace = strings.Replace(join, `\",\"`, `","`, -1)
	//replace = strings.Replace(join, `'`, `\'`, -1)
	var s string
	s = `"` + replace + `"`

	sql := "SELECT sellers,id FROM product_details WHERE sellers IN (" + s + ") and (distribution = \"Walmart.com\" OR distribution  = \"Walmart\") order by sellers"
	err := db.Select(&temps, sql)
	if err != nil {
		//os.Exit(1)
		panic(err)
	}

	return
}
