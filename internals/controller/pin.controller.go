package controller

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
)

// PinSvc is the interface the controller depends on.
type PinSvc interface {
	HasPin(ctx context.Context, email string) (bool, error)
	SetPin(ctx context.Context, email, rawPin string) error
	VerifyPin(ctx context.Context, email, rawPin string) error
}

// PinController handles PIN-related HTTP requests.
type PinController struct {
	pinSvc PinSvc
}

func NewPinController(pinSvc PinSvc) *PinController {
	return &PinController{pinSvc: pinSvc}
}

// SetPin godoc
// @Summary      Set or update user PIN
// @Description  Hashes and stores a 6-digit PIN for the authenticated user.
// @Tags         pin
// @Accept       json
// @Produce      json
// @Param        body  body  dto.SetPinRequest  true  "PIN payload"
// @Success      200   {object}  dto.Response
// @Failure      400   {object}  dto.Response
// @Failure      500   {object}  dto.Response
// @Security     BearerAuth
// @Router       /pin [post]
func (p *PinController) SetPin(ctx *gin.Context) {
	claims, ok := p.claimsFromCtx(ctx)
	if !ok {
		return
	}

	var req dto.SetPinRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Bad Request",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if req.PinHash == nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Bad Request",
			Success: false,
			Error:   "pin_hash is required",
		})
		return
	}

	if err := p.pinSvc.SetPin(ctx.Request.Context(), claims.Email, *req.PinHash); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to set PIN",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: "PIN set successfully",
		Success: true,
	})
}

// VerifyPin godoc
// @Summary      Verify user PIN
// @Description  Checks a supplied 6-digit PIN against the stored argon2id hash.
// @Tags         pin
// @Accept       json
// @Produce      json
// @Param        body  body  dto.VerifyPinRequest  true  "PIN to verify"
// @Success      200   {object}  dto.Response
// @Failure      400   {object}  dto.Response
// @Failure      401   {object}  dto.Response
// @Security     BearerAuth
// @Router       /pin/verify [post]
func (p *PinController) VerifyPin(ctx *gin.Context) {
	claims, ok := p.claimsFromCtx(ctx)
	if !ok {
		return
	}

	var req dto.VerifyPinRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Bad Request",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := p.pinSvc.VerifyPin(ctx.Request.Context(), claims.Email, req.Pin); err != nil {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Invalid PIN",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: "PIN verified",
		Success: true,
	})
}

// CheckPin godoc
// @Summary      Check whether the user has a PIN set
// @Description  Returns a boolean indicating if the user has already configured a PIN.
// @Tags         pin
// @Produce      json
// @Success      200  {object}  dto.Response
// @Failure      500  {object}  dto.Response
// @Security     BearerAuth
// @Router       /pin/status [get]
func (p *PinController) CheckPin(ctx *gin.Context) {
	claims, ok := p.claimsFromCtx(ctx)
	if !ok {
		return
	}

	hasPin, err := p.pinSvc.HasPin(ctx.Request.Context(), claims.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Error",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: "OK",
		Success: true,
		Data:    gin.H{"has_pin": hasPin},
	})
}

// claimsFromCtx extracts pkg.Claims set by VerifyTokenWithDB middleware.
// Returns false and writes an error response if claims are missing.
func (p *PinController) claimsFromCtx(ctx *gin.Context) (pkg.Claims, bool) {
	raw, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{
			Message: "Unauthorized",
			Success: false,
			Error:   "Missing JWT claims",
		})
		return pkg.Claims{}, false
	}
	claims, ok := raw.(pkg.Claims)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Error",
			Success: false,
			Error:   "Internal Server Error",
		})
		return pkg.Claims{}, false
	}
	return claims, true
}
