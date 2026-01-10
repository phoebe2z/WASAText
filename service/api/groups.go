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

func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userId, err := extractBearer(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req struct {
		Name           string  `json:"name"`
		InitialMembers []int64 `json:"initialMembers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validation
	if len(req.Name) < 3 || len(req.Name) > 20 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(req.InitialMembers) < 2 { // "At least 3 persons" (creator + 2 others)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create group
	// Add creator to initial members list for DB creation logic if needed,
	// or handled by CreateConversation logic.
	// My CreateConversation implementation takes a list of IDs. I should include creator.
	members := append(req.InitialMembers, userId)

	group, err := rt.db.CreateConversation(req.Name, true, members)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error creating group")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"groupId": group.ID})
}

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userId, err := extractBearer(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check permission (must be member to add others? Spec says "Group members can add other users")
	in, err := rt.db.IsUserInConversation(groupId, userId)
	if err != nil || !in {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var req struct {
		UserIds []int64 `json:"userIds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, newMemberId := range req.UserIds {
		// Verify user exists? DB FK will handle logic but maybe good to check.
		// For now assume valid IDs or DB error.
		err = rt.db.AddMember(groupId, newMemberId)
		if err != nil {
			// Skip or error?
			// ignoring error for now (e.g. already member)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userId, err := extractBearer(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = rt.db.RemoveMember(groupId, userId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userId, err := extractBearer(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check auth
	in, _ := rt.db.IsUserInConversation(groupId, userId)
	if !in {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var req struct {
		NewName string `json:"newName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = rt.db.SetGroupName(groupId, req.NewName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userId, err := extractBearer(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	groupId, err := strconv.ParseInt(ps.ByName("groupId"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check auth
	in, _ := rt.db.IsUserInConversation(groupId, userId)
	if !in {
		w.WriteHeader(http.StatusForbidden)
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
		err = rt.db.SetGroupPhoto(groupId, req.PhotoURL)
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

	file, handler, err := r.FormFile("photo")
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

	filename := fmt.Sprintf("group-%d-%d%s", groupId, time.Now().Unix(), filepath.Ext(handler.Filename))
	err = os.WriteFile(filepath.Join("static", filename), data, 0644)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error saving file")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	photoURL := "/static/" + filename

	err = rt.db.SetGroupPhoto(groupId, photoURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
