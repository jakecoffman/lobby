package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/gorest"
	"github.com/jakecoffman/lobby/games/spyfall"
	"github.com/jakecoffman/lobby/lib"
	"github.com/jakecoffman/lobby/server"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http/pprof"
	"time"
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

	spyfall.Install(r, lib.DB)

	r.GET("/debug/pprof/", func(ctx *gin.Context) {
		pprof.Index(ctx.Writer, ctx.Request)
	})
	r.GET("/debug/pprof/goroutine", func(ctx *gin.Context) {
		pprof.Handler("goroutine").ServeHTTP(ctx.Writer, ctx.Request)
	})
	r.Run("0.0.0.0:8444")
}

func bootstrap() {
	id := bson.NewObjectIdWithTime(time.Time{})
	lib.InsertUser(&lib.User{ID: id, Name: "admin"})
}

func route(db *mgo.Database) *gin.Engine {
	router := gin.Default()
	router.Use(server.UserMiddleware)
	userRoute := router.Group("/users")
	{
		controller := server.UserController{
			MongoController: gorest.MongoController{
				C:        db.C(lib.USER),
				Resource: &lib.User{},
			},
		}
		userRoute.GET("/", controller.List)
		userRoute.GET("/:id", controller.Get)

		userRoute.POST("/", controller.Create)
		userRoute.PUT("/:id", controller.Update)
		userRoute.DELETE("/:id", controller.Delete)
	}
	return router
}
