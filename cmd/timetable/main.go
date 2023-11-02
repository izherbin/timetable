package main

import (
	"log"
	"net/http"
	"time"

	ttAPI "github.com/sergrom/timetable/internal/api"

	"github.com/gin-gonic/gin"
)

const (
	APP_PORT = ":8899"
)

func main() {
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Static("/css", "web/css")
	router.Static("/js", "web/js")
	router.Static("/img", "web/img")
	router.Static("/fonts", "web/fonts")
	router.StaticFile("/favicon.ico", "web/favicon.ico")
	router.LoadHTMLGlob("web/*.html")

	for route, handler := range ttAPI.NewTimetableAPI().GetHandlers() {
		router.Handle(handler.Method, route, handler.Fn)
	}

	serv := &http.Server{
		Addr:        APP_PORT,
		Handler:     router,
		ReadTimeout: 3 * time.Second,
	}

	if err := serv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}

	log.Println("Program exited. Bye.")
}
