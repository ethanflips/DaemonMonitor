package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func StartWebService() {
	router := gin.Default()
	router.GET("/data", GetSimData)

	router.Run("localhost:1234")
}

func GetSimData(c *gin.Context) {
	c.JSON(http.StatusOK, latestStates)
}
