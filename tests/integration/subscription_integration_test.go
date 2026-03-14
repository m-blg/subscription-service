package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"subscription-service/internal/handler"
	"subscription-service/internal/model"
	"subscription-service/internal/repository"
	"subscription-service/internal/service"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testPool *pgxpool.Pool
	router   *gin.Engine
)

func TestMain(m *testing.M) {
	// Setup
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:secret@localhost:5433/subs_db?sslmode=disable"
	}

	// Wait for DB to be ready
	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("pgx", dbURL)
		if err == nil {
			if err = db.Ping(); err == nil {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Fatalf("could not connect to test database: %v", err)
	}

	// Run migrations
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}
	// Path to migrations from the perspective of this test file
	if err := goose.Up(db, "../../migrations"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	db.Close()

	// Connect with pgxpool
	ctx := context.Background()
	testPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("unable to connect to database pool: %v", err)
	}
	defer testPool.Close()

	// Initialize handlers
	repo := repository.NewSubscriptionRepo(testPool)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := service.NewSubscriptionService(repo, logger)
	h := handler.NewSubscriptionHandler(svc)

	gin.SetMode(gin.TestMode)
	router = gin.Default()
	v1 := router.Group("/api/v1")
	{
		subs := v1.Group("/subscriptions")
		h.DefineRoutesGIN(subs)
	}

	code := m.Run()

	// TODO: Cleanup

	os.Exit(code)
}

func TestSubscriptionLifecycle(t *testing.T) {
	userID := uuid.New()
	serviceName := "Netflix"
	price := model.RUB(1000)
	startDateStr := "01-2024"
	endDateStr := "12-2024"

	// Create Subscription
	reqBody := handler.SubscriptionRequest{
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   parseMonthYear(t, startDateStr),
		EndDate:     func() *model.MonthYear { my := parseMonthYear(t, endDateStr); return &my }(),
	}

	jsonReq, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/subscriptions/", bytes.NewBuffer(jsonReq))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdSub model.Subscription
	err := json.Unmarshal(w.Body.Bytes(), &createdSub)
	require.NoError(t, err)
	assert.Equal(t, serviceName, createdSub.ServiceName)
	assert.Equal(t, price, createdSub.Price)
	assert.Equal(t, userID, createdSub.UserID)
	assert.NotEqual(t, uuid.Nil, createdSub.ID)
	assert.False(t, createdSub.CreatedAt.IsZero())
	assert.False(t, createdSub.UpdatedAt.IsZero())

	subID := createdSub.ID
	initialUpdatedAt := createdSub.UpdatedAt

	// Get Subscription
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/subscriptions/"+subID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var fetchedSub model.Subscription
	err = json.Unmarshal(w.Body.Bytes(), &fetchedSub)
	require.NoError(t, err)
	assert.Equal(t, subID, fetchedSub.ID)
	assert.Equal(t, initialUpdatedAt.Unix(), fetchedSub.UpdatedAt.Unix())

	// List Subscriptions
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/subscriptions/?user_id="+userID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var subList []model.Subscription
	err = json.Unmarshal(w.Body.Bytes(), &subList)
	require.NoError(t, err)
	assert.Len(t, subList, 1)
	assert.Equal(t, subID, subList[0].ID)

	// Update Subscription
	time.Sleep(time.Second) // Ensure UpdatedAt will be different
	newPrice := model.RUB(1200)
	updateReqBody := handler.SubscriptionRequest{
		ServiceName: serviceName,
		Price:       newPrice,
		UserID:      userID,
		StartDate:   parseMonthYear(t, startDateStr),
	}
	jsonUpdateReq, _ := json.Marshal(updateReqBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/subscriptions/"+subID.String(), bytes.NewBuffer(jsonUpdateReq))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var updatedSub model.Subscription
	err = json.Unmarshal(w.Body.Bytes(), &updatedSub)
	require.NoError(t, err)
	assert.Equal(t, newPrice, updatedSub.Price)
	assert.True(t, updatedSub.UpdatedAt.After(initialUpdatedAt))

	// Total Cost
	w = httptest.NewRecorder()
	// startDateStr := "01-2024" ; endDateStr := "12-2024"
	// total?user_id=...&from=01-2024&until=02-2024
	req, _ = http.NewRequest("GET", fmt.Sprintf("/api/v1/subscriptions/total?user_id=%s&from=01-2024&until=02-2024", userID.String()), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var totalResp handler.TotalCostResponse
	err = json.Unmarshal(w.Body.Bytes(), &totalResp)
	require.NoError(t, err)
	assert.Equal(t, newPrice, totalResp.Total)

	// Delete Subscription
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/subscriptions/"+subID.String(), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify Delete
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/subscriptions/"+subID.String(), nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func parseMonthYear(t *testing.T, s string) model.MonthYear {
	var my model.MonthYear
	err := my.UnmarshalJSON([]byte(`"` + s + `"`))
	require.NoError(t, err)
	return my
}
