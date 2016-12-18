package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/lobby/games/spyfall"
	"github.com/jakecoffman/lobby/lib"
	"github.com/jakecoffman/lobby/server"
	"gopkg.in/mgo.v2"
	"log"
	"net/http/pprof"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error
	lib.Session, err = mgo.Dial("127.0.0.1:27017")
	if err != nil {
		log.Fatal(err)
	}
	lib.DB = lib.Session.DB("lobby")

	bootstrap()

	r := route(lib.DB)

	registry := lib.NewInMemoryRegistry()
	spyfall.Install(r, lib.DB, registry)

	r.GET("/debug/pprof/", func(ctx *gin.Context) {
		pprof.Index(ctx.Writer, ctx.Request)
	})
	r.GET("/debug/pprof/goroutine", func(ctx *gin.Context) {
		pprof.Handler("goroutine").ServeHTTP(ctx.Writer, ctx.Request)
	})
	r.Run("0.0.0.0:8444")
}

func bootstrap() {
	lib.DB.DropDatabase()
	lib.InsertUser(&lib.User{ID: "admin", Name: "admin"})
}

func route(db *mgo.Database) *gin.Engine {
	router := gin.Default()
	router.Use(server.UserMiddleware)
	router.GET("/me", func(ctx *gin.Context) {
		user := ctx.MustGet("player").(*lib.User)
		ctx.JSON(200, user)
	})
	return router
}
