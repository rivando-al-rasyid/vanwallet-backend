package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/dto"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/model"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/pkg"
	"github.com/rivando-al-rasyid/vanwallet-backend/internals/service"
)

type TopupController struct {
	topupService *service.TopupService
}

func NewTopupController(topupService *service.TopupService) *TopupController {
	return &TopupController{topupService: topupService}
}

// CreateTopup godoc
// @Summary      Initiate wallet top-up via Midtrans
// @Description  Creates a pending top-up and returns a Midtrans Snap token for payment.
// @Tags         Topups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body           body      dto.TopupRequest  true  "Top-up payload"
// @Success      201            {object}  dto.Response{data=dto.TopupResponse}
// @Failure      400            {object}  dto.Response{error}
// @Failure      401            {object}  dto.Response{error}
// @Failure      403            {object}  dto.Response{error}
// @Failure      500            {object}  dto.Response{error}
// @Router       /transaction/topup [post]
func (t *TopupController) CreateTopup(ctx *gin.Context) {
	email, ok := pkg.CurrentUserEmail(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, dto.NewError("Unauthorized", errors.New("missing user context")))
		return
	}

	var body dto.TopupRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err))
		return
	}

	walletID, err := uuid.Parse(body.WalletID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid wallet_id", errors.New("wallet_id must be a valid UUID")))
		return
	}

	if !checkWalletOwnership(ctx, t.topupService, email, walletID) {
		return
	}

	pm := model.PaymentMethod(body.PaymentMethod)
	result, err := t.topupService.CreateTopup(ctx.Request.Context(), email, model.Topup{
		WalletID:      walletID,
		Amount:        body.Amount,
		PaymentMethod: &pm,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to create top-up", err))
		return
	}

	topup := result.Topup
	extRef := ""
	if topup.ExternalReference != nil {
		extRef = *topup.ExternalReference
	}
	pmStr := ""
	if topup.PaymentMethod != nil {
		pmStr = string(*topup.PaymentMethod)
	}

	ctx.JSON(http.StatusCreated, dto.NewSuccess("Top-up initiated successfully", dto.TopupResponse{
		ID:                topup.ID.String(),
		WalletID:          topup.WalletID.String(),
		Amount:            topup.Amount,
		PaymentMethod:     pmStr,
		ExternalReference: extRef,
		Status:            string(topup.Status),
		SnapToken:         result.SnapToken,
		RedirectURL:       result.RedirectURL,
		ClientKey:         result.ClientKey,
		CreatedAt:         topup.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}))
}

type MidtransWebhookController struct {
	topupService *service.TopupService
}

func NewMidtransWebhookController(topupService *service.TopupService) *MidtransWebhookController {
	return &MidtransWebhookController{topupService: topupService}
}

// HandleNotification godoc
// @Summary      Midtrans payment notification webhook
// @Description  Receives Midtrans payment status updates and settles successful top-ups.
// @Tags         Webhooks
// @Accept       json
// @Produce      json
// @Param        body  body  dto.MidtransNotification  true  "Midtrans notification payload"
// @Success      200   {object}  dto.Response
// @Failure      400   {object}  dto.Response{error}
// @Failure      401   {object}  dto.Response{error}
// @Failure      500   {object}  dto.Response{error}
// @Router       /webhooks/midtrans [post]
func (m *MidtransWebhookController) HandleNotification(ctx *gin.Context) {
	rawBody, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid request body", err))
		return
	}

	var notification dto.MidtransNotification
	if err := json.Unmarshal(rawBody, &notification); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.NewError("Invalid notification payload", err))
		return
	}

	if err := m.topupService.HandleMidtransNotification(ctx.Request.Context(), notification, rawBody); err != nil {
		if err.Error() == "invalid midtrans signature" {
			ctx.JSON(http.StatusUnauthorized, dto.NewError("Invalid signature", err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, dto.NewError("Failed to process notification", err))
		return
	}

	ctx.JSON(http.StatusOK, dto.NewSuccessNoData("Notification processed"))
}
