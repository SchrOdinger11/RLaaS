package handlers

import (
	"net/http"
)

// CheckHandler handles the /check endpoint.
func CheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Request Allowed"))
}
