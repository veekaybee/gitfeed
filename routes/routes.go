package routes

import (
	"gitfeed/handlers"
	"net/http"
)

func CreateRoutes(postService *handlers.PostService) {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("GET /static/favicon.ico", fs)
	http.Handle("GET /", fs)

	/*Post Routes*/
	http.HandleFunc("DELETE /api/v1/post/{id}", postService.PostDeleteHandler)
	http.HandleFunc("DELETE /api/v1/posts", postService.PostsDeleteHandler)
	http.HandleFunc("GET /api/v1/post/{id}", postService.PostGetHandler)

	http.HandleFunc("GET /api/v1/posts", postService.PostsGetHandler)
	http.HandleFunc("GET /api/v1/timestamp", postService.TimeStampGetHandler)
	http.HandleFunc("GET /api/v1/github/{username}/{repository}", handlers.HandleGitHubRepo)

}
