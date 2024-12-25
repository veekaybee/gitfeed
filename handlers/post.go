package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gitfeed/db"
	"io"
	"log"
	"net/http"
)

type PostRequest struct {
	Post []db.ATPost `json:"posts"`
}

type PostService struct {
	PostRepository db.PostRepo
}

func ExtractUri(p db.ATPost) string {
	var uri string
	for _, facet := range p.Commit.Record.Facets {
		for _, feature := range facet.Features {
			if feature.Type == "app.bsky.richtext.facet#link" {
				uri = feature.URI
			}
		}
	}
	return uri
}

func (ps *PostService) PostWriteHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var req PostRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return

	}

	for i, p := range req.Post {
		var langs sql.Null[string]
		if len(p.Commit.Record.Langs) > 0 {
			langs.Valid = true
			langs.V = p.Commit.Record.Langs[0]
		}
		uri := ExtractUri(p)
		if uri != "" {
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

			err = ps.PostRepository.WritePost(post)
			if err != nil {
				log.Fatalf("Failed to write row: %v", err)
			}
			log.Printf("Wrote Post %d %s", i, post.Did)
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Received %d posts successfully", len(req.Post))))

}

func (ps *PostService) PostDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := ps.PostRepository.DeletePost(id); err != nil {
		log.Printf("Failed to delete post: %s because %v", id, err)
		w.WriteHeader(500)
		return
	}
	log.Printf("Deleted post %s\n", id)

	w.WriteHeader(http.StatusOK)

}

func (ps *PostService) PostsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if err := ps.PostRepository.DeletePosts(); err != nil {
		log.Printf("Failed to delete post because %v", err)
		w.WriteHeader(500)
		return
	}
	log.Printf("Deleted all posts\n")

	w.WriteHeader(http.StatusOK)

}

func (ps *PostService) PostGetHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	post, err := ps.PostRepository.GetPost(id)
	if err != nil {
		http.Error(w, "Error fetching post", http.StatusBadRequest)
		return
	}
	log.Printf("Fetch post %v+\n %v+ ", id, post)

	w.WriteHeader(http.StatusOK)

}

func (ps *PostService) TimeStampGetHandler(w http.ResponseWriter, r *http.Request) {
	ts, err := ps.PostRepository.GetTimeStamp()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error fetching timestamp", http.StatusBadRequest)
		return
	}
	log.Printf("Fetch timestamp %d\n ", ts)

	w.WriteHeader(http.StatusOK)
	response := map[string]int64{"timestamp": ts / 1000}
	json.NewEncoder(w).Encode(response)
}

func (us *PostService) PostsGetHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Host, r.Method, r.RequestURI, r.RemoteAddr)
	posts, err := us.PostRepository.GetAllPosts()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error fetching posts", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(posts); err != nil {
		log.Printf("Error encoding posts to JSON: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Printf("Fetched and returned %d posts\n", len(posts))
}
