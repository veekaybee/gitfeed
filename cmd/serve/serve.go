package main

import (
	"fmt"
	"gitfeed/db"
	"gitfeed/routes"
	"log"
	"net/http"
	"strings"

	"gitfeed/handlers"

	"github.com/gorilla/websocket"
)

func FindMatches(text, pattern string) bool {

	return strings.Contains(text, pattern)

}

func GetPost(c *websocket.Conn) db.ATPost {
	for {
		post := db.ATPost{}
		err := c.ReadJSON(&post)
		if err != nil {
			log.Println("read:", err)
		}
		found := FindMatches(post.Commit.Record.Text, "github.com")
		if found {
			return post
		}
	}

}

func main() {

	fmt.Println("Starting DB...")

	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Start post service
	pr := db.NewPostRepository(database)
	postService := &handlers.PostService{PostRepository: pr}

	// Create web routes
	routes.CreateRoutes(postService)

	log.Printf("Starting gitfeed server...")
	log.Fatal(http.ListenAndServe(":8000", nil))

}
