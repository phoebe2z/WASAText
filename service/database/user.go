package database

func (db *appdbimpl) ListUsers(query string) ([]User, error) {
	var users []User
	sqlQuery := "SELECT id, name, IFNULL(photo_url, '') FROM users"
	var args []interface{}

	if query != "" {
		sqlQuery += " WHERE name LIKE ?"
		args = append(args, "%"+query+"%")
	}

	rows, err := db.c.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		err = rows.Scan(&u.ID, &u.Name, &u.PhotoURL)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (db *appdbimpl) CreateUser(name string) (User, error) {
	var u User
	res, err := db.c.Exec("INSERT INTO users (name) VALUES (?)", name)
	if err != nil {
		return u, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return u, err
	}
	u.ID = id
	u.Name = name
	return u, nil
}

func (db *appdbimpl) GetUser(id int64) (User, error) {
	var u User
	err := db.c.QueryRow("SELECT id, name, IFNULL(photo_url, '') FROM users WHERE id = ?", id).Scan(&u.ID, &u.Name, &u.PhotoURL)
	return u, err
}

func (db *appdbimpl) GetUserByName(name string) (User, error) {
	var u User
	err := db.c.QueryRow("SELECT id, name, IFNULL(photo_url, '') FROM users WHERE name = ?", name).Scan(&u.ID, &u.Name, &u.PhotoURL)
	return u, err
}

func (db *appdbimpl) SetUserName(id int64, name string) error {
	_, err := db.c.Exec("UPDATE users SET name = ? WHERE id = ?", name, id)
	return err
}

func (db *appdbimpl) SetUserPhoto(id int64, photoURL string) error {
	_, err := db.c.Exec("UPDATE users SET photo_url = ? WHERE id = ?", photoURL, id)
	return err
}
