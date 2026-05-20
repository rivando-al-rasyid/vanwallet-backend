package service

import (
	"context"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
)

type DashboardRepo interface {
	GetData(ctx context.Context, email string) (model.Dashboard, error)
}

type DashboardService struct {
	dashboardRepo DashboardRepo
}

func NewDashboardService(dashboardRepo DashboardRepo) *DashboardService {
	return &DashboardService{dashboardRepo: dashboardRepo}
}

func (d *DashboardService) GetData(ctx context.Context, email string) (model.Dashboard, error) {
	return d.dashboardRepo.GetData(ctx, email)
}
