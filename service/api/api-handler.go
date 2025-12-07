package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {

	// Public
	rt.router.POST("/session", rt.wrap(rt.login))

	// User
	rt.router.PUT("/user/name", rt.wrap(rt.setUserName))
	rt.router.PUT("/user/photo", rt.wrap(rt.setUserPhoto))

	// Conversations
	rt.router.GET("/conversations", rt.wrap(rt.getConversations))
	rt.router.GET("/conversations/:conversationId", rt.wrap(rt.getConversation))

	// Messages
	rt.router.POST("/messages", rt.wrap(rt.setMessage))
	rt.router.DELETE("/messages/:messageId", rt.wrap(rt.deleteMessage))
	rt.router.POST("/messages/:messageId/forward", rt.wrap(rt.forwardMessage))
	rt.router.POST("/messages/:messageId/reaction", rt.wrap(rt.addReaction))
	rt.router.DELETE("/messages/:messageId/reaction", rt.wrap(rt.removeReaction))

	// Groups
	rt.router.POST("/groups", rt.wrap(rt.addGroup))
	rt.router.POST("/groups/:groupId/members", rt.wrap(rt.setGroupMembers))
	rt.router.DELETE("/groups/:groupId/me", rt.wrap(rt.leaveGroup))
	rt.router.PUT("/groups/:groupId/name", rt.wrap(rt.setGroupName))
	rt.router.PUT("/groups/:groupId/photo", rt.wrap(rt.setGroupPhoto))

	// Special
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
