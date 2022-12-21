package controller

import (
	"github.com/gin-gonic/gin"
)

type RestController interface {
	Get(ctx *gin.Context)
	List(ctx *gin.Context)
}
