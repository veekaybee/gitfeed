package main

import (
	"context"
	"database/sql"
	"fmt"
	"gitfeed/db"
	"gitfeed/handlers"
	"time"

	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/gorilla/websocket"
)

func FindMatches(text, pattern string) bool {

	return strings.Contains(text, pattern)

}

// https://github.com/gorilla/websocket/blob/main/examples/command/main.go
const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Time to wait before force close on connection.
	closeGracePeriod = 10 * time.Second
)

func IngestPosts(ws *websocket.Conn, pr *db.PostRepository) {
	defer ws.Close()
	ws.SetReadLimit(maxMessageSize)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		p := db.ATPost{}
		err := ws.ReadJSON(&p)
		if err != nil {
			break
		}

		found := FindMatches(p.Commit.Record.Text, "github.com")
		if found {
			log.Printf("Post: %v", p)

			var langs sql.Null[string]
			if len(p.Commit.Record.Langs) > 0 {
				langs.Valid = true
				langs.V = p.Commit.Record.Langs[0]
			}
			uri := handlers.ExtractUri(p)

			if uri != "" && FindMatches(uri, "github.com") {
				post := db.DBPost{
					Did:        p.Did,
					TimeUs:     p.TimeUs,
					Kind:       p.Kind,
					Operation:  p.Commit.Operation,
					Collection: p.Commit.Collection,
					Rkey:       p.Commit.Rkey,
					Cid:        p.Commit.Cid,
					Type:       p.Commit.Record.Type,
					CreatedAt:  p.Commit.Record.CreatedAt,
					Langs:      langs,
					Text:       p.Commit.Record.Text,
					URI:        uri,
				}

				err = pr.WritePost(post)
				if err != nil {
					log.Fatalf("Failed to write row: %v", err)
				}
				log.Printf("Wrote Post %v", post.Did)
			}
		}

	}
}

func main() {
	fmt.Println("Starting DB...")

	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	defer database.Close()

	pr := db.NewPostRepository(database)

	postTableColumns := map[string]string{
		"id":                "INTEGER PRIMARY KEY AUTOINCREMENT",
		"did":               "TEXT NOT NULL",
		"time_us":           "INTEGER NOT NULL",
		"kind":              "TEXT NOT NULL",
		"commit_rev":        "TEXT NOT NULL",
		"commit_operation":  "TEXT NOT NULL",
		"commit_collection": "TEXT NOT NULL",
		"commit_rkey":       "TEXT NOT NULL",
		"commit_cid":        "TEXT NOT NULL",
		"record_type":       "TEXT NOT NULL",
		"record_created_at": "DATETIME NOT NULL",
		"record_langs":      "TEXT",
		"record_text":       "TEXT",
		"record_uri":        "TEXT",
	}

	// Create the table if it doesn't exist
	err = pr.CreateTableIfNotExists("posts", postTableColumns)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// start collection
	fmt.Println("Starting feed...")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Instantiatiate JetStream Feed
	u := "wss://jetstream2.us-west.bsky.network/subscribe?wantedCollections=app.bsky.feed.post"

	log.Printf("connecting to %s", u)

	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		log.Fatalln("dial:", err)
	}
	defer c.Close()

	go IngestPosts(c, pr)

	<-ctx.Done()
}
