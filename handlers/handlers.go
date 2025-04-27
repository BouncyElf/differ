package handlers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func InitApp() *gin.Engine {
	var app *gin.Engine
	gin.SetMode(gin.ReleaseMode)
	app = gin.New()
	app.Use(gin.Recovery())
	// LoggerWithFormatter middleware will write the logs to gin.DefaultWriter
	// By default gin.DefaultWriter = os.Stdout
	app.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		if param.StatusCode < 400 {
			return ""
		}

		// your custom format
		return fmt.Sprintf("[%s] %+v %d %s %s\n",
			param.TimeStamp.Format(time.RFC3339Nano),
			param.Request,
			param.StatusCode,
			param.Latency,
			param.ErrorMessage,
		)
	}))

	app.Any("/*path", Differ)

	return app
}
