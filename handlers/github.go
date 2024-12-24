package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

const githubAPIURL = "https://api.github.com"

func HandleGitHubRepo(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	repository := r.PathValue("repository")

	resp, err := http.Get(fmt.Sprintf("%s/repos/%s/%s", githubAPIURL, username, repository))
	if err != nil {
		log.Printf("Error getting GitHub repo data: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != 200 {
		log.Printf("Error getting GitHub repo data: %v", resp.Status)
		w.WriteHeader(resp.StatusCode)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error getting bytes: %v", err)
		w.WriteHeader(500)
	}
	w.Write(bytes)
	defer resp.Body.Close()
}
