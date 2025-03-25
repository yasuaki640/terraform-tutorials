package main

import (
	"fmt"
	"github.com/yasuaki640/terraform-tutorials/aws-cognito-sample/sample-application-client/config"
	"github.com/yasuaki640/terraform-tutorials/aws-cognito-sample/sample-application-client/handlers"
	"log"
	"net/http"
)

func main() {
	config.Init()

	http.HandleFunc("/", handlers.HandleHome)
	http.HandleFunc("/login", handlers.HandleLogin)
	http.HandleFunc("/logout", handlers.HandleLogout)
	http.HandleFunc("/callback", handlers.HandleCallback)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
