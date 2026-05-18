package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

type ProfileController struct {
	profileservice *service.ProfileService
}

func NewProfileController(profileservice *service.ProfileService) *ProfileController {
	return &ProfileController{profileservice: profileservice}
}

func (p *ProfileController) GetProfile(ctx *gin.Context) {
	email := ctx.Query("email")
	if email == "" {
		log.Printf("[ProfileController.GetProfile] Missing email parameter\n")
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid request payload",
			Success: false,
			Error:   "Email parameter is required",
		})
		return
	}

	user, profile, err := p.profileservice.GetProfileByEmail(ctx.Request.Context(), email)
	if err != nil {
		log.Printf("[ProfileController.GetProfile] Service error: %v\n", err)
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to fetch profile",
			Success: false,
			Error:   "data Tidak Ditemukan",
		})
		return
	}

	res := dto.Profile{
		ID:        user.Id,
		Email:     user.Email,
		FullName:  profile.FullName,
		Phone:     profile.Phone,
		Photo:     profile.Photo,
		CreatedAt: profile.CreatedAt,
		UpdatedAt: profile.UpdatedAt,
	}
	ctx.JSON(http.StatusOK, dto.Response{
		Data:    res,
		Message: "Profile successfully retrieved",
		Success: true,
	})
}
