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
	IsUserInConversation(conversationId int64, userId int64) (bool, error)
	FindOneOnOneConversation(userId1, userId2 int64) (int64, error)

	// Group Specific
	SetGroupName(id int64, name string) error
	SetGroupPhoto(id int64, photoURL string) error
	AddMember(groupId int64, userId int64) error
	RemoveMember(groupId int64, userId int64) error
	GetConversationMembers(conversationId int64) ([]int64, error)
	GetConversationMembersDetailed(conversationId int64) ([]User, error)
	UpdateParticipantLastRead(conversationId, userId int64) error

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
			last_read_at DATETIME,
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
			is_deleted BOOLEAN NOT NULL DEFAULT 0,
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

	// Migrations
	_, _ = db.Exec("ALTER TABLE messages ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT 0")
	_, _ = db.Exec("ALTER TABLE conversations ADD COLUMN last_message_at DATETIME")
	_, _ = db.Exec("ALTER TABLE participants ADD COLUMN last_read_at DATETIME")
	_, _ = db.Exec("UPDATE messages SET status = 1 WHERE status = 0")

	// Cleanup duplicate 1-on-1 conversations
	_, _ = db.Exec(`
		DELETE FROM conversations
		WHERE id IN (
			SELECT c1.id
			FROM conversations c1
			JOIN participants p1a ON c1.id = p1a.conversation_id
			JOIN participants p1b ON c1.id = p1b.conversation_id
			WHERE c1.is_group = 0 AND p1a.user_id < p1b.user_id
			AND EXISTS (
				SELECT 1 FROM conversations c2
				JOIN participants p2a ON c2.id = p2a.conversation_id
				JOIN participants p2b ON c2.id = p2b.conversation_id
				WHERE c2.is_group = 0 AND p2a.user_id = p1a.user_id AND p2b.user_id = p1b.user_id
				AND (c2.last_message_at > c1.last_message_at OR (c2.last_message_at = c1.last_message_at AND c2.id > c1.id))
			)
		)
	`)

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
	ID                    int64     `json:"conversationId"`
	Name                  string    `json:"name"`
	IsGroup               bool      `json:"isGroup"`
	PhotoURL              string    `json:"photoUrl"`
	LastMessageAt         time.Time `json:"latestMessageTime"`
	LatestMessagePreview  string    `json:"latestMessagePreview"`
	LatestMessageStatus   int       `json:"latestMessageStatus"`
	LatestMessageSenderId int64     `json:"latestMessageSenderId"`
	LatestMessageDeleted  bool      `json:"latestMessageDeleted"`
	UnreadCount           int       `json:"unreadCount"`
}

type Message struct {
	ID             int64      `json:"id"`
	ConversationId int64      `json:"conversationId"`
	SenderId       int64      `json:"senderId"`
	SenderName     string     `json:"senderName"`
	Content        string     `json:"content"`
	ContentType    string     `json:"contentType"`
	ReplyToId      *int64     `json:"replyToId"`
	TimeStamp      time.Time  `json:"timeStamp"`
	Status         int        `json:"status"`
	IsDeleted      bool       `json:"isDeleted"`
	Reactions      []Reaction `json:"reactions"`
}

type Reaction struct {
	MessageID   int64  `json:"-"`
	UserID      int64  `json:"-"`
	ReactorName string `json:"reactorName"`
	Emoticon    string `json:"emoticon"`
}
