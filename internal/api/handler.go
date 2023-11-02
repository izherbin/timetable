package api

import "github.com/gin-gonic/gin"

// Handler ...
type Handler struct {
	Method string
	Fn     func(*gin.Context)
}
