package routes

import (
	"gitfeed/handlers"
	"net/http"
)

func CreateRoutes(postService *handlers.PostService) {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("GET /", fs)

	http.HandleFunc("GET /api/v1/hello/", handlers.HelloHandler)

	/*Post Routes*/
	http.HandleFunc("POST /api/v1/post/", postService.PostWriteHandler)
	http.HandleFunc("DELETE /api/v1/post/{id}", postService.PostDeleteHandler)
	http.HandleFunc("GET /api/v1/post/{id}", postService.PostGetHandler)

	http.HandleFunc("GET /api/v1/posts", postService.PostsGetHandler)
	http.HandleFunc("GET /api/v1/timestamp", postService.TimeStampGetHandler)

}
