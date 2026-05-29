package controller

import (
	"log"
	"net/http"
	"strings"

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
// @Summary      Register a new systemic user entity
// @Description  Creates an absolute identity space record within the backend infrastructure dataset[cite: 3].
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterRequest  true  "Sign up identification credentials payload"
// @Success      201   {object}  dto.Response{data=dto.UserResponse}
// @Failure      400   {object}  dto.Response{error}
// @Failure      409   {object}  dto.Response{error}
// @Failure      500   {object}  dto.Response{error}
// @Router       /auth/register [post]
func (a *AuthController) Register(ctx *gin.Context) {
	var body dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Printf("[AuthController.Register] bind error: %v\n", err)
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request payload", "Please ensure your input matches the required format"))
		return
	}

	res, err := a.authservice.Register(ctx.Request.Context(), body)
	if err != nil {
		log.Printf("[AuthController.Register] service error: %v\n", err)
		status := http.StatusInternalServerError
		errDetail := "Internal server error"
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			status = http.StatusConflict
			errDetail = "Email already exists"
		}
		ctx.JSON(status, dto.NewError("Registration failed", errDetail))
		return
	}

	ctx.JSON(http.StatusCreated, dto.NewSuccess("User successfully registered", res))
}

// Login godoc
// @Summary      Authenticate profile session token
// @Description  Verifies input entries against structural data records to issue signed verification JWTs[cite: 3].
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Identity access block data definitions"
// @Success      200   {object}  dto.Response{data=dto.UserResponse}
// @Failure      400   {object}  dto.Response{error}
// @Failure      401   {object}  dto.Response{error}
// @Router       /auth/login [post]
func (a *AuthController) Login(ctx *gin.Context) {
	var body dto.LoginRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		log.Printf("[AuthController.Login] bind error: %v\n", err)
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request payload", "Please ensure your input matches the required format"))
		return
	}

	token, err := a.authservice.Login(ctx.Request.Context(), body)
	if err != nil {
		log.Printf("[AuthController.Login] service error: %v\n", err)
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Login failed", "Incorrect email or password"))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccess("Login successful", token))
}

// Logout godoc
// @Summary      Terminates system session tokens
// @Description  Invalidates the currently active authorization signature directly inside redis/database memory pools[cite: 3].
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security		BearerAuth
// @Success      200            {object}  dto.Response{data=object}
// @Failure      401            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /auth/logout [post]
func (a *AuthController) Logout(ctx *gin.Context) {
	claimsRaw, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	email := claimsRaw.(pkg.Claims).Email

	rawToken, exists := ctx.Get("raw_token")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Logout failed", "raw token not found in context"))
		return
	}

	if err := a.authservice.Logout(ctx.Request.Context(), rawToken.(string), email); err != nil {
		log.Printf("[AuthController.Logout] service error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Logout failed", "Internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessNoData("Logged out successfully"))
}

// GetPIN godoc
// @Summary      Get internal system PIN hash sequence
// @Description  Fetches the secure transaction authentication pin configuration status block[cite: 3].
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security		BearerAuth
// @Success      200            {object}  dto.Response{data=string}
// @Failure      401            {object}  dto.Response{error}
// @Failure      404            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /auth/pin [get]
func (a *AuthController) GetPIN(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}

	email := claims.(pkg.Claims).Email
	pin, err := a.authservice.GetUserPin(ctx.Request.Context(), email)
	if err != nil {
		if err.Error() == "user pin not found" {
			ctx.JSON(http.StatusNotFound, dto.NewError("Failed to fetch PIN", "PIN not set for this user"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to fetch PIN", "Internal server error"))
		return
	}
	if pin.PinHash == nil || len(*pin.PinHash) == 0 {
		ctx.JSON(http.StatusNotFound, dto.NewError("Failed to fetch PIN", "PIN not set for this user"))
		return
	}
	ctx.JSON(http.StatusOK, dto.NewSuccess("PIN successfully retrieved", pin.PinHash))
}

// VerifyPIN godoc
// @Summary      Verify validity of PIN entry blocks
// @Description  Evaluates transactional authorization structures before processing balance updates[cite: 3].
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security		BearerAuth
// @Param        body           body      dto.VerifyPinRequest  true  "Pin block parameters validation request"
// @Success      200            {object}  dto.Response{data=object}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Router       /auth/pin/verify [post]
func (a *AuthController) VerifyPIN(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	email := claims.(pkg.Claims).Email

	var body dto.VerifyPinRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err.Error()))
		return
	}

	if err := a.authservice.VerifyPin(ctx.Request.Context(), email, body.Pin); err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Invalid PIN", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessNoData("PIN verified successfully"))
}
