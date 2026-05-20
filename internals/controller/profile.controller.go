package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

type ProfileController struct {
	profileservice *service.ProfileService
}

func NewProfileController(profileservice *service.ProfileService) *ProfileController {
	return &ProfileController{profileservice: profileservice}
}

// GetProfile godoc
//
//	@Summary		Get user profile
//	@Description	Retrieve the profile information of the authenticated user
//	@Tags			Profile
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	dto.Response{data=dto.ProfileResponse}
//	@Failure		401	{object}	dto.Response
//	@Failure		404	{object}	dto.Response
//	@Failure		500	{object}	dto.Response
//	@Router			/profile/ [get]
func (p *ProfileController) GetProfile(ctx *gin.Context) {
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

	profile, err := p.profileservice.GetProfile(ctx.Request.Context(), email)
	if err != nil {
		if err.Error() == "user profile not found" {
			ctx.JSON(http.StatusNotFound, dto.Response{
				Message: "Failed to fetch profile",
				Success: false,
				Error:   "Data tidak ditemukan",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to fetch profile",
			Success: false,
			Error:   "Internal server error",
		})
		return
	}

	res := dto.ProfileResponse{
		FullName: profile.FullName,
		Phone:    profile.Phone,
		Photo:    profile.Photo,
	}
	ctx.JSON(http.StatusOK, dto.Response{
		Data:    res,
		Message: "Profile successfully retrieved",
		Success: true,
	})
}

// EditProfile godoc
//
//	@Summary		Update user profile
//	@Description	Update one or more profile fields (full_name, phone, photo) of the authenticated user
//	@Tags			Profile
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			body	body		dto.UpdateProfileRequest	true	"Profile update payload (all fields optional)"
//	@Success		200		{object}	dto.Response{data=dto.ProfileResponse}
//	@Failure		400		{object}	dto.Response
//	@Failure		401		{object}	dto.Response
//	@Failure		500		{object}	dto.Response
//	@Router			/profile/ [post]
func (p *ProfileController) EditProfile(ctx *gin.Context) {
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

	var body dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid request body",
			Success: false,
			Error:   "data tidak ada",
		})
		return
	}

	updates := map[string]any{}
	if body.FullName != nil {
		updates["full_name"] = body.FullName
	}
	if body.Phone != nil {
		updates["phone"] = body.Phone
	}
	if body.Photo != nil {
		updates["photo"] = body.Photo
	}

	if len(updates) == 0 {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "No fields to update",
			Success: false,
			Error:   "Request body is empty",
		})
		return
	}

	profile, err := p.profileservice.EditProfile(ctx.Request.Context(), email, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to update profile",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data: dto.ProfileResponse{
			FullName: profile.FullName,
			Phone:    profile.Phone,
			Photo:    profile.Photo,
		},
		Message: "Profile successfully updated",
		Success: true,
	})
}

// EditPin godoc
//
//	@Summary		Update user PIN
//	@Description	Update the PIN of the authenticated user
//	@Tags			Profile
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			body	body		dto.SetPinRequest	true	"PIN update payload"
//	@Success		200		{object}	dto.Response
//	@Failure		400		{object}	dto.Response
//	@Failure		401		{object}	dto.Response
//	@Failure		500		{object}	dto.Response
//	@Router			/profile/pin [post]
func (p *ProfileController) EditPin(ctx *gin.Context) {
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

	var body dto.SetPinRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid request body",
			Success: false,
			Error:   "data tidak ada",
		})
		return
	}

	updates := map[string]any{}
	if body.PinHash != nil {
		updates["pin_hash"] = body.PinHash
	}
	if body.FailedAttempts != nil {
		updates["failed_attempts"] = body.FailedAttempts
	}
	if body.LockedUntil != nil {
		updates["locked_until"] = body.LockedUntil
	}
	if len(updates) == 0 {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "No fields to update",
			Success: false,
			Error:   "Request body is empty",
		})
		return
	}

	_, err := p.profileservice.EditPin(ctx.Request.Context(), email, updates)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to update Pin",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: "Profile successfully updated",
		Success: true,
	})
}
