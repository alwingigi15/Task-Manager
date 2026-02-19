package utils

import (
	"encoding/json"
	"net/http"
)

func HandleError(w http.ResponseWriter, err error, msg string, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func SuccessResponse(w http.ResponseWriter, msg string, data interface{}, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": msg, "data": data})
}
