package handlers

import (
	"net/http"
	"strconv"
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

type listSubscriptionsResponse struct {
	Items  []subscriptionResponse `json:"items"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
	Total  int64                  `json:"total"`
}

type totalResponse struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	UserID      *string `json:"user_id,omitempty"`
	ServiceName *string `json:"service_name,omitempty"`
	Total       int64   `json:"total"`
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

func (h *SubscriptionHandler) Get(c echo.Context) error {
	id := c.Param("id")

	row, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil {
		var r errorResponse
		r.Error.Code = "NOT_FOUND"
		r.Error.Message = "subscription not found"
		return c.JSON(http.StatusNotFound, r)
	}

	var endDate *string
	if row.EndMonth != nil {
		s := util.FormatMonth(*row.EndMonth)
		endDate = &s
	}

	resp := subscriptionResponse{
		ID:          row.ID,
		ServiceName: row.ServiceName,
		Price:       row.Price,
		UserID:      row.UserID,
		StartDate:   util.FormatMonth(row.StartMonth),
		EndDate:     endDate,
		CreatedAt:   row.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   row.UpdatedAt.Format(time.RFC3339),
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *SubscriptionHandler) Delete(c echo.Context) error {
	id := c.Param("id")

	err := h.repo.Delete(c.Request().Context(), id)
	if err != nil {
		var r errorResponse
		r.Error.Code = "NOT_FOUND"
		r.Error.Message = "subscription not found"
		return c.JSON(http.StatusNotFound, r)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *SubscriptionHandler) List(c echo.Context) error {
	limit := 20
	offset := 0

	if v := strings.TrimSpace(c.QueryParam("limit")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return writeValidationError(c, "limit must be a positive integer", "limit")
		}
		if n > 100 {
			n = 100
		}
		limit = n
	}
	if v := strings.TrimSpace(c.QueryParam("offset")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return writeValidationError(c, "offset must be a non-negative integer", "offset")
		}
		offset = n
	}

	var userID *string
	if v := strings.TrimSpace(c.QueryParam("user_id")); v != "" {
		if _, err := uuid.Parse(v); err != nil {
			return writeValidationError(c, "user_id must be a valid UUID", "user_id")
		}
		userID = &v
	}

	var serviceName *string
	if v := strings.TrimSpace(c.QueryParam("service_name")); v != "" {
		serviceName = &v
	}

	res, err := h.repo.List(c.Request().Context(), postgres.ListSubscriptionsParams{
		UserID:      userID,
		ServiceName: serviceName,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		var r errorResponse
		r.Error.Code = "INTERNAL_ERROR"
		r.Error.Message = "failed to list subscriptions"
		return c.JSON(http.StatusInternalServerError, r)
	}

	items := make([]subscriptionResponse, 0, len(res.Items))
	for _, row := range res.Items {
		var endDate *string
		if row.EndMonth != nil {
			s := util.FormatMonth(*row.EndMonth)
			endDate = &s
		}

		items = append(items, subscriptionResponse{
			ID:          row.ID,
			ServiceName: row.ServiceName,
			Price:       row.Price,
			UserID:      row.UserID,
			StartDate:   util.FormatMonth(row.StartMonth),
			EndDate:     endDate,
			CreatedAt:   row.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:   row.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	return c.JSON(http.StatusOK, listSubscriptionsResponse{
		Items:  items,
		Limit:  limit,
		Offset: offset,
		Total:  res.Total,
	})
}

func (h *SubscriptionHandler) Update(c echo.Context) error {
	id := c.Param("id")

	// валидируем id как UUID (чтобы не отдавать странные ошибки от БД)
	if _, err := uuid.Parse(id); err != nil {
		return writeValidationError(c, "id must be a valid UUID", "id")
	}

	var req createSubscriptionRequest // reuse (same fields)
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
	if req.EndDate != nil && strings.TrimSpace(*req.EndDate) != "" {
		em, err := util.ParseMonth(strings.TrimSpace(*req.EndDate))
		if err != nil {
			return writeValidationError(c, "end_date must be in MM-YYYY format", "end_date")
		}
		if em.Before(startMonth) {
			return writeValidationError(c, "end_date must be >= start_date", "end_date")
		}
		endMonthPtr = &em
	}

	row, err := h.repo.Update(c.Request().Context(), postgres.UpdateSubscriptionParams{
		ID:          id,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartMonth:  startMonth,
		EndMonth:    endMonthPtr,
	})
	if err != nil {
		var r errorResponse
		r.Error.Code = "NOT_FOUND"
		r.Error.Message = "subscription not found"
		return c.JSON(http.StatusNotFound, r)
	}

	var endDate *string
	if row.EndMonth != nil {
		s := util.FormatMonth(*row.EndMonth)
		endDate = &s
	}

	resp := subscriptionResponse{
		ID:          row.ID,
		ServiceName: row.ServiceName,
		Price:       row.Price,
		UserID:      row.UserID,
		StartDate:   util.FormatMonth(row.StartMonth),
		EndDate:     endDate,
		CreatedAt:   row.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   row.UpdatedAt.UTC().Format(time.RFC3339),
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *SubscriptionHandler) Total(c echo.Context) error {
	fromStr := strings.TrimSpace(c.QueryParam("from"))
	toStr := strings.TrimSpace(c.QueryParam("to"))
	if fromStr == "" {
		return writeValidationError(c, "from is required (MM-YYYY)", "from")
	}
	if toStr == "" {
		return writeValidationError(c, "to is required (MM-YYYY)", "to")
	}

	fromMonth, err := util.ParseMonth(fromStr)
	if err != nil {
		return writeValidationError(c, "from must be in MM-YYYY format", "from")
	}
	toMonth, err := util.ParseMonth(toStr)
	if err != nil {
		return writeValidationError(c, "to must be in MM-YYYY format", "to")
	}
	if toMonth.Before(fromMonth) {
		return writeValidationError(c, "to must be >= from", "to")
	}

	// filters
	var userID *string
	if v := strings.TrimSpace(c.QueryParam("user_id")); v != "" {
		if _, err := uuid.Parse(v); err != nil {
			return writeValidationError(c, "user_id must be a valid UUID", "user_id")
		}
		userID = &v
	}

	var serviceName *string
	if v := strings.TrimSpace(c.QueryParam("service_name")); v != "" {
		serviceName = &v
	}

	total, err := h.repo.Total(c.Request().Context(), postgres.TotalParams{
		FromMonth:   fromMonth,
		ToMonth:     toMonth,
		UserID:      userID,
		ServiceName: serviceName,
	})
	if err != nil {
		var r errorResponse
		r.Error.Code = "INTERNAL_ERROR"
		r.Error.Message = "failed to calculate total"
		return c.JSON(http.StatusInternalServerError, r)
	}

	resp := totalResponse{
		From:        util.FormatMonth(fromMonth),
		To:          util.FormatMonth(toMonth),
		UserID:      userID,
		ServiceName: serviceName,
		Total:       total,
	}
	return c.JSON(http.StatusOK, resp)
}
