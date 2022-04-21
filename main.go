package main

import "github.com/gin-gonic/gin"

func main() {

	r := gin.Default()

	r.GET("/chat:name", chat) //聊天

	r.POST("/create", CreateRoom) //创建房间

	r.Run(":80")

}
