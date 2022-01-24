package main

type Endpoint string

const (
	//generics
	login      Endpoint = "/login"
	register   Endpoint = "/register"
	home       Endpoint = "/"
	overview   Endpoint = "/overview"
	searchPage Endpoint = "/search"

	//users
	getUser      Endpoint = "/users/{id}"
	getUserBlobs Endpoint = "/users/{id}/blobs"
	followUser   Endpoint = "/users/{id}/follow"
	unfollowUser Endpoint = "/users/{id}/unfollow"
	searchUsers  Endpoint = "/users/search/{query}"
	modifyUser   Endpoint = "/users/modify"
	deleteUser   Endpoint = "/users/delete"
	//TODO
	getUserPage Endpoint = "/users/page/{id}"

	//blobs
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
