package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userId, err := extractBearer(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conversations, err := rt.db.GetConversations(userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting conversations")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if conversations == nil {
		_, _ = w.Write([]byte("[]"))
		return
	}
	_ = json.NewEncoder(w).Encode(conversations)
}

func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userId, err := extractBearer(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conversationId, err := strconv.ParseInt(ps.ByName("conversationId"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if user is in conversation
	in, err := rt.db.IsUserInConversation(conversationId, userId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error checking membership")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !in {
		w.WriteHeader(http.StatusNotFound) // or Forbidden
		return
	}

	messages, err := rt.db.GetMessages(conversationId)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error getting messages")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if messages == nil {
		_, _ = w.Write([]byte("[]"))
		return
	}

	_ = json.NewEncoder(w).Encode(messages)
}

func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userId, err := extractBearer(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var req struct {
		RecipientName string `json:"recipientName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Look up user by name
	user, err := rt.db.GetUserByName(req.RecipientName)
	if err != nil {
		// Assume error means not found (or check specifics if possible, but for now 404 is safe for user not found)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Prevent self-chat
	if user.ID == userId {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "cannot chat with yourself"})
		return
	}

	members := []int64{userId, user.ID}

	// Check if 1-on-1 conversation already exists
	existingId, err := rt.db.FindOneOnOneConversation(userId, user.ID)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error finding existing conversation")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if existingId != 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"message":        "conversation already exists",
			"conversationId": existingId,
		})
		return
	}

	// Create conversation
	conversation, err := rt.db.CreateConversation("", false, members)
	if err != nil {
		rt.baseLogger.WithError(err).Error("error creating conversation")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(conversation)
}
