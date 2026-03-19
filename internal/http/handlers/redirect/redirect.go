package redirect

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/storage"
)

// 1. НАПИШИ ИНТЕРФЕЙС URLGetter
type URLGetter interface {
	GetURL(ctx context.Context, alias string) (string, error)
}

// 2. НАПИШИ КОНСТРУКТОР ХЕНДЛЕРА
func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 3. ДОСТАНЬ АЛИАС ИЗ ПУТИ
		// подсказка: alias := r.PathValue("alias")
		alias := r.PathValue("alias")
		// 4. СХОДИ В ИНТЕРФЕЙС
		if alias != "" {
			http.Error(w, "Invalid Request", http.StatusBadRequest)
			return
		}

		url, err := urlGetter.GetURL(context.Background(), alias)
		// 5. ОБРАБОТАЙ ОШИБКИ (404 или 500)

		if errors.Is(err, storage.ErrURLNotFound) {

			http.Error(w, http.StatusNotFound, 404)
			return
		} else {
			http.Error(w, http.StatusInternalServerError)
			return
		}
		// 6. СДЕЛАЙ РЕДИРЕКТ
	}
}
