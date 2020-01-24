package forum

import (
	"github.com/jackc/pgx"
)

type ForumService struct {
	db *pgx.ConnPool
}

func NewForumService(db *pgx.ConnPool) *ForumService {
	return &ForumService{db: db}
}

func (fs *ForumService) SelectFullForumBySlug(slug string) (forum Forum, err error) {
	sqlQuery := `SELECT f.slug, f.title, f.user FROM forum as f where f.slug=$1`
	err = fs.db.QueryRow(sqlQuery, slug).Scan(&forum.Slug, &forum.Title, &forum.User)
	if err != nil {
		return
	}
	sqlQuery = `SELECT count(*) FROM thread as t where t.forum=$1`
	err = fs.db.QueryRow(sqlQuery, slug).Scan(&forum.Threads)
	if err != nil {
		return
	}
	sqlQuery = `
	SELECT count(*) FROM post as p where p.forum=$1`
	err = fs.db.QueryRow(sqlQuery, slug).Scan(&forum.Posts)
	return
}

func (fs *ForumService) SelectForumBySlug(slug string) (forum Forum, err error) {
	sqlQuery := `
	SELECT f.id, f.slug, f.title, f.user, f.threads, f.posts FROM forum as f where f.slug = $1`
	err = fs.db.QueryRow(sqlQuery, slug).Scan(&forum.Id, &forum.Slug, &forum.Title, &forum.User, &forum.Threads, &forum.Posts)
	return
}

func (fs *ForumService) InsertForum(forum Forum) (err error) {
	sqlQuery := `INSERT INTO forum (slug, title, "user") VALUES ($1,$2,$3)`
	_, err = fs.db.Exec(sqlQuery, forum.Slug, forum.Title, forum.User)
	return
}

func (fs *ForumService) Clean() (err error) {
	sqlQuery := `TRUNCATE vote, post, thread, forum, "user", forum_user RESTART IDENTITY CASCADE;`
	_, err = fs.db.Exec(sqlQuery)
	return
}

func (fs *ForumService) SelectStatus() (status Status, err error) {
	sqlQuery := `
	SELECT *
	FROM (SELECT COUNT(*) AS post FROM post) AS Post,
		 (SELECT COUNT(*) AS thread FROM thread) AS Thread,
		 (SELECT COUNT(*) AS forum FROM forum) AS Forum,
		 (SELECT COUNT(*) AS "user" FROM "user") AS Users;`
	err = fs.db.QueryRow(sqlQuery).Scan(&status.Post, &status.Thread, &status.Forum, &status.User)
	return
}

func (fs *ForumService) UpdateThreadCount(forumId int) (err error) {
	sqlQuery := `
	UPDATE forum SET threads=threads+1 WHERE forum.id=$1`
	_, err = fs.db.Exec(sqlQuery, forumId)
	return
}

func (fs *ForumService) UpdatePostCount(forum string, count int) (err error) {
	sqlQuery := `
	UPDATE forum SET posts=posts+$2 WHERE forum.slug=$1`
	_, err = fs.db.Exec(sqlQuery, forum, count)
	return
}

func (fs *ForumService) InsertForumUser(forumId int, userId int) (err error) {
	sqlQuery := `
	INSERT INTO forum_user (forum_id, user_id) VALUES ($1,$2)`
	_, err = fs.db.Exec(sqlQuery, forumId, userId)
	return
}
