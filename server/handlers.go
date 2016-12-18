package server

import (
	"log"
	"net/http"

	"errors"
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
		// TODO: insecure, just finding user by their ID. need proper login cookie ID maker
		if p, err := lib.FindUser(cookie.Value); err == nil {
			ctx.Set("player", p)
			return
		}
		log.Println("Found cookie but couldn't find player", cookie.Value)
		fallthrough
	case err == http.ErrNoCookie:
		user := lib.NewUser()
		cookie = &http.Cookie{Name: GAMECOOKIE, Value: user.ID}
		if err := lib.InsertUser(user); err != nil {
			log.Println(err)
		}
		http.SetCookie(ctx.Writer, cookie)
		ctx.Set("player", user)
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
	log.Println("Cookie is", cookie)
	if err != nil {
		user = lib.NewUser()
		cookie = &http.Cookie{Name: GAMECOOKIE, Value: user.ID}
		if err := lib.InsertUser(user); err != nil {
			log.Println(err)
			return nil, err
		}
		user.Connect(&lib.WsConn{Conn: conn}, registry)
		if err = sendCookie(user); err != nil {
			return nil, err
		}
	} else {
		user, err = lib.FindUser(cookie.Value)
		if err != nil {
			user = lib.NewUser()
			cookie = &http.Cookie{Name: GAMECOOKIE, Value: user.ID}
			if err := lib.InsertUser(user); err != nil {
				log.Println(err)
				return nil, err
			}
			user.Connect(&lib.WsConn{Conn: conn}, registry)
			if err = sendCookie(user); err != nil {
				return nil, err
			}
		} else {
			user.Connect(&lib.WsConn{Conn: conn}, registry)
		}
	}

	return user, nil
}

func sendCookie(user *lib.User) error {
	if err := user.Send(map[string]string{
		"Type":   "cookie",
		"Cookie": fmt.Sprintf("%s=%s; path=/;", GAMECOOKIE, user.ID),
	}); err != nil {
		return errors.New(fmt.Sprintln("Trouble sending updated cookie", err))
	}
	return nil
}
