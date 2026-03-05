package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iampsih/subscriptions-service/internal/repository/postgres"
	"github.com/iampsih/subscriptions-service/internal/util"
	"github.com/labstack/echo/v4"
)

type SubscriptionHandler struct {
	repo *postgres.SubscriptionRepo
}

func NewSubscriptionHandler(repo *postgres.SubscriptionRepo) *SubscriptionHandler {
	return &SubscriptionHandler{repo: repo}
}

type createSubscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

type subscriptionResponse struct {
	ID          string  `json:"id"`
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type errorResponse struct {
	Error struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	} `json:"error"`
}

func writeValidationError(c echo.Context, msg string, field string) error {
	var r errorResponse
	r.Error.Code = "VALIDATION_ERROR"
	r.Error.Message = msg
	if field != "" {
		r.Error.Details = map[string]string{"field": field}
	}
	return c.JSON(http.StatusBadRequest, r)
}

func (h *SubscriptionHandler) Create(c echo.Context) error {
	var req createSubscriptionRequest
	if err := c.Bind(&req); err != nil {
		return writeValidationError(c, "invalid json body", "")
	}

	req.ServiceName = strings.TrimSpace(req.ServiceName)
	if req.ServiceName == "" {
		return writeValidationError(c, "service_name is required", "service_name")
	}
	if len(req.ServiceName) > 255 {
		return writeValidationError(c, "service_name is too long (max 255)", "service_name")
	}
	if req.Price <= 0 {
		return writeValidationError(c, "price must be > 0", "price")
	}
	if _, err := uuid.Parse(req.UserID); err != nil {
		return writeValidationError(c, "user_id must be a valid UUID", "user_id")
	}

	startMonth, err := util.ParseMonth(req.StartDate)
	if err != nil {
		return writeValidationError(c, "start_date must be in MM-YYYY format", "start_date")
	}

	var endMonthPtr *time.Time
	var endDateStr *string
	if req.EndDate != nil && strings.TrimSpace(*req.EndDate) != "" {
		em, err := util.ParseMonth(strings.TrimSpace(*req.EndDate))
		if err != nil {
			return writeValidationError(c, "end_date must be in MM-YYYY format", "end_date")
		}
		if em.Before(startMonth) {
			return writeValidationError(c, "end_date must be >= start_date", "end_date")
		}
		endMonthPtr = &em
		s := util.FormatMonth(em)
		endDateStr = &s
	}

	row, err := h.repo.Create(c.Request().Context(), postgres.CreateSubscriptionParams{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartMonth:  startMonth,
		EndMonth:    endMonthPtr,
	})
	if err != nil {
		var r errorResponse
		r.Error.Code = "INTERNAL_ERROR"
		r.Error.Message = "failed to create subscription"
		return c.JSON(http.StatusInternalServerError, r)
	}

	resp := subscriptionResponse{
		ID:          row.ID,
		ServiceName: row.ServiceName,
		Price:       row.Price,
		UserID:      row.UserID,
		StartDate:   util.FormatMonth(row.StartMonth),
		EndDate:     endDateStr,
		CreatedAt:   row.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   row.UpdatedAt.UTC().Format(time.RFC3339),
	}

	return c.JSON(http.StatusCreated, resp)
}
