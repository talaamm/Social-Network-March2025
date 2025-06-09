package main

import (
	"log"
	"net/http"

	"social-network/internal/config"
	"social-network/internal/handlers"
	"social-network/internal/repositories"
	"social-network/internal/websocket"

	han "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	db := config.InitDB()
	defer config.CloseDB()

	userRepo := repositories.NewUserRepository(db)
	groupRepo := repositories.NewGroupRepository(db)
	chatRepo := repositories.NewChatRepository(db)

	handlers.InitHandlers(userRepo, groupRepo, chatRepo)

	r := mux.NewRouter()

	r.HandleFunc("/", handlers.CheckSession).Methods("GET") // to check session validation
	r.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	r.HandleFunc("/login", handlers.LoginUser).Methods("POST")
	r.HandleFunc("/logout", handlers.LogoutUser).Methods("POST")

	r.HandleFunc("/api/posts", handlers.CreatePostHandler).Methods("POST")
	r.HandleFunc("/all-posts", handlers.GetAllPostsHandler).Methods("GET")

	r.HandleFunc("/api/comments", handlers.GetCommentsForPostHandler).Methods("GET")
	r.HandleFunc("/api/comments", handlers.CreateCommentHandler).Methods("POST")

	r.HandleFunc("/api/like", handlers.LikePost).Methods("POST")

	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	r.PathPrefix("/group_uploads/").Handler(http.StripPrefix("/group_uploads/", http.FileServer(http.Dir("group_uploads"))))

	r.HandleFunc("/api/user-posts", handlers.GetUserPostsHandler).Methods("GET")
	r.HandleFunc("/api/user-data", handlers.GetUserData).Methods("GET")
	r.HandleFunc("/api/myself", handlers.CurrentUser).Methods("GET")
	r.HandleFunc("/api/private", handlers.UpdatePrivacy).Methods("POST")
	r.HandleFunc("/api/discover-people", handlers.GetAllUsers).Methods("GET")
	r.HandleFunc("/api/follow", handlers.FollowUser).Methods("POST")
	r.HandleFunc("/api/unfollow", handlers.UnfollowUser).Methods("POST")
	r.HandleFunc("/api/followers", handlers.GetFollowers).Methods("GET")
	r.HandleFunc("/api/following", handlers.GetFollowing).Methods("GET")
	r.HandleFunc("/api/follow-counts", handlers.GetFollowCounts).Methods("GET")
	r.HandleFunc("/api/follow-status", handlers.GetFollowStatus).Methods("GET")

	r.HandleFunc("/api/notifications", handlers.GetNotificationsHandler).Methods("GET")
	r.HandleFunc("/ws/notifications", handlers.WebSocketNotificationHandler)
	r.HandleFunc("/api/mark-notification-read", handlers.MarkNotificationsAsReadHandler).Methods("POST")
	r.HandleFunc("/api/clear-notifications", handlers.ClearNotifications).Methods("POST")

	hub := handlers.NewHub()
	go hub.Run()
	setupWebSocketRoutes(r, hub)

	r.HandleFunc("/api/chat/recent", handlers.GetRecentChats).Methods("GET")
	r.HandleFunc("/api/chat/users", handlers.GetAvailableChatUsers).Methods("GET")
	r.HandleFunc("/api/chat/history", handlers.GetChatHistoryHandler).Methods("GET")

	r.HandleFunc("/api/follow-requests", handlers.GetFollowRequests).Methods("GET")
	r.HandleFunc("/api/update-follow-request", handlers.UpdateFollowRequest).Methods("POST")

	r.HandleFunc("/api/groups", handlers.GetGroupDetailsHandler).Methods("GET")
	r.HandleFunc("/api/groups/posts", handlers.CreateGroupPostHandler).Methods("POST")
	r.HandleFunc("/api/groups/posts", handlers.GetGroupPostsHandler).Methods("GET")
	r.HandleFunc("/api/groups/comments", handlers.GetGroupPostCommentsHandler).Methods("GET")
	r.HandleFunc("/api/groups/comments", handlers.AddGroupPostCommentHandler).Methods("POST")
	r.HandleFunc("/api/groups/like", handlers.LikeGroupPostHandler).Methods("POST")
	r.HandleFunc("/api/groups/leave", handlers.LeaveGroupHandler).Methods("POST")
	r.HandleFunc("/api/groups-created-by-me", handlers.GetUserCreatedGroupsHandler).Methods("GET")
	r.HandleFunc("/api/my-groups", handlers.GetUserGroupsHandler).Methods("GET")
	r.HandleFunc("/api/discover-groups", handlers.GetNonMemberGroupsHandler).Methods("GET")
	r.HandleFunc("/api/groups/join", handlers.RequestToJoinGroupHandler).Methods("POST")
	r.HandleFunc("/api/groups/pending-requests", handlers.GetPendingGroupRequestsHandler).Methods("GET")
	r.HandleFunc("/api/groups", handlers.CreateGroupHandler).Methods("POST")
	r.HandleFunc("/api/groups/members", handlers.GetGroupMembersHandler).Methods("GET")
	r.HandleFunc("/api/groups/invite", handlers.InviteUserToGroupHandler).Methods("POST")
	r.HandleFunc("/api/followers-to-invite", handlers.GetFollowersToInviteHandler).Methods("GET")
	r.HandleFunc("/api/groups-to-chat", handlers.GetUserMemberGroupsHandler).Methods("GET")

	r.HandleFunc("/api/groups/approve", handlers.ApproveMembershipHandler).Methods("POST")
	r.HandleFunc("/api/groups/reject", handlers.RejectMembershipHandler).Methods("POST")

	r.HandleFunc("/api/groups/invitations", handlers.GetGroupInvitationsHandler).Methods("GET")          // Fetch invitations
	r.HandleFunc("/api/groups/accept-invitation", handlers.AcceptGroupInvitationHandler).Methods("POST") // Accept invitation
	r.HandleFunc("/api/groups/reject-invitation", handlers.RejectGroupInvitationHandler).Methods("POST") // Reject invitation

	r.HandleFunc("/api/groups/events", handlers.CreateGroupEventHandler).Methods("POST")
	r.HandleFunc("/api/groups/events/rsvp", handlers.RSVPEventHandler).Methods("POST")
	r.HandleFunc("/api/groups/events/rsvp/count", handlers.GetRSVPCountHandler).Methods("GET")
	r.HandleFunc("/api/groups/events", handlers.GetGroupEventsHandler).Methods("GET")

	r.HandleFunc("/api/group/chat/history", handlers.GetGroupChatHistoryHandler).Methods("GET")
	r.HandleFunc("/api/selected-users", handlers.GetSelectedUsersHandler).Methods("GET")
	r.HandleFunc("/api/update-selected-users", handlers.UpdateSelectedUsersHandler).Methods("POST")

	groupHub := websocket.NewGroupHub()
	go groupHub.Run() // ✅ Run the WebSocket hub in a goroutine

	setupWebSocketRoutesG(r, groupHub)

	corsOptions := han.CORS(
		han.AllowedOrigins([]string{"http://localhost:5173"}), // Allow Vue.js frontend
		han.AllowedMethods([]string{"GET", "POST", "OPTIONS", "PUT", "DELETE"}),
		han.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		han.AllowCredentials(), // ✅ This is MANDATORY for cookies/sessions
	)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./frontend/src/components"))))

	log.Println("✅ Server running on :8080")
	http.ListenAndServe(":8080", corsOptions(r))
}

func setupWebSocketRoutes(r *mux.Router, hub *handlers.Hub) {
	r.HandleFunc("/ws/chat", func(w http.ResponseWriter, r *http.Request) {
		handlers.ServeWs(hub, w, r)
	})
}

func setupWebSocketRoutesG(r *mux.Router, groupHub *websocket.GroupHub) {
	r.HandleFunc("/ws/groupchat", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeGroupChatWs(groupHub, w, r)
	})
}
