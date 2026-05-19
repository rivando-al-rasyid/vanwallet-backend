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

	_, profile, err := p.profileservice.GetProfileByEmail(ctx.Request.Context(), email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to fetch profile",
			Success: false,
			Error:   "data Tidak Ditemukan",
		})
		return
	}

	res := dto.ProfileResponse{
		ID:       profile.ID,
		UserID:   profile.UserID,
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
