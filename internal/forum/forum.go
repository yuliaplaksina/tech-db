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
	sqlQuery := `SELECT f.slug, f.title, f.user
	FROM public.forum as f
	where lower(f.slug) = lower($1)`
	err = fs.db.QueryRow(sqlQuery, slug).Scan(&forum.Slug, &forum.Title, &forum.User)
	if err != nil {
		return
	}
	sqlQuery = `SELECT count(*)
	FROM public.thread as t
	where lower(t.forum) = lower($1)`
	err = fs.db.QueryRow(sqlQuery, slug).Scan(&forum.Threads)
	if err != nil {
		return
	}
	sqlQuery = `
	SELECT count(*)
	FROM public.post as p
	where lower(p.forum) = lower($1)`
	err = fs.db.QueryRow(sqlQuery, slug).Scan(&forum.Posts)
	return
}

func (fs *ForumService) SelectForumBySlug(slug string) (forum Forum, err error) {
	sqlQuery := `
	SELECT f.id, f.slug, f.title, f.user, f.threads, f.posts
	FROM public.forum as f
	where lower(f.slug) = lower($1)`
	err = fs.db.QueryRow(sqlQuery, slug).Scan(&forum.Id, &forum.Slug, &forum.Title, &forum.User, &forum.Threads, &forum.Posts)
	return
}

func (fs *ForumService) InsertForum(forum Forum) (err error) {
	sqlQuery := `INSERT INTO public.forum (slug, title, "user")
	VALUES ($1,$2,$3)`
	_, err = fs.db.Exec(sqlQuery, forum.Slug, forum.Title, forum.User)
	return
}

func (fs *ForumService) Clean() (err error) {
	sqlQuery := `TRUNCATE public.vote, public.post, public.thread, public.forum, public.user, public.forum_user RESTART IDENTITY CASCADE;`
	_, err = fs.db.Exec(sqlQuery)
	return
}

func (fs *ForumService) SelectStatus() (status Status, err error) {
	sqlQuery := `
	SELECT *
	FROM (SELECT COUNT(*) AS post FROM public.post) AS Post,
		 (SELECT COUNT(*) AS thread FROM public.thread) AS Thread,
		 (SELECT COUNT(*) AS forum FROM public.forum) AS Forum,
		 (SELECT COUNT(*) AS "user" FROM public.user) AS Users;`
	err = fs.db.QueryRow(sqlQuery).Scan(&status.Post, &status.Thread, &status.Forum, &status.User)
	return
}

func (fs *ForumService) UpdateThreadCount(forumId int) (err error) {
	sqlQuery := `
	UPDATE public.forum SET threads = threads + 1
	WHERE forum.id = $1`
	_, err = fs.db.Exec(sqlQuery, forumId)
	return
}

func (fs *ForumService) UpdatePostCount(forum string, count int) (err error) {
	sqlQuery := `
	UPDATE public.forum SET posts = posts + $2
	WHERE Lower(forum.slug) = Lower($1)`
	_, err = fs.db.Exec(sqlQuery, forum, count)
	return
}

func (fs *ForumService) InsertForumUser(forumId int, userId int) (err error) {
	sqlQuery := `
	INSERT INTO public.forum_user (forum_id, user_id)
	VALUES ($1,$2)`
	_, err = fs.db.Exec(sqlQuery, forumId, userId)
	return
}
