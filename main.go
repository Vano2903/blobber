package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Post struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Content  string `json:"content,omitempty"`
}

//* middlewares
//functions in golang can take functions as parameters
//the middleware function is executed before the handler function, basically it's a wrapper
//this function returns a handlerFunc which is a method of http function that can be used to serve the request
//the return is a handlerFunc which take as arguments the response and the request, check if the jwt exists
//and, if so, it executes the actual handler (which is given as argument and is called "next")
func JWTAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log.Println(r.Method, r.RequestURI)
		//check if the cookie "JWT" exists
		_, err := r.Cookie("JWT")
		if err != nil {
			// OLD: if err is not nil it means that the cookie was not found so we return a 401 unauthorized
			// returnError(w, http.StatusUnauthorized, "missing 'JWT' cookie")

			//if err is not nil then redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		} else {
			next(w, r)
		}
	})
}

//* generic's handlers
//return the login html page
func loginPage(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	//read the file
	page, err := ioutil.ReadFile("pages/login.html")
	if err != nil {
		//check for possible errors, if so return 503
		returnError(w, http.StatusServiceUnavailable, "error, reason: "+err.Error())
		return
	}

	//set the content type and write the file
	w.Header().Set("Content-Type", "text/html")
	w.Write(page)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	var post Post
	//decode the json given in the post body (r.Body) and parse it into the post struct (as pointer)
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		//check for errors, if so return 400
		returnError(w, http.StatusBadRequest, "Invalid json, "+err.Error())
		return
	}

	//hash the password with sha256
	hashedPassword := sha256.Sum256([]byte(post.Password))
	user, err := QueryUserByUsername(post.Username, 0)
	if err != nil {
		//internal server error
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}
	if user.Password == fmt.Sprintf("%x", hashedPassword) {
		//create a jwt with the info and the expiration time
		token, err := NewJWT(user.Username, user.ID, time.Now().Add(time.Hour*time.Duration(2)).Unix())
		if err != nil {
			returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
			return
		}

		//set the jwt even as a cookie
		cookie := &http.Cookie{
			Name:     "JWT",
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(time.Hour * time.Duration(2)),
			HttpOnly: true,
		}

		user.Password = "-hidden-"
		userJson, _ := json.Marshal(user)

		http.SetCookie(w, cookie)
		//and as a header
		w.Header().Add("Authorization", "Bearer "+token)
		returnSuccessJson(w, http.StatusOK, "Successfully logged in", "user", userJson)
		return
	}
	//return unauthorized
	returnError(w, http.StatusUnauthorized, "Invalid credentials")
}

func registerPage(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	//read the file
	page, err := ioutil.ReadFile("pages/register.html")
	if err != nil {
		//check for possible errors, if so return 503
		returnError(w, http.StatusServiceUnavailable, "error, reason: "+err.Error())
		return
	}

	//set the content type and write the file
	w.Header().Set("Content-Type", "text/html")
	w.Write(page)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	//read from the post body the json data and fill the post struct
	var post Post
	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		//check for errors, if so return 400
		returnError(w, http.StatusBadRequest, "Invalid json, "+err.Error())
		return
	}

	//hash the password with sha256
	hashedPassword := fmt.Sprintf("%x", sha256.Sum256([]byte(post.Password)))
	err = AddUser(post.Username, hashedPassword, "")
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	returnSuccess(w, http.StatusCreated, "successfully registered, you can now login")
}

func homePage(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	user, err := QueryUserByID(jwtContent.UserID, 0)
	if err != nil {
		returnError(w, http.StatusBadRequest, "User not found, error: "+err.Error())
		return
	}

	data := struct {
		Username string
		ID       int
		Bio      string
	}{
		Username: jwtContent.Username,
		ID:       jwtContent.UserID,
		Bio:      user.Description,
	}

	tmpl, err := template.ParseFiles("pages/home.html")
	if err != nil {
		returnError(w, http.StatusServiceUnavailable, "Internal server error: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

func searchPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	page, err := ioutil.ReadFile("pages/search.html")
	if err != nil {
		//check for possible errors, if so return 503
		returnError(w, http.StatusServiceUnavailable, "error, reason: "+err.Error())
		return
	}

	//set the content type and write the file
	w.Header().Set("Content-Type", "text/html")
	w.Write(page)
}

func userPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid id")
		return
	}

	user, err := QueryUserByID(id, jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusNotFound, "User not found")
		return
	}

	followsButton := "Follow"
	if user.Follows {
		followsButton = "Un-Follow"
	}
	if user.ID == jwtContent.UserID {
		followsButton = "remove"
	}
	blobs, _ := user.GetBlobs(false, jwtContent.UserID)

	data := struct {
		Username      string
		ID            int
		FollowsButton string
		Likes         int
		Blobs         int
		Followers     int
		Followings    int
		Description   string
	}{
		Username:      user.Username,
		ID:            user.ID,
		FollowsButton: followsButton,
		Likes:         user.LikesCount,
		Blobs:         len(blobs),
		Followers:     user.FollowersCount,
		Followings:    user.FollowingCount,
		Description:   user.Description,
	}

	tmpl, err := template.ParseFiles("pages/user.html")
	if err != nil {
		returnError(w, http.StatusServiceUnavailable, "Internal server error: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

//* user's handlers
func getUserBlobsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid id")
		return
	}

	user, err := QueryUserByID(id, jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	blobs, err := user.GetBlobs(true, jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	blobsJson, _ := json.Marshal(blobs)
	returnSuccessJson(w, http.StatusOK, "Successfully retrieved blobs", "blobs", blobsJson)
}

func modifyUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	var post Post
	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid json, "+err.Error())
		return
	}

	user, err := QueryUserByID(jwtContent.UserID, 0)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}
	user.ModifyDescription(post.Content)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	returnSuccess(w, http.StatusOK, "content modified successfully")
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	user, err := QueryUserByID(jwtContent.UserID, 0)
	user.Delete()
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	returnSuccess(w, http.StatusOK, "user deleted successfully")
}

func overviewHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	user, err := QueryUserByID(jwtContent.UserID, 0)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	overview, err := user.GetOverview()
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	overviewJson, _ := json.Marshal(overview)
	returnSuccessJson(w, http.StatusOK, "Successfully retrieved overview", "overview", overviewJson)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid user id")
		return
	}

	user, err := QueryUserByID(id, jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusNotFound, "User not found")
		return
	}

	user.Password = "-hidden-"
	userJSON, _ := json.Marshal(user)
	returnSuccessJson(w, http.StatusOK, "Successfully retrieved user", "data", userJSON)
}

func searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	search := mux.Vars(r)["query"]
	fmt.Println(search)
	users, err := QueryUsersBySubstring(search, jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusNotFound, "no users found")
		return
	}

	usersJSON, _ := json.Marshal(users)
	returnSuccessJson(w, http.StatusOK, "Successfully retrieved users", "users", usersJSON)
}

func followUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid user id")
		return
	}

	user, err := QueryUserByID(jwtContent.UserID, 0)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	err = user.Follow(id)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	returnSuccess(w, http.StatusOK, "Successfully followed user")
}

func unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid user id")
		return
	}

	user, err := QueryUserByID(jwtContent.UserID, 0)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	err = user.Unfollow(id)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	returnSuccess(w, http.StatusOK, "Successfully unfollowed user")
}

//* blob's handlers
func getBlobHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid blob id")
		return
	}

	blob, err := QueryBlobByID(id, 0)
	if err != nil {
		returnError(w, http.StatusNotFound, "Blob not found")
		return
	}

	blobJSON, _ := json.Marshal(blob)
	returnSuccessJson(w, http.StatusOK, "Successfully retrieved blob", "blob", blobJSON)
}

func addBlobHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	var post Post
	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	err = AddBlob(jwtContent.UserID, post.Content)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	returnSuccess(w, http.StatusOK, "Successfully added blob")
}

func modifyBlobHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	var post Post
	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid blob id")
		return
	}

	blob, err := QueryBlobByID(id, jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusNotFound, "Blob not found")
		return
	}

	if jwtContent.UserID != blob.UserID {
		returnError(w, http.StatusUnauthorized, "You are not authorized to modify this blob, only the owner can modify it")
		return
	}

	err = blob.Modify(post.Content)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	returnSuccess(w, http.StatusOK, "Successfully modified blob")
}

func deleteBlobHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid blob id")
		return
	}

	blob, err := QueryBlobByID(id, jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusNotFound, "Blob not found")
		return
	}

	if jwtContent.UserID != blob.UserID {
		returnError(w, http.StatusUnauthorized, "You are not authorized to delete this blob, only the owner can delete it")
		return
	}

	err = blob.Delete()
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	returnSuccess(w, http.StatusOK, "Successfully deleted blob")
}

func addLikeBlobHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid blob id")
		return
	}

	blob, err := QueryBlobByID(id, jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusNotFound, "Blob not found")
		return
	}

	err = blob.Like(jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	returnSuccess(w, http.StatusOK, "Successfully liked blob")
}

func removeLikeBlobHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid blob id")
		return
	}

	blob, err := QueryBlobByID(id, jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusNotFound, "Blob not found")
		return
	}

	err = blob.Unlike(jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	returnSuccess(w, http.StatusOK, "Successfully unliked blob")
}

func toggleLikeBlobHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method, r.RequestURI)
	jwtContent, err := checkJWT(w, r)
	if err != nil {
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		returnError(w, http.StatusBadRequest, "Invalid blob id")
		return
	}

	blob, err := QueryBlobByID(id, jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusNotFound, "Blob not found")
		return
	}

	err = blob.ToggleLike(jwtContent.UserID)
	if err != nil {
		returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
		return
	}

	returnSuccess(w, http.StatusOK, "Successfully toggled like blob")
}

func main() {
	r := mux.NewRouter()

	//*generics
	r.HandleFunc("/images/blob", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RequestURI)
		fileBytes, err := ioutil.ReadFile("blob.png")
		if err != nil {
			returnError(w, http.StatusInternalServerError, "Internal server error: "+err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(fileBytes)
	}).Methods("GET")
	//pages
	r.HandleFunc(login.String(), loginPage).Methods("GET")
	r.HandleFunc(register.String(), registerPage).Methods("GET")
	//this just means that the homepage is available only if the client has a jwt cookie
	r.HandleFunc(home.String(), JWTAuthMiddleware(homePage)).Methods("GET")
	r.HandleFunc(searchPage.String(), JWTAuthMiddleware(searchPageHandler)).Methods("GET")

	r.HandleFunc(getUserPage.String(), JWTAuthMiddleware(userPageHandler)).Methods("GET") //JWTAuthMiddleware(getUserPageHandler)

	//api
	r.HandleFunc(login.String(), loginHandler).Methods("POST")
	r.HandleFunc(register.String(), registerHandler).Methods("POST")
	r.HandleFunc(overview.String(), JWTAuthMiddleware(overviewHandler)).Methods("GET")

	//*users (all api)
	r.HandleFunc(getUser.String(), JWTAuthMiddleware(getUserHandler)).Methods("GET")
	r.HandleFunc(getUserBlobs.String(), JWTAuthMiddleware(getUserBlobsHandler)).Methods("GET")
	r.HandleFunc(searchUsers.String(), JWTAuthMiddleware(searchUsersHandler)).Methods("GET")
	r.HandleFunc(followUser.String(), JWTAuthMiddleware(followUserHandler)).Methods("GET")
	r.HandleFunc(unfollowUser.String(), JWTAuthMiddleware(unfollowUserHandler)).Methods("GET")
	r.HandleFunc(modifyUser.String(), JWTAuthMiddleware(modifyUserHandler)).Methods("POST")
	r.HandleFunc(deleteUser.String(), JWTAuthMiddleware(deleteUserHandler)).Methods("GET")

	//*blobs (all pi)
	r.HandleFunc(getBlob.String(), getBlobHandler).Methods("GET")
	r.HandleFunc(addBlob.String(), JWTAuthMiddleware(addBlobHandler)).Methods("POST")
	r.HandleFunc(modifyBlob.String(), JWTAuthMiddleware(modifyBlobHandler)).Methods("POST")
	r.HandleFunc(deleteBlob.String(), JWTAuthMiddleware(deleteBlobHandler)).Methods("GET")
	r.HandleFunc(addLikeBlob.String(), JWTAuthMiddleware(addLikeBlobHandler)).Methods("GET")
	r.HandleFunc(removeLikeBlob.String(), JWTAuthMiddleware(removeLikeBlobHandler)).Methods("GET")
	r.HandleFunc(toggleLikeBlob.String(), JWTAuthMiddleware(toggleLikeBlobHandler)).Methods("GET")

	//*start the server
	log.Fatal(http.ListenAndServe(":8080", r))
}
