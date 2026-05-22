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
//	@Accept			mpfd
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			full_name	formData	string	false	"Full name"
//	@Param			phone		formData	string	false	"Phone number (E.164)"
//	@Param			photo		formData	file	false	"Profile photo (jpg/png/webp, max 2MB)"
//	@Success		200			{object}	dto.Response{data=dto.ProfileResponse}
//	@Failure		400			{object}	dto.Response
//	@Failure		401			{object}	dto.Response
//	@Failure		500			{object}	dto.Response
//	@Router			/profile/ [post]
func (p *ProfileController) validateAndSavePhoto(ctx *gin.Context, photo *multipart.FileHeader, email string) (*string, error) {
	if e := p.profileservice.ValidateUpload(2*config.MB, photo); e != nil {
		log.Println(e.Error())
		if errors.Is(e, config.ErrFileTooLarge) {
			ctx.JSON(http.StatusUnprocessableEntity, dto.Response{
				Message: "File too large",
				Success: false,
				Error:   "Photo must be under 2MB",
			})
			return nil, e
		}
		if errors.Is(e, config.ErrExtNotAllowed) {
			ctx.JSON(http.StatusUnprocessableEntity, dto.Response{
				Message: "Invalid file type",
				Success: false,
				Error:   "Only .jpg, .jpeg, .png, .webp are allowed",
			})
			return nil, e
		}
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Error",
			Success: false,
			Error:   "Internal Server Error",
		})
		return nil, e
	}

	ext := path.Ext(photo.Filename)
	filename := fmt.Sprintf("%s_photo_%d%s", strings.ToLower(strings.ReplaceAll(email, "@", "_")), time.Now().UnixNano(), ext)
	dst := filepath.Join("public", "img", filename)
	if err := ctx.SaveUploadedFile(photo, dst); err != nil {
		log.Println("error: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Error",
			Success: false,
			Error:   "Internal Server Error",
		})
		return nil, err
	}

	photoURL := fmt.Sprintf("/img/%s", filename)
	return &photoURL, nil
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
	if err := ctx.ShouldBindWith(&body, binding.FormMultipart); err != nil {
		log.Println("error: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Error",
			Success: false,
			Error:   "Internal Server Error",
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
		photoURL, err := p.validateAndSavePhoto(ctx, body.Photo, email)
		if err != nil {
			return
		}
		updates["photo"] = photoURL
	}

	_, err := p.profileservice.EditProfile(ctx, email, updates)
	if err != nil {
		log.Println("error: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Error",
			Success: false,
			Error:   "Internal Server Error",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: "OK",
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

	_, err := p.profileservice.EditPin(ctx.Request.Context(), email, *body.PinHash)
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

func (p *ProfileController) EditPassword(ctx *gin.Context) {
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

	var body dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid request body",
			Success: false,
			Error:   "data tidak ada",
		})
		return
	}

	_, err := p.profileservice.EditPassword(ctx.Request.Context(), email, body.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to update Password",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: "Password successfully updated",
		Success: true,
	})
}
