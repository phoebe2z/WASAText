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
		_ = tx.Rollback()
		return message, err
	}

	// Create Message
	res, err := tx.Exec(`
		INSERT INTO messages (conversation_id, sender_id, content, content_type, reply_to_id, created_at, status)
		VALUES (?, ?, ?, ?, ?, ?, 1)
	`, conversationId, senderId, content, contentType, replyToId, time.Now())

	if err != nil {
		_ = tx.Rollback()
		return message, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return message, err
	}

	// Update Conversation LastMessageAt
	_, err = tx.Exec("UPDATE conversations SET last_message_at = ? WHERE id = ?", time.Now(), conversationId)
	if err != nil {
		_ = tx.Rollback()
		return message, err
	}

	err = tx.Commit()
	if err != nil {
		return message, err
	}

	message.ID = id
	message.ConversationId = conversationId
	message.SenderId = senderId
	message.SenderName = senderName
	message.Content = content
	message.ContentType = contentType
	message.ReplyToId = replyToId
	message.TimeStamp = time.Now()
	message.Status = 0

	return message, nil
}

func (db *appdbimpl) GetMessages(conversationId int64) ([]Message, error) {
	rows, err := db.c.Query(`
		SELECT 
			m.id, 
			m.conversation_id, 
			m.sender_id, 
			u.name as sender_name,
			m.created_at, 
			m.content, 
			m.content_type, 
			m.reply_to_id, 
			CASE 
				WHEN m.status >= 2 THEN 2
				WHEN NOT EXISTS (
					SELECT 1 FROM participants p 
					WHERE p.conversation_id = m.conversation_id 
					AND p.user_id != m.sender_id 
					AND (p.last_read_at IS NULL OR p.last_read_at < m.created_at)
				) THEN 2
				ELSE m.status
			END as status,
			m.is_deleted
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = ?
		ORDER BY m.created_at ASC
	`, conversationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		var replyTo sql.NullInt64
		if err := rows.Scan(&m.ID, &m.ConversationId, &m.SenderId, &m.SenderName, &m.TimeStamp, &m.Content, &m.ContentType, &replyTo, &m.Status, &m.IsDeleted); err != nil {
			return nil, err
		}
		if replyTo.Valid {
			m.ReplyToId = &replyTo.Int64
		}
		messages = append(messages, m)
	}

	// Fetch reactions for each message
	for i := range messages {
		reactions, err := db.GetReactions(messages[i].ID)
		if err != nil {
			return nil, err
		}
		messages[i].Reactions = reactions
	}

	return messages, rows.Err()
}

func (db *appdbimpl) GetMessage(id int64) (Message, error) {
	var m Message
	var replyTo sql.NullInt64
	err := db.c.QueryRow(`
		SELECT m.id, m.conversation_id, m.sender_id, u.name, m.content, m.content_type, m.reply_to_id, m.created_at, m.status, m.is_deleted
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.id = ?
	`, id).Scan(&m.ID, &m.ConversationId, &m.SenderId, &m.SenderName, &m.Content, &m.ContentType, &replyTo, &m.TimeStamp, &m.Status, &m.IsDeleted)

	if replyTo.Valid {
		m.ReplyToId = &replyTo.Int64
	}
	return m, err
}

func (db *appdbimpl) DeleteMessage(id int64) error {
	_, err := db.c.Exec("UPDATE messages SET is_deleted = 1 WHERE id = ?", id)
	return err
}

func (db *appdbimpl) SetMessageStatus(id int64, status int) error {
	_, err := db.c.Exec("UPDATE messages SET status = ? WHERE id = ?", status, id)
	return err
}
