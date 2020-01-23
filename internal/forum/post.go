package forum

import (
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type PostService struct {
	db *pgx.ConnPool
}

func NewPostService(db *pgx.ConnPool) *PostService {
	return &PostService{db: db}
}

func (ps *PostService) SelectPostById(id int) (post Post, err error) {
	sqlQuery := `SELECT p.author, p.created, p.forum, p.id, p.is_edited, p.message, p.parent, p.thread
	FROM public.post as p
	where p.id = $1`
	err = ps.db.QueryRow(sqlQuery, id).Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)
	return
}

func (ps *PostService) FindPostById(id int, thread int) (err error) {
	sqlQuery := `SELECT p.id
	FROM public.post as p
	where p.id = $1 AND p.thread = $2`
	var postId int64
	err = ps.db.QueryRow(sqlQuery, id, thread).Scan(&postId)
	return
}

func (ps *PostService) InsertPost(post Post) (lastId int, err error) {
	sqlQuery := `INSERT INTO public.post (author, created, forum, message, parent, thread)
	VALUES ($1,$2,$3,$4,$5,$6)
	RETURNING id`
	err = ps.db.QueryRow(sqlQuery, post.Author, post.Created, post.Forum, post.Message, post.Parent, post.Thread).Scan(&lastId)
	return
}

func (ps *PostService) UpdatePostMessage(newMessage string, id int) (countUpdateString int64, err error) {
	sqlQuery := `UPDATE public.post SET message = $1,
                       is_edited = true
	where post.id = $2`
	result, err := ps.db.Exec(sqlQuery, newMessage, id)
	if err != nil {
		return
	}
	countUpdateString = result.RowsAffected()
	return
}

func (ps *PostService) CreatePosts(thread Thread, forumId int, created string, posts []Post) (post []Post, err error) {
	/*	tx, err := ps.db.Begin()
		if err != nil {
			return nil, err
		}
	defer tx.Rollback()*/

	sqlStr := "INSERT INTO public.post(id, parent, thread, forum, author, created, message, path) VALUES "
	vals := []interface{}{}
	for _, post := range posts {
		var authorId int
		err = ps.db.QueryRow(`SELECT "user".id FROM public."user" WHERE LOWER("user".nick_name) = LOWER($1)`,
			post.Author,
		).Scan(&authorId)
		if err != nil {
			return nil, errors.New("404")
		}
		sqlQuery := `
		INSERT INTO public.forum_user (forum_id, user_id)
		VALUES ($1,$2)`
		_, _ = ps.db.Exec(sqlQuery, forumId, authorId)

		if post.Parent == 0 {
			sqlStr += "(nextval('public.post_id_seq'::regclass), ?, ?, ?, ?, ?, ?, " +
				"ARRAY[currval(pg_get_serial_sequence('public.post', 'id'))::bigint]),"
			vals = append(vals, post.Parent, thread.Id, thread.Forum, post.Author, created, post.Message)
		} else {
			var parentThreadId int32
			err = ps.db.QueryRow("SELECT post.thread FROM public.post WHERE post.id = $1",
				post.Parent,
			).Scan(&parentThreadId)
			if err != nil {
				return nil, err
			}
			if parentThreadId != int32(thread.Id) {
				return nil, errors.New("Parent post was created in another thread")
			}

			sqlStr += " (nextval('public.post_id_seq'::regclass), ?, ?, ?, ?, ?, ?, " +
				"(SELECT post.path FROM public.post WHERE post.id = ? AND post.thread = ?) || " +
				"currval(pg_get_serial_sequence('public.post', 'id'))::bigint),"

			vals = append(vals, post.Parent, thread.Id, thread.Forum, post.Author, created, post.Message, post.Parent, thread.Id)
		}

	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")

	sqlStr += " RETURNING  id, parent, thread, forum, author, created, message, is_edited "

	sqlStr = ReplaceSQL(sqlStr, "?")
	if len(posts) > 0 {
		rows, err := ps.db.Query(sqlStr, vals...)
		if err != nil {
			return nil, err
		}
		scanPost := Post{}
		for rows.Next() {
			err := rows.Scan(
				&scanPost.Id,
				&scanPost.Parent,
				&scanPost.Thread,
				&scanPost.Forum,
				&scanPost.Author,
				&scanPost.Created,
				&scanPost.Message,
				&scanPost.IsEdited,
			)
			if err != nil {
				rows.Close()
				return nil, err
			}
			post = append(post, scanPost)
		}
		rows.Close()
	}
	return post, nil
}

func ReplaceSQL(old, searchPattern string) string {
	tmpCount := strings.Count(old, searchPattern)
	for m := 1; m <= tmpCount; m++ {
		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
	}
	return old
}

/*func (ps *PostService) CheckPosts(threadId int, posts []Post) (err error) {
	_, err := h.UserService.FindUserByNickName(newPosts[i].Author)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can'ps find user"})
	}
	if newPosts[i].Parent != 0 {
		err = h.PostService.FindPostById(newPosts[i].Parent, newPosts[i].Thread)
		if err != nil {
			return ctx.JSON(http.StatusConflict, forum.ErrorMessage{Message: "Can'ps find post"})
		}
	}
}*/
