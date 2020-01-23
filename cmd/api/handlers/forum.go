package handlers

import (
	"fmt"
	"github.com/jackc/pgx"
	"github.com/labstack/echo"
	"math"
	"net/http"
	"strconv"
	"tech-db/internal/forum"
)

type Forum struct {
	ForumService  *forum.ForumService
	UserService   *forum.UserService
	ThreadService *forum.ThreadService
}

func (h *Forum) CreateForum(ctx echo.Context) (Err error) {
	newForum := forum.Forum{}
	if err := ctx.Bind(&newForum); err != nil {
		ctx.Logger().Warn(err)
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	fullForum, err := h.ForumService.SelectForumBySlug(newForum.Slug)
	if err == nil {
		return ctx.JSON(http.StatusConflict, fullForum)
	}
	if err != pgx.ErrNoRows {
		ctx.Logger().Warn(err)
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	user, err := h.UserService.FindUserByNickName(newForum.User)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find user"})
		}
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}

	newForum.User = user.NickName

	if err = h.ForumService.InsertForum(newForum); err != nil {
		ctx.Logger().Warn(err)
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	return ctx.JSON(http.StatusCreated, newForum)
}

func (h *Forum) CreateThread(ctx echo.Context) (Err error) {
	slug := ctx.Param("slug")
	if slug == "" {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	newThread := forum.Thread{}
	if err := ctx.Bind(&newThread); err != nil {
		fmt.Println(err)
		ctx.Logger().Warn(err)
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	newThread.Forum = slug

	threadForum, err := h.ForumService.SelectForumBySlug(newThread.Forum)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find forum"})
		}
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}
	newThread.Forum = threadForum.Slug
	newThread.ForumId = threadForum.Id

	author, err := h.UserService.FindUserByNickName(newThread.Author)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find user"})
		}
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}

	newThread.Author = author.NickName

	if newThread.Slug != "" {
		thread, err := h.ThreadService.SelectThreadBySlug(newThread.Slug)
		if err == nil {
			return ctx.JSON(http.StatusConflict, thread)
		}
		if err != pgx.ErrNoRows {
			ctx.Logger().Warn(err)
			return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		}
	}

	threadId, err := h.ThreadService.InsertThread(newThread)
	if err != nil {
		ctx.Logger().Warn(err)
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	newThread.Id = threadId

	err = h.ForumService.UpdateThreadCount(newThread.ForumId)
	if err != nil {
		ctx.Logger().Warn(err)
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	err = h.ForumService.InsertForumUser(newThread.ForumId, author.Id)
	if err != nil {
		ctx.Logger().Warn(err)
	}
	/*	thread, err := h.ThreadService.SelectThreadById(id)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, "")
		}*/

	return ctx.JSON(http.StatusCreated, newThread)
}

func (h *Forum) GetForumDetails(ctx echo.Context) error {
	slug := ctx.Param("slug")
	if slug == "" {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	fullForum, err := h.ForumService.SelectForumBySlug(slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find forum"})
		}
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}

	return ctx.JSON(http.StatusOK, fullForum)
}

func (h *Forum) GetForumThreads(ctx echo.Context) error {
	slug := ctx.Param("slug")
	if slug == "" {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	limitStr := ctx.QueryParam("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.Logger().Warn(err)
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}
	if limit < 0 {
		ctx.Logger().Warn(err)
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}

	since := ctx.QueryParam("since")

	descStr := ctx.QueryParam("desc")
	desc, err := strconv.ParseBool(descStr)
	if err != nil {
		desc = false
	}

	threads, err := h.ThreadService.SelectThreadByForum(slug, limit, since, desc)
	if err != nil {
		ctx.Logger().Warn(err)
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}

	if len(threads) == 0 {
		_, err := h.ForumService.SelectForumBySlug(slug)
		if err != nil {
			if err == pgx.ErrNoRows {
				return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
			}
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
		}

		threads := []forum.Thread{}
		return ctx.JSON(http.StatusOK, threads)
	}

	return ctx.JSON(http.StatusOK, threads)
}

func (h *Forum) GetForumUsers(ctx echo.Context) error {
	slug := ctx.Param("slug")
	if slug == "" {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	usersForum, err := h.ForumService.SelectForumBySlug(slug)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find forum"})
	}
	limitStr := ctx.QueryParam("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = math.MaxInt32
	}
	if limit < 0 {
		ctx.Logger().Warn(err)
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}

	since := ctx.QueryParam("since")

	desc := ctx.QueryParam("desc")
	if desc == "" {
		desc = "false"
	}

	users, err := h.UserService.SelectUsersByForum(usersForum.Id, limit, since, desc)
	if err != nil {
		ctx.Logger().Warn(err)
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}
	if users == nil {
		nullUsers := []User{}
		return ctx.JSON(http.StatusOK, nullUsers)
	}

	return ctx.JSON(http.StatusOK, users)
}

func (h *Forum) Clean(ctx echo.Context) error {
	err := h.ForumService.Clean()
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, nil)
}

func (h *Forum) Status(ctx echo.Context) error {

	status, err := h.ForumService.SelectStatus()
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, status)
}
