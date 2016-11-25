package main

import (
	"log"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/lobby"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	r := gin.Default()
	r.Use(lobby.RegisterHandler)
	r.StaticFile("/", "./static/index.html")
	r.Static("/static", "./static")
	r.GET("/ws", lobby.WebSocketHandler)
	r.GET("/debug/pprof/", func(ctx *gin.Context) { pprof.Index(ctx.Writer, ctx.Request) })
	r.GET("/debug/pprof/goroutine", func(ctx *gin.Context) { pprof.Handler("goroutine").ServeHTTP(ctx.Writer, ctx.Request) })
	r.Run("0.0.0.0:8444")
}
