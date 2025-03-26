package main

import (
	"go-final/controller"

	"github.com/gin-gonic/gin"
)

func main() {
	// Release mode
	gin.SetMode(gin.ReleaseMode)
	controller.StartServer()
}
