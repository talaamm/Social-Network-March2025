<script setup>
import { ref, onMounted } from "vue";
import axios from "axios";
import config from "@/config";
import { useRouter } from "vue-router";
import Navbar from "@/components/Navbar.vue";

const router = useRouter();
const recentChats = ref([]);
const availableUsers = ref([]);
const userGroups = ref([]); // ✅ Store user groups here

axios.defaults.withCredentials = true; // ✅ Ensures cookies are sent & received

const fetchRecentChats = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/chat/recent`, {
      withCredentials: true,
    });
    recentChats.value = response.data || [];
  } catch (error) {
    console.error("❌ Error fetching recent chats:", error);
  }
};

const formatDate = (isoString) => {
  return new Date(isoString).toLocaleString(); // ✅ Convert timestamp to readable format
};

const fetchAvailableUsers = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/chat/users`, {
      withCredentials: true,
    });
    availableUsers.value = response.data || [];
  } catch (error) {
    console.error("❌ Error fetching chat users:", error);
  }
};

// **Open Chat with Selected User**
const openChat = (userID , nick) => {
  router.push({ name: "ChatPage", params: { id: userID , nickname: nick } });
};

// ✅ Fetch groups the user is a member of
const fetchUserGroups = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/groups-to-chat`, {
      withCredentials: true,
    });
    userGroups.value = response.data || [];
  } catch (error) {
    console.error("❌ Error fetching groups:", error);
  }
};

const openGroupChat = (groupID, groupName) => {
  router.push({ name: "GroupChat", params: { groupid: groupID, name: groupName } });
};

onMounted(() => {
  fetchRecentChats();
  fetchAvailableUsers();
  fetchUserGroups();
});
</script>

<template>
  <div class="container">
  <Navbar/>
  <div class="chat-list-container">
 
 
    <h2>Group Chats</h2>
      <div v-if="userGroups.length === 0">No groups yet.</div>
      <div v-else>
        <div
          v-for="group in userGroups"
          :key="group.id"
          class="group-item"
          @click="openGroupChat(group.id, group.name)"
        >
          <strong>{{ group.name }}</strong> <br />
          <small>Created by: {{ group.creator_nickname }}</small> <br />
          <small>Members: {{ group.member_count }}</small>
        </div>
      </div>
 
    <h2>Recent Chats</h2>
    <div v-if="recentChats.length === 0">No recent chats.</div>
    <ul>
      <li v-for="user in recentChats" :key="user.id" @click="openChat(user.id , user.nickname)">
        <strong>@{{ user.nickname }}</strong> ({{ user.first_name }} {{ user.last_name }})
        <span class="last-message-time">{{ formatDate(user.last_message_time) }}</span>
      </li>
    </ul>

    <h2>Available Users</h2>
    <div v-if="availableUsers.length === 0">No users available to chat.</div>
    <ul>
      <li v-for="user in availableUsers" :key="user.id" @click="openChat(user.id , user.nickname)">
        <strong>@{{ user.nickname }}</strong> ({{ user.first_name }} {{ user.last_name }})
      </li>
    </ul>
  </div>
  </div>
</template>

<style scoped>
.chat-list-container {
  width: 400px;
  margin: 0 auto;
  padding: 15px;
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

ul {
  list-style: none;
  padding: 0;
}

li {
  padding: 10px;
  border-bottom: 1px solid #ddd;
  cursor: pointer;
}

li:hover {
  background: #f0f0f0;
}

.last-message-time {
  font-size: 12px;
  color: gray;
  margin-left: 10px;
}
.chats-container {
  padding: 20px;
}
.chat-item, .group-item {
  padding: 10px;
  margin: 5px 0;
  border: 1px solid #ddd;
  cursor: pointer;
}
.chat-item:hover, .group-item:hover {
  background-color: #f0f0f0;
}
</style>
