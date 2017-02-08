package server

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/lobby/lib"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

const (
	GAMECOOKIE = "GAME"
)

func UserMiddleware(ctx *gin.Context) {
	cookie, err := ctx.Request.Cookie(GAMECOOKIE)
	switch {
	// user has a GAME cookie
	case err == nil:
		// TODO: insecure, just finding user by their ID. need proper login cookie ID maker
		if p, err := lib.FindUser(cookie.Value); err == nil {
			ctx.Set("player", p)
			return
		}
		log.Println("Found cookie but couldn't find player", cookie.Value)
		fallthrough
	// user has no GAME cookie
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
func WSMiddleware(conn *websocket.Conn) (*lib.User, error) {
	var user *lib.User
	var err error
	cookie, err := conn.Request().Cookie(GAMECOOKIE)
	// user has no GAME cookie
	if err != nil {
		user = lib.NewUser()
		cookie = &http.Cookie{Name: GAMECOOKIE, Value: user.ID}
		if err := lib.InsertUser(user); err != nil {
			log.Println(err)
			return nil, err
		}
		user.Connect(&lib.WsConn{Conn: conn})
		if err = sendCookie(user); err != nil {
			return nil, err
		}
	} else {
		// user has a GAME cookie
		user, err = lib.FindUser(cookie.Value)
		if err != nil {
			// woah, can't find user!
			user = lib.NewUser()
			cookie = &http.Cookie{Name: GAMECOOKIE, Value: user.ID}
			if err := lib.InsertUser(user); err != nil {
				log.Println(err)
				return nil, err
			}
			user.Connect(&lib.WsConn{Conn: conn})
			// TODO does the client need to clear their current cookie?
			if err = sendCookie(user); err != nil {
				return nil, err
			}
		} else {
			// happiest path
			user.Connect(&lib.WsConn{Conn: conn})
		}
	}

	user.Send(map[string]interface{}{"Type": "me", "Me": user})

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
