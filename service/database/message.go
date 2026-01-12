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

	// Fetch sender name
	var senderName string
	err = tx.QueryRow("SELECT name FROM users WHERE id = ?", senderId).Scan(&senderName)
	if err != nil {
		tx.Rollback()
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
	message.SenderName = senderName
	message.Content = content
	message.ContentType = contentType
	message.ReplyToID = replyToId
	message.CreatedAt = time.Now()
	message.Status = 0

	return message, nil
}

func (db *appdbimpl) GetMessages(conversationId int64) ([]Message, error) {
	rows, err := db.c.Query(`
		SELECT m.id, m.conversation_id, m.sender_id, u.name, m.content, m.content_type, m.reply_to_id, m.created_at, m.status
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = ?
		ORDER BY m.created_at DESC
	`, conversationId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		var replyTo sql.NullInt64
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.SenderName, &m.Content, &m.ContentType, &replyTo, &m.CreatedAt, &m.Status); err != nil {
			return nil, err
		}
		if replyTo.Valid {
			id := replyTo.Int64
			m.ReplyToID = &id
		}
		messages = append(messages, m)
	}

	return messages, rows.Err()
}

func (db *appdbimpl) GetMessage(id int64) (Message, error) {
	var m Message
	var replyTo sql.NullInt64
	err := db.c.QueryRow(`
		SELECT m.id, m.conversation_id, m.sender_id, u.name, m.content, m.content_type, m.reply_to_id, m.created_at, m.status
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.id = ?
	`, id).Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.SenderName, &m.Content, &m.ContentType, &replyTo, &m.CreatedAt, &m.Status)

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
