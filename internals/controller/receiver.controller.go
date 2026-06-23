package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

type ReceiverController struct {
	receiverService *service.ReceiverService
}

func NewReceiverController(receiverService *service.ReceiverService) *ReceiverController {
	return &ReceiverController{receiverService: receiverService}
}

// FindReceivers godoc
// @Summary      List or search transfer receivers
// @Description  Returns available receiver profiles. Supports optional search by full name, email, phone, wallet label, or wallet id.
// @Tags         Receivers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        q      query     string  false  "Optional search keyword"
// @Param        page   query     int     false  "Page number" default(1)
// @Param        limit  query     int     false  "Items per page" default(10)
// @Success      200    {object}  dto.Response{data=dto.ReceiverListResponse}
// @Failure      401    {object}  dto.Response{error}
// @Failure      500    {object}  dto.Response{error}
// @Router       /transaction/receiver [get]
func (t *ReceiverController) FindReceivers(ctx *gin.Context) {
	email, ok := pkg.CurrentUserEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", errors.New("missing user context")))
		return
	}

	query := ctx.Query("q")
	if query == "" {
		query = ctx.Query("query")
	}

	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	results, total, err := t.receiverService.SearchReceivers(ctx.Request.Context(), email, query, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Search failed", err))
		return
	}

	resp := make([]dto.ReceiverResult, 0, len(results))
	for _, r := range results {
		item := dto.ReceiverResult{
			UserID:      r.UserID.String(),
			Email:       r.Email,
			WalletID:    r.WalletID.String(),
			WalletLabel: r.WalletLabel,
		}
		if r.FullName != nil {
			item.FullName = *r.FullName
		}
		if r.Phone != nil {
			item.Phone = *r.Phone
		}
		if r.Photo != nil {
			item.Photo = *r.Photo
		}
		resp = append(resp, item)
	}

	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	ctx.JSON(http.StatusOK, dto.NewSuccess("Receivers fetched successfully", dto.ReceiverListResponse{
		Data:       resp,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}))
}
