package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

type settings struct {
	dataurl  string
	estopurl string
	errorurl string
	interval int
}

func StartWebService() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"10.10.0.0/255", "10.101.20.0/255", "10.101.10.0/255"})
	r.GET("/data", GetSimData)
	r.LoadHTMLGlob("assets/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"setinterval": fetchInterval,
			"setestopurl": ntfyEstopURL,
			"seterrorurl": ntfyErrURL,
			"setdataurl":  dataURL,
			"active":      activeSims,
			"idle":        idleSims,
			"errorlist":   errorList,
			"datatime":    latestFetch,
		})
	})
	r.GET("/settings", func(c *gin.Context) {
		c.HTML(200, "settings.html", gin.H{
			"setinterval": fetchInterval,
			"setestopurl": ntfyEstopURL,
			"seterrorurl": ntfyErrURL,
			"setdataurl":  dataURL,
		})
	})

	r.POST("/", func(c *gin.Context) {
		s := settings{
			dataurl:  c.PostForm("dataurl"),
			estopurl: "https://ntfy.sh/" + c.PostForm("estopurl"),
			errorurl: "https://ntfy.sh/" + c.PostForm("errorurl"),
			interval: ParseInt(c.PostForm("interval")),
		}
		if s.dataurl != "" {
			dataURL = s.dataurl
		} else if strings.ToLower(s.dataurl) == "default" {
			dataURL = "http://10.101.20.10:3000/game-servers/daemon-states"
		}
		if s.interval > 4 {
			fetchInterval = time.Duration(s.interval) * time.Second
		}
		if s.estopurl != "https://ntfy.sh/" {
			ntfyEstopURL = s.estopurl
		}
		if s.errorurl != "https://ntfy.sh/" {
			ntfyErrURL = s.errorurl
		}
		c.HTML(200, "index.html", gin.H{
			"setinterval": fetchInterval,
			"setestopurl": ntfyEstopURL,
			"seterrorurl": ntfyErrURL,
			"setdataurl":  dataURL,
			"active":      activeSims,
			"idle":        idleSims,
			"errorlist":   errorList,
		})

	})
	r.Run(":1234")
}

func GetSimData(c *gin.Context) {
	c.JSON(http.StatusOK, latestStates)
}
