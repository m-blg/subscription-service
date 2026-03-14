package handler

import (
	"subscription-service/internal/model"

	"github.com/google/uuid"
)

// type Subscription struct {
// 	ID          uuid.UUID        `json:"id"`
// 	ServiceName string           `json:"service_name"`
// 	Price       model.RUB        `json:"price"`
// 	UserID      uuid.UUID        `json:"user_id"`
// 	StartDate   model.MonthYear  `json:"start_date"`
// 	EndDate     *model.MonthYear `json:"end_date,omitempty"`
// }

type SubscriptionRequest struct {
	ServiceName string           `json:"service_name" binding:"required"`
	Price       model.RUB        `json:"price" binding:"required,gt=0"`
	UserID      uuid.UUID        `json:"user_id" binding:"required"`
	StartDate   model.MonthYear  `json:"start_date" binding:"required" swaggertype:"string" example:"07-2025"`
	EndDate     *model.MonthYear `json:"end_date,omitempty" swaggertype:"string" example:"12-2025"`
}

type TotalCostResponse struct {
	Total model.RUB `json:"total"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
