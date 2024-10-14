package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 打开 where.csv 文件
	f, err := os.Open("name.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	// 解析 CSV 文件
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 遍历 img 目录下的所有文件
	imgDir := "img"
	files, err := os.ReadDir(imgDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range files {
		// 遍历 CSV 文件中的所有记录
		for _, record := range records {
			// 如果 name1 和文件名匹配，重命名文件
			if strings.HasPrefix(file.Name(), record[0]+".") {
				newName := record[1] + ".jpg"
				oldPath := filepath.Join(imgDir, file.Name())
				newPath := filepath.Join(imgDir, newName)
				err = os.Rename(oldPath, newPath)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Printf("重命名 %s 为 %s\n", oldPath, newPath)
				}
			}
		}
	}
}
