package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhp12543/fil-address/http/api"
	"net/http"
)

func Init() *gin.Engine {
	// 修改模式
	gin.SetMode(gin.ReleaseMode)
	// 创建路由
	engine := gin.New()
	// gin.Recovery() 作用是防止程序的panic导致程序崩溃，gin.Recovery()捕获panic，返回500
	engine.Use(gin.Recovery())
	// 404默认值
	engine.NoRoute(NoResponse)
	// MiddleWare中间件-解决跨域
	//engine.Use(middlewares.Cors())
	// 设置请求映射
	initRouter(engine)
	return engine
}

func initRouter(router *gin.Engine) {
	router.GET("api/ping", ping)
	router.GET("/", api.Index)
	router.GET("/filAddress", api.FilAddress)
	router.GET("/send/filAddress", api.FilSendAddress)
	router.GET("/eosDecodeHex", api.EosDecodeHex)
}

func ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func NoResponse(c *gin.Context) {
	// 返回404状态码
	c.JSON(http.StatusNotFound, gin.H{
		"status": 404,
		"error":  "404, page not exists!",
	})
}
