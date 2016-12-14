package server

import (
	"log"
	"net/http"

	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/lobby"
	"github.com/jakecoffman/lobby/lib"
	"golang.org/x/net/websocket"
)

const (
	playerId = "GAME"
)

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
	player.Connect(&lib.WsConn{Conn: conn})
	if err = player.Send(map[string]string{
		"type":   "cookie",
		"cookie": fmt.Sprintf("%s=%s; path=/;", playerId, player.ID),
	}); err != nil {
		log.Println("Player didn't get updated cookie", err)
		return
	}

	lobby.Lobby.Play(player)
}
