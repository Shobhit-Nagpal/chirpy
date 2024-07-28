package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Shobhit-Nagpal/chirpy/internal/database"
	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, req *http.Request) {
	type ReqBody struct {
		Body string `json:"body"`
	}

	type Err struct {
		Error string `json:"error"`
	}

	type Response struct {
		Valid bool `json:"valid"`
	}

	type CleanedResponse struct {
		Id       int    `json:"id"`
		Body     string `json:"body"`
		AuthorId int    `json:"author_id"`
	}

	authorization := req.Header.Get("Authorization")
	if authorization == "" {
		//return with err
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tokenString := strings.Fields(authorization)
	token, err := jwt.ParseWithClaims(tokenString[1], &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil
	})

	if err != nil {
		log.Printf("Error parsing token: %s", err)
		err = respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	bodyFromReq, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	reqBody := ReqBody{}

	err = json.Unmarshal(bodyFromReq, &reqBody)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(reqBody.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedMsg := cleanMessage(reqBody.Body)

	//Insert into db here
	path, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	db, err := database.NewDB(path + "/database.json")
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	userId, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Error getting user id: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	userIdInt, err := strconv.Atoi(userId)

	chirp, err := db.CreateChirp(cleanedMsg, userIdInt)
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	response := CleanedResponse{
		Id:       chirp.Id,
		Body:     chirp.Body,
		AuthorId: chirp.AuthorId,
	}

	err = respondWithJSON(w, http.StatusCreated, response)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	return
}

func cleanMessage(msg string) string {
	words := strings.Split(msg, " ")
	for idx, word := range words {
		switch strings.ToLower(word) {
		case "kerfuffle":
			words[idx] = "****"
		case "sharbert":
			words[idx] = "****"
		case "fornax":
			words[idx] = "****"
		}
	}

	cleanMsg := strings.Join(words, " ")
	return cleanMsg
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, req *http.Request) {

	path, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	db, err := database.NewDB(path + "/database.json")
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	chirps, err := db.GetChirps()
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps")
		return
	}

	err = respondWithJSON(w, http.StatusOK, chirps)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	return
}

func handleGetChirpById(w http.ResponseWriter, req *http.Request) {
	idStr := req.PathValue("chirpId")

	path, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	db, err := database.NewDB(path + "/database.json")
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error converting string to integer for id: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	chirp, err := db.GetChirpById(id)
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		if err.Error() == "Chirp not found" {
			err = respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			err = respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps")
		}
		return
	}

	err = respondWithJSON(w, http.StatusOK, chirp)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	return
}

func handleCreateUser(w http.ResponseWriter, req *http.Request) {
	type Request struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds *int   `json:"expires_in_seconds"`
	}

	body := Request{}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	path, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	db, err := database.NewDB(path + "/database.json")
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	user, err := db.CreateUser(body.Email, body.Password)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		if err.Error() == "User exists" {
			err = respondWithError(w, http.StatusBadRequest, err.Error())
		} else {
			err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		}
		return
	}

	type Response struct {
		Id          int    `json:"id"`
		Email       string `json:"email"`
		IsChirpyRed bool   `json:"is_chirpy_red"`
	}

	resp := Response{
		Email:       user.Email,
		Id:          user.Id,
		IsChirpyRed: user.IsChirpyRed,
	}
	err = respondWithJSON(w, http.StatusCreated, resp)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	return
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, req *http.Request) {
	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	body := Request{}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	path, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	db, err := database.NewDB(path + "/database.json")
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	type Response struct {
		Id           int    `json:"id"`
		Email        string `json:"email"`
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed  bool   `json:"is_chirpy_red"`
	}

	login, user, refreshToken, err := db.LoginUser(body.Email, body.Password)
	if err != nil {
		log.Printf("Error logging in user: %s", err)
		if err.Error() == "User not found" {
			err = respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		}
		return
	}

	if !login {
		err = respondWithError(w, http.StatusUnauthorized, "Unauthorized")
	} else {
		claims := &jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			Subject:   strconv.Itoa(user.Id),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		ss, err := token.SignedString(cfg.jwtSecret)
		if err != nil {
			log.Printf("Error creating jwt: %s", err)
			err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		}
		resp := Response{
			Email:        user.Email,
			Id:           user.Id,
			Token:        ss,
			RefreshToken: refreshToken,
			IsChirpyRed:  user.IsChirpyRed,
		}
		err = respondWithJSON(w, http.StatusOK, resp)
		if err != nil {
			log.Printf("Error encoding to json: %s", err)
			err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		}
	}
	return
}

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, req *http.Request) {
	type Request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	body := Request{}

	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	authorization := req.Header.Get("Authorization")
	if authorization == "" {
		//return with err
		err = respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tokenString := strings.Fields(authorization)
	token, err := jwt.ParseWithClaims(tokenString[1], &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil
	})

	if err != nil {
		log.Printf("Error parsing token: %s", err)
		err = respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userId, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Error getting user id: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	userIdInt, err := strconv.Atoi(userId)
	fmt.Println(userId)
	if err != nil {
		log.Printf("Error converting user id: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	path, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	db, err := database.NewDB(path + "/database.json")
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	user, err := db.UpdateUser(userIdInt, body.Email, body.Password)
	if err != nil {
		log.Printf("Error updating user: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	type Response struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}

	resp := Response{
		Email: user.Email,
		Id:    user.Id,
	}
	err = respondWithJSON(w, http.StatusOK, resp)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	return
}

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, req *http.Request) {
	authorization := req.Header.Get("Authorization")
	if authorization == "" {
		//return with err
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tokenString := strings.Fields(authorization)

	if len(tokenString) < 2 {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	path, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	db, err := database.NewDB(path + "/database.json")
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	fmt.Println(tokenString[1])

	valid, id, err := db.ValidateRefreshToken(tokenString[1])
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	if !valid {
		err = respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	type Response struct {
		Token string `json:"token"`
	}

	claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		Subject:   strconv.Itoa(id),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(cfg.jwtSecret)
	if err != nil {
		log.Printf("Error creating jwt: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	resp := Response{
		Token: ss,
	}
	err = respondWithJSON(w, http.StatusOK, resp)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	return
}

func (cfg *apiConfig) handleRevokeToken(w http.ResponseWriter, req *http.Request) {
	authorization := req.Header.Get("Authorization")
	if authorization == "" {
		//return with err
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tokenString := strings.Fields(authorization)

	if len(tokenString) < 2 {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	path, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	db, err := database.NewDB(path + "/database.json")
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	err = db.RevokeToken(tokenString[1])
	if err != nil {
		log.Printf("Error revoking token: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	type Response struct {
	}

	err = respondWithJSON(w, http.StatusNoContent, Response{})
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	return
}

func (cfg *apiConfig) handleDeleteChirpById(w http.ResponseWriter, req *http.Request) {

	authorization := req.Header.Get("Authorization")
	if authorization == "" {
		//return with err
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tokenString := strings.Fields(authorization)
	token, err := jwt.ParseWithClaims(tokenString[1], &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil
	})

	if err != nil {
		log.Printf("Error parsing token: %s", err)
		err = respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := req.PathValue("chirpId")

	path, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	db, err := database.NewDB(path + "/database.json")
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error converting string to integer for id: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	userId, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("Error getting user id: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	userIdInt, err := strconv.Atoi(userId)

	err = db.DeleteChirp(id, userIdInt)
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		if err.Error() == "Forbidden" {
			err = respondWithError(w, http.StatusForbidden, err.Error())
		} else {
			err = respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps")
		}
		return
	}

	type Response struct {
	}

	err = respondWithJSON(w, http.StatusNoContent, Response{})
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	return
}

func (cfg *apiConfig) handlePolkaWebhook(w http.ResponseWriter, req *http.Request) {

	type UserId struct {
		Id int `json:"user_id"`
	}
	type Request struct {
		Event string `json:"event"`
		Data  UserId `json:"data"`
	}

	authorization := req.Header.Get("Authorization")
	if authorization == "" {
		//return with err
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tokenString := strings.Fields(authorization)

	if tokenString[1] != cfg.polka {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	bodyFromReq, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	reqBody := Request{}

	err = json.Unmarshal(bodyFromReq, &reqBody)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	event := reqBody.Event

	type Response struct {
	}

	if event != "user.upgraded" {
		err = respondWithJSON(w, http.StatusNoContent, Response{})
		return
	}

	userId := reqBody.Data.Id

	path, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	db, err := database.NewDB(path + "/database.json")
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	err = db.UpgradeUser(userId)
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		if err.Error() == "No user found" {
			err = respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			err = respondWithError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	err = respondWithJSON(w, http.StatusNoContent, Response{})
}
