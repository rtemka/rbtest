package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rtemka/rbtest/pkg/repo/memdb"
)

func TestAPI(t *testing.T) {
	api := New(memdb.New(), log.New(io.Discard, "", 0), 1*time.Minute)

	tb, err := json.Marshal(map[string]any{"id": 1, "name": "upd test"})
	if err != nil {
		t.Fatalf("TestAPI = err %v", err)
	}

	tests := []struct {
		name           string
		path           string
		method         string
		body           io.Reader
		wantStatusCode int
	}{
		{
			name:           "itemsHandlerList",
			path:           "/items",
			method:         http.MethodGet,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "itemsHandlerDelete",
			path:           "/items/1",
			method:         http.MethodDelete,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "itemsHandlerGet",
			path:           "/items/1",
			method:         http.MethodGet,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "itemsHandlerPutError",
			path:           "/items",
			method:         http.MethodPut,
			body:           strings.NewReader(`{id:990, name: "fail upd test"}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "itemsHandlerPut",
			path:           "/items",
			method:         http.MethodPut,
			body:           bytes.NewReader(tb),
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, tt.body)
			rr := httptest.NewRecorder()

			api.router.ServeHTTP(rr, req)

			resp := rr.Result()

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("%s() resp code = %d, want %d", tt.name, resp.StatusCode, tt.wantStatusCode)
			}
		})
	}

}
