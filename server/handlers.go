package server

import (
	"log"
	"net/http"

	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/lobby/lib"
	"golang.org/x/net/websocket"
)

const (
	GAMECOOKIE = "GAME"
)

var registry *lib.InMemoryRegistry

func UserMiddleware(ctx *gin.Context) {
	cookie, err := ctx.Request.Cookie(GAMECOOKIE)
	switch {
	case err == nil:
		if _, err := lib.FindUser(cookie.Value); err == nil {
			// good, we have a cookie and a player
			return
		}
		log.Println("Database was reset?")
		fallthrough
	case err == http.ErrNoCookie:
		user := &lib.User{}
		cookie = &http.Cookie{Name: GAMECOOKIE, Value: user.ID.Hex()}
		if err := lib.InsertUser(user); err != nil {
			log.Println(err)
		}
		http.SetCookie(ctx.Writer, cookie)
	default:
		log.Println(err)
		ctx.AbortWithStatus(400)
		return
	}
}

// for games to use to allow anonymous users to connect
func Connect(conn *websocket.Conn) (*lib.User, error) {
	var user *lib.User
	var err error
	cookie, err := conn.Request().Cookie(GAMECOOKIE)
	if err != nil {
		user = lib.NewUser()
		cookie = &http.Cookie{Name: GAMECOOKIE, Value: user.ID.Hex()}
		if err := lib.InsertUser(user); err != nil {
			log.Println(err)
			return nil, err
		}
	} else {
		user, err = lib.FindUser(cookie.Value)
		if err != nil {
			user = lib.NewUser()
			cookie = &http.Cookie{Name: GAMECOOKIE, Value: user.ID.Hex()}
			if err := lib.InsertUser(user); err != nil {
				log.Println(err)
				return nil, err
			}
		}
	}
	user.Connect(&lib.WsConn{Conn: conn}, registry)
	if err = user.Send(map[string]string{
		"Type":   "cookie",
		"Cookie": fmt.Sprintf("%s=%s; path=/;", GAMECOOKIE, user.ID.Hex()),
	}); err != nil {
		log.Println("Player didn't get updated cookie", err)
		return nil, err
	}

	return user, nil
}
