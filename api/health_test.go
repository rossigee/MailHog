package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/pat"
	"github.com/smartystreets/goconvey/convey"

	"github.com/mailhog/MailHog-Server/config"
	"github.com/mailhog/data"
	"github.com/mailhog/storage"
)

func TestHealthAPI(t *testing.T) {
	convey.Convey("Health API should work correctly", t, func() {
		// Create test configuration
		conf := &config.Config{
			WebPath:      "",
			StorageType:  "memory",
			Storage:      storage.CreateInMemory(),
			MessageChan:  make(chan *data.Message),
			OutgoingSMTP: make(map[string]*config.OutgoingSMTP),
		}

		// Create router and health API
		router := pat.New()
		healthAPI := CreateHealthAPI(conf, router, "1.0.0-test")

		convey.So(healthAPI, convey.ShouldNotBeNil)
		convey.So(healthAPI.config, convey.ShouldEqual, conf)
		convey.So(healthAPI.startTime, convey.ShouldHappenBefore, time.Now())

		convey.Convey("Health endpoint should return OK", func() {
			req, err := http.NewRequest("GET", "/health", nil)
			convey.So(err, convey.ShouldBeNil)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			convey.So(rr.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(rr.Header().Get("Content-Type"), convey.ShouldEqual, "application/json")

			var response HealthResponse
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			convey.So(err, convey.ShouldBeNil)
			convey.So(response.Status, convey.ShouldEqual, "ok")
			convey.So(response.Version, convey.ShouldEqual, "1.0.0-test")
			convey.So(response.Timestamp, convey.ShouldNotBeEmpty)
			convey.So(response.Uptime, convey.ShouldNotBeEmpty)
		})

		convey.Convey("Readiness endpoint should return ready with memory storage", func() {
			req, err := http.NewRequest("GET", "/ready", nil)
			convey.So(err, convey.ShouldBeNil)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			convey.So(rr.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(rr.Header().Get("Content-Type"), convey.ShouldEqual, "application/json")

			var response ReadinessResponse
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			convey.So(err, convey.ShouldBeNil)
			convey.So(response.Status, convey.ShouldEqual, "ready")
			convey.So(response.Storage, convey.ShouldEqual, "memory")
			convey.So(response.StorageOk, convey.ShouldBeTrue)
			convey.So(response.Timestamp, convey.ShouldNotBeEmpty)
		})

		convey.Convey("Metrics endpoint should return prometheus format", func() {
			// Add a test message to storage
			msg := &data.Message{
				ID: "test-message",
				From: &data.Path{
					Relays:  []string{},
					Mailbox: "test",
					Domain:  "example.com",
				},
				To: []*data.Path{{
					Relays:  []string{},
					Mailbox: "recipient",
					Domain:  "example.com",
				}},
				Content: &data.Content{
					Headers: map[string][]string{
						"Subject": {"Test Subject"},
					},
					Body: "Test body",
				},
			}
			conf.Storage.Store(msg)

			req, err := http.NewRequest("GET", "/metrics", nil)
			convey.So(err, convey.ShouldBeNil)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			convey.So(rr.Code, convey.ShouldEqual, http.StatusOK)
			convey.So(rr.Header().Get("Content-Type"), convey.ShouldEqual, "text/plain")

			body := rr.Body.String()
			convey.So(body, convey.ShouldContainSubstring, "# HELP mailhog_messages_total")
			convey.So(body, convey.ShouldContainSubstring, "# TYPE mailhog_messages_total counter")
			convey.So(body, convey.ShouldContainSubstring, "mailhog_messages_total 1")
			convey.So(body, convey.ShouldContainSubstring, "# HELP mailhog_uptime_seconds")
			convey.So(body, convey.ShouldContainSubstring, "# TYPE mailhog_uptime_seconds gauge")
			convey.So(body, convey.ShouldContainSubstring, "mailhog_uptime_seconds")
			convey.So(body, convey.ShouldContainSubstring, "# HELP mailhog_memory_usage_bytes")
			convey.So(body, convey.ShouldContainSubstring, "# TYPE mailhog_memory_usage_bytes gauge")
			convey.So(body, convey.ShouldContainSubstring, "mailhog_memory_usage_bytes")
			convey.So(body, convey.ShouldContainSubstring, "# HELP mailhog_goroutines")
			convey.So(body, convey.ShouldContainSubstring, "# TYPE mailhog_goroutines gauge")
			convey.So(body, convey.ShouldContainSubstring, "mailhog_goroutines")
			convey.So(body, convey.ShouldContainSubstring, "mailhog_storage_type_info{storage_type=\"memory\"} 1")
			convey.So(body, convey.ShouldContainSubstring, "mailhog_storage_connected 1")
		})

		convey.Convey("WebPath prefix should be respected", func() {
			// Create config with WebPath
			confWithPath := &config.Config{
				WebPath:      "/mailhog",
				StorageType:  "memory",
				Storage:      storage.CreateInMemory(),
				MessageChan:  make(chan *data.Message),
				OutgoingSMTP: make(map[string]*config.OutgoingSMTP),
			}

			routerWithPath := pat.New()
			CreateHealthAPI(confWithPath, routerWithPath, "1.0.0-test")

			// Test health endpoint with prefix
			req, err := http.NewRequest("GET", "/mailhog/health", nil)
			convey.So(err, convey.ShouldBeNil)

			rr := httptest.NewRecorder()
			routerWithPath.ServeHTTP(rr, req)

			convey.So(rr.Code, convey.ShouldEqual, http.StatusOK)

			// Test that endpoint without prefix returns 404
			req, err = http.NewRequest("GET", "/health", nil)
			convey.So(err, convey.ShouldBeNil)

			rr = httptest.NewRecorder()
			routerWithPath.ServeHTTP(rr, req)

			convey.So(rr.Code, convey.ShouldEqual, http.StatusNotFound)
		})
	})
}

func TestReadinessWithStorageFailure(t *testing.T) {
	convey.Convey("Readiness endpoint should handle storage failures", t, func() {
		// Create a mock storage that fails
		mockStorage := &MockFailingStorage{}

		conf := &config.Config{
			WebPath:      "",
			StorageType:  "mock",
			Storage:      mockStorage,
			MessageChan:  make(chan *data.Message),
			OutgoingSMTP: make(map[string]*config.OutgoingSMTP),
		}

		router := pat.New()
		CreateHealthAPI(conf, router, "1.0.0-test")

		req, err := http.NewRequest("GET", "/ready", nil)
		convey.So(err, convey.ShouldBeNil)

		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		convey.So(rr.Code, convey.ShouldEqual, http.StatusServiceUnavailable)

		var response ReadinessResponse
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		convey.So(err, convey.ShouldBeNil)
		convey.So(response.Status, convey.ShouldEqual, "not_ready")
		convey.So(response.StorageOk, convey.ShouldBeFalse)
	})
}

// MockFailingStorage is a mock storage that always fails
type MockFailingStorage struct{}

func (m *MockFailingStorage) Store(msg *data.Message) (string, error) {
	return "", nil
}

func (m *MockFailingStorage) Count() int {
	return 0
}

func (m *MockFailingStorage) Search(kind, query string, start, limit int) (*data.Messages, int, error) {
	return nil, 0, nil
}

func (m *MockFailingStorage) List(start, limit int) (*data.Messages, error) {
	// This will cause the readiness check to fail
	return nil, fmt.Errorf("storage error")
}

func (m *MockFailingStorage) DeleteOne(id string) error {
	return nil
}

func (m *MockFailingStorage) DeleteAll() error {
	return nil
}

func (m *MockFailingStorage) Load(id string) (*data.Message, error) {
	return nil, nil
}
