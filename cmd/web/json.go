package main

import (
	"encoding/json"
	"net/http"
)

func jsonEncode(w http.ResponseWriter, body any) error {
	return json.NewEncoder(w).Encode(body)
}
