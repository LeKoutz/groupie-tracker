package handlers

import (
	"fmt"
	"groupie-tracker/api"
	"groupie-tracker/models"
	"groupie-tracker/services"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

// Parse templates once before server startup
var (
    index_tmpl  = template.Must(template.ParseFiles("templates/index.html"))
    artist_tmpl = template.Must(template.ParseFiles("templates/artist_detail.html"))
    error_tmpl  = template.Must(template.ParseFiles("templates/error.html"))
)

// HandleErrors renders an error page with the given status code, error message, and response.
func HandleErrors(w http.ResponseWriter, statusCode int, message, response string) {
	errorData := struct {
		StatusCode 	int
		Message    	string
		Response string
	}{
		StatusCode:  statusCode,
		Message:     http.StatusText(statusCode),
		Response: 	 response,
	}	
	w.WriteHeader(statusCode)
	if err := error_tmpl.Execute(w, errorData); err != nil {
		http.Error(w, fmt.Sprintf("Error %d: %s %s", statusCode, message, response), statusCode)
		return
	}
}