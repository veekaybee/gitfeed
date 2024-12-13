package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gitfeed/db"
	"gitfeed/handlers"
	"os"
	"os/signal"
	"sync"
	"time"

	"log"
	"strings"

	"github.com/gorilla/websocket"
)

var (
	ErrMaxReconnectsExceeded = errors.New("maximum reconnection attempts exceeded")
)

const (
	maxMessageSize = 512 * 1024 // 512KB
	pongWait       = 60 * time.Second
)

type WebSocketManager struct {
	url            string
	reconnectDelay time.Duration
	maxReconnects  int
	writeWait      time.Duration
	readWait       time.Duration
	pingPeriod     time.Duration

	conn           *websocket.Conn
	mu             sync.Mutex
	done           chan struct{}
	isConnected    bool
	reconnectCount int

	messageHandler func([]byte)
	errorHandler   func(error)

	postRepo *db.PostRepository
}

func NewWebSocketManager(url string, postRepo *db.PostRepository) *WebSocketManager {
	wsm := &WebSocketManager{
		url:            url,
		reconnectDelay: 3 * time.Second,
		writeWait:      10 * time.Second,
		pingPeriod:     (pongWait * 9) / 10,
		done:           make(chan struct{}),
		postRepo:       postRepo,
		errorHandler:   func(err error) { log.Printf("Error: %v", err) },
	}

	return wsm
}

func (w *WebSocketManager) Connect(ctx context.Context) {

	for !w.isConnected {

		time.Sleep(w.reconnectDelay)
		log.Printf("Connecting to %s", w.url)

		dialer := websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		}

		conn, _, err := dialer.DialContext(ctx, w.url, nil)
		if err != nil {
			w.isConnected = false
			continue
		}

		w.conn = conn
		w.isConnected = true
	}
}

// GitHub matches
func FindMatches(text, pattern string) bool {

	return strings.Contains(text, pattern)

}

func (w *WebSocketManager) readPump(ctx context.Context) {

	w.Connect(ctx)
	counter := 0
	for {
		select {
		case <-ctx.Done():
			log.Printf("Exiting readPump: got kill signal\n")
			return
		default:
			var post db.ATPost
			if err := w.conn.ReadJSON(&post); err != nil {
				w.Connect(ctx)
				continue
			}
			counter++
			if counter%100 == 0 {
				log.Printf("Read %d posts\n", counter)
			}

			// Process the post
			if found := FindMatches(post.Commit.Record.Text, "github.com"); found {
				log.Printf("Post: %v", post)

				var langs sql.Null[string]
				if len(post.Commit.Record.Langs) > 0 {
					langs.Valid = true
					langs.V = post.Commit.Record.Langs[0]
				}

				uri := handlers.ExtractUri(post)
				if uri != "" && FindMatches(uri, "github.com") {
					dbPost := db.DBPost{
						Did:        post.Did,
						TimeUs:     post.TimeUs,
						Kind:       post.Kind,
						Operation:  post.Commit.Operation,
						Collection: post.Commit.Collection,
						Rkey:       post.Commit.Rkey,
						Cid:        post.Commit.Cid,
						Type:       post.Commit.Record.Type,
						CreatedAt:  post.Commit.Record.CreatedAt,
						Langs:      langs,
						Text:       post.Commit.Record.Text,
						URI:        uri,
					}

					if err := w.postRepo.WritePost(dbPost); err != nil {
						w.errorHandler(fmt.Errorf("failed to write post: %v", err))
						continue
					}
					log.Printf("Wrote Post %v", dbPost.Did)
				}
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

	wsManager := NewWebSocketManager(
		"wss://jetstream2.us-west.bsky.network/subscribe?wantedCollections=app.bsky.feed.post",
		pr,
	)
	wsManager.reconnectDelay = 5 * time.Second

	log.Printf("connecting to %s\n", wsManager.url)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	wsManager.readPump(ctx)
}
