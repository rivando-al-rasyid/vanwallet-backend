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
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to fetch profile",
			Success: false,
			Error:   "data Tidak Ditemukan",
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
