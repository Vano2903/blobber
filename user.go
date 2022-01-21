package main

import (
	"fmt"
	"sort"
)

type User struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Description    string `json:"description"`
	LikesCount     int    `json:"likes"`
	FollowersCount int    `json:"followers"`
	FollowingCount int    `json:"following"`
	Follows        bool   `json:"follows"`
}

func (u User) ModifyBlob(id int, content string) error {
	Blob, err := QueryBlobByID(id, u.ID)
	if err != nil {
		return nil
	}

	return Blob.Modify(content)
}

func (u User) DeleteBlob(id int) error {
	Blob, err := QueryBlobByID(id, u.ID)
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
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT count(*) FROM likes WHERE ID_user = ? AND ID_blob = ?", u.ID, id).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (u User) GetBlobs(sorted bool) ([]Blob, error) {
	db, err := connectToDB()
	if err != nil {
		return []Blob{}, err
	}
	defer db.Close()

	//to sort by date i could use a query but in this function i used the sort package just to show it
	//to see the sql query that also sort by date is in the GetOverview method
	rows, err := db.Query("SELECT b.*, u.username FROM blobs b join users u on b.ID_user = u.ID WHERE b.ID_user = ?", u.ID)
	if err != nil {
		return []Blob{}, err
	}
	var blobs []Blob
	for rows.Next() {
		var blob Blob
		err = rows.Scan(&blob.ID, &blob.UserID, &blob.Content, &blob.AddedDate, &blob.Username)
		if err != nil {
			return []Blob{}, err
		}
		blob.HasLiked(u.ID)
		blob.CountLikes()
		blobs = append(blobs, blob)
	}

	//sort blobs by added date
	if sorted {
		//basically sort.Slice take a slice as argument and a function that takes 2 integers and returns a boolean
		//the function returns true if the first integer is smaller than the second integer
		sort.Slice(blobs, func(i, j int) bool {
			return blobs[i].AddedDate.After(blobs[j].AddedDate)
		})
	}

	return blobs, nil
}

//user relations related
func (u User) ModifyDescription(description string) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Query("UPDATE users SET description = ? WHERE ID = ?", description, u.ID)
	return err
}

func (u User) Delete() error {
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Query("DELETE FROM users WHERE ID = ?", u.ID)
	return err
}

func (u User) GetOverview() ([]Blob, error) {
	db, err := connectToDB()
	if err != nil {
		return []Blob{}, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT b.*, u.username FROM follows join blobs b on ID_user_followed = b.ID_user join users u on ID_user_followed = u.ID WHERE ID_user_follower = ? ORDER BY b.added_date DESC", u.ID)
	if err != nil {
		return []Blob{}, err
	}
	var blobs []Blob
	for rows.Next() {
		var blob Blob
		err = rows.Scan(&blob.ID, &blob.UserID, &blob.Content, &blob.AddedDate, &blob.Username)
		if err != nil {
			return []Blob{}, err
		}
		blob.CountLikes()
		blob.HasLiked(u.ID)
		blobs = append(blobs, blob)
	}
	return blobs, nil
}

func (u User) Follow(id int) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Query("INSERT INTO follows (ID_user_follower, ID_user_followed) VALUES (?, ?)", u.ID, id)
	return err
}

func (u User) Unfollow(id int) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Query("DELETE FROM follows WHERE ID_user_follower = ? AND ID_user_followed = ?", u.ID, id)
	return err
}

func (u User) GetFollowers() ([]User, error) {
	db, err := connectToDB()
	if err != nil {
		return []User{}, err
	}
	defer db.Close()

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

		user.Info(u.ID)
		user.Password = "-hidden-"
		users = append(users, user)
	}
	return users, nil
}

func (u User) GetFollowings() ([]User, error) {
	db, err := connectToDB()
	if err != nil {
		return []User{}, err
	}
	defer db.Close()

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

		user.Info(u.ID)
		user.Password = "-hidden-"
		users = append(users, user)
	}
	return users, nil
}

func (u *User) Info(requesterID int) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.QueryRow("SELECT COUNT(ID_user) FROM likes WHERE ID_user=?", u.ID).Scan(&u.LikesCount); err != nil {
		return err
	}

	if err := db.QueryRow("SELECT COUNT(ID_user_followed) FROM follows WHERE ID_user_followed=?", u.ID).Scan(&u.FollowersCount); err != nil {
		return err
	}

	if err := db.QueryRow("SELECT COUNT(ID_user_follower) FROM follows WHERE ID_user_follower=?", u.ID).Scan(&u.FollowingCount); err != nil {
		return err
	}

	//check if the requester is following the user
	// var following bool
	return db.QueryRow("SELECT COUNT(ID_user_followed) FROM follows WHERE ID_user_followed=? AND ID_user_follower=?", u.ID, requesterID).Scan(&u.Follows)
	// if err != nil {
	// 	return err
	// }
}

//not methods
func AddUser(username, password, description string) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = QueryUserByUsername(username, 0)
	if err == nil {
		return fmt.Errorf("user already exists")
	}

	_, err = db.Query("INSERT INTO users (username, password, description) VALUES (?, ?, ?)", username, password, description)
	return err
}

func QueryUserByID(id int, requesterID int) (User, error) {
	db, err := connectToDB()
	if err != nil {
		return User{}, err
	}
	defer db.Close()

	var user User
	err = db.QueryRow("SELECT id, username, password, description FROM users WHERE id = ?", id).Scan(&user.ID, &user.Username, &user.Password, &user.Description)
	if err != nil {
		return User{}, err
	}
	user.Info(requesterID)
	return user, nil
}

func QueryUserByUsername(username string, requesterID int) (User, error) {
	db, err := connectToDB()
	if err != nil {
		return User{}, err
	}
	defer db.Close()

	var user User
	err = db.QueryRow("SELECT id, username, password, description FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password, &user.Description)
	if err != nil {
		return User{}, err
	}

	user.Info(requesterID)
	return user, nil
}

func QueryUsersBySubstring(usernameSubstring string, requesterID int) ([]User, error) {
	db, err := connectToDB()
	if err != nil {
		return []User{}, err
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, username, password, description FROM users WHERE username LIKE CONCAT('%', ?, '%')", usernameSubstring)
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
		if user.ID != requesterID {
			user.Info(requesterID)
			user.Password = "-hidden-"
			users = append(users, user)
		}
	}

	return users, nil
}
