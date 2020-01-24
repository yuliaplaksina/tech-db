package main

import (
	"fmt"
	"github.com/jackc/pgx"
	_ "github.com/jackc/pgx"
	"github.com/labstack/echo"
	"io/ioutil"
	"tech-db/cmd/api/handlers"
	"tech-db/internal/forum"
)

var (
	connectionString = "postgres://forum_user:testpass@localhost:5432/forum_db?sslmode=disable"
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

	err = LoadSchemaSQL(db)
	if err != nil {
		fmt.Println(err)
	}
	userService := forum.NewUserService(db)
	threadService := forum.NewThreadService(db)
	forumService := forum.NewForumService(db)
	postService := forum.NewPostService(db)

	user := handlers.User{UserService: userService}
	forum := handlers.Forum{ForumService: forumService, UserService: userService, ThreadService: threadService}
	post := handlers.Post{PostService: postService, ForumService: forumService, UserService: userService, ThreadService: threadService}

	e := echo.New()
	prefix := "/api"
	e.POST(prefix+"/user/:nickname/create", user.CreateUser)
	e.GET(prefix+"/user/:nickname/profile", user.GetProfile)
	e.POST(prefix+"/user/:nickname/profile", user.EditProfile)

	e.POST(prefix+"/forum/create", forum.CreateForum)
	e.POST(prefix+"/forum/:slug/create", forum.CreateThread)
	e.GET(prefix+"/forum/:slug/details", forum.GetForumDetails)
	e.GET(prefix+"/forum/:slug/threads", forum.GetForumThreads)
	e.GET(prefix+"/forum/:slug/users", forum.GetForumUsers)

	e.GET(prefix+"/post/:id/details", post.GetFullPost)
	e.POST(prefix+"/post/:id/details", post.EditMessage)

	e.GET(prefix+"/thread/:slug_or_id/details", post.GetThread)
	e.POST(prefix+"/thread/:slug_or_id/details", post.EditThread)
	e.GET(prefix+"/thread/:slug_or_id/posts", post.GetPosts)
	e.POST(prefix+"/thread/:slug_or_id/create", post.CreatePosts)
	e.POST(prefix+"/thread/:slug_or_id/vote", post.CreateVote)

	e.POST(prefix+"/service/clear", forum.Clean)
	e.GET(prefix+"/service/status", forum.Status)

	e.Logger.Warnf("start listening on %s", host)
	if err := e.Start(host); err != nil {
		e.Logger.Errorf("server error: %s", err)
	}

	e.Logger.Warnf("shutdown")

}

const dbSchema = "dum_hw_pdb.sql"

func LoadSchemaSQL(db *pgx.ConnPool) error {

	content, err := ioutil.ReadFile(dbSchema)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(string(content)); err != nil {
		return err
	}
	tx.Commit()
	return nil
}
