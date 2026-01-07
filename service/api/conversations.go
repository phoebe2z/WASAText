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
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(conversations)
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

	// Enrich messages with reactions (N+1 query issue here potentially, but okay for homework)
	for i := range messages {
		reactions, err := rt.db.GetReactions(messages[i].ID)
		if err == nil {
			messages[i].Reactions = reactions // Assuming struct has Reactions field (I added it in database.go)
		}
	}

	w.WriteHeader(http.StatusOK)
	if messages == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(messages)
}
