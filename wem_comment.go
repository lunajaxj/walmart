package main

import (
	"fmt"
	xlsx "github.com/tealeg/xlsx"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"log"
	"strconv"
	"strings"
)

var db *sqlx.DB

type Comment struct {
	Comment_ID           int    `db:"comment_ID"`
	Comment_post_ID      int    `db:"comment_post_ID"`
	Comment_author       string `db:"comment_author"`
	Comment_author_email string `db:"comment_author_email"`
	Comment_author_url   string `db:"comment_author_url"`
	Comment_author_IP    string `db:"comment_author_IP"`
	Comment_date         string `db:"comment_date"`
	Comment_date_gmt     string `db:"comment_date_gmt"`
	Comment_content      string `db:"comment_content"`
	Comment_karma        int    `db:"comment_karma"`
	Comment_approved     string `db:"comment_approved"`
	Comment_agent        string `db:"comment_agent"`
	Comment_type         string `db:"comment_type"`
	Comment_parent       int    `db:"comment_parent"`
	User_id              int    `db:"user_id"`
}

func main() {
	var err error
	db, err = sqlx.Open("mysql", "www_simiya_net:XYrXC2HKsr@tcp(45.12.109.14:3306)/www_simiya_net?charset=utf8mb4")
	if err != nil {
		panic(err)
	}
	log.Println("开始读取文件...")
	coms := open()
	log.Println("开始上传...")
	put(coms)
	log.Println("上传成功")

}
wem_comment.go
func open() []Comment {
	var coms []Comment
	// 打开XLS文件
	xlFile, err := xlsx.OpenFile("评论.xlsx")
	if err != nil {
		panic(err)
	}

	// 获取第一个工作表
	sheet := xlFile.Sheets[0]

	// 遍历所有行
	for i := 1; i <= sheet.MaxRow; i++ {
		var er error
		row := sheet.Row(i)
		if len(row.Cells) == 0 {
			break
		}

		ID, er := strconv.Atoi(strings.ToLower(row.Cells[0].String()))
		if row.Cells[0].String() != "" && er == nil {
			layout := "02/01/2006" // 定义输入日期字符串的格式
			dateStr := row.Cells[3].String()
			t, err := time.Parse(layout, dateStr)
			if err != nil {
				fmt.Println("日期解析失败：", err)
				panic(err)
			}
			mysqlDate := t.Format("2006-01-02 15:04:05")
			xj, err := strconv.Atoi(strings.ToLower(row.Cells[4].String()))
			if err != nil {
				panic(err.Error())
			}
			var xx = ""
			switch xj {
			case 0:
				xx = "✰✰✰✰✰"
			case 1:
				xx = "★✰✰✰✰"
			case 2:
				xx = "★★✰✰✰"
			case 3:
				xx = "★★★✰✰"
			case 4:
				xx = "★★★★✰"
			case 5:
				xx = "★★★★★"
			}
			coms = append(coms, Comment{
				Comment_post_ID:      ID,
				Comment_author:       row.Cells[1].String(),
				Comment_author_email: row.Cells[2].String(),
				Comment_author_url:   "",
				Comment_author_IP:    "192.168.2.1",
				Comment_date:         mysqlDate,
				Comment_date_gmt:     mysqlDate,
				//Comment_content:      "评分 " + xx + "\r\n" + "标题 " + row.Col(5) + "\r\n" + "正文 " + row.Col(6),
				Comment_content:  xx + "\r\n" + row.Cells[5].String() + "\r\n" + row.Cells[6].String(),
				Comment_karma:    0,
				Comment_approved: "1",
				Comment_agent:    "Mozilla/5.0 (Windows NT 6.1; ) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.85 Safari/537.36",
				Comment_type:     "comment",
				Comment_parent:   0,
				User_id:          0,
			})
		}

	}
	return coms

}

func put(coms []Comment) error {
	_, err := db.NamedExec(`INSERT INTO wp_comments ( comment_post_ID,comment_author,comment_author_email,comment_author_url,comment_author_IP,comment_date,comment_date_gmt,comment_content,comment_karma,comment_approved,comment_agent,comment_type,comment_parent,user_id)
VALUES (:comment_post_ID,:comment_author,:comment_author_email,:comment_author_url,:comment_author_IP,:comment_date,:comment_date_gmt,:comment_content,:comment_karma,:comment_approved,:comment_agent,:comment_type,:comment_parent,:user_id)`, coms)
	if err != nil {
		panic(err)
		return err
	}
	return nil
}
