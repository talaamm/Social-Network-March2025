<template>
  <div class="welcome-container">
    <!-- <h1>Welcome to NTLink</h1> -->
    <h1 v-if="!auth.isLoggedIn">Welcome to SOCIAL-NETWORK</h1>
    <div v-if="auth.isLoggedIn" class="notification-container">
      <div id="welcomnot">
        <h1>Welcome to SOCIAL-NETWORK</h1>
        <button class="notification-btn" @click="toggleNotifications">
          🔔 <span v-if="unreadCount > 0" class="badge">{{ unreadCount }}</span>
        </button>
      </div>
      <div v-if="showNotifications" class="notification-dropdown">
        <h3>Notifications</h3>
        <p v-if="notifications.length === 0" class="no-notifications">
          No notifications yet.
        </p>

        <ul v-else>
          <li
            v-for="notification in notifications"
            :key="notification.id"
            :class="{ unread: !notification.is_read }"
            @click="markAsRead(notification.id)"
          >
            <span>{{ notification.message }}</span>
            <!-- <small>{{ formatDate(notification.created_at) }}</small> -->
          </li>
        </ul>

        <button @click="clearNotifications" class="clear-btn">Clear All</button>
      </div>
    </div>

    <!-- {{ !auth.isLoggedIn }} -->

    <nav v-if="!auth.isLoggedIn">
      <router-link to="/login">Login</router-link> |
      <router-link to="/register">Register</router-link>
    </nav>
    <nav v-else @loadstart="homepage"></nav>
    <!-- <button v-if="auth.isLoggedIn" @click="auth.logout">Logout</button> -->
    <router-view class="router-view" />
  </div>
</template>

<script setup>
import { useAuthStore } from "@/authStore"; // ✅ Import the store
import { useRouter } from "vue-router";
import { ref, onMounted, onUnmounted } from "vue";
import axios from "axios";
import config from "@/config";
import { eventBus } from "@/eventBus"; // Import the Event Bus

const router = useRouter();
const auth = useAuthStore();
let heartbeatInterval = null; // ✅ Ping-pong keep-alive

let notifWs = null; // ✅ Rename WebSocket for clarity
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
const notifications = ref([]);
const unreadCount = ref(0);
const showNotifications = ref(false);

const toggleNotifications = () => {
  showNotifications.value = !showNotifications.value;
  if (showNotifications.value == true) {
    fetchNotifications();
  }
};

const fetchNotifications = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/notifications`, {
      withCredentials: true,
    });

    notifications.value = response.data || []; // ✅ Prevents errors if API returns null
    unreadCount.value = notifications.value.filter((n) => !n.is_read).length;
  } catch (error) {
    console.error(
      "❌ Error fetching notifications:",
      error.response ? error.response.data : error
    );
    notifications.value = []; // ✅ Prevents crashes
  }
};

const markAsRead = async (id) => {
  try {
    await axios.post(
      `${config.API_URL}/api/mark-notification-read`,
      { id },
      { withCredentials: true }
    );

    const notification = notifications.value.find((n) => n.id === id);
    if (notification) notification.is_read = true;

    unreadCount.value = notifications.value.filter((n) => !n.is_read).length;
  } catch (error) {
    console.error("Error marking notification as read:", error);
  }
};

const clearNotifications = async () => {
  try {
    await axios.post(
      `${config.API_URL}/api/clear-notifications`,
      {},
      { withCredentials: true }
    );
    notifications.value = [];
    unreadCount.value = 0;
  } catch (error) {
    console.error("Error clearing notifications:", error);
  }
};
let loggedin_id = 0;
async function fetcCurrUserData() {
  try {
    const response = await axios.get(`${config.API_URL}/api/myself`, {
      withCredentials: true,
    });
    console.log(response.data);
    loggedin_id = response.data.id;

    console.log(loggedin_id);
  } catch (error) {
    console.error("Error fetching user data:", error);
  }
}

// ✅ **Real-time Notifications with WebSocket**
const setupWebSocket = () => {
  const wsUrl = `${config.API_URL.replace(
    /^http/,
    "ws"
  )}/ws/notifications?user_id=${loggedin_id}`;

  notifWs = new WebSocket(wsUrl);

  notifWs.onopen = () => {
    console.log("✅ WebSocket connected for notifications.");
    reconnectAttempts = 0; // ✅ Reset reconnect attempts

// ✅ Start sending heartbeat pings every 30 seconds
if (heartbeatInterval) clearInterval(heartbeatInterval);
heartbeatInterval = setInterval(() => {
  if (notifWs.readyState === WebSocket.OPEN) {
    notifWs.send(JSON.stringify({ type: "ping" }));
    console.log("📡 Sent heartbeat ping");
  }
}, 30000);
  };
  

notifWs.onmessage = (event) => {
  try {
    const data = JSON.parse(event.data);

    if (data.type === "post_update") {
      console.log("🔥 Live Post Update:", data);

      // ✅ Broadcast the update to all components that need it
      eventBus.emit("postUpdate", data);
    } else if (data.type === "group_post_update") {
      console.log("🚀 New Group Post:", data);
      eventBus.emit("groupPostUpdate", data);
    } else if (data.type === "new_group_event") {
      console.log("🚀 New Group Event:", data);
      eventBus.emit("groupEvent", data);
    } else {
      // ✅ Handle normal notifications
      notifications.value.unshift(data);
      unreadCount.value += 1;
    }
  } catch (error) {
    console.error("❌ Error parsing WebSocket message:", error);
  }
};


  notifWs.onclose = () => {
    console.log("🔴 WebSocket disconnected. Attempting to reconnect...");
   
    if (reconnectAttempts < maxReconnectAttempts) {
      setTimeout(() => {
        reconnectAttempts++;
        console.log(`🔄 Reconnecting attempt ${reconnectAttempts}...`);
        setupWebSocket();
      }, 5000);
    } else {
      console.error(
        "🚨 Max reconnect attempts reached. WebSocket won't reconnect."
      );
    }
   
    // setTimeout(setupWebSocket, 5000); // Try reconnecting after 5 seconds
  };

  notifWs.onerror = (error) => {
    console.error("❌ WebSocket error:", error);
    console.error("🔴 WebSocket connection to notifications failed. Retrying...");
  };
};

async function waitForUserId() {
  let retries = 10; // ⏳ Max wait time (10 * 500ms = 5 seconds)
  while (loggedin_id <= 0 && retries > 0) {
    console.warn(`⏳ Waiting for User ID... (${10 - retries}/10)`);
    await new Promise((resolve) => setTimeout(resolve, 500)); // Wait 500ms
    retries--;
  }

  if (loggedin_id > 0) {
    console.log(`✅ Starting WebSocket connection for User ID: ${loggedin_id}`);
    fetchNotifications();
    setupWebSocket();
  } else {
    console.error("❌ WebSocket setup failed: User ID not available.");
  }
}

// Call function after fetching user data
// await fetcCurrUserData();


onMounted(async () => {
  if (auth.isLoggedIn) {
    await fetcCurrUserData();
    waitForUserId();
    router.push("/home");
  }
  
  
  // if (loggedin_id > 0) {
  //   console.log(`✅ Starting WebSocket connection for User ID: ${loggedin_id}`);
  //   fetchNotifications();
  //   setupWebSocket();
  // } else {
  //   console.error("❌ WebSocket setup delayed: User ID not available.");
  // }

  if (!auth.isLoggedIn) {
    console.log(`❌ User logged out. 🔴Notification websocket closed `);
  if (notifWs) notifWs.close();
}

});

</script>

<style scoped>
/* Welcome Page Styles */
.welcome-container {
  display: flex;
  flex-direction: column; /* Arrange items vertically */
  align-items: center; /* Center horizontally */
  text-align: center; /* Center text within elements */
}

.welcome-container h1 {
  color: #007bff; /* Blue heading */
  margin-bottom: 20px;
}

.welcome-container nav {
  margin-bottom: 30px;
}

.welcome-container nav a {
  margin: 0 10px; /* Space between links */
  text-decoration: none; /* Remove underlines from links */
  color: #333; /* Dark gray link color */
  font-weight: bold;
  transition: color 0.3s; /* Smooth color transition */
}

.welcome-container nav a:hover {
  color: #0056b3; /* Darker blue on hover */
}

.welcome-container {
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 0 5px rgba(0, 0, 0, 0.1);
}
/* Styles for the router view area */
.welcome-container .router-view {
  /* You can add styles specific to the router view content here */
  /* For example: */
  background-color: #fff;

  width: 100%; /* Occupy most of the container's width */
  max-width: 120%; /* Set a maximum width */
}
/* Example styles for content inside the router-view */
.welcome-container .router-view h2 {
  color: #333;
}

.welcome-container .router-view p {
  color: #555;
  line-height: 1.6;
}

.notification-container {
  position: relative;
  display: inline-block;
}

.notification-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  position: relative;
}

.badge {
  position: absolute;
  top: -5px;
  right: -5px;
  background: red;
  color: white;
  font-size: 12px;
  padding: 4px 7px;
  border-radius: 50%;
}

.notification-dropdown {
  margin-top: 75px;
  position: absolute;
  right: 0;
  width: 250px;
  background: white;
  box-shadow: 0px 4px 8px rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  padding: 10px;
  z-index: 1000;
}

.notification-dropdown h3 {
  font-size: 16px;
  margin-bottom: 10px;
  text-align: center;
}

.no-notifications {
  text-align: center;
  color: gray;
}

ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

li {
  padding: 10px;
  border-bottom: 1px solid #ddd;
  cursor: pointer;
  transition: background 0.3s;
}

li.unread {
  font-weight: bold;
  background: #f9f9f9;
}

li:hover {
  background: #ececec;
}

small {
  display: block;
  color: gray;
  font-size: 12px;
  margin-top: 5px;
}

.clear-btn {
  width: 100%;
  padding: 8px;
  background: #dc3545;
  color: white;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  margin-top: 10px;
}

.clear-btn:hover {
  background: #c82333;
}
.notification-container {
  width: 100%; /* Ensures full width */
  display: flex;
  justify-content: flex-end; /* Moves content (button) to the right */
}

#welcomnot {
  display: flex;
  align-items: center;
  justify-content: space-between; /* Pushes elements apart */
  width: 100%; /* Makes sure the div takes full width */
}

.notification-btn {
  margin-left: auto; /* Pushes the button to the right */
  display: flex; /* Ensure it behaves correctly */
  align-items: center;
}

#welcomnot {
  display: flex;
  align-items: center;
  justify-content: center; /* Center everything horizontally */
  width: 100%; /* Ensure full width */
  position: relative; /* Helps position elements properly */
}

#welcomnot h1 {
  flex-grow: 1; /* Makes h1 take up all available space */
  text-align: center; /* Centers the text */
}

.notification-btn {
  position: absolute;
  right: 0; /* Moves the button to the right */
}
</style>
