package database

func (db *appdbimpl) AddReaction(messageId int64, userId int64, emoticon string) error {
	_, err := db.c.Exec("REPLACE INTO reactions (message_id, user_id, emoticon) VALUES (?, ?, ?)", messageId, userId, emoticon)
	return err
}

func (db *appdbimpl) RemoveReaction(messageId int64, userId int64) error {
	_, err := db.c.Exec("DELETE FROM reactions WHERE message_id = ? AND user_id = ?", messageId, userId)
	return err
}

func (db *appdbimpl) GetReactions(messageId int64) ([]Reaction, error) {
	rows, err := db.c.Query(`
		SELECT r.user_id, u.name, r.emoticon 
		FROM reactions r
		JOIN users u ON r.user_id = u.id
		WHERE r.message_id = ?
	`, messageId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []Reaction
	for rows.Next() {
		var r Reaction
		r.MessageID = messageId
		if err := rows.Scan(&r.UserID, &r.ReactorName, &r.Emoticon); err != nil {
			return nil, err
		}
		reactions = append(reactions, r)
	}
	return reactions, rows.Err()
}
