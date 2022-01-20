package main

type User struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Description string `json:"description"`
}

//blobs related
func (u User) AddBlob(content string) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	_, err = db.Query("INSERT INTO Blobs (ID_user, content) VALUES (?, ?)", u.ID, content)
	return err
}

func (u User) ModifyBlob(id int, content string) error {
	Blob, err := QueryBlobByID(id)
	if err != nil {
		return nil
	}

	return Blob.Modify(content)
}

func (u User) DeleteBlob(id int) error {
	Blob, err := QueryBlobByID(id)
	if err != nil {
		return nil
	}

	return Blob.Delete()
}

func (u User) HasLiked(id int) (bool, error) {
	db, err := connectToDB()
	if err != nil {
		return false, err
	}

	var count int
	err = db.QueryRow("SELECT count(*) FROM likes WHERE ID_user = ? AND ID_blob = ?", u.ID, id).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

//user relations related
func (u User) Follow(id int) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	_, err = db.Query("INSERT INTO follows (ID_user_follower, ID_user_followed) VALUES (?, ?)", u.ID, id)
	return err
}

func (u User) Unfollow(id int) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	_, err = db.Query("DELETE FROM Blobs WHERE ID_user_follower = ? AND ID_user_followed = ?", u.ID, id)
	return err
}

func (u User) GetFollowers() ([]User, error) {
	db, err := connectToDB()
	if err != nil {
		return []User{}, err
	}

	rows, err := db.Query("SELECT follower.* FROM follows f inner join users followed on f.ID_user_followed = followed.ID join users follower on f.ID_user_follower = follower.ID WHERE followed.ID = ?", u.ID)
	if err != nil {
		return []User{}, err
	}
	var users []User
	for rows.Next() {
		var user User
		err = rows.Scan(&user)
		if err != nil {
			return []User{}, err
		}

		users = append(users, user)
	}
	return users, nil
}

//not methods
func AddUser(username, password, description string) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	_, err = db.Query("INSERT INTO users (username, password, description) VALUES (?, ?, ?)", username, password, description)
	return err
}

func QueryUserByID(id int) (User, error) {
	db, err := connectToDB()
	if err != nil {
		return User{}, err
	}

	var user User
	err = db.QueryRow("SELECT id, username, password, description FROM users WHERE id = ?", id).Scan(&user.ID, &user.Username, &user.Password, &user.Description)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func QueryUserByUsername(username string) (User, error) {
	db, err := connectToDB()
	if err != nil {
		return User{}, err
	}

	var user User
	err = db.QueryRow("SELECT id, username, password, description FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password, &user.Description)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func QueryUsersBySubstring(usernameSubstring string) ([]User, error) {
	db, err := connectToDB()
	if err != nil {
		return []User{}, err
	}

	rows, err := db.Query("SELECT id, username, password, description FROM users WHERE username LIKE %?%", usernameSubstring)
	if err != nil {
		return []User{}, err
	}

	var users []User
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.Username, &user.Password, &user.Description)
		if err != nil {
			return []User{}, err
		}
		users = append(users, user)
	}

	return users, nil
}
