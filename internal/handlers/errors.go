package handlers

import "net/http"

func handleInternalServerError(w http.ResponseWriter) {
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
