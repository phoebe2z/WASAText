package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

// Helper to extract Bearer token
func extractBearer(r *http.Request) (int64, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, http.ErrNoCookie // abuse this error for missing auth
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, http.ErrNoCookie
	}
	return strconv.ParseInt(parts[1], 10, 64)
}

func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userId, err := extractBearer(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req struct {
		NewName string `json:"newName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(req.NewName) < 3 || len(req.NewName) > 16 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check for duplication
	_, err = rt.db.GetUserByName(req.NewName)
	if err == nil {
		w.WriteHeader(http.StatusConflict) // 409
		return
	}

	err = rt.db.SetUserName(userId, req.NewName)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error setting user name")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userId, err := extractBearer(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if JSON
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var req struct {
			PhotoURL string `json:"photoUrl"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = rt.db.SetUserPhoto(userId, req.PhotoURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// Multipart
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("newPhoto")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save file
	data, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("user-%d-%d%s", userId, time.Now().Unix(), filepath.Ext(handler.Filename))
	err = os.WriteFile(filepath.Join("static", filename), data, 0644)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error saving file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	photoURL := "/static/" + filename

	err = rt.db.SetUserPhoto(userId, photoURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"photoUrl": photoURL})
}

func (rt *_router) getMyProfile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userId, err := extractBearer(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := rt.db.GetUser(userId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}
func (rt *_router) listUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	_, err := extractBearer(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	users, err := rt.db.ListUsers(query)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error listing users")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}
