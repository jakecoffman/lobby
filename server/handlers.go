package server

import (
	"log"
	"net/http"

	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/lobby/games/lobby"
	"github.com/jakecoffman/lobby/games/spyfall"
	"github.com/jakecoffman/lobby/lib"
	"golang.org/x/net/websocket"
)

const (
	playerId = "GAME"
)

var registry *lib.InMemoryRegistry

func init() {
	registry = lib.NewInMemoryRegistry()
	lob := &lobby.Lobby{}
	lob.Init()
	registry.Register(lob, "lobby")
	registry.Singleton(lob, "lobby")
	go lob.Run(registry)

	registry.Register(&spyfall.Spyfall{}, "spyfall")
}

func UserHandler(ctx *gin.Context) {
	cookie, err := ctx.Request.Cookie(playerId)
	switch {
	case err == nil:
		if _, err := lib.FindPlayer(cookie.Value); err == nil {
			// good, we have a cookie and a player
			return
		}
		log.Println("Database was reset?")
		fallthrough
	case err == http.ErrNoCookie:
		player := lib.NewPlayer()
		cookie = &http.Cookie{Name: playerId, Value: player.ID}
		if err := lib.InsertPlayer(player); err != nil {
			log.Println(err)
		}
		http.SetCookie(ctx.Writer, cookie)
	default:
		log.Println(err)
		ctx.AbortWithStatus(400)
		return
	}
}

func WebSocketHandler(ctx *gin.Context) {
	websocket.Handler(webSocketHandler).ServeHTTP(ctx.Writer, ctx.Request)
}

func webSocketHandler(conn *websocket.Conn) {
	var player *lib.Player
	var err error
	cookie, err := conn.Request().Cookie(playerId)
	if err != nil {
		player = lib.NewPlayer()
		cookie = &http.Cookie{Name: playerId, Value: player.ID}
		if err := lib.InsertPlayer(player); err != nil {
			log.Println(err)
		}
	} else {
		player, err = lib.FindPlayer(cookie.Value)
		if err != nil {
			player = lib.NewPlayer()
			cookie = &http.Cookie{Name: playerId, Value: player.ID}
			if err := lib.InsertPlayer(player); err != nil {
				log.Println(err)
			}
		}
	}
	player.Connect(&lib.WsConn{Conn: conn}, registry)
	if err = player.Send(map[string]string{
		"Type":   "cookie",
		"Cookie": fmt.Sprintf("%s=%s; path=/;", playerId, player.ID),
	}); err != nil {
		log.Println("Player didn't get updated cookie", err)
		return
	}

	player.Run(registry)
}
