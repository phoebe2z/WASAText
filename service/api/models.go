package api

import (
	"time"
)
type ErrorResponse struct {
	Error string `json:"error"`
	Code  *int   `json:"code,omitempty"` 
}

// Reaction 
type Reaction struct {
	ReactorName string `json:"reactorName"`
	Emoticon    string `json:"emoticon"`
}

// Message
type Message struct {
	ID             int        `json:"id"`
	ConversationID int        `json:"conversationId"`
	SenderName     string     `json:"senderName"`
	TimeStamp      time.Time  `json:"timeStamp"`
	Content        string     `json:"content"`
	ContentType    string     `json:"contentType"` // "text" or "photo"
	ReplyToID      *int       `json:"replyToId,omitempty"` 
	Status         int        `json:"status"`      // 0: Sent, 1: Received, 2: Read
	Reactions      []Reaction `json:"reactions"`
}

// ConversationInfo 
type ConversationInfo struct {
	ConversationID         int       `json:"conversationId"`
	Name                   string    `json:"name"`
	PhotoURL               *string   `json:"photoURL,omitempty"`
	IsGroup                bool      `json:"isGroup"`
	LatestMessageTime      time.Time `json:"latestMessageTime"`
	LatestMessagePreview   string    `json:"latestMessagePreview"`
	UnreadCount            int       `json:"unreadCount"`
}

// POST /session (doLogin) requestbody
type LoginRequest struct {
	Name string `json:"name"` 
}

// POST /session (doLogin) response (201 Created)
type LoginResponse struct {
	Identifier int `json:"identifier"` 
}

// PUT /user/name (setMyUserName) requestbody
type UpdateUserNameRequest struct {
	NewName string `json:"newName"`
}

// POST /messages (sendMessage) requestbody
type SendMessageRequest struct {
	ConversationID int    `json:"conversationId"`
	Content        string `json:"content"`
	ContentType    string `json:"contentType"`
	ReplyToID      *int   `json:"replyToId,omitempty"`
}

// POST /groups (createGroup) requestbody
type CreateGroupRequest struct {
	Name     string `json:"name"`
	MemberIDs []int  `json:"memberIds"`
}

// POST /groups/{groupId}/members (addToGroup) requestbody
type UserIDsRequest struct {
	UserIDs []int `json:"userIds"`
}

// POST /messages/{messageId}/forward (forwardMessage) requestbody
type ForwardMessageRequest struct {
	ConversationIDs []int `json:"conversationIds"` 
}

type User struct {
	ID       int
	Name     string
	PhotoURL string
	
}

type Conversation struct {
	ID       int
	Name     string
	IsGroup  bool
	PhotoURL string
	Members  []int // List of User IDs
}

var (
	NextUserID         int = 1002
	NextConversationID int = 1
	NextMessageID      int = 1
)

// Simple in-memory storage for demonstration
var (
	// Users: [ID] -> User. 
	Users = make(map[int]User)
	// Conversations: [ID] -> Conversation details.
	Conversations = make(map[int]Conversation)
	// Messages: [ConversationID] -> []Message.
	Messages = make(map[int][]Message)
)
