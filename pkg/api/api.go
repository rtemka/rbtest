package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rtemka/rbtest/domain"
)

type repo = domain.Repository
type item = domain.Item

var (
	ErrInternal = errors.New("internal server error")
	ErrBadInput = errors.New("invalid input")
)

// API приложения.
type API struct {
	router *mux.Router
	repo   repo
	logger *log.Logger
}

// Возвращает новый объект *API
func New(repo repo, logger *log.Logger, cacheInterval time.Duration) *API {
	api := API{
		router: mux.NewRouter(),
		repo:   repo,
		logger: logger,
	}
	api.endpoints()

	return &api
}

// Router возвращает маршрутизатор запросов.
func (api *API) Router() *mux.Router {
	return api.router
}

func (api *API) endpoints() {
	api.router.Use(
		api.logRequestMiddleware,
		api.closerMiddleware,
		api.headersMiddleware,
	)

	api.router.HandleFunc("/items", api.itemsHandlerList()).Methods(http.MethodGet, http.MethodOptions)
	api.router.HandleFunc("/items/{id}", api.itemsHandlerGet()).Methods(http.MethodGet, http.MethodOptions)
	api.router.HandleFunc("/items/{id}", api.itemsHandlerDelete()).Methods(http.MethodDelete, http.MethodOptions)
	api.router.HandleFunc("/items", api.itemsHandlerPut()).Methods(http.MethodPut, http.MethodOptions)
}

// headersMiddleware задает обычные заголовки для всех ответов.
func (api *API) headersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// closerMiddleware считывает и закрывает тело запроса
// для повторного использования TCP-соединения.
func (api *API) closerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	})
}

// logRequestMiddleware логирует request
func (api *API) logRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		api.logger.Printf("method=%s path=%s query=%s vars=%s remote=%s",
			r.Method, r.URL.Path, r.URL.Query(), mux.Vars(r), r.RemoteAddr)
	})
}

func (api *API) WriteJSONError(w http.ResponseWriter, err error, code int) {
	w.WriteHeader(code)
	msg := map[string]string{"error": err.Error()}
	_ = json.NewEncoder(w).Encode(&msg)
}

func (api *API) WriteJSON(w http.ResponseWriter, data any, code int) {
	w.WriteHeader(code)
	if data == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(data)
}

// itemsHandlerList возвращает список сущностьей из БД.
func (api *API) itemsHandlerList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		items, err := api.repo.Items(ctx)
		if err != nil {
			api.WriteJSONError(w, ErrInternal, http.StatusInternalServerError)
			return
		}
		api.WriteJSON(w, items, http.StatusOK)
	}
}

// itemsHandlerDelete удлаляет сущность из БД.
func (api *API) itemsHandlerDelete() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		s := mux.Vars(r)["id"]
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			api.WriteJSONError(w, fmt.Errorf("%w: bad 'id' path parameter", ErrBadInput), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = api.repo.DeleteItem(ctx, id)
		if err != nil {
			api.WriteJSONError(w, ErrInternal, http.StatusInternalServerError)
			return
		}

		api.WriteJSON(w, map[string]any{"deleted": map[string]int64{"id": id}}, http.StatusOK)
	}
}

// itemsHandlerDelete получает из БД по id.
func (api *API) itemsHandlerGet() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		s := mux.Vars(r)["id"]
		id, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			api.WriteJSONError(w, fmt.Errorf("%w: bad 'id' path parameter", ErrBadInput), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		item, err := api.repo.Item(ctx, id)
		if err != nil {
			api.WriteJSONError(w, ErrInternal, http.StatusInternalServerError)
			return
		}

		api.WriteJSON(w, item, http.StatusOK)
	}
}

// itemsHandlerPut обновляет сущность в БД.
func (api *API) itemsHandlerPut() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var item item
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil || item.ID == 0 {
			api.WriteJSONError(w, fmt.Errorf("%w: bad JSON string in request body", ErrBadInput), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err = api.repo.UpdateItem(ctx, item)
		if err != nil {
			api.WriteJSONError(w, ErrInternal, http.StatusInternalServerError)
			return
		}

		api.WriteJSON(w, map[string]any{"updated": map[string]int64{"id": item.ID}}, http.StatusOK)
	}
}
