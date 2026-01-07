package database

import (
	"database/sql"
	"time"
)

func (db *appdbimpl) SendMessage(conversationId int64, senderId int64, content string, contentType string, replyToId *int64) (Message, error) {
	var message Message
	tx, err := db.c.Begin()
	if err != nil {
		return message, err
	}

	// Create Message
	res, err := tx.Exec(`
		INSERT INTO messages (conversation_id, sender_id, content, content_type, reply_to_id, created_at, status)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`, conversationId, senderId, content, contentType, replyToId, time.Now())

	if err != nil {
		tx.Rollback()
		return message, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return message, err
	}

	// Update Conversation LastMessageAt
	_, err = tx.Exec("UPDATE conversations SET last_message_at = ? WHERE id = ?", time.Now(), conversationId)
	if err != nil {
		tx.Rollback()
		return message, err
	}

	err = tx.Commit()
	if err != nil {
		return message, err
	}

	message.ID = id
	message.ConversationID = conversationId
	message.SenderID = senderId
	message.Content = content
	message.ContentType = contentType
	message.ReplyToID = replyToId
	message.CreatedAt = time.Now()
	message.Status = 0

	return message, nil
}

func (db *appdbimpl) GetMessages(conversationId int64) ([]Message, error) {
	rows, err := db.c.Query(`
		SELECT id, conversation_id, sender_id, content, content_type, reply_to_id, created_at, status
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at DESC 
	`, conversationId) // API says reverse chronologically, usually means newest first, but let's check spec details or standard chat app. Usually API returns history. API Spec says "reverse chronologically".
	// NOTE: Spec says "Array of messages... reverse chronologically". This usually means newest first.
	// But chat apps usually display oldest at top. If the API client appends to bottom, it might want chronological.
	// However, if spec says reverse chronological, I should follow spec.

	// Correcting to DESC based on spec "reverse chronologically"

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		var replyTo sql.NullInt64
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.ContentType, &replyTo, &m.CreatedAt, &m.Status); err != nil {
			return nil, err
		}
		if replyTo.Valid {
			id := replyTo.Int64
			m.ReplyToID = &id
		}
		messages = append(messages, m)
	}

	// If spec expects reverse chronological (Newest first), we should iterate and sort or change query.
	// I'll stick to DESC in DB query if that's what's meant.
	// Re-reading spec: "Array of messages in the conversation, reverse chronologically"

	return messages, rows.Err()
}

func (db *appdbimpl) GetMessage(id int64) (Message, error) {
	var m Message
	var replyTo sql.NullInt64
	err := db.c.QueryRow(`
		SELECT id, conversation_id, sender_id, content, content_type, reply_to_id, created_at, status
		FROM messages WHERE id = ?
	`, id).Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.ContentType, &replyTo, &m.CreatedAt, &m.Status)

	if replyTo.Valid {
		rid := replyTo.Int64
		m.ReplyToID = &rid
	}
	return m, err
}

func (db *appdbimpl) DeleteMessage(id int64) error {
	_, err := db.c.Exec("DELETE FROM messages WHERE id = ?", id)
	return err
}

func (db *appdbimpl) SetMessageStatus(id int64, status int) error {
	_, err := db.c.Exec("UPDATE messages SET status = ? WHERE id = ?", status, id)
	return err
}
