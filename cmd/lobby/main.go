package main

import (
	"log"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/lobby/server"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	r := gin.Default()
	r.Use(server.UserHandler)

	r.GET("/ws", server.WebSocketHandler)
	r.GET("/debug/pprof/", func(ctx *gin.Context) { pprof.Index(ctx.Writer, ctx.Request) })
	r.GET("/debug/pprof/goroutine", func(ctx *gin.Context) { pprof.Handler("goroutine").ServeHTTP(ctx.Writer, ctx.Request) })
	r.Run("0.0.0.0:8444")
}
