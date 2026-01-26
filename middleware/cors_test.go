package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sets CORS headers on GET request", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// CORSヘッダーの確認
		origin := w.Header().Get("Access-Control-Allow-Origin")
		if origin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", origin)
		}

		methods := w.Header().Get("Access-Control-Allow-Methods")
		if methods == "" {
			t.Error("Access-Control-Allow-Methods header should be set")
		}

		headers := w.Header().Get("Access-Control-Allow-Headers")
		if headers == "" {
			t.Error("Access-Control-Allow-Headers header should be set")
		}
	})

	t.Run("handles OPTIONS preflight request", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// OPTIONSリクエストは204を返す
		if w.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
		}

		// CORSヘッダーが設定されていることを確認
		origin := w.Header().Get("Access-Control-Allow-Origin")
		if origin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", origin)
		}
	})

	t.Run("allows Authorization header", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		headers := w.Header().Get("Access-Control-Allow-Headers")
		if headers == "" {
			t.Error("Access-Control-Allow-Headers header should be set")
		}

		// Authorizationヘッダーが許可されていることを確認
		if !containsString(headers, "Authorization") {
			t.Error("Access-Control-Allow-Headers should include 'Authorization'")
		}
	})

	t.Run("allows X-User-ID header", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		headers := w.Header().Get("Access-Control-Allow-Headers")

		// X-User-IDヘッダーが許可されていることを確認
		if !containsString(headers, "X-User-ID") {
			t.Error("Access-Control-Allow-Headers should include 'X-User-ID'")
		}
	})

	t.Run("allows required HTTP methods", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		methods := w.Header().Get("Access-Control-Allow-Methods")

		requiredMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		for _, method := range requiredMethods {
			if !containsString(methods, method) {
				t.Errorf("Access-Control-Allow-Methods should include '%s'", method)
			}
		}
	})
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstr(s, substr)))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
