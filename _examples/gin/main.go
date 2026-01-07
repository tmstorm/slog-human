package main

import (
	"log/slog"
	"net/http"

	logger "github.com/tmstorm/slog-human"
	"github.com/tmstorm/slog-human/middleware"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func main() {
	l := logger.NewDefaultLogger()
	slog.SetDefault(l)

	r := gin.New()
	r.Use(gin.Recovery())
	// slog-simple's gin middleware uses gin-contrib/requestid
	// under the hood. If you would like the requestid to show
	// this middleware must be in use, otherwise it will be ignored.
	r.Use(requestid.New())
	r.Use(middleware.GinLogger(slog.Default()))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	err := r.Run(":8000")
	if err != nil {
		slog.Error("[GIN]", "error", err.Error())
	}
}
