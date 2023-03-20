package controller

import (
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"log"
	"net/http"
)

func SetupSocketIO(app *gin.Engine) {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Println("CONNECTED TO SOCKET IO", s.ID())
		return nil
	})

	server.OnEvent("/", "join", func(s socketio.Conn, msg string) {
		log.Println("joined", msg)
		//s.Emit("reply", "have "+msg)
	})

	server.OnEvent("/", "leave", func(s socketio.Conn, msg string) string {
		s.SetContext(msg)
		log.Println("joined", msg)
		return "recv " + msg
	})

	server.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	//server.OnDisconnect("/", func(s socketio.Conn, reason string) {
	//	log.Println("closed", reason)
	//})

	go server.Serve()
	defer server.Close()

	app.GET("/socket.io/*any", gin.WrapH(server))
	app.POST("/socket.io/*any", gin.WrapH(server))
	app.StaticFS("/public", http.Dir("../asset"))
}
