package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

type AuthController struct {
	authservice *service.AuthService
}

func NewAuthController(authservice *service.AuthService) *AuthController {
	return &AuthController{authservice: authservice}
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Create a new account with email and password. Automatically creates a linked profile, PIN slot, and default wallet in a single transaction.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RegisterRequest				true	"Register payload"
//	@Success		201		{object}	dto.Response{data=dto.UserResponse}	"User created"
//	@Failure		400		{object}	dto.Response						"Invalid request body"
//	@Failure		500		{object}	dto.Response						"Email already exists or internal error"
//	@Router			/auth/register [post]
func (a *AuthController) Register(ctx *gin.Context) {
	var body dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Printf("[AuthController.Register] bind error: %v\n", err)
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid request payload",
			Success: false,
			Error:   "Please ensure your input matches the required format",
		})
		return
	}

	res, err := a.authservice.Register(ctx.Request.Context(), body)
	if err != nil {
		log.Printf("[AuthController.Register] service error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Registration failed",
			Success: false,
			Error:   "Email already exists",
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.Response{
		Data:    res,
		Message: "User successfully registered",
		Success: true,
	})
}

// Login godoc
//
//	@Summary		Login
//	@Description	Authenticate with email and password. Returns a signed JWT (24 h). The token is persisted in the `tokens` table and can be revoked via logout.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.LoginRequest			true	"Login payload"
//	@Success		200		{object}	dto.Response{data=string}	"Bearer JWT token"
//	@Failure		400		{object}	dto.Response				"Invalid request body"
//	@Failure		401		{object}	dto.Response				"Wrong email or password"
//	@Router			/auth/login [post]
func (a *AuthController) Login(ctx *gin.Context) {
	var body dto.LoginRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Printf("[AuthController.Login] bind error: %v\n", err)
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid request payload",
			Success: false,
			Error:   "Please ensure your input matches the required format",
		})
		return
	}

	token, err := a.authservice.Login(ctx.Request.Context(), body)
	if err != nil {
		log.Printf("[AuthController.Login] service error: %v\n", err)
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Login failed",
			Success: false,
			Error:   "wrong Email Or Password",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data:    token,
		Message: "Login successful",
		Success: true,
	})
}

// Logout godoc
//
//	@Summary		Logout
//	@Description	Revokes the current Bearer token by setting `is_revoked = true` in the tokens table. The token becomes invalid immediately on all subsequent requests — no need to wait for expiry.
//	@Tags			Auth
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	dto.Response				"Logged out"
//	@Failure		401	{object}	dto.Response				"Unauthorized or missing token"
//	@Failure		500	{object}	dto.Response				"Internal server error"
//	@Router			/auth/logout [post]
func (a *AuthController) Logout(ctx *gin.Context) {
	_, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Unauthorized",
			Success: false,
			Error:   "Missing claims",
		})
		return
	}

	rawToken, exists := ctx.Get("raw_token")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Logout failed",
			Success: false,
			Error:   "raw token not found in context",
		})
		return
	}

	if err := a.authservice.Logout(ctx.Request.Context(), rawToken.(string)); err != nil {
		log.Printf("[AuthController.Logout] service error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Logout failed",
			Success: false,
			Error:   "Internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: "Logged out successfully",
		Success: true,
	})
}

// GetPIN godoc
//
//	@Summary		Get PIN hash
//	@Description	Returns the bcrypt-hashed PIN of the authenticated user. Use this to verify whether a PIN has already been set before prompting the user to create one.
//	@Tags			Auth
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	dto.Response{data=string}	"Bcrypt PIN hash"
//	@Failure		401	{object}	dto.Response				"Unauthorized"
//	@Failure		404	{object}	dto.Response				"PIN not set for this user"
//	@Failure		500	{object}	dto.Response
//	@Router			/auth/pin [get]
func (a *AuthController) GetPIN(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Unauthorized",
			Success: false,
			Error:   "Missing claims",
		})
		return
	}

	email := claims.(pkg.Claims).Email
	pin, err := a.authservice.GetUserPin(ctx.Request.Context(), email)
	if err != nil {
		if err.Error() == "user pin not found" {
			ctx.JSON(http.StatusNotFound, dto.Response{
				Message: "Failed to fetch PIN",
				Success: false,
				Error:   "Data tidak ditemukan",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to fetch PIN",
			Success: false,
			Error:   "Internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data:    pin.PinHash,
		Message: "Pin successfully retrieved",
		Success: true,
	})
}
