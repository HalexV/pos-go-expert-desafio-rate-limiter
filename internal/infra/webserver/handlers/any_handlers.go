package handlers

import (
	"net/http"
)

type AnyHandler struct{}

func NewAnyHandler() *AnyHandler {
	return &AnyHandler{}
}

func (h *AnyHandler) GetAny(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Nice! Try Again if you can!"))
}
