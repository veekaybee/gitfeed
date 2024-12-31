package main

import (
	"fmt"
	"gitfeed/db"
	"gitfeed/routes"
	"log"
	"net/http"

	"gitfeed/handlers"
)

func main() {

	fmt.Println("Starting DB...")

	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start post service
	fmt.Println("Connect to post service...")
	pr := db.NewPostRepository(database)
	postService := &handlers.PostService{PostRepository: pr}

	// Create web routes
	routes.CreateRoutes(postService)

	log.Printf("Starting gitfeed server...")
	log.Fatal(http.ListenAndServe(":80", nil))

}
