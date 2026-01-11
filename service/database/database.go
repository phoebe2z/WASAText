/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	// User
	CreateUser(name string) (User, error)
	GetUser(id int64) (User, error)
	GetUserByName(name string) (User, error)
	SetUserName(id int64, name string) error
	SetUserPhoto(id int64, photoURL string) error
	ListUsers(query string) ([]User, error)

	// Conversation
	CreateConversation(name string, isGroup bool, initialMembers []int64) (Conversation, error)
	GetConversations(userId int64) ([]Conversation, error)
	GetConversation(id int64) (Conversation, error)

	// Group Specific
	SetGroupName(id int64, name string) error
	SetGroupPhoto(id int64, photoURL string) error
	AddMember(groupId int64, userId int64) error
	RemoveMember(groupId int64, userId int64) error
	IsUserInConversation(conversationId int64, userId int64) (bool, error)
	GetConversationMembers(conversationId int64) ([]int64, error)
	GetConversationMembersDetailed(conversationId int64) ([]User, error)

	// Message
	SendMessage(conversationId int64, senderId int64, content string, contentType string, replyToId *int64) (Message, error)
	GetMessages(conversationId int64) ([]Message, error)
	GetMessage(id int64) (Message, error)
	DeleteMessage(id int64) error
	SetMessageStatus(id int64, status int) error

	// Reaction
	AddReaction(messageId int64, userId int64, emoticon string) error
	RemoveReaction(messageId int64, userId int64) error
	GetReactions(messageId int64) ([]Reaction, error)

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Create tables
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			photo_url TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			is_group BOOLEAN NOT NULL DEFAULT 0,
			photo_url TEXT,
			last_message_at DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS participants (
			conversation_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			PRIMARY KEY (conversation_id, user_id),
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id INTEGER NOT NULL,
			sender_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			content_type TEXT NOT NULL,
			reply_to_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			status INTEGER DEFAULT 0,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS reactions (
			message_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			emoticon TEXT NOT NULL,
			PRIMARY KEY (message_id, user_id),
			FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);`,
	}

	for _, stmt := range tables {
		_, err := db.Exec(stmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

// Models

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	PhotoURL string `json:"photoUrl"`
}

type Conversation struct {
	ID            int64     `json:"conversationId"`
	Name          string    `json:"name"`
	IsGroup       bool      `json:"isGroup"`
	PhotoURL      string    `json:"photoUrl"`
	LastMessageAt time.Time `json:"latestMessageTime"`
}

type Message struct {
	ID             int64      `json:"id"`
	ConversationID int64      `json:"conversationId"`
	SenderID       int64      `json:"senderId"` // NOTE: API spec says senderName, DB stores ID
	Content        string     `json:"content"`
	ContentType    string     `json:"contentType"`
	ReplyToID      *int64     `json:"replyToId"` // Pointer to handle null
	CreatedAt      time.Time  `json:"timeStamp"`
	Status         int        `json:"status"`
	Reactions      []Reaction `json:"reactions"`
}

type Reaction struct {
	MessageID int64  `json:"-"`
	UserID    int64  `json:"-"`
	Emoticon  string `json:"emoticon"`
}
