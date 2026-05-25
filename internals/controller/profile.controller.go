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

// GetProfile godoc
//
//	@Summary		Get profile
//	@Description	Retrieve the authenticated user's profile fields: full_name, phone, and photo URL.
//	@Tags			Profile
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.Response{data=dto.ProfileResponse}	"Profile data"
//	@Failure		401	{object}	dto.Response							"Unauthorized or missing token"
//	@Failure		404	{object}	dto.Response							"Profile not found"
//	@Failure		500	{object}	dto.Response							"Internal server error"
//	@Router			/profile/ [get]
func (p *ProfileController) GetProfile(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}

	email := claims.(pkg.Claims).Email
	profile, err := p.profileservice.GetProfile(ctx.Request.Context(), email)
	if err != nil {
		if err.Error() == "user profile not found" {
			ctx.JSON(http.StatusNotFound, dto.Response{Message: "Failed to fetch profile", Success: false, Error: "Data tidak ditemukan"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Failed to fetch profile", Success: false, Error: "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data:    dto.ProfileResponse{FullName: profile.FullName, Phone: profile.Phone, Photo: profile.Photo},
		Message: "Profile successfully retrieved",
		Success: true,
	})
}

func (p *ProfileController) validateAndSavePhoto(ctx *gin.Context, photo *multipart.FileHeader, email string) (*string, error) {
	if e := p.profileservice.ValidateUpload(2*config.MB, photo); e != nil {
		log.Println(e.Error())
		if errors.Is(e, config.ErrFileTooLarge) {
			ctx.JSON(http.StatusUnprocessableEntity, dto.Response{Message: "File too large", Success: false, Error: "Photo must be under 2MB"})
			return nil, e
		}
		if errors.Is(e, config.ErrExtNotAllowed) {
			ctx.JSON(http.StatusUnprocessableEntity, dto.Response{Message: "Invalid file type", Success: false, Error: "Only .jpg, .jpeg, .png, .webp are allowed"})
			return nil, e
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Error", Success: false, Error: "Internal Server Error"})
		return nil, e
	}

	ext := path.Ext(photo.Filename)
	filename := fmt.Sprintf("%s_photo_%d%s", strings.ToLower(strings.ReplaceAll(email, "@", "_")), time.Now().UnixNano(), ext)
	dst := filepath.Join("public", "img", filename)
	if err := ctx.SaveUploadedFile(photo, dst); err != nil {
		log.Println("error: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Error", Success: false, Error: "Internal Server Error"})
		return nil, err
	}

	photoURL := fmt.Sprintf("/img/%s", filename)
	return &photoURL, nil
}

// EditProfile godoc
//
//	@Summary		Update profile
//	@Description	Update one or more profile fields: full_name, phone, and/or photo. All fields are optional — omit any field to leave it unchanged. Accepts multipart/form-data.
//	@Tags			Profile
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			full_name	formData	string	false	"Display name"
//	@Param			phone		formData	string	false	"Phone number in E.164 format (e.g. +628123456789)"
//	@Param			photo		formData	file	false	"Profile photo — JPEG, PNG, or WebP; max 2 MB"
//	@Success		200			{object}	dto.Response							"Profile updated successfully"
//	@Failure		400			{object}	dto.Response							"Invalid or malformed form data"
//	@Failure		401			{object}	dto.Response							"Unauthorized or missing token"
//	@Failure		422			{object}	dto.Response							"Photo exceeds 2 MB or has unsupported file type"
//	@Failure		500			{object}	dto.Response							"Internal server error"
//	@Router			/profile/ [post]
func (p *ProfileController) EditProfile(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}
	email := claims.(pkg.Claims).Email

	var body dto.UpdateProfileRequest
	if err := ctx.ShouldBindWith(&body, binding.FormMultipart); err != nil {
		log.Println("error: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Error", Success: false, Error: "Internal Server Error"})
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
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Error", Success: false, Error: "Internal Server Error"})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Message: "Profile successfully updated", Success: true})
}

// EditPin godoc
//
//	@Summary		Set / update PIN
//	@Description	Store a new 6-digit PIN for the authenticated user. Send the PIN as a bcrypt hash (`pin_hash`). Any previously set PIN is replaced.
//	@Tags			Profile
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.SetPinRequest	true	"PIN payload (pin_hash = bcrypt hash of the 6-digit PIN)"
//	@Success		200		{object}	dto.Response		"PIN updated successfully"
//	@Failure		400		{object}	dto.Response		"Invalid or missing request body"
//	@Failure		401		{object}	dto.Response		"Unauthorized or missing token"
//	@Failure		500		{object}	dto.Response		"Internal server error"
//	@Router			/profile/change/pin [post]
func (p *ProfileController) EditPin(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}
	email := claims.(pkg.Claims).Email

	var body dto.SetPinRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid request body", Success: false, Error: "data tidak ada"})
		return
	}

	_, err := p.profileservice.EditPin(ctx.Request.Context(), email, *body.PinHash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Failed to update Pin", Success: false, Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Message: "PIN successfully updated", Success: true})
}

// EditPassword godoc
//
//	@Summary		Change password
//	@Description	Verifies the current password then replaces it with the new one. Both `old_password` and `password` are required.
//	@Tags			Profile
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.ChangePasswordRequest	true	"Password change payload"
//	@Success		200		{object}	dto.Response				"Password updated successfully"
//	@Failure		400		{object}	dto.Response				"Invalid or missing request body"
//	@Failure		401		{object}	dto.Response				"Unauthorized or incorrect current password"
//	@Failure		500		{object}	dto.Response				"Internal server error"
//	@Router			/profile/change/password [post]
func (p *ProfileController) EditPassword(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}
	email := claims.(pkg.Claims).Email

	var body dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{Message: "Invalid request body", Success: false, Error: "data tidak ada"})
		return
	}

	_, err := p.profileservice.EditPassword(ctx.Request.Context(), email, body.OldPassword, body.Password)
	if err != nil {
		if err.Error() == "old password is incorrect" {
			ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Failed to update password", Success: false, Error: "Old password is incorrect"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Failed to update Password", Success: false, Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{Message: "Password successfully updated", Success: true})
}

// GetUserInfo godoc
//
//	@Summary		User info for app header
//	@Description	Returns email, full_name, phone, photo, and current_balance (sum across all wallets). Intended for populating the persistent app header/navbar without a full dashboard call.
//	@Tags			Profile
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.Response{data=dto.UserInfoResponse}	"User info"
//	@Failure		401	{object}	dto.Response							"Unauthorized or missing token"
//	@Failure		500	{object}	dto.Response							"Internal server error"
//	@Router			/profile/me [get]
func (p *ProfileController) GetUserInfo(ctx *gin.Context) {
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.Response{Message: "Unauthorized", Success: false, Error: "Missing claims"})
		return
	}
	email := claims.(pkg.Claims).Email

	profile, balance, err := p.profileservice.GetUserInfo(ctx.Request.Context(), email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{Message: "Failed to fetch user info", Success: false, Error: "Internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data: dto.UserInfoResponse{
			Email:          email,
			FullName:       profile.FullName,
			Phone:          profile.Phone,
			Photo:          profile.Photo,
			CurrentBalance: balance,
		},
		Message: "User info successfully retrieved",
		Success: true,
	})
}
