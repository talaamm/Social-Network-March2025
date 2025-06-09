package middlewars

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"social-network/internal/config"
	"social-network/internal/models"

	uuid "github.com/satori/go.uuid"
)

func GetUserIDFromSession(w http.ResponseWriter, r *http.Request) int {
	usr := GetUserbySession(w, r)
	return usr.ID
}

func CreateNewSession(w http.ResponseWriter, u models.User) error {
	db := config.GetDB()

	sessionID := uuid.NewV4() // generate a new uuid
	_, err := db.Exec("INSERT INTO sessions (sessionUUID, username , userID) VALUES (?, ?, ?)", sessionID, u.Nickname, u.ID)
	if err != nil {
		log.Println(err, "cannot create session")
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID.String(),
		MaxAge:   24 * 60 * 60,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,                // ✅ Keep false for localhost, change to true in production (HTTPS)
		SameSite: http.SameSiteLaxMode, // ✅ Fix for localhost, prevents cross-origin issues
	})

	fmt.Println("SESSION CREATED FOR USER: ", u)
	return nil
}

func DeleteSession(id int) {
	db := config.GetDB()

	query := "DELETE FROM sessions WHERE userID = ?"
	_, err := db.Exec(query, id)
	if err != nil {
		log.Println(err, "cannot delete past sessions")
	}
}

func GetUserBy_username(u string) models.User {
	db := config.GetDB()
	query := "SELECT * FROM users WHERE nickname = ?"
	var theUser models.User
	err := db.QueryRow(query, u).Scan(
		&theUser.ID, &theUser.Nickname, &theUser.Email, &theUser.Password,
		&theUser.Age, &theUser.Gender, &theUser.FirstName, &theUser.LastName, &theUser.Birthdate,
		&theUser.IsPrivate)
	if err != nil {
		log.Println(err, "error in querying usr")
	}
	return theUser
}

func GetUserbySession(w http.ResponseWriter, r *http.Request) models.User {
	db := config.GetDB()
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Valid() != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Println("No session cookie found in browser", err)
		return models.User{}
	}
	var username string
	err = db.QueryRow("SELECT username FROM sessions WHERE sessionUUID=?;", cookie.Value).Scan(&username) //&id, &sessionID, &userID, &username
	if err != nil {
		log.Println(err, "error retreiving session by sess id")
	}
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Println("No session cookie found in db")
		return models.User{}
	}
	return GetUserBy_username(username)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	us := GetUserbySession(w, r)
	DeleteSession(us.ID)
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1, // 🔥 This deletes the cookie immediately
	})
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}
