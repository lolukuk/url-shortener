package save

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"time"
)

type URLSaver interface {
	SaveURL(ctx context.Context, urlToSave string, alias string) (int64, error)
}

type Request struct {
	URL   string `json:"url"`
	Alias string `json:"alias"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Alias  string `json:"alias,omitempty"`
}

const (
	statusOK    = "OK"
	statusError = "Error"
)

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("failed to decode request", slog.Any("err", err))
			respondJSON(w, log, http.StatusBadRequest, Response{Status: statusError, Error: "invalid json"})
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = randomString(6)
		}

		_, err := urlSaver.SaveURL(r.Context(), req.URL, alias)
		if err != nil {
			log.Error("failed to save url", slog.Any("err", err))
			respondJSON(w, log, http.StatusInternalServerError, Response{Status: statusError, Error: "internal error"})
			return
		}

		respondJSON(w, log, http.StatusOK, Response{Status: statusOK, Alias: alias})
	}
}

func respondJSON(w http.ResponseWriter, log *slog.Logger, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error("failed to encode response", slog.Any("err", err))
	}
}

const randAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randAlphabet[rand.Intn(len(randAlphabet))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
