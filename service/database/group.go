package database

func (db *appdbimpl) SetGroupName(id int64, name string) error {
	_, err := db.c.Exec("UPDATE conversations SET name = ? WHERE id = ? AND is_group = 1", name, id)
	return err
}

func (db *appdbimpl) SetGroupPhoto(id int64, photoURL string) error {
	_, err := db.c.Exec("UPDATE conversations SET photo_url = ? WHERE id = ? AND is_group = 1", photoURL, id)
	return err
}

func (db *appdbimpl) AddMember(groupId int64, userId int64) error {
	_, err := db.c.Exec("INSERT INTO participants (conversation_id, user_id) VALUES (?, ?)", groupId, userId)
	return err
}

func (db *appdbimpl) RemoveMember(groupId int64, userId int64) error {
	_, err := db.c.Exec("DELETE FROM participants WHERE conversation_id = ? AND user_id = ?", groupId, userId)
	return err
}
