package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"social-network/internal/config"
	"social-network/internal/middlewars"
	"social-network/internal/models"
	"social-network/internal/websocket"
)

func FollowUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()

	// Parse request body
	var UsertoFollow models.User
	if err := json.NewDecoder(r.Body).Decode(&UsertoFollow); err != nil {
		log.Println("Invalid request data", err)
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	if UsertoFollow.ID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	fmt.Println("received body", r.Body)
	fmt.Println("received load", UsertoFollow)

	err := db.QueryRow("SELECT is_private FROM users WHERE id = ?", UsertoFollow.ID).Scan(&UsertoFollow.IsPrivate)
	if err == sql.ErrNoRows {
		fmt.Println("user doesnt exist", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Determine follow status based on privacy setting
	status := "pending"
	if !UsertoFollow.IsPrivate {
		status = "accepted"
	}

	// Insert follow request
	_, err = db.Exec("INSERT INTO followers (follower_id, following_id, status) VALUES (?, ?, ?)", user.ID, UsertoFollow.ID, status)
	if err != nil {
		http.Error(w, "Failed to follow user", http.StatusInternalServerError)
		return
	}

	// Insert notification for the followed user
	notificationMsg := user.Nickname + " has requested to follow you."
	if !UsertoFollow.IsPrivate {
		notificationMsg = user.Nickname + " is now following you."
	}

	// InsertoNotif(w, UsertoFollow.ID, "follow", notificationMsg)

	websocket.SendNotification(UsertoFollow.ID, "follow", notificationMsg)

	// Send response
	response := map[string]string{"status": status, "message": "Follow request sent successfully"}
	json.NewEncoder(w).Encode(response)
}

func UnfollowUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()

	// Parse request body
	var requestData models.User
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		log.Println("Invalid request data:", err)
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	if requestData.ID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check the follow status before deleting
	var followStatus string
	err := db.QueryRow("SELECT status FROM followers WHERE follower_id = ? AND following_id = ?", user.ID, requestData.ID).Scan(&followStatus)
	if err == sql.ErrNoRows {
		http.Error(w, "No follow request found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Remove follow request or unfollow the user
	result, err := db.Exec("DELETE FROM followers WHERE follower_id = ? AND following_id = ?", user.ID, requestData.ID)
	if err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Failed to unfollow user", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "No follow request found", http.StatusNotFound)
		return
	}

	// Determine notification message
	notificationMsg := fmt.Sprintf("%s has unfollowed you.", user.Nickname)
	if followStatus == "pending" {
		notificationMsg = fmt.Sprintf("%s has canceled their follow request.", user.Nickname)
	}

	// Insert notification for the unfollowed user
	// InsertoNotif(w, requestData.ID, "unfollow", notificationMsg)

	websocket.SendNotification(requestData.ID, "unfollow", notificationMsg)

	response := map[string]string{"status": "unfollowed", "message": "Follow request removed"}
	json.NewEncoder(w).Encode(response)
}

func GetFollowers(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()

	rows, err := db.Query(`
        SELECT u.id, u.nickname, u.first_name, u.last_name 
        FROM followers f 
        JOIN users u ON f.follower_id = u.id 
        WHERE f.following_id = ? AND f.status = 'accepted'
    `, user.ID)
	if err != nil {
		http.Error(w, "Failed to retrieve followers", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var followers []models.User
	for rows.Next() {
		var follower models.User
		err := rows.Scan(&follower.ID, &follower.Nickname, &follower.FirstName, &follower.LastName)
		if err != nil {
			log.Println("Error scanning follower:", err)
			continue
		}
		followers = append(followers, follower)
	}

	json.NewEncoder(w).Encode(followers)
}

// GetFollowing - Retrieve users the logged-in user follows
func GetFollowing(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()

	rows, err := db.Query(`
        SELECT u.id, u.nickname, u.first_name, u.last_name, f.status 
        FROM followers f 
        JOIN users u ON f.following_id = u.id 
        WHERE f.follower_id = ?
    `, user.ID)
	if err != nil {
		http.Error(w, "Failed to retrieve following list", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var following []struct {
		models.User
		Status string `json:"status"`
	}
	for rows.Next() {
		var follow struct {
			models.User
			Status string `json:"status"`
		}
		err := rows.Scan(&follow.ID, &follow.Nickname, &follow.FirstName, &follow.LastName, &follow.Status)
		if err != nil {
			log.Println("Error scanning following:", err)
			continue
		}
		following = append(following, follow)
	}

	json.NewEncoder(w).Encode(following)
}

func GetFollowCounts(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()

	var followerCount, followingCount int

	// Get number of followers
	err := db.QueryRow("SELECT COUNT(*) FROM followers WHERE following_id = ? AND status = 'accepted'", user.ID).Scan(&followerCount)
	if err != nil {
		log.Println("Error fetching followers count:", err)
		followerCount = 0
	}

	// Get number of people the user is following
	err = db.QueryRow("SELECT COUNT(*) FROM followers WHERE follower_id = ?", user.ID).Scan(&followingCount)
	if err != nil {
		log.Println("Error fetching following count:", err)
		followingCount = 0
	}

	response := map[string]int{
		"followers": followerCount,
		"following": followingCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetFollowStatus(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()

	targetUserID := r.URL.Query().Get("user_id")
	if targetUserID == "" {
		http.Error(w, "Missing user_id parameter", http.StatusBadRequest)
		return
	}

	var status string
	err := db.QueryRow(
		"SELECT status FROM followers WHERE follower_id = ? AND following_id = ?",
		user.ID, targetUserID,
	).Scan(&status)

	if err == sql.ErrNoRows {
		status = "not-following" // Default if no record exists
	} else if err != nil {
		log.Println("Error checking follow status:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	response := map[string]string{"status": status}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateFollowRequest(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var requestData struct {
		RequestID  int    `json:"requestId"`
		Status     string `json:"status"`
		FollowerID int    `json:"followerId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		log.Println("❌ Invalid input:", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	db := config.GetDB()

	if requestData.Status == "accepted" {
		// ✅ Update status to 'accepted'
		_, err := db.Exec(`UPDATE followers SET status = 'accepted' WHERE id = ?`, requestData.RequestID)
		if err != nil {
			log.Println("❌ Failed to update follow request:", err)
			http.Error(w, "Failed to update request", http.StatusInternalServerError)
			return
		}
	} else if requestData.Status == "rejected" {
		// ✅ Delete request if rejected
		_, err := db.Exec(`DELETE FROM followers WHERE id = ?`, requestData.RequestID)
		if err != nil {
			log.Println("❌ Failed to delete follow request:", err)
			http.Error(w, "Failed to delete request", http.StatusInternalServerError)
			return
		}
	}

	// ✅ Send notification to the follower
	var notificationMessage string
	if requestData.Status == "accepted" {
		notificationMessage = fmt.Sprintf("%s accepted your follow request!", user.Nickname)
	} else {
		notificationMessage = fmt.Sprintf("%s rejected your follow request.", user.Nickname)
	}
	websocket.SendNotification(requestData.FollowerID, "follow", notificationMessage)

	log.Printf("📢 Follow request %d marked as %s", requestData.RequestID, requestData.Status)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Follow request updated successfully"})
}

func GetFollowRequests(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()
	var requests []struct {
		ID               int    `json:"id"`
		FollowerID       int    `json:"follower_id"`
		FollowerNickname string `json:"follower_nickname"`
	}

	query := `
        SELECT f.id, f.follower_id, u.nickname AS follower_nickname
        FROM followers f
        JOIN users u ON f.follower_id = u.id
        WHERE f.following_id = ? AND f.status = 'pending'
    `

	rows, err := db.Query(query, user.ID)
	if err != nil {
		log.Println("❌ Error fetching follow requests:", err)
		http.Error(w, "Failed to fetch follow requests", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var request struct {
			ID               int    `json:"id"`
			FollowerID       int    `json:"follower_id"`
			FollowerNickname string `json:"follower_nickname"`
		}
		if err := rows.Scan(&request.ID, &request.FollowerID, &request.FollowerNickname); err != nil {
			log.Println("❌ Error scanning request:", err)
			continue
		}
		requests = append(requests, request)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

func GetSelectedUsersHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	db := config.GetDB()
	rows, err := db.Query("SELECT user_id FROM posts_visibility WHERE post_creator = ?", user.ID)
	if err != nil {
		http.Error(w, "Error fetching selected users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var selectedUsers []models.PostVisibility
	for rows.Next() {
		var entry models.PostVisibility
		if err := rows.Scan(&entry.UserID); err != nil {
			continue
		}
		selectedUsers = append(selectedUsers, entry)
	}

	json.NewEncoder(w).Encode(selectedUsers)
}

func UpdateSelectedUsersHandler(w http.ResponseWriter, r *http.Request) {
	user := middlewars.GetUserbySession(w, r)
	if user.ID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// ✅ Expect an array of user IDs from frontend
	var requestData struct {
		UserIDs []int `json:"user_ids"` // ✅ Accepts multiple user IDs
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	db := config.GetDB()

	// ✅ Remove existing selected users for the current user
	_, err := db.Exec("DELETE FROM posts_visibility WHERE post_creator = ?", user.ID)
	if err != nil {
		http.Error(w, "Error clearing existing selected users", http.StatusInternalServerError)
		return
	}

	// ✅ Insert new selected users
	for _, userID := range requestData.UserIDs {
		_, err := db.Exec("INSERT INTO posts_visibility (post_creator, user_id) VALUES (?, ?)", user.ID, userID)
		if err != nil {
			http.Error(w, "Error adding selected users", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Updated successfully"})
}
