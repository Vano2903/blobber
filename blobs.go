package main

import "time"

type Blob struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	AddedDate time.Time `json:"added_date"`
}

func (b Blob) Like(LikerID int) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	_, err = db.Query("INSERT INTO likes (ID_user, ID_blob) VALUES (?, ?)", LikerID, b.ID)
	return err
}

func (b Blob) Unlike(LikerID int) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	_, err = db.Query("DELETE FROM likes WHERE ID_user = ? AND ID_blob = ?", LikerID, b.ID)
	return err
}

func (b Blob) ToggleLike(LikerID int) error {
	user, err := QueryUserByID(LikerID)
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

	_, err = db.Query("DELETE FROM Blobs WHERE ID = ?", b.ID)
	return err
}

func (b *Blob) Modify(content string) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	_, err = db.Query("UPDATE Blobs SET content = ? WHERE ID = ?", content, b.ID)
	return err
}

func AddBlob(userID int, content string) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	_, err = db.Query("INSERT INTO Blobs (ID_user, content) VALUES (?, ?)", userID, content)
	return err
}

func QueryBlobByID(id int) (Blob, error) {
	db, err := connectToDB()
	if err != nil {
		return Blob{}, err
	}

	var Blob Blob
	db.QueryRow("SELECT * FROM Blobs WHERE ID = ?", id).Scan(&Blob.ID, &Blob.UserID, &Blob.Content, &Blob.AddedDate)
	return Blob, nil
}
