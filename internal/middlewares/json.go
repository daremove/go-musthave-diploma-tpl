package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const parsedJSONDataField = "parsedJSONDataField"

type ModelParameter interface {
	interface{} | []interface{}
}

type RequestWithModel[Model ModelParameter] struct {
	*http.Request
	data Model
}

func JSONMiddleware[Model ModelParameter](next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type is not application/json", http.StatusUnsupportedMediaType)
			return
		}

		var parsedData Model
		var buf bytes.Buffer

		if _, err := buf.ReadFrom(r.Body); err != nil {
			http.Error(w, fmt.Sprintf("Error occurred during reading from the body: %s", err.Error()), http.StatusBadRequest)
			return
		}

		if err := json.Unmarshal(buf.Bytes(), &parsedData); err != nil {
			http.Error(w, fmt.Sprintf("Error occurred during unmarshaling data %s", err.Error()), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), parsedJSONDataField, parsedData)))
	})
}

func GetParsedJSONData[Model ModelParameter](w http.ResponseWriter, r *http.Request) Model {
	data, ok := r.Context().Value(parsedJSONDataField).(Model)

	if !ok {
		http.Error(w, "Could not retrieve data from context", http.StatusInternalServerError)
		var empty Model
		return empty
	}

	return data
}
