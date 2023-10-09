package api

import (
	"encoding/json"
	"net/http"
)

type Jsonable interface {
}

func JsonResponse(j Jsonable, resposeCode int, w http.ResponseWriter) {
	b, err := json.Marshal(j)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(resposeCode)
	w.Write(b)
}
