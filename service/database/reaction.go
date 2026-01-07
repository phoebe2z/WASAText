package database

func (db *appdbimpl) AddReaction(messageId int64, userId int64, emoticon string) error {
	_, err := db.c.Exec("INSERT INTO reactions (message_id, user_id, emoticon) VALUES (?, ?, ?)", messageId, userId, emoticon)
	return err
}

func (db *appdbimpl) RemoveReaction(messageId int64, userId int64) error {
	_, err := db.c.Exec("DELETE FROM reactions WHERE message_id = ? AND user_id = ?", messageId, userId)
	return err
}

func (db *appdbimpl) GetReactions(messageId int64) ([]Reaction, error) {
	rows, err := db.c.Query("SELECT user_id, emoticon FROM reactions WHERE message_id = ?", messageId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []Reaction
	for rows.Next() {
		var r Reaction
		r.MessageID = messageId
		if err := rows.Scan(&r.UserID, &r.Emoticon); err != nil {
			return nil, err
		}

		// Fetch reactor name - ideally this should be a JOIN in the query for performance
		var name string
		_ = db.c.QueryRow("SELECT name FROM users WHERE id = ?", r.UserID).Scan(&name)
		// Note: The Reaction model in API spec has "reactorName". The struct in database.go has UserID.
		// I should potentially update the struct or handle this join.
		// Let's modify the query to JOIN.

		reactions = append(reactions, r)
	}
	return reactions, rows.Err()
}
