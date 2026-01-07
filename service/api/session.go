package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse request body
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate Name
	if len(req.Name) < 3 || len(req.Name) > 16 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Check regex (simplified, just allowing chars)
	// pattern: '^.*?$' (which means anything, but minLength 3 implies non-empty)

	// Check if user exists
	user, err := rt.db.GetUserByName(req.Name)
	if err == nil {
		// User exists, return ID
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int64{"identifier": user.ID})
		return
	}

	// User does not exist, create new
	newUser, err := rt.db.CreateUser(req.Name)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error creating user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"identifier": newUser.ID})
}
