package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

type AuthController struct {
	authservice *service.AuthService
}

func NewAuthController(authservice *service.AuthService) *AuthController {
	return &AuthController{
		authservice: authservice,
	}

}

func (a *AuthController) Register(ctx *gin.Context) {
	var body dto.NewUser
	if err := ctx.ShouldBindBodyWithJSON(body); err != nil {
		log.Println("Error: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Error", Success: false, Error: "internal server error"})
		return
	}
	res, err := a.authservice.Register(ctx.Request.Context(), body)
	if err != nil {
		log.Println("Error: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Error", Success: false, Error: "internal server error"})
		return
	}
	ctx.JSON(http.StatusCreated, dto.Response{
		Data:    res,
		Message: "User Registered",
		Success: true,
	})

}
