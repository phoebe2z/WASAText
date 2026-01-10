package database

import (
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
		tx.Rollback()
		return conversation, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return conversation, err
	}

	// Add Participants (Unique)
	stmt, err := tx.Prepare("INSERT INTO participants (conversation_id, user_id) VALUES (?, ?)")
	if err != nil {
		tx.Rollback()
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
			tx.Rollback()
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
			c.last_message_at
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
		if err := rows.Scan(&c.ID, &c.Name, &c.IsGroup, &c.PhotoURL, &c.LastMessageAt); err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	return conversations, rows.Err()
}

func (db *appdbimpl) GetConversation(id int64) (Conversation, error) {
	var c Conversation
	err := db.c.QueryRow(`
		SELECT id, IFNULL(name, ''), is_group, IFNULL(photo_url, ''), last_message_at
		FROM conversations WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.IsGroup, &c.PhotoURL, &c.LastMessageAt)
	return c, err
}

func (db *appdbimpl) IsUserInConversation(conversationId int64, userId int64) (bool, error) {
	var count int
	err := db.c.QueryRow("SELECT COUNT(*) FROM participants WHERE conversation_id = ? AND user_id = ?", conversationId, userId).Scan(&count)
	return count > 0, err
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
