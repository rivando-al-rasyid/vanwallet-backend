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

	Dashboard, err := d.dashboardservice.GetData(ctx, email)
	if err != nil {
		if err.Error() == "Data Not not found" {
			ctx.JSON(http.StatusNotFound, dto.Response{
				Message: "Failed to fetch transaction Data ",
				Success: false,
				Error:   "Data tidak ditemukan",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Message: "Failed to fetch transaction ",
			Success: false,
			Error:   "Internal server error",
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Data:    Dashboard,
		Message: "Profile successfully retrieved",
		Success: true,
	})
}
