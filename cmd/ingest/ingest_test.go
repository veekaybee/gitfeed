package main

import (
	"database/sql"
	"gitfeed/db"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestProcessPost(t *testing.T) {

	post := db.ATPost{
		Did:    "did:plc:7ywxd6gcvpmgw3q33dg6xnxf",
		TimeUs: 1703088300000000, // Dec 20, 2024 15:45:00 UTC
		Type:   "create",
		Kind:   "app.bsky.feed.post",
		Commit: struct {
			Rev        string `json:"rev"`
			Type       string `json:"type"`
			Operation  string `json:"operation"`
			Collection string `json:"collection"`
			Rkey       string `json:"rkey"`
			Record     struct {
				Type      string    `json:"$type"`
				CreatedAt time.Time `json:"createdAt"`
				Embed     struct {
					Type     string `json:"$type"`
					External struct {
						Description string `json:"description"`
						Title       string `json:"title"`
						URI         string `json:"uri"`
					} `json:"external"`
				} `json:"embed"`
				Facets []struct {
					Features []struct {
						Type string `json:"$type", omitempty`
						URI  string `json:"uri", omitempty`
					} `json:"features"`
					Index struct {
						ByteEnd   int `json:"byteEnd"`
						ByteStart int `json:"byteStart"`
					} `json:"index"`
				} `json:"facets"`
				Langs []string `json:"langs", omitempty`
				Text  string   `json:"text"`
			} `json:"record"`
			Cid string `json:"cid"`
		}(struct {
			Rev        string `json:"rev"`
			Type       string `json:"type"`
			Operation  string `json:"operation"`
			Collection string `json:"collection"`
			Rkey       string `json:"rkey"`
			Record     struct {
				Type      string    `json:"$type"`
				CreatedAt time.Time `json:"createdAt"`
				Embed     struct {
					Type     string `json:"$type"`
					External struct {
						Description string `json:"description"`
						Title       string `json:"title"`
						URI         string `json:"uri"`
					} `json:"external"`
				} `json:"embed"`
				Facets []struct {
					Features []struct {
						Type string `json:"$type,omitempty"`
						URI  string `json:"uri,omitempty"`
					} `json:"features"`
					Index struct {
						ByteEnd   int `json:"byteEnd"`
						ByteStart int `json:"byteStart"`
					} `json:"index"`
				} `json:"facets"`
				Langs []string `json:"langs,omitempty"`
				Text  string   `json:"text"`
			} `json:"record"`
			Cid string `json:"cid"`
		}{
			Rev:        "3jdkeis8fj",
			Type:       "app.bsky.feed.post",
			Operation:  "create",
			Collection: "app.bsky.feed.post",
			Rkey:       "3jsu47dlw9",
			Record: struct {
				Type      string    `json:"$type"`
				CreatedAt time.Time `json:"createdAt"`
				Embed     struct {
					Type     string `json:"$type"`
					External struct {
						Description string `json:"description"`
						Title       string `json:"title"`
						URI         string `json:"uri"`
					} `json:"external"`
				} `json:"embed"`
				Facets []struct {
					Features []struct {
						Type string `json:"$type,omitempty"`
						URI  string `json:"uri,omitempty"`
					} `json:"features"`
					Index struct {
						ByteEnd   int `json:"byteEnd"`
						ByteStart int `json:"byteStart"`
					} `json:"index"`
				} `json:"facets"`
				Langs []string `json:"langs,omitempty"`
				Text  string   `json:"text"`
			}{
				Type:      "app.bsky.feed.post",
				CreatedAt: time.Date(2024, 12, 20, 15, 45, 0, 0, time.UTC),
				Embed: struct {
					Type     string `json:"$type"`
					External struct {
						Description string `json:"description"`
						Title       string `json:"title"`
						URI         string `json:"uri"`
					} `json:"external"`
				}{
					Type: "app.bsky.embed.external",
					External: struct {
						Description string `json:"description"`
						Title       string `json:"title"`
						URI         string `json:"uri"`
					}{
						Description: "Discover the latest advances in distributed systems and their practical applications in modern software architecture.",
						Title:       "Understanding Distributed Systems in 2024",
						URI:         "https://tech-articles.example.com/distributed-systems-2024",
					},
				},
				Facets: []struct {
					Features []struct {
						Type string `json:"$type,omitempty"`
						URI  string `json:"uri,omitempty"`
					} `json:"features"`
					Index struct {
						ByteEnd   int `json:"byteEnd"`
						ByteStart int `json:"byteStart"`
					} `json:"index"`
				}{
					{
						Features: []struct {
							Type string `json:"$type,omitempty"`
							URI  string `json:"uri,omitempty"`
						}{
							{
								Type: "app.bsky.richtext.facet#mention",
								URI:  "at://did:plc:4xj4pq5yuxxy6yh6tropical/profile",
							},
						},
						Index: struct {
							ByteEnd   int `json:"byteEnd"`
							ByteStart int `json:"byteStart"`
						}{
							ByteStart: 0,
							ByteEnd:   8,
						},
					},
					{
						Features: []struct {
							Type string `json:"$type,omitempty"`
							URI  string `json:"uri,omitempty"`
						}{
							{
								Type: "app.bsky.richtext.facet#link",
								URI:  "https://github.com/distributed-systems-2024",
							},
						},
						Index: struct {
							ByteEnd   int `json:"byteEnd"`
							ByteStart int `json:"byteStart"`
						}{
							ByteStart: 64,
							ByteEnd:   127,
						},
					},
				},
				Langs: []string{"en"},
				Text:  "@xzy Check out this fascinating article on distributed systems! https://github.com/distributed-systems-2024 #tech #distributed",
			},
			Cid: "bafyreib2rxk3rqpbswxhicg4x3nqwfxwyfqrj5luzb7pwxixphv5a2",
		}),
	}
	want := db.DBPost{
		Did:        "did:plc:7ywxd6gcvpmgw3q33dg6xnxf",
		TimeUs:     1703088300000000,
		Kind:       "app.bsky.feed.post",
		Operation:  "create",
		Collection: "app.bsky.feed.post",
		Rkey:       "3jsu47dlw9",
		Cid:        "bafyreib2rxk3rqpbswxhicg4x3nqwfxwyfqrj5luzb7pwxixphv5a2",
		Type:       "app.bsky.feed.post",
		CreatedAt:  time.Date(2024, 12, 20, 15, 45, 0, 0, time.UTC),
		Langs: sql.Null[string]{
			V:     "en",
			Valid: true,
		},
		Text: "@xzy Check out this fascinating article on distributed systems! https://github.com/distributed-systems-2024 #tech #distributed",
		URI:  "https://github.com/distributed-systems-2024",
	}
	got := ProcessPost(post)
	assert.Equal(t, want, got, "values should match")

}
