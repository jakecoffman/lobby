package server

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/lobby"
	"github.com/jakecoffman/lobby/lib"
	"golang.org/x/net/websocket"
)

const (
	playerId = "GAME"
)

var fileStore *lib.FileStore

func init() {
	fileStore = lib.NewFileStore()
}

func RegisterHandler(ctx *gin.Context) {
	cookie, err := ctx.Request.Cookie(playerId)
	switch {
	case err == nil:
		if _, ok := fileStore.Get(cookie.Value); ok {
			// good, we have a cookie and a player
			return
		}
		log.Println("Database was reset?")
		fileStore.Delete(cookie.Value)
		fallthrough
	case err == http.ErrNoCookie:
		player := lib.NewPlayer()
		cookie = &http.Cookie{Name: playerId, Value: player.Id()}
		fileStore.Set(*player)
		// this works somehow even though we're about to upgrade to a websocket
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
	player, ok := fileStore.Get(cookie.Value)
	if !ok {
		websocket.JSON.Send(conn, map[string]string{"error": "No player for cookie"})
		conn.Close()
		return
	}
	player.Connect(conn)

	lobby.Lobby.Play(&player, fileStore)
}
