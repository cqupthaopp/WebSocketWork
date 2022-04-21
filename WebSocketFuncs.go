package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
	"unsafe"
)

var upGrader = websocket.Upgrader{

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var cnt = 1
var lock = sync.Mutex{}
var MsgCh map[string]chan Msg
var room = sync.Map{}

func upDateMsg(roomName string, username string, c *websocket.Conn) {

	for {

		mt, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}

		if mt == websocket.TextMessage {

			MsgCh[roomName] <- Msg{
				time:    time.Now(),
				User:    username,
				Content: message,
			}

		}

		if err != nil {
			fmt.Println(err)
			break
		}

	}

}

func BroadCast(roomName string) {

	for {

		var err error

		select {

		case msg := <-MsgCh[roomName]:

			if msg.time.Unix()-time.Now().Unix() >= 50000 {
				continue
			}

			m := append(S2B((msg.time.Format(time.Stamp))+"  用户"+msg.User+" 说:\n"), msg.Content...)
			room.Range(func(key, value any) bool {
				conn := value.(*websocket.Conn)
				err = conn.WriteMessage(websocket.TextMessage, m)
				if err != nil {
					log.Println("conn.WriteMessage err: ", err)
				}
				return true
			})
		}
	}

}

func S2B(str string) (bytes []byte) {
	x := *(*[2]uintptr)(unsafe.Pointer(&str))
	bytes = *(*[]byte)(unsafe.Pointer(&[3]uintptr{x[0], x[1], x[1]}))
	return
}

func chat(c *gin.Context) {

	roomName := c.Param("name")
	if _, ok := MsgCh[roomName]; !ok {
		c.JSON(401, gin.H{
			"status": 10003,
			"info":   "No Room",
		})
		return
	}

	conn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	go BroadCast(roomName)
	defer conn.Close()

	if err != nil {
		fmt.Println(err)
		c.JSON(200, gin.H{
			"status": 10002,
			"info":   "failed",
		})
		return
	}

	lock.Lock()
	username := getUserName()
	lock.Unlock()

	room.Store(username, conn)
	upDateMsg(roomName, username, conn)

	c.JSON(200, gin.H{
		"status": 10000,
		"info":   "successed",
	})
	return

}

func JudgePermission(token string) bool {
	return true
}

func CreateRoom(c *gin.Context) {

	roomName := c.PostForm("roomName")
	var Token string //…………

	if !JudgePermission(Token) {
		c.JSON(401, gin.H{
			"status": 10005,
			"info":   "No permission",
		}) //无权限
		return
	}

	if _, ok := MsgCh[roomName]; ok {
		c.JSON(405, gin.H{
			"status": 10006,
			"info":   "Room exist",
		}) //房间已经存在
		return
	}

	MsgCh[roomName] = make(chan Msg, 50) //make chan

	c.JSON(200, gin.H{
		"status": 10005,
		"info":   "successed",
	})
	return
}

func getUserName() string {
	cnt++
	return "Hao_pp" + string('0'+cnt)
}
