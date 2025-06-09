<template>
  <div class="container">
    <Navbar />
    <div class="group-chat-container">
      <h2 class="chat-title">Group Chat: {{ groupName }}</h2>

      <div class="chat-messages" ref="chatBox">
        <div
          v-for="msg in messages"
          :key="msg.id"
          :class="[
            'message',
            msg.sender_id === loggedin_id ? 'sent' : 'received',
          ]"
        >
          <p class="message-content">{{ msg.content }}</p>
          <small class="message-meta"
            >{{ msg.sender_nickname }} - {{ formatDate(msg.sent_at) }}</small
          >
        </div>
      </div>

      <div class="chat-input">
        <input
          v-model="newMessage"
          type="text"
          placeholder="Type a message..."
          @keyup.enter="sendMessage"
        />
        <button @click="sendMessage">Send</button>
      </div>
    </div>
  </div>
</template>
<script setup>
import { ref, onMounted, nextTick, onUnmounted, watch } from "vue";
import axios from "axios";
import config from "@/config";
import { useRoute } from "vue-router";
import Navbar from "@/components/Navbar.vue";

const route = useRoute();
const groupID = ref(route.params.groupid);
const groupName = ref(route.params.name);
const messages = ref([]);
const newMessage = ref("");
let ws = null;
let loggedin_id = 0;
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
let heartbeatInterval = null; // ✅ Ping-pong keep-alive
const chatBox = ref(null); // ✅ Reference for auto-scrolling

const scrollToBottom = async () => {
  await nextTick(); // ✅ Ensure Vue updates the DOM first
  setTimeout(() => {
    if (chatBox.value) {
      chatBox.value.scrollTop = chatBox.value.scrollHeight; // ✅ Scroll to bottom
    }
  }, 100); // ✅ Slight delay to ensure smooth scrolling
};

// ✅ Fetch Group Chat History
const fetchChatHistory = async () => {
  try {
    const response = await axios.get(
      `${config.API_URL}/api/group/chat/history?group_id=${groupID.value}`,
      {
        withCredentials: true,
      }
    );
    console.log("📜 Fetched Messages:", response.data); // ✅ Debugging log
    messages.value = response.data
      ? response.data.map((msg) => ({
          ...msg,
          nickname: msg.nickname || "Group Member", // ✅ Prevents frontend null values
        }))
      : [];
    scrollToBottom();
  } catch (error) {
    console.error("❌ Error fetching chat history:", error);
    messages.value = [];
  }
};

// ✅ Setup WebSocket Connection with Auto-Reconnect & Heartbeat
const setupWebSocket = () => {
  if (ws && ws.readyState === WebSocket.OPEN) {
    console.log("⚠️ WebSocket already connected.");
    return;
  }

  if (!loggedin_id || !groupID.value) {
    console.error(
      "⚠️ User ID or GroupID is not set. WebSocket connection delayed."
    );
    return;
  }

  const wsUrl = `${config.API_URL.replace(
    /^http/,
    "ws"
  )}/ws/groupchat?user_id=${loggedin_id}&group_id=${groupID.value}`;

  ws = new WebSocket(wsUrl);
  console.log("🌐 Connecting to WebSocket:", wsUrl);

  ws.onopen = () => {
    console.log("✅ WebSocket connected for GROUP Chat:", groupID.value);
    reconnectAttempts = 0; // ✅ Reset reconnect attempts

    // ✅ Start sending heartbeat pings every 30 seconds
    if (heartbeatInterval) clearInterval(heartbeatInterval);
    heartbeatInterval = setInterval(() => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "ping" }));
        console.log("📡 Sent heartbeat ping");
      }
    }, 30000);
  };

  ws.onmessage = (event) => {
    try {
      console.log("📩 Raw WebSocket Message:", event.data);

      const data = JSON.parse(event.data);
      console.log("✅ Parsed Message:", data);

      if (!data || !data.content || !data.sender_id) {
        console.error("❌ Invalid message received:", data);
        return;
      }

      if (!Array.isArray(messages.value)) {
        console.error("❌ messages is not an array! Resetting.");
        messages.value = [];
      }

      messages.value.push(data);
      scrollToBottom();
    } catch (error) {
      console.error("❌ Error parsing WebSocket message:", error);
    }
  };

  ws.onclose = (event) => {
    console.warn(
      `🔴 WebSocket closed. Code: ${event.code}, Reason: ${event.reason}`
    );

    // ✅ Attempt to reconnect with a limit
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
  };

  ws.onerror = (error) => {
    console.error("❌ WebSocket error:", error);
  };
};
let loggedin_nickname = "";
// ✅ Fetch current user ID
async function fetchCurrUserData() {
  try {
    const response = await axios.get(`${config.API_URL}/api/myself`, {
      withCredentials: true,
    });
    loggedin_id = response.data.id;
    loggedin_nickname = response.data.nickname;
    console.log("👤 Logged-in User ID:", loggedin_id);
  } catch (error) {
    console.error("❌ Error fetching user data:", error);
  }
}

// ✅ Send Message to Group Chat
const sendMessage = () => {
  if (!newMessage.value.trim()) return;
  if (!ws || ws.readyState !== WebSocket.OPEN) {
    console.error("❌ WebSocket is not open. Message not sent.");
    return;
  }

  const messageData = {
    sender_id: loggedin_id,
    sender_nickname: loggedin_nickname, // ✅ Include nickname
    group_id: parseInt(groupID.value, 10),
    content: newMessage.value,
    sent_at: new Date().toISOString(),
  };

  // ✅ Send to WebSocket
  ws.send(JSON.stringify(messageData));
  if (!Array.isArray(messages.value)) {
    console.error("❌ messages is not an array! Resetting.");
    messages.value = [];
  }
  // ✅ Only update UI *if WebSocket doesn't send it back*
  setTimeout(() => {
    if (
      !messages.value.find(
        (m) => m.content === messageData.content && m.sender_id === loggedin_id
      )
    ) {
      messages.value.push(messageData);
    }
  }, 300);

  scrollToBottom();
  newMessage.value = "";
};

// ✅ Watch for Route Changes & Ensure WebSocket Works Properly
watch(
  () => route.params.groupid,
  async (newGroupId) => {
    if (newGroupId) {
      groupID.value = newGroupId;
      console.log("🔄 Group ID updated:", newGroupId);
      await fetchChatHistory();
      setupWebSocket();
    }
  }
);

// ✅ Initialize on Component Mount
onMounted(async () => {
  await fetchCurrUserData();
  // if (route.params.groupid) {
  //   groupID.value = route.params.groupid;
  // } else {
  //   console.error("❌ No group ID provided in route.");
  // }

  if (loggedin_id > 0 && groupID.value > 0) {
    fetchChatHistory();
    setupWebSocket();
  } else {
    console.error("❌ WebSocket setup delayed: User ID not available.");
  }
});

// ✅ Cleanup WebSocket on Component Unmount
// ✅ Cleanup WebSocket on Component Unmount
onUnmounted(() => {
  console.log("🔴 Closing WebSocket connection...");

  if (ws) {
    ws.onclose = null; // ✅ Prevent auto-reconnect loop
    ws.onerror = null;
    ws.onmessage = null;
    ws.close(); // ✅ Close WebSocket properly
    ws = null; // ✅ Clear reference to prevent reconnection
  }

  if (heartbeatInterval) {
    clearInterval(heartbeatInterval); // ✅ Stop heartbeat ping
    heartbeatInterval = null;
  }
});

const formatDate = (isoString) => {
  return new Date(isoString).toLocaleString(); // ✅ Convert timestamp to readable format
};
</script>

<style scoped>
/* ✅ Chat container */
.group-chat-container {
  width: 500px;
  margin: 0 auto;
  background: #fff;
  border-radius: 10px;
  box-shadow: 0px 4px 10px rgba(0, 0, 0, 0.1);
  padding: 20px;
}

/* ✅ Chat title */
.chat-title {
  text-align: center;
  font-size: 20px;
  font-weight: bold;
  color: #333;
  margin-bottom: 10px;
}

/* ✅ Messages container */
.chat-messages {
  max-height: 400px;
  overflow-y: auto;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background: #f9f9f9;
}

/* ✅ Individual message */
.message {
  max-width: 75%;
  padding: 10px;
  border-radius: 15px;
  margin: 5px 0;
  font-size: 14px;
  position: relative;
}

/* ✅ Sent messages */
.sent {
  align-self: flex-end;
  background: #007bff;
  color: white;
  text-align: right;
  margin-left: auto;
}

/* ✅ Received messages */
.received {
  align-self: flex-start;
  background: #e5e5ea;
  color: black;
  text-align: left;
  margin-right: auto;
}

/* ✅ Message content */
.message-content {
  margin: 0;
}

/* ✅ Message metadata (nickname + time) */
.message-meta {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.6);
  margin-top: 5px;
  display: block;
}

/* ✅ Chat input container */
.chat-input {
  display: flex;
  margin-top: 15px;
  gap: 10px;
}

/* ✅ Input field */
.chat-input input {
  flex: 1;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 8px;
  font-size: 14px;
}

/* ✅ Send button */
.chat-input button {
  background: #007bff;
  color: white;
  border: none;
  padding: 10px 15px;
  border-radius: 8px;
  cursor: pointer;
  transition: 0.2s;
}

/* ✅ Button hover effect */
.chat-input button:hover {
  background: #0056b3;
}
.chat-messages {
  height: 400px; /* ✅ Adjust based on your layout */
  overflow-y: auto; /* ✅ Enable scrolling */
  display: flex;
  flex-direction: column;
}

.container {
  display: flex;
  width: 100%;
  /* Occupy most of the screen */
  max-width: 1200px;
  /* Set a maximum width */
  margin: 20px auto;
  /* Center the container */
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
  /* Subtle shadow */
  border-radius: 8px;
  /* Rounded corners */
  overflow: hidden;
  /* Hide overflowing content */
}
</style>
