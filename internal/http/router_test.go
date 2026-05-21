package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthRouteReturnsOK(t *testing.T) {
	router := NewRouter()

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("health status = %q, want %q", body["status"], "ok")
	}
}

func TestRouterRegistersGatewayHandler(t *testing.T) {
	router := NewRouter(WithGatewayHandler(fakeGatewayRouteRegistrar{}))

	request := httptest.NewRequest(http.MethodGet, "/gateway/test", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestRouterRegistersEvalHandler(t *testing.T) {
	router := NewRouter(WithEvalHandler(fakeEvalRouteRegistrar{}))

	request := httptest.NewRequest(http.MethodGet, "/eval/test", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}
}

type fakeGatewayRouteRegistrar struct{}

func (fakeGatewayRouteRegistrar) RegisterRoutes(router *gin.Engine) {
	router.GET("/gateway/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "gateway"})
	})
}

type fakeEvalRouteRegistrar struct{}

func (fakeEvalRouteRegistrar) RegisterRoutes(router *gin.Engine) {
	router.GET("/eval/test", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "eval"})
	})
}
