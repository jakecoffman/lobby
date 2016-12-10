package lobby

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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
	cookie, err := conn.Request().Cookie(playerId)
	if err != nil {
		websocket.JSON.Send(conn, map[string]string{"error": "Invalid cookie"})
		conn.Close()
		return
	}
	player, err := lib.FindPlayer(cookie.Value)
	if err != nil {
		websocket.JSON.Send(conn, map[string]string{"error": "No player for cookie"})
		conn.Close()
		return
	}
	player.Connect(conn)

	Lobby.Play(player)
}
