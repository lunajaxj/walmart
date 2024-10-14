package main

import (
	"github.com/go-vgo/robotgo"
)

func main() {
	// 移动鼠标到指定位置
	robotgo.MoveMouse(100, 200)

	// 单击
	robotgo.Click("left", true)

	// 右键点击
	robotgo.Click("right", false)

}
