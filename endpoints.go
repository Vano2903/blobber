package main

type Endpoint string

const (
	//generics
	login    Endpoint = "/login"
	register Endpoint = "/register"
	home     Endpoint = "/"
	overview Endpoint = "/overview"

	//users
	getUser      Endpoint = "/users/{id}"
	followUser   Endpoint = "/users/{id}/follow"
	unfollowUser Endpoint = "/users/{id}/unfollow"
	searchUsers  Endpoint = "/users/search/{query}"
	//TODO
	modifyUser Endpoint = "/users/{id}/modify"
	deleteUser Endpoint = "/users/{id}/delete"

	//Blobs
	addBlob    Endpoint = "/blob/add"
	getBlob    Endpoint = "/blob/{id}"
	modifyBlob Endpoint = "/blob/{id}/modify"
	deleteBlob Endpoint = "/blob/{id}/delete"

	addLikeBlob    Endpoint = "/blob/{id}/like/add"
	removeLikeBlob Endpoint = "/blob/{id}/like/remove"
	toggleLikeBlob Endpoint = "/blob/{id}/like/toggle"
)

func (e Endpoint) String() string {
	return string(e)
}
