package redirect

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"url-shortener/internal/storage"
)

type URLGetter interface {
	GetURL(ctx context.Context, alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		alias := r.PathValue("alias")
		if alias == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		url, err := urlGetter.GetURL(r.Context(), alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("url not found", slog.String("alias", alias))
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}

			log.Error("failed to get url", slog.String("alias", alias), slog.Any("err", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		log.Info("redirecting", slog.String("alias", alias), slog.String("url", url))
		http.Redirect(w, r, url, http.StatusFound)
	}
}
