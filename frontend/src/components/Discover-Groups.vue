<template>
  <div class="container">
    <Navbar />
    <div class="discover-container">
      <header class="discover-header">
        <h1>Discover New Groups</h1>
        <div v-if="showPopup" :style="popupStyle" class="popup">
            🚀 Join the group to view posts!
          </div>
      </header>

      <div v-if="loading" class="loading">Loading groups...</div>
      <div v-if="error" class="error">{{ error }}</div>

      <!-- 🔵 Pending Requests Section -->
      <div v-if="pendingGroups.length" class="pending-section">
        <h2>Pending Group Requests</h2>
        <div class="groups-list">
          <div v-for="group in pendingGroups" :key="group.id" class="group-card pending">
            <div class="group-info" @click="showJoinPopup(group.id, $event)">
              <h3>{{ group.name }}</h3>
              <p>{{ group.description }}</p>
              <p>Creator: @{{ group.creator_nickname }}</p>
              <p>Members: {{ group.member_count }}</p>
            </div>
            <button class="join-btn pending-btn" disabled>⏳ Pending Approval</button>
          </div>
        </div>
      </div>

      <!-- 🟢 Available Groups to Join -->
      <h2>Available Groups</h2>
      <div class="groups-list">
        <div v-if="availableGroups.length == 0"><p>No Available Groups</p></div>
        <div v-else v-for="group in availableGroups" :key="group.id" class="group-card">
          <div class="group-info" @click="showJoinPopup(group.id, $event)">
            <h3>{{ group.name }}</h3>
            <p>{{ group.description }}</p>
            <p>Creator: @{{ group.creator_nickname }}</p>
            <p>Members: {{ group.member_count }}</p>
          </div>

          <button
            class="join-btn"
            :class="{ requested: group.requestSent }"
            @click.stop="requestToJoin(group)"
            :disabled="group.requestSent"
          >
            {{ group.requestSent ? "✅ Request Sent" : "➕ Request to Join" }}
          </button>
        
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import axios from "axios";
import config from "@/config";
import Navbar from "@/components/Navbar.vue";
import { useRouter } from "vue-router";

axios.defaults.withCredentials = true;
const router = useRouter();
const showPopup = ref(false);
const popupStyle = ref({ top: "0px", left: "0px" });

const availableGroups = ref([]); // Groups the user can request to join
const pendingGroups = ref([]); // Groups with pending requests
const loading = ref(false);
const error = ref("");

const fetchGroups = async () => {
  try {
    loading.value = true;
    const response = await axios.get(`${config.API_URL}/api/discover-groups`);

    if (response.data) {
      availableGroups.value =
        response.data.available_groups.map((group) => ({
          ...group,
          requestSent: false,
        })) || [];

      pendingGroups.value = response.data.pending_groups || [];
    }
  } catch (err) {
    error.value = "Failed to load groups.";
    console.error("Error fetching groups:", err);
  } finally {
    loading.value = false;
  }
};
const requestToJoin = async (group) => {
  if (group.requestSent) return;

  try {
    // ✅ 1. Instantly move the group to pending in the frontend

    await axios.post(`${config.API_URL}/api/groups/join?group_id=${group.id}`);
    group.requestSent = true;
    availableGroups.value = availableGroups.value.filter((g) => g.id !== group.id);
    pendingGroups.value = [...pendingGroups.value, group];
    console.log(`Join request sent for group ID ${group.id}`);
  } catch (error) {
    group.requestSent = false;
    console.error("Error requesting to join group:", error);
  }
};

// ✅ Show Temporary Popup Near Click
const showJoinPopup = (groupId, event) => {
  const { clientX, clientY } = event; // Get click position

  popupStyle.value = {
    top: `${clientY}px`,
    left: `${clientX}px`,
  };

  showPopup.value = true;

  setTimeout(() => {
    showPopup.value = false;
  }, 2000); // Auto-hide after 2 seconds
};

onMounted(fetchGroups);
</script>

<style scoped>
.container {
  display: flex;
  width: 95%;
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

.discover-container {
  width: 90%;
  max-width: 1000px;
  margin-top: 20px;
  text-align: center;
}

.discover-header {
  margin-bottom: 20px;
}

.groups-list {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
  justify-content: center;
}

.group-card {
  background: white;
  padding: 20px;
  border-radius: 10px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out;
  width: 280px;
  cursor: pointer;
}

.group-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.15);
}

.group-info {
  text-align: left;
}

.join-btn {
  background-color: #4caf50;
  color: white;
  font-size: 14px;
  padding: 8px 12px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  margin-top: 10px;
  transition: background 0.3s;
}

.join-btn:hover {
  background-color: #45a049;
}

.loading {
  font-size: 18px;
  color: gray;
}

.error {
  font-size: 16px;
  color: red;
}

/* ✅ Button when request is sent */
.join-btn.requested {
  background-color: #999;
  cursor: not-allowed;
  cursor: not-allowed; /* 🔥 Disables pointer events & shows '🚫' cursor */
  opacity: 0.7; /* Optional: Makes the button look inactive */
}
.pending-btn {
  background-color: #ff9800;
  cursor: not-allowed;
}
.popup {
  position: absolute;
  background-color: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 10px;
  border-radius: 5px;
  font-size: 14px;
  z-index: 1000;
  transition: opacity 0.3s ease-in-out;
  opacity: 1;
}

.popup.hide {
  opacity: 0;
}
</style>
