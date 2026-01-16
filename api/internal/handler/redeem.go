package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whatsapp-promo-poc/api/internal/service"
	"github.com/whatsapp-promo-poc/pkg/models"
)

type RedeemHandler struct {
	promoService *service.PromoService
}

func NewRedeemHandler(promoService *service.PromoService) *RedeemHandler {
	return &RedeemHandler{
		promoService: promoService,
	}
}

func (h *RedeemHandler) Handle(c *gin.Context) {
	var req models.RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.RedeemResponse{
			Status:    "error",
			ErrorCode: "INVALID_REQUEST",
			Message: models.Message{
				FR: "Requete invalide",
				EN: "Invalid request",
			},
		})
		return
	}

	// Default language to French if not specified
	if req.Language == "" {
		req.Language = "fr"
	}

	result, err := h.promoService.RedeemCode(
		c.Request.Context(),
		req.PhoneNumber,
		req.Code,
		req.Language,
		req.MessageID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.RedeemResponse{
			Status:    "error",
			ErrorCode: "SYSTEM_ERROR",
			Message: models.Message{
				FR: "Une erreur s'est produite. Veuillez reessayer plus tard.",
				EN: "An error occurred. Please try again later.",
			},
		})
		return
	}

	if result.Success {
		c.JSON(http.StatusOK, models.RedeemResponse{
			Status:        "success",
			TransactionID: result.TransactionID,
			Reward: &models.Reward{
				Type:        result.RewardType,
				Amount:      result.RewardDesc,
				Description: result.RewardDesc,
			},
			Message: models.Message{
				FR: result.MessageFR,
				EN: result.MessageEN,
			},
		})
	} else {
		c.JSON(http.StatusOK, models.RedeemResponse{
			Status:        "error",
			TransactionID: result.TransactionID,
			ErrorCode:     string(result.ErrorCode),
			Message: models.Message{
				FR: result.MessageFR,
				EN: result.MessageEN,
			},
		})
	}
}
