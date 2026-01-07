package database

func (db *appdbimpl) CreateUser(name string) (User, error) {
	var user User
	res, err := db.c.Exec("INSERT INTO users (name) VALUES (?)", name)
	if err != nil {
		return user, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return user, err
	}
	user.ID = id
	user.Name = name
	return user, nil
}

func (db *appdbimpl) GetUser(id int64) (User, error) {
	var user User
	err := db.c.QueryRow("SELECT id, name, IFNULL(photo_url, '') FROM users WHERE id = ?", id).Scan(&user.ID, &user.Name, &user.PhotoURL)
	return user, err
}

func (db *appdbimpl) GetUserByName(name string) (User, error) {
	var user User
	err := db.c.QueryRow("SELECT id, name, IFNULL(photo_url, '') FROM users WHERE name = ?", name).Scan(&user.ID, &user.Name, &user.PhotoURL)
	return user, err
}

func (db *appdbimpl) SetUserName(id int64, name string) error {
	_, err := db.c.Exec("UPDATE users SET name = ? WHERE id = ?", name, id)
	return err
}

func (db *appdbimpl) SetUserPhoto(id int64, photoURL string) error {
	_, err := db.c.Exec("UPDATE users SET photo_url = ? WHERE id = ?", photoURL, id)
	return err
}
