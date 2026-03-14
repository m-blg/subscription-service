package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"subscription-service/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubscriptionService interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	List(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	CalculateTotalCost(ctx context.Context, userID *uuid.UUID, serviceName string, from, until time.Time) (model.RUB, error)
}

type SubscriptionHandler struct {
	svc SubscriptionService
}

func NewSubscriptionHandler(svc SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc}
}

func (h *SubscriptionHandler) HandleServiceError(c *gin.Context, err error, internalErrMsg ...string) {
	if errors.Is(err, model.ErrNotFound) {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
	} else if errors.Is(err, model.ErrValidation) {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	} else if errors.Is(err, context.DeadlineExceeded) {
		c.JSON(http.StatusGatewayTimeout, ErrorResponse{Error: "request took too long"})
	} else {
		msg := "internal error"
		if len(internalErrMsg) > 0 {
			msg = internalErrMsg[0]
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: msg})
	}
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Creates a new subscription for a user
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body SubscriptionRequest true "Subscription data"
// @Success 201 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	sub := &model.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if err := h.svc.Create(c.Request.Context(), sub); err != nil {
		h.HandleServiceError(c, err, "failed to create subscription")
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// GetByID godoc
// @Summary Get subscription by ID
// @Description Retrieves a subscription by its UUID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}
	sub, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.HandleServiceError(c, err, "failed to get subscription")
		return
	}
	c.JSON(http.StatusOK, sub)
}

// ListSubscriptions godoc
// @Summary List subscriptions
// @Description Returns a list of subscriptions, optionally filtered by user ID, with pagination
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User UUID"
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {array} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [get]
func (h *SubscriptionHandler) List(c *gin.Context) {
	userIDStr := c.Query("user_id")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	var userID *uuid.UUID
	if userIDStr != "" {
		parsedID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
			return
		}
		userID = &parsedID
	}

	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	subs, err := h.svc.List(c.Request.Context(), userID, limit, offset)
	if err != nil {
		h.HandleServiceError(c, err, "failed to list subscriptions")
		return
	}
	c.JSON(http.StatusOK, subs)
}

// Update godoc
// @Summary Update a subscription
// @Description Updates an existing subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID (UUID)"
// @Param input body SubscriptionRequest true "Subscription data"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	subID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid subscription id"})
		return
	}

	var req SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	sub := &model.Subscription{
		ID:          subID,
		UserID:      req.UserID,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if err := h.svc.Update(c.Request.Context(), sub); err != nil {
		h.HandleServiceError(c, err, "failed to update subscription")
		return
	}

	c.JSON(http.StatusOK, sub)
}

// DeleteSubscription godoc
// @Summary Delete a subscription
// @Description Deletes a subscription by ID
// @Tags subscriptions
// @Param id path string true "Subscription ID (UUID)"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.HandleServiceError(c, err, "failed to delete subscription")
		return
	}
	c.Status(http.StatusNoContent)
}

// GetTotalCost godoc
// @Summary Calculate total cost for period
// @Description Calculates the total cost of subscriptions for a user within a specified time range
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "Filter by User UUID"
// @Param service_name query string false "Filter by service name"
// @Param from query string true "Start period (MM-YYYY)"
// @Param until query string true "End period (MM-YYYY)"
// @Success 200 {object} TotalCostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/total [get]
func (h *SubscriptionHandler) GetTotalCost(c *gin.Context) {
	userIDStr := c.Query("user_id")
	serviceName := c.Query("service_name")
	fromStr := c.Query("from")
	untilStr := c.Query("until")

	var userID *uuid.UUID
	if userIDStr != "" {
		parsedID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
			return
		}
		userID = &parsedID
	}

	layout := model.LayoutMMYYYY
	from, err := time.Parse(layout, fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid 'from' date"})
		return
	}

	until, err := time.Parse(layout, untilStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid 'until' date"})
		return
	}

	total, err := h.svc.CalculateTotalCost(c.Request.Context(), userID, serviceName, from, until)
	if err != nil {
		h.HandleServiceError(c, err, "failed to calculate total cost")
		return
	}

	c.JSON(http.StatusOK, TotalCostResponse{Total: total})
}
