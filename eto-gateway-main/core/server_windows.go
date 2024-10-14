package core

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// initServer 初始化服务器（Windows平台）
func initServer(address string, router *gin.Engine) server {
	return &http.Server{
		Addr:           address,
		Handler:        router,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}
