package server

import (
	"github.com/gin-gonic/gin"
	"github.com/jakecoffman/gorest"
	"github.com/jakecoffman/lobby/lib"
	"gopkg.in/mgo.v2/bson"
)

type UserController struct {
	gorest.MongoController
}

func (c *UserController) Delete(ctx *gin.Context) {
	id := bson.ObjectIdHex(ctx.Param("id"))

	user := c.Resource.New().(*lib.User)
	if err := c.C.FindId(id).One(user); err != nil {
		ctx.JSON(404, bson.M{"error": err.Error()})
		return
	}

	user.IsDeleted = true
	if err := c.C.UpdateId(id, user); err != nil {
		ctx.JSON(404, bson.M{"error": "not found"})
		return
	}

	ctx.JSON(200, user)
}
