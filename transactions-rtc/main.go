package main

import (
	"fmt"
	"net/http"
	api "transactions_rtc/api"
)

// @title           Transactions API
// @version         1.0
// @description     Simple get/post transactions server

// @host      localhost:8080
// @BasePath  /
// @accept plain
// @produce json
// @schemes http https

func main() {
	http.Handle("/", api.NewMyApi())
	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
