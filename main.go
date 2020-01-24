package main

import (
	"fmt"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx"
	"github.com/labstack/echo"
	"tech-db/cmd/api/handlers"
	"tech-db/internal/forum"
)

var (
	connectionString = "postgres://forum:forum@localhost:5432/forum?sslmode=disable"
	host             = "0.0.0.0:5000"
	maxConn = 2000
)

func main() {
	config, err := pgx.ParseURI(connectionString)
	if err != nil {
		fmt.Println(err)
		return
	}

	db, err := pgx.NewConnPool(
		pgx.ConnPoolConfig{
			ConnConfig:     config,
			MaxConnections: maxConn,
		})
	if err != nil {
		fmt.Println(err)
		return
	}

	userService := forum.NewUserService(db)
	threadService := forum.NewThreadService(db)
	forumService := forum.NewForumService(db)
	postService := forum.NewPostService(db)

	user := handlers.User{UserService: userService}
	forum := handlers.Forum{ForumService: forumService, UserService: userService, ThreadService: threadService}
	post := handlers.Post{PostService: postService, ForumService: forumService, UserService: userService, ThreadService: threadService}

	e := echo.New()

	e.POST("/api/user/:nickname/create", user.CreateUser)
	e.GET("/api/user/:nickname/profile", user.GetProfile)
	e.POST("/api/user/:nickname/profile", user.EditProfile)

	e.POST("/api/forum/create", forum.CreateForum)
	e.POST("/api/forum/:slug/create", forum.CreateThread)
	e.GET("/api/forum/:slug/details", forum.GetForumDetails)
	e.GET("/api/forum/:slug/threads", forum.GetForumThreads)
	e.GET("/api/forum/:slug/users", forum.GetForumUsers)

	e.GET("/api/post/:id/details", post.GetFullPost)
	e.POST("/api/post/:id/details", post.EditMessage)

	e.GET("/api/thread/:slug_or_id/details", post.GetThread)
	e.POST("/api/thread/:slug_or_id/details", post.EditThread)
	e.GET("/api/thread/:slug_or_id/posts", post.GetPosts)
	e.POST("/api/thread/:slug_or_id/create", post.CreatePosts)
	e.POST("/api/thread/:slug_or_id/vote", post.CreateVote)

	e.POST("/api/service/clear", forum.Clean)
	e.GET("/api/service/status", forum.Status)

	e.Logger.Warnf("start listening on %s", host)
	if err := e.Start(host); err != nil {
		e.Logger.Errorf("server error: %s", err)
	}

	e.Logger.Warnf("shutdown")

}
