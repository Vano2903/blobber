package main

import (
	"fmt"
	"strings"
	"time"
)

type Blob struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Username    string    `json:"username"`
	Content     string    `json:"content"`
	AddedDate   time.Time `json:"added_date"`
	LikesCounts int       `json:"likes"`
	Liked       bool      `json:"liked"`
	IsOwner     bool      `json:"is_owner"`
}

func (b *Blob) Info(requesterID int) {
	db, err := connectToDB()
	if err != nil {
		return
	}
	defer db.Close()

	var count int
	db.QueryRow("SELECT COUNT(*) FROM likes WHERE ID_blob = ? AND ID_user = ?", b.ID, requesterID).Scan(&count)
	fmt.Println("count:", count)
	if count > 0 {
		fmt.Println("liked")
		b.Liked = true
	}

	db.QueryRow("SELECT COUNT(*) FROM blobs WHERE ID = ? AND ID_user = ?", b.ID, requesterID).Scan(&count)
	if count > 0 {
		b.IsOwner = true
	}

	db.QueryRow("SELECT COUNT(*) FROM likes WHERE ID_blob = ?", b.ID).Scan(&b.LikesCounts)
}

func (b Blob) Like(LikerID int) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Query("INSERT INTO likes (ID_user, ID_blob) VALUES (?, ?)", LikerID, b.ID)
	return err
}

func (b Blob) Unlike(LikerID int) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Query("DELETE FROM likes WHERE ID_user = ? AND ID_blob = ?", LikerID, b.ID)
	return err
}

func (b Blob) ToggleLike(LikerID int) error {
	user, err := QueryUserByID(LikerID, 0)
	if err != nil {
		return err
	}
	liked, err := user.HasLiked(b.ID)
	if err != nil {
		return err
	}
	if liked {
		return b.Unlike(LikerID)
	}
	return b.Like(LikerID)
}

func (b Blob) Delete() error {
	db, err := connectToDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Query("DELETE FROM blobs WHERE ID = ?", b.ID)
	return err
}

func (b *Blob) Modify(content string) error {
	content = strings.Trim(content, " ")
	if content == "" {
		return fmt.Errorf("bad request: content can't be empty")
	}

	if b.Content == content {
		return fmt.Errorf("bad request: nothing to change")
	}

	db, err := connectToDB()
	if err != nil {
		return fmt.Errorf("internal server error: %v", err)
	}
	defer db.Close()

	_, err = db.Query("UPDATE blobs SET content = ? WHERE ID = ?", content, b.ID)
	if err != nil {
		return fmt.Errorf("internal server error: %v", err)
	}
	return nil
}

func AddBlob(userID int, content string) error {
	content = strings.Trim(content, " ")
	if content == "" {
		return fmt.Errorf("bad request: content can't be empty")
	}

	db, err := connectToDB()
	if err != nil {
		return fmt.Errorf("internal server error: %v", err)
	}
	defer db.Close()

	_, err = db.Query("INSERT INTO blobs (ID_user, content) VALUES (?, ?)", userID, content)
	if err != nil {
		return fmt.Errorf("internal server error: %v", err)
	}
	return nil
}

func QueryBlobByID(id, requesterID int) (Blob, error) {
	db, err := connectToDB()
	if err != nil {
		return Blob{}, err
	}
	defer db.Close()

	var blob Blob
	// var date []byte
	db.QueryRow("SELECT b.*, u.username FROM blobs b join users u on b.ID_user = u.ID WHERE b.ID = ?", id).Scan(&blob.ID, &blob.UserID, &blob.Content, &blob.AddedDate, &blob.Username)
	if blob.ID == 0 {
		return Blob{}, fmt.Errorf("Blob with id %d not found", id)
	}

	blob.Info(requesterID)
	return blob, nil
}
