package main

import (
	"fmt"
	"gotalk-api/config"
)

func main() {
	config.LoadEnv()
	port := config.GetEnv("APP_PORT", "3000")
	fmt.Println("PORT:", port)
}
