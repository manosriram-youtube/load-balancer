package main

import (
	"log"
	"simple-load-balancer/pkg/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// should init all variables
	err := handler.Init()
	if err != nil {
		log.Fatal("error initializing yaml file ", err)
	}

	// should listen for incoming requests and load balance them according to the algorithm given
	// to be implemented: round-robin, stickyIP
	r.GET("/", handler.Balance)
	r.Run(":8081")
}
