package database

import (
	"database/sql"
	"time"
)

func (db *appdbimpl) CreateConversation(name string, isGroup bool, initialMembers []int64) (Conversation, error) {
	var conversation Conversation
	// Transaction to ensure atomicity
	tx, err := db.c.Begin()
	if err != nil {
		return conversation, err
	}

	// Create Conversation
	res, err := tx.Exec("INSERT INTO conversations (name, is_group, last_message_at) VALUES (?, ?, ?)", name, isGroup, time.Now())
	if err != nil {
		_ = tx.Rollback()
		return conversation, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return conversation, err
	}

	// Add Participants (Unique)
	stmt, err := tx.Prepare("INSERT INTO participants (conversation_id, user_id) VALUES (?, ?)")
	if err != nil {
		_ = tx.Rollback()
		return conversation, err
	}
	defer stmt.Close()

	seen := make(map[int64]bool)
	for _, memberId := range initialMembers {
		if seen[memberId] {
			continue
		}
		seen[memberId] = true
		_, err = stmt.Exec(id, memberId)
		if err != nil {
			_ = tx.Rollback()
			return conversation, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return conversation, err
	}

	conversation.ID = id
	conversation.Name = name
	conversation.IsGroup = isGroup
	conversation.LastMessageAt = time.Now() // Approximate
	return conversation, nil
}

func (db *appdbimpl) GetConversations(userId int64) ([]Conversation, error) {
	rows, err := db.c.Query(`
		SELECT 
			c.id, 
			CASE 
				WHEN c.is_group = 1 THEN IFNULL(c.name, '') 
				ELSE (
					SELECT name FROM users WHERE id = COALESCE(
						(SELECT user_id FROM participants WHERE conversation_id = c.id AND user_id != ? LIMIT 1),
						?
					)
				)
			END, 
			c.is_group, 
			CASE 
				WHEN c.is_group = 1 THEN IFNULL(c.photo_url, '') 
				ELSE (
					SELECT IFNULL(photo_url, '') FROM users WHERE id = COALESCE(
						(SELECT user_id FROM participants WHERE conversation_id = c.id AND user_id != ? LIMIT 1),
						?
					)
				)
			END, 
			c.last_message_at,
			(SELECT content FROM messages WHERE conversation_id = c.id ORDER BY created_at DESC LIMIT 1) as latest_preview,
			(SELECT sender_id FROM messages WHERE conversation_id = c.id ORDER BY created_at DESC LIMIT 1) as latest_sender,
			(SELECT 
				CASE 
					WHEN m_inner.status >= 2 THEN 2
					WHEN NOT EXISTS (
						SELECT 1 FROM participants p 
						WHERE p.conversation_id = m_inner.conversation_id 
						AND p.user_id != m_inner.sender_id 
						AND (p.last_read_at IS NULL OR p.last_read_at < m_inner.created_at)
					) THEN 2
					ELSE m_inner.status
				END
			FROM messages m_inner WHERE m_inner.conversation_id = c.id ORDER BY m_inner.created_at DESC LIMIT 1) as latest_status,
			(SELECT is_deleted FROM messages WHERE conversation_id = c.id ORDER BY created_at DESC LIMIT 1) as latest_deleted,
			(SELECT COUNT(*) FROM messages m 
			WHERE m.conversation_id = c.id 
			AND m.sender_id != p_me.user_id
			AND (p_me.last_read_at IS NULL OR m.created_at > p_me.last_read_at)) as unread_count
		FROM conversations c
		JOIN participants p_me ON c.id = p_me.conversation_id
		WHERE p_me.user_id = ?
		ORDER BY c.last_message_at DESC
	`, userId, userId, userId, userId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var c Conversation
		var preview sql.NullString
		var senderId sql.NullInt64
		var status sql.NullInt64
		var deleted sql.NullBool
		var lastAt sql.NullTime
		if err := rows.Scan(&c.ID, &c.Name, &c.IsGroup, &c.PhotoURL, &lastAt, &preview, &senderId, &status, &deleted, &c.UnreadCount); err != nil {
			return nil, err
		}
		if lastAt.Valid {
			c.LastMessageAt = lastAt.Time
		}
		if preview.Valid {
			c.LatestMessagePreview = preview.String
		}
		if senderId.Valid {
			c.LatestMessageSenderId = senderId.Int64
		}
		if status.Valid {
			c.LatestMessageStatus = int(status.Int64)
		}
		if deleted.Valid {
			c.LatestMessageDeleted = deleted.Bool
		}
		conversations = append(conversations, c)
	}
	return conversations, rows.Err()
}

func (db *appdbimpl) GetConversation(id int64) (Conversation, error) {
	var c Conversation
	var lastAt sql.NullTime
	err := db.c.QueryRow(`
		SELECT id, IFNULL(name, ''), is_group, IFNULL(photo_url, ''), last_message_at
		FROM conversations WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.IsGroup, &c.PhotoURL, &lastAt)
	if lastAt.Valid {
		c.LastMessageAt = lastAt.Time
	}
	return c, err
}

func (db *appdbimpl) IsUserInConversation(conversationId int64, userId int64) (bool, error) {
	var count int
	err := db.c.QueryRow("SELECT COUNT(*) FROM participants WHERE conversation_id = ? AND user_id = ?", conversationId, userId).Scan(&count)
	return count > 0, err
}

func (db *appdbimpl) FindOneOnOneConversation(userId1, userId2 int64) (int64, error) {
	var id int64
	err := db.c.QueryRow(`
		SELECT c.id 
		FROM conversations c
		JOIN participants p1 ON c.id = p1.conversation_id
		JOIN participants p2 ON c.id = p2.conversation_id
		WHERE c.is_group = 0 
		AND p1.user_id = ? 
		AND p2.user_id = ?
	`, userId1, userId2).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

func (db *appdbimpl) GetConversationMembers(conversationId int64) ([]int64, error) {
	rows, err := db.c.Query("SELECT user_id FROM participants WHERE conversation_id = ?", conversationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		members = append(members, id)
	}
	return members, rows.Err()
}

func (db *appdbimpl) GetConversationMembersDetailed(conversationId int64) ([]User, error) {
	rows, err := db.c.Query(`
		SELECT u.id, u.name, IFNULL(u.photo_url, '')
		FROM users u
		JOIN participants p ON u.id = p.user_id
		WHERE p.conversation_id = ?
	`, conversationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.PhotoURL); err != nil {
			return nil, err
		}
		members = append(members, u)
	}
	return members, rows.Err()
}

func (db *appdbimpl) UpdateParticipantLastRead(conversationId, userId int64) error {
	_, err := db.c.Exec("UPDATE participants SET last_read_at = ? WHERE conversation_id = ? AND user_id = ?", time.Now(), conversationId, userId)
	return err
}
