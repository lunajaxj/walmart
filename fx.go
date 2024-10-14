package main

import (
	"encoding/csv"
	"github.com/agnivade/levenshtein"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
)

type Product struct {
	ID    string
	Title string
}

func main() {
	// 打开CSV文件
	file, err := os.Open("data.csv")
	log.Println("开始数据相似度匹配...")
	if err != nil {
		panic(err)
	}

	// 创建CSV读取器
	reader := csv.NewReader(file)
	reader.Comma = ','         // 指定分隔符为逗号
	reader.Comment = '#'       // 忽略以#开头的注释行
	reader.FieldsPerRecord = 2 // 指定每条记录有两个字段

	// 读取CSV数据
	var products []Product

	var simila []Product
	var pr Product
	var coun int
	for {
		record, err := reader.Read()
		if err != nil {
			// 如果已到文件结尾，则跳出循环
			if err == io.EOF {
				break
			}
			panic(err)
		}
		// 跳过头部
		if record[0] == "id" {
			continue
		}
		if pr.ID == "" {
			pr = Product{
				ID:    record[0],
				Title: record[1],
			}
			simila = append(simila, pr)
			continue
		}
		coun++
		pr2 := Product{
			ID:    record[0],
			Title: record[1],
		}

		distance := levenshtein.ComputeDistance(pr.Title, pr2.Title)
		similarity := 1 - float64(distance)/float64(len(pr.Title)+len(pr2.Title))
		// 如果相似度超过6成，就把它们视为同一款产品
		if similarity > 0.7 {
			//log.Printf("第1轮对比，第%d次对比结果：%f \n", coun, similarity)
			simila = append(simila, pr2)
		} else {
			products = append(products, Product{
				ID:    record[0],
				Title: record[1],
			})
		}
	}
	file.Close()
	log.Printf("第%d轮匹配结束, 有%d个数据匹配成功 当前剩余数据总数%d\n", 1, len(simila), len(products))
	// 创建一个新的文件
	reg := regexp.MustCompile("[^\\w\\s]+")
	name := reg.ReplaceAllString(simila[0].Title, "")
	file2, err2 := os.Create("file/" + strconv.Itoa(len(simila)) + "个_" + name + ".csv")
	if err2 != nil {
		panic(err2)
	}
	defer file2.Close()

	// 创建一个CSV写入器
	writer := csv.NewWriter(file2)
	for i := range simila {
		writer.Write([]string{simila[i].ID, simila[i].Title})
	}
	// 刷新缓冲区
	writer.Flush()

	file2.Close()
	// 遍历商品，找到相似的商品
	lun := 2
	for {
		coun = 0
		index := []int{0}
		var similar []Product
		similar = append(similar, products[0])
		for i := 0; i < len(products)-1; i++ {
			coun++
			// 计算标题相似度
			distance := levenshtein.ComputeDistance(products[0].Title, products[i+1].Title)
			similarity := 1 - float64(distance)/float64(len(products[0].Title)+len(products[i+1].Title))
			// 如果相似度超过6成，就把它们视为同一款产品
			if similarity > 0.7 {
				//log.Printf("第%d轮对比，第%d次对比结果：%f id:%s \n", lun, coun, similarity, products[i+1].ID)
				index = append(index, i+1)
				similar = append(similar, products[i+1])
			}
		}
		for _, v := range index {
			products = deleteSlice(products, v)
		}
		log.Printf("第%d轮匹配结束, 有%d个数据匹配成功, 当前剩余数据总数%d \n", lun, len(similar), len(products))
		// 创建一个新的文件
		reg := regexp.MustCompile("[^\\w\\s]+")
		name := reg.ReplaceAllString(similar[0].Title, "")
		file3, err3 := os.Create("file/" + strconv.Itoa(len(simila)) + "个_" + name + ".csv")
		if err3 != nil {
			panic(err3)
		}

		// 创建一个CSV写入器
		writer2 := csv.NewWriter(file3)
		for i := range similar {
			writer2.Write([]string{similar[i].ID, similar[i].Title})
		}
		// 刷新缓冲区
		writer2.Flush()
		file3.Close()
		lun++
	}
	log.Println("全部完成")

	//// 输出相似的商品// levenshtein.RatioForStrings是一个函数，用于计算两个字符串的相似度
	//for _, similarProduct := range similarProducts {
	//	fmt.Println(similarProduct)
	//}
}

func deleteSlice(s []Product, index int) []Product {
	copy(s[index:], s[index+1:])
	return s[:len(s)-1]
}
