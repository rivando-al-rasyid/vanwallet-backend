package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

type DashboardController struct {
	dashboardservice *service.DashboardService
}

func NewDashboardController(dashboardservice *service.DashboardService) *DashboardController {
	return &DashboardController{dashboardservice: dashboardservice}
}

// GetData godoc
//
//	@Summary		Get user dashboard information
//	@Description	Returns the authenticated user's current balance, total income, and total expense
//	@Tags			Dashboard
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	dto.Response{data=dto.DashboardResponse}
//	@Failure		401	{object}	dto.Response
//	@Failure		404	{object}	dto.Response
//	@Failure		500	{object}	dto.Response
//	@Router			/dashboard/ [get]
func (d *DashboardController) GetData(ctx *gin.Context) {
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

	dashboard, err := d.dashboardservice.GetData(ctx.Request.Context(), email)
	if err != nil {
		if err.Error() == "user profile not found" {
			ctx.JSON(http.StatusNotFound, dto.Response{
				Message: "Failed to fetch dashboard data",
				Success: false,
				Error:   "Data tidak ditemukan",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to fetch dashboard data",
			Success: false,
			Error:   "Internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data: dto.DashboardResponse{
			CurrentBalance: dashboard.CurrentBalance,
			TotalIncome:    dashboard.TotalIncome,
			TotalExpense:   dashboard.TotalExpense,
		},
		Message: "Dashboard data successfully retrieved",
		Success: true,
	})
}

// GetTransactionReport godoc
//
//	@Summary		Get transaction report (graph)
//	@Description	Returns income vs expense chart data grouped by day (7days) or by week (30days)
//	@Tags			Dashboard
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			range	query		string	false	"Time range: 7days (default) or 30days"	Enums(7days, 30days)
//	@Success		200		{object}	dto.Response{data=dto.TransactionReportResponse}
//	@Failure		401		{object}	dto.Response
//	@Failure		500		{object}	dto.Response
//	@Router			/dashboard/report [get]
func (d *DashboardController) GetTransactionReport(ctx *gin.Context) {
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

	rangeParam := ctx.DefaultQuery("range", "7days")
	if rangeParam != "7days" && rangeParam != "30days" {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Message: "Invalid range parameter",
			Success: false,
			Error:   "range must be '7days' or '30days'",
		})
		return
	}

	points, err := d.dashboardservice.GetTransactionReport(ctx.Request.Context(), email, rangeParam)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to fetch transaction report",
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Map model.ChartPoint → dto.ChartPointResponse
	resp := make([]dto.ChartPointResponse, 0, len(points))
	for _, p := range points {
		resp = append(resp, dto.ChartPointResponse{
			Label:   p.Label,
			Income:  p.Income,
			Expense: p.Expense,
		})
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data: dto.TransactionReportResponse{
			Range:  rangeParam,
			Points: resp,
		},
		Message: "Transaction report successfully retrieved",
		Success: true,
	})
}
