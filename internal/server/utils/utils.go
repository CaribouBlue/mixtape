package utils

import (
	"encoding/json"

	"net/http"

	"github.com/a-h/templ"
)

func HandleJsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
	}
}

func HandleHtmlResponse(r *http.Request, w http.ResponseWriter, component templ.Component) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	component.Render(r.Context(), w)
}

func HandleRedirect(w http.ResponseWriter, r *http.Request, redirect string) error {
	metadata, err := ContextValue(r.Context(), RequestMetaDataCtxKey)
	if err != nil {
		return err

	}

	var isHtmxRequest bool
	if metadata == nil {
		isHtmxRequest = false
	} else {
		isHtmxRequest = metadata.IsHtmxRequest
	}

	if isHtmxRequest {
		w.Header().Add("HX-Redirect", redirect)
		w.WriteHeader(http.StatusFound)
		return nil
	} else {
		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}
}
