package handlers

import (
	"tech-db/internal/forum"
	"github.com/jackc/pgx"
	"github.com/labstack/echo"
	"net/http"
)

type User struct {
	UserService *forum.UserService
}

func (h *User) CreateUser(ctx echo.Context) (Err error) {
	nickName := ctx.Param("nickname")
	if nickName == "" {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}
	newUser := forum.User{}
	if err := ctx.Bind(&newUser); err != nil {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}
	newUser.NickName = nickName

	userSlice, err := h.UserService.SelectUserByNickNameOrEmail(newUser.NickName, newUser.Email)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	if len(userSlice) > 0 {
		return ctx.JSON(http.StatusConflict, userSlice)
	}

	if err = h.UserService.InsertUser(newUser); err != nil {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	return ctx.JSON(http.StatusCreated, newUser)
}

func (h *User) GetProfile(ctx echo.Context) (Err error) {
	nickName := ctx.Param("nickname")
	if nickName == "" {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}
	user, err := h.UserService.SelectUserByNickName(nickName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Can't find user"})
		}
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}

	return ctx.JSON(http.StatusOK, user)
}

func (h *User) EditProfile(ctx echo.Context) (Err error) {
	nickName := ctx.Param("nickname")
	if nickName == "" {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}
	editUser := forum.User{}
	if err := ctx.Bind(&editUser); err != nil {
		ctx.Logger().Warn(err)
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}
	editUser.NickName = nickName

	userSlice, err := h.UserService.SelectUserByNickNameOrEmail(editUser.NickName, editUser.Email)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}
	if len(userSlice) > 1 {
		return ctx.JSON(http.StatusConflict, forum.ErrorMessage{Message: "This email is already registered by user"})
	}
	if len(userSlice) == 0 {
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "User not found"})
	}
	if userSlice[0].NickName != editUser.NickName {
		return ctx.JSON(http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
	}
	editUser.Id = userSlice[0].Id
	if editUser.About == "" {
		editUser.About = userSlice[0].About
	}
	if editUser.Email == "" {
		editUser.Email = userSlice[0].Email
	}
	if editUser.FullName == "" {
		editUser.FullName = userSlice[0].FullName
	}

	if err = h.UserService.UpdateUser(editUser); err != nil {
		return ctx.JSON(http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	return ctx.JSON(http.StatusOK, editUser)
}
