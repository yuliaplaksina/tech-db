package handlers

import (
	"github.com/jackc/pgx"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
	"strings"
	"tech-db/internal/forum"
	"time"
)

type Post struct {
	ForumService  *forum.ForumService
	UserService   *forum.UserService
	ThreadService *forum.ThreadService
	PostService   *forum.PostService
}

func (h *Post) GetFullPost(ctx echo.Context) error {
	idStr := ctx.Param("id")
	if idStr == "" {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0{
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}

	related := ctx.QueryParam("related")

	post, err := h.PostService.SelectPostById(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find post"})
		}
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}

	fullPost := forum.FullPost{Post: post}

	if strings.Contains(related, "user") {
		user, err := h.UserService.SelectUserByNickName(post.Author)
		if err != nil {
			if err == pgx.ErrNoRows {
				return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find user"})
			}
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
		}
		fullPost.Author = user
	}

	if strings.Contains(related, "forum") {
		fullForum, err := h.ForumService.SelectForumBySlug(post.Forum)
		if err != nil {
			if err == pgx.ErrNoRows {
				return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find forum"})
			}
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
		}
		fullPost.Forum = fullForum
	}

	if strings.Contains(related, "thread") {
		thread, err := h.ThreadService.SelectThreadById(post.Thread)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, "")
		}
		fullPost.Thread = thread
	}

	return ctx.JSON(http.StatusOK, fullPost)
}

func (h *Post) EditMessage(ctx echo.Context) error {
	idStr := ctx.Param("id")
	if idStr == "" {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}
	if id < 0 {
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}
	editMessage := forum.Message{}
	if err := ctx.Bind(&editMessage); err != nil {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}
	post, err := h.PostService.SelectPostById(id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}
	if editMessage.Message != "" && editMessage.Message != post.Message {
		num, err := h.PostService.UpdatePostMessage(editMessage.Message, id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
		}
		if num != 1 {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find post"})
		}
		post.Message = editMessage.Message
		post.IsEdited = true
	}

	return ctx.JSON(http.StatusOK, post)
}

func (h *Post) CreatePosts(ctx echo.Context) error {
	createdTime := time.Now().Format(time.RFC3339Nano)
	slugOrIdStr := ctx.Param("slug_or_id")
	newPosts := []forum.Post{}
	if err := ctx.Bind(&newPosts); err != nil {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}
	var thread forum.Thread
	id, err := strconv.Atoi(slugOrIdStr)
	if err != nil {
		slug := slugOrIdStr
		thread, err = h.ThreadService.FindThreadBySlug(slug)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
		}
	} else {
		thread, err = h.ThreadService.FindThreadById(id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
		}
	}
	if len(newPosts) == 0 {
		return ctx.JSON(http.StatusCreated, newPosts)
	}
	forumPosts, err := h.ForumService.SelectForumBySlug(thread.Forum)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find forum"})
	}

	posts, err := h.PostService.CreatePosts(thread,forumPosts.Id, createdTime, newPosts)
	if err != nil {
		if err.Error() == "404" {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find post author by nickname:"})
		}
		return ctx.JSON(http.StatusConflict, forum.ErrorMessage{Message: "Parent post was created in another thread"})
	}

	err = h.ForumService.UpdatePostCount(thread.Forum, len(newPosts))
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Unexpected error"})
	}

	return ctx.JSON(http.StatusCreated, posts)
}

func (h *Post) EditThread(ctx echo.Context) error {
	slugOrIdStr := ctx.Param("slug_or_id")
	var editThread forum.Thread
	if err := ctx.Bind(&editThread); err != nil {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}
	var thread forum.Thread
	id, err := strconv.Atoi(slugOrIdStr)
	if err != nil {
		slug := slugOrIdStr
		thread, err = h.ThreadService.FindThreadBySlug(slug)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
		}
	} else {
		thread, err = h.ThreadService.FindThreadById(id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
		}
	}
	if editThread.Message != "" {
		thread.Message = editThread.Message
	}
	if editThread.Title != "" {
		thread.Title = editThread.Title
	}
	if editThread.Message == "" && editThread.Title == "" {
		return ctx.JSON(http.StatusOK, thread)
	}
	err = h.ThreadService.UpdateThread(thread)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't update thread"})
	}
	return ctx.JSON(http.StatusOK, thread)
}
func (h *Post) CreateVote(ctx echo.Context) error {
	slugOrIdStr := ctx.Param("slug_or_id")
	var newVote forum.Vote
	if err := ctx.Bind(&newVote); err != nil {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}
	var thread forum.Thread
	id, err := strconv.Atoi(slugOrIdStr)
	if err != nil {
		slug := slugOrIdStr
		thread, err = h.ThreadService.FindThreadBySlug(slug)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
		}
	} else {
		thread, err = h.ThreadService.FindThreadById(id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
		}
	}

	user, err := h.UserService.FindUserByNickName(newVote.NickName)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find user"})
	}
	newVote.ThreadId = thread.Id
	newVote.UserId = user.Id
	vote, err := h.ThreadService.SelectVote(newVote)
	if err != nil {
		err = h.ThreadService.InsertVote(newVote)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't vote"})
		}
		err = h.ThreadService.UpdateVoteCount(newVote)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't vote"})
		}
	} else {
		_, err = h.ThreadService.UpdateVote(newVote)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't vote"})
		}
		if vote.Voice == -1 && newVote.Voice == 1 {
			newVote.Voice = 2
		} else {
			if vote.Voice == 1 && newVote.Voice == -1 {
				newVote.Voice = -2
			} else {
				newVote.Voice = 0
			}
		}
		err = h.ThreadService.UpdateVoteCount(newVote)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't vote"})
		}
	}

	thread, err = h.ThreadService.SelectThreadById(newVote.ThreadId)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
	}
	return ctx.JSON(http.StatusOK, thread)
}

func (h *Post) GetThread(ctx echo.Context) error {
	slugOrIdStr := ctx.Param("slug_or_id")

	var thread forum.Thread
	id, err := strconv.Atoi(slugOrIdStr)
	if err != nil {
		thread, err = h.ThreadService.SelectThreadBySlug(slugOrIdStr)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
		}
	} else {
		thread, err = h.ThreadService.SelectThreadById(id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
		}
	}

	return ctx.JSON(http.StatusOK, thread)
}

func (h *Post) GetPosts(ctx echo.Context) error {
	slugOrIdStr := ctx.Param("slug_or_id")

	limit := ctx.QueryParam("limit")
	since := ctx.QueryParam("since")
	sort := ctx.QueryParam("sort")
	desc := ctx.QueryParam("desc")

	if limit == "" {
		limit = "100"
	}

	if sort == "" {
		sort = "flat"
	}

	if desc == "true" {
		desc = "desc"
	} else {
		desc = ""
	}

	var thread forum.Thread
	id, err := strconv.Atoi(slugOrIdStr)
	if err != nil {
		thread, err = h.ThreadService.SelectThreadBySlug(slugOrIdStr)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
		}
		id = thread.Id
	}

	posts, err := h.ThreadService.SelectPosts(id, limit, since, sort, desc)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't read posts"})
	}
	if len(posts) == 0 {

		thread, err = h.ThreadService.SelectThreadById(id)
		if err != nil {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
		}

		postss := []Post{}
		return ctx.JSON(http.StatusOK, postss)
	}
	return ctx.JSON(http.StatusOK, posts)
}
