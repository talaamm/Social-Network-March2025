<template>
  <div class="container">
    <Navbar />
    <div class="chat-container">
      <h2>Chat with @{{ receiverNickname }}</h2>

      <div class="chat-messages" ref="chatBox">
        <div
          v-if="messages"
          v-for="msg in messages"
          :key="msg.id"
          :class="msg.receiver_id != loggedin_id ? 'sent' : 'received'"
        >
          <p>{{ msg.content }}</p>

          <small v-if="msg.sender_id === loggedin_id">
             {{ formatDate(msg.sent_at) }}</small
          >
          <small v-else> {{ formatDate(msg.sent_at) }} </small>
        </div>
      </div>

      <div class="chat-input">
        <input v-model="newMessage" type="text" placeholder="Type a message..." />
        <button @click="sendMessage">Send</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, onUnmounted } from "vue";
import { useRoute } from "vue-router";
import axios from "axios";
import config from "@/config";
import Navbar from "@/components/Navbar.vue";

const route = useRoute();
const receiverID = ref(route.params.id);
const receiverNickname = ref(route.params.nickname); // ✅ Get nickname
const chatBox = ref(null); // ✅ Reference for auto-scrolling

const messages = ref([]);
const newMessage = ref("");
let ws = null;

const scrollToBottom = async () => {
  await nextTick(); // ✅ Ensure Vue updates the DOM first
  setTimeout(() => {
    if (chatBox.value) {
      chatBox.value.scrollTop = chatBox.value.scrollHeight; // ✅ Scroll to bottom
    }
  }, 100); // ✅ Slight delay to ensure smooth scrolling
};
const setupWebSocket = () => {
  if (ws && ws.readyState === WebSocket.OPEN) {
    console.log("⚠️ Chat WebSocket already connected.");
    return;
  }

  if (!loggedin_id) {
    console.error("⚠️ User ID not set. WebSocket connection delayed.");
    return;
  }
  const wsUrl = `${config.API_URL.replace(/^http/, "ws")}/ws/chat?user_id=${loggedin_id}`;

  ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    console.log("✅ WebSocket connected for messaging.");
  };

  ws.onmessage = (event) => {
    try {
      console.log("📩 Raw WebSocket Message:", event.data); // ✅ Log message

      const data = JSON.parse(event.data);

      console.log("✅ Parsed Message:", data); // ✅ Log parsed message

      if (!data || !data.content || !data.sender_id) {
        console.error("❌ Invalid message received:", data);
        return;
      }

      // ✅ Ensure messages is an array before pushing
      if (!Array.isArray(messages.value)) {
        console.error("❌ messages is not an array! Resetting.");
        messages.value = [];
      }

      messages.value.push(data); // ✅ Add message to array
      scrollToBottom();
    } catch (error) {
      console.error("❌ Error parsing WebSocket message:", error);
    }
  };

  ws.onclose = (event) => {
    console.warn(`🔴 WebSocket closed. Code: ${event.code}, Reason: ${event.reason}`);
    setTimeout(setupWebSocket, 5000); // ✅ Reconnect after 5s
  };

  ws.onerror = (error) => {
    console.error("❌ WebSocket error:", error);
  };
};

const sendMessage = () => {
  if (!ws || ws.readyState !== WebSocket.OPEN) {
    console.error("❌ WebSocket is not open. Message not sent.");
    return;
  }

  if (!newMessage.value.trim()) return;

  const messageData = {
    receiver_id: parseInt(receiverID.value, 10), // ✅ Ensure it's an integer
    content: newMessage.value,
    sent_at: new Date().toISOString(),
  };
  console.log("📩 Sending Message:", messageData); // ✅ Debugging log

  ws.send(JSON.stringify(messageData)); // ✅ Send message via WebSocket

  if (!Array.isArray(messages.value)) {
    console.error("❌ messages is not an array! Resetting.");
    messages.value = [];
  }

  setTimeout(() => {
    if (!messages.value.find((m) => m.content === messageData.content)) {
      messages.value.push(messageData);
    }
  }, 300);
  scrollToBottom();
  newMessage.value = "";
};

// **📜 Fetch Chat History**
const fetchChatHistory = async () => {
  try {
    const response = await axios.get(
      `${config.API_URL}/api/chat/history?receiver_id=${receiverID.value}`,
      {
        withCredentials: true,
      }
    );
    messages.value = Array.isArray(response.data) ? response.data : []; // ✅ Ensure it's always an array
    scrollToBottom();
  } catch (error) {
    console.error("❌ Error fetching chat history:", error);
    messages.value = []; // ✅ Always default to an array
  }
};
let loggedin_id = 0;
let loggedin_nickname = "";
async function fetcCurrUserData() {
  try {
    const response = await axios.get(`${config.API_URL}/api/myself`, {
      withCredentials: true,
    });
    console.log(response.data);
    loggedin_id = response.data.id;
    loggedin_nickname = response.data.nickname;
    console.log(loggedin_id);
  } catch (error) {
    console.error("Error fetching user data:", error);
  }
}

const formatDate = (isoString) => {
  return new Date(isoString).toLocaleString(); // ✅ Convert timestamp to readable format
};

onMounted(async () => {
  await fetcCurrUserData();
  if (loggedin_id > 0) {
    setupWebSocket(); // ✅ Only initialize WebSocket if loggedin_id is valid
    // setupWebSocket();
    fetchChatHistory();
  } else {
    console.error("❌ WebSocket setup delayed: User ID not available.");
  }
});

onUnmounted(() => {
  if (ws) {
    ws.close();
    console.log("🔴 WebSocket closed.");
  }
});
</script>

<style scoped>
.chat-container {
  width: 400px;
  margin: 0 auto;
  background: #f9f9f9;
  padding: 15px;
  border-radius: 10px;
}

.chat-messages {
  height: 300px;
  overflow-y: auto;
  background: white;
  padding: 10px;
  border-radius: 5px;
}

.sent {
  text-align: right;
  background: #d1e7fd;
  padding: 8px;
  border-radius: 5px;
  margin: 5px 0;
}

.received {
  text-align: left;
  background: #e6e6e6;
  padding: 8px;
  border-radius: 5px;
  margin: 5px 0;
}

.chat-input {
  display: flex;
  margin-top: 10px;
}

.chat-input input {
  flex: 1;
  padding: 8px;
}

.chat-input button {
  background: #007bff;
  color: white;
  border: none;
  padding: 8px;
  cursor: pointer;
}

.container {
  display: flex;
  width: 90%;
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
.chat-messages {
  height: 400px; /* ✅ Adjust based on your layout */
  overflow-y: auto; /* ✅ Enable scrolling */
  display: flex;
  flex-direction: column;
}
</style>
