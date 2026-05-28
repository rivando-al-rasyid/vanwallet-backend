package controller

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/config"
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
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}

	email := claims.(pkg.Claims).Email
	profile, err := p.profileservice.GetProfile(ctx.Request.Context(), email)
	if err != nil {
		if err.Error() == "user profile not found" {
			ctx.JSON(http.StatusNotFound, dto.NewError("Failed to fetch profile", "Profile not found"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to fetch profile", "Internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccess("Profile successfully retrieved", dto.ProfileResponse{
		FullName: profile.FullName,
		Phone:    profile.Phone,
		Photo:    profile.Photo,
	}))
}

func (p *ProfileController) validateAndSavePhoto(ctx *gin.Context, photo *multipart.FileHeader, email string) (*string, error) {
	if e := p.profileservice.ValidateUpload(2*config.MB, photo); e != nil {
		log.Println(e.Error())
		if errors.Is(e, config.ErrFileTooLarge) {
			ctx.JSON(http.StatusUnprocessableEntity, dto.NewError("File too large", "Photo must be under 2MB"))
			return nil, e
		}
		if errors.Is(e, config.ErrExtNotAllowed) {
			ctx.JSON(http.StatusUnprocessableEntity, dto.NewError("Invalid file type", "Only .jpg, .jpeg, .png, .webp are allowed"))
			return nil, e
		}
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error", "Internal server error"))
		return nil, e
	}

	ext := path.Ext(photo.Filename)
	filename := fmt.Sprintf("%s_photo_%d%s", strings.ToLower(strings.ReplaceAll(email, "@", "_")), time.Now().UnixNano(), ext)
	dst := filepath.Join("public", "img", filename)
	if err := ctx.SaveUploadedFile(photo, dst); err != nil {
		log.Println("error: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error", "Internal server error"))
		return nil, err
	}

	photoURL := fmt.Sprintf("/img/%s", filename)
	return &photoURL, nil
}

func (p *ProfileController) EditProfile(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	email := claims.(pkg.Claims).Email

	var body dto.UpdateProfileRequest
	if err := ctx.ShouldBindWith(&body, binding.FormMultipart); err != nil {
		log.Println("error: ", err.Error())
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err.Error()))
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
		photoURL, err := p.validateAndSavePhoto(ctx, body.Photo, email)
		if err != nil {
			return
		}
		updates["photo"] = photoURL
	}

	_, err := p.profileservice.EditProfile(ctx, email, updates)
	if err != nil {
		log.Println("error: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Error", "Internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessNoData("Profile successfully updated"))
}

func (p *ProfileController) EditPin(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	email := claims.(pkg.Claims).Email

	var body dto.SetPinRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err.Error()))
		return
	}

	_, err := p.profileservice.EditPin(ctx.Request.Context(), email, *body.PinHash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to update PIN", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessNoData("PIN successfully updated"))
}

func (p *ProfileController) EditPassword(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	email := claims.(pkg.Claims).Email

	var body dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err.Error()))
		return
	}

	_, err := p.profileservice.EditPassword(ctx.Request.Context(), email, body.OldPassword, body.Password)
	if err != nil {
		if err.Error() == "old password is incorrect" {
			ctx.JSON(http.StatusUnauthorized, dto.NewError("Failed to update password", "Old password is incorrect"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to update password", err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessNoData("Password successfully updated"))
}

func (p *ProfileController) GetUserInfo(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", "Missing claims"))
		return
	}
	claimsTyped := claims.(pkg.Claims)
	email := claimsTyped.Email

	profile, balance, err := p.profileservice.GetUserInfo(ctx.Request.Context(), email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to fetch user info", "Internal server error"))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccess("User info successfully retrieved", dto.UserInfoResponse{
		ID:             claimsTyped.ID.String(),
		Email:          email,
		FullName:       profile.FullName,
		Phone:          profile.Phone,
		Photo:          profile.Photo,
		CurrentBalance: balance,
	}))
}
