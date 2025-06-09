package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"social-network/internal/config"
	"social-network/internal/middlewars"
	"social-network/internal/models"
	"social-network/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

func CheckSession(w http.ResponseWriter, r *http.Request) {
	u := middlewars.GetUserbySession(w, r)
	empty := models.User{}
	if u == empty {
		http.Error(w, "user unauthorized", http.StatusUnauthorized)
		return
	}
	// w.WriteHeader(http.StatusOK)
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Println("❌ Error decoding request body:", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	fmt.Println("received load: ", user)
	// Hash password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("❌ Error hashing password:", err)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	// Get database connection
	db := config.GetDB()
	userRepo := repositories.NewUserRepository(db)

	// Insert user into database
	err = userRepo.CreateUser(&user)
	if err != nil {
		log.Println("❌ Error inserting user into database:", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}
	authUser := middlewars.GetUserBy_username(user.Nickname)
	err = middlewars.CreateNewSession(w, authUser)
	if err != nil {
		http.Error(w, "error creating new session", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	// json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully & Session created successfuly"})
	json.NewEncoder(w).Encode(map[string]any{
		"message":  "User registered successfully & Session created successfuly",
		"user_id":  authUser.ID,
		"nickname": authUser.Nickname,
	})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	fmt.Println("\nReceived load: ", user)
	db := config.GetDB()
	userRepo := repositories.NewUserRepository(db)

	storedUser, storedPassword, err := userRepo.GetUserByEmailOrNickname(user.Email)
	if err != nil || storedUser == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(user.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	middlewars.DeleteSession(storedUser.ID)
	// Set session
	err = middlewars.CreateNewSession(w, *storedUser)
	if err != nil {
		http.Error(w, "error creating new session", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"message":  "Login successful",
		"user_id":  storedUser.ID,
		"nickname": storedUser.Nickname,
	})
}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
	middlewars.Logout(w, r)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func GetUserData(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	viewingUserID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil || viewingUserID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	db := config.GetDB()
	repo := repositories.NewUserRepository(db)
	userData, err := repo.GetUserDataById(viewingUserID)
	if err != nil {
		log.Println("Error retrieving userData:", err)
		http.Error(w, "Failed to retrieve userData", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(userData)
}

func CurrentUser(w http.ResponseWriter, r *http.Request) {
	u := middlewars.GetUserbySession(w, r)
	empty := models.User{}
	if u == empty {
		http.Error(w, "user unauthorized", http.StatusUnauthorized)
		return
	}
	// w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(u)
}

func UpdatePrivacy(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()

	// Parse the JSON request body
	var req models.User
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("error decoding body received load\n:", r.Body)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Fetch the current privacy setting from the database
	currentPrivacy := user.IsPrivate

	// If the new value is the same as the current one, no update is needed
	if currentPrivacy == req.IsPrivate {
		w.WriteHeader(http.StatusNotModified) // 304 Not Modified
		w.Write([]byte(`{"message": "No changes needed"}`))
		return
	}

	// Update the privacy setting in the database
	_, err = db.Exec("UPDATE users SET is_private = ? WHERE id = ?", req.IsPrivate, user.ID)
	if err != nil {
		http.Error(w, "Database update failed", http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Privacy setting updated successfully"}`))
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	// Ensure only GET requests are allowed
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed. Use GET.", http.StatusMethodNotAllowed)
		return
	}

	// Get logged-in user
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get database connection
	db := config.GetDB()
	userRepo := repositories.NewUserRepository(db)

	// Fetch users not being followed
	users, err := userRepo.GetUsersNotFollowed(user.ID)
	if err != nil {
		log.Println("❌ Error fetching users:", err)
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	// Send response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
