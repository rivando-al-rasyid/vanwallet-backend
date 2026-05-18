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
	return &AuthController{authservice: authservice}
}

func (a *AuthController) Register(ctx *gin.Context) {
	var body dto.NewUser

	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Printf("[AuthController.Register] JSON binding error: %v\n", err)
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid request payload",
			Success: false,
			Error:   "Please ensure your input matches the required format",
		})
		return
	}

	res, err := a.authservice.Register(ctx.Request.Context(), body)
	if err != nil {
		log.Printf("[AuthController.Register] Service error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Registration failed",
			Success: false,
			Error:   "Email Already Exist",
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.Response{
		Data:    res,
		Message: "User successfully registered",
		Success: true,
	})
}

func (a *AuthController) Login(ctx *gin.Context) {
	var body dto.NewUser

	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Printf("[AuthController.Login] JSON binding error: %v\n", err)
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid request payload",
			Success: false,
			Error:   "Please ensure your input matches the required format",
		})
		return
	}

	res, err := a.authservice.Login(ctx.Request.Context(), body)
	if err != nil {
		log.Printf("[AuthController.Login] Service error: %v\n", err)
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Login failed",
			Success: false,
			Error:   "wrong Email Or Password",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data:    res,
		Message: "Login successful",
		Success: true,
	})
}
