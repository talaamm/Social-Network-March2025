<template>
  <div class="discover-container">
    <Navbar />
    <main class="content">
      <div v-if="loading" class="loading">Loading users...</div>
      <div v-if="error" class="error">{{ error }}</div>
      <div v-if="users.length === 0 && !loading" class="no-users">
        No users found to discover.
      </div>

      <div v-for="user in users" :key="user.id" class="user-card">
        <!-- <div class="user-avatar">{{ user.nickname[0] }}</div> -->
        <div class="user-avatar" :style="{ backgroundColor: getRandomColor(user.id) }">
          {{ user.nickname[0] }}
        </div>
        <div class="user-info" @click="showProfile(user.id)">
          <h2>{{ user.first_name }} {{ user.last_name }}</h2>
          <h3>@{{ user.nickname }}</h3>
          <p>{{ user.isprivate ? "🔒 Private" : "🌍 Public" }}</p>
          <span class="hidden-user-id">{{ user.id }}</span>
        </div>
        <button
          class="follow-btn"
          :class="getFollowClass(user.following_status)"
          @click.stop="toggleFollow(user)"
        >
          {{ getFollowText(user.following_status) }}
        </button>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import axios from "axios";
import config from "@/config";
import Navbar from "@/components/Navbar.vue";
import { useRouter } from "vue-router";
const router = useRouter();

const users = ref([]);
const loading = ref(true);
const error = ref("");

axios.defaults.withCredentials = true;

let loggedin_id = 0;

const getRandomColor = (id) => {
  const colors = [
    "#5733ff",
    "#33ff57",
    "#f1c40f",
    "#8e44ad",
    "#3498db",
    "#ff1493",
    "#7fffd4",
    "#dc143c",
    "#1e90ff",
    "#32cd32",
    "#ff4500",
    "#ff8c00",
    "#00ced1",
    "#ff69b4",
    "#00fa9a",
    "#e74c3c",
    "#2ecc71",
    "#4682b4",
    "#da70d6",
    "#ff5733",
    "#b8860b",
    "#ff6347",
    "#40e0d0",
    "#9b59b6",
  ];
  // console.log(id, id % colors.length);
  return colors[id % colors.length];
};

async function fetcCurrUserData() {
  try {
    const response = await axios.get(`${config.API_URL}/api/myself`, {
      withCredentials: true,
    });
    loggedin_id = response.data.id;
    // privateP.value = response.data.isprivate ? "private" : "public";

    console.log(loggedin_id);
  } catch (error) {
    console.error("Error fetching user data:", error);
  }
}

const fetchUsers = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/discover-people`, {
      withCredentials: true,
    });
    users.value = response.data.filter((user) => user.id !== loggedin_id);
  } catch (err) {
    error.value = "Failed to load users.";
    console.error("Error fetching users:", err);
  } finally {
    loading.value = false;
  }
};

function showProfile(userId) {
  if (userId == loggedin_id) {
    router.push("/my-profile");
  } else {
    router.push({ name: "UserProfile", params: { id: userId } });
  }
}
const toggleFollow = async (user) => {
  try {
    if (user.following_status === "accepted" || user.following_status === "pending") {
      // Unfollow or cancel follow request
      const response = await axios.post(
        `${config.API_URL}/api/unfollow`,
        { id: user.id },
        { withCredentials: true }
      );
      if (response.status === 200) {
        user.following_status = "not-following"; // Update UI
        console.log("Unfollowed:", response.data);
      }
    } else {
      // Follow user
      const response = await axios.post(
        `${config.API_URL}/api/follow`,
        { id: user.id },
        { withCredentials: true }
      );
      if (response.status === 200) {
        user.following_status = response.data.status; // 'accepted' or 'pending'
        console.log("Followed:", response.data);
      }
    }
  } catch (err) {
    console.error("Error toggling follow status:", err);
  }
};

const getFollowClass = (status) => {
  return status === "following" || status === "accepted"
    ? "following"
    : status === "pending"
    ? "pending"
    : "not-following";
};

const getFollowText = (status) => {
  return status === "accepted"
    ? "Unfollow"
    : status === "pending"
    ? "Cancel Request"
    : "Follow";
};

onMounted(async () => {
  await fetcCurrUserData();
  fetchUsers();
});
</script>

<style scoped>
.discover-container {
  display: flex;
  width: 90%;
  max-width: 1200px;
  margin: 20px auto;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  background: white;
}

.main-content {
  display: flex;
  flex: 1;
}

.content {
  flex: 1;
  padding: 20px;
  background-color: #fff;
}

.loading,
.error,
.no-users {
  text-align: center;
  font-size: 16px;
  color: gray;
  margin-top: 10px;
}

.user-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 15px;
  border-bottom: 1px solid #ddd;
  transition: background 0.3s;
  cursor: pointer;
}

.user-card:hover {
  background: #f0f0f0;
}

.user-info {
  flex: 1;
}

.hidden-user-id {
  display: none;
}

.follow-btn {
  padding: 8px 12px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.3s ease-in-out;
}

.follow-btn.not-following {
  background-color: #007bff;
  color: white;
}

.follow-btn.following {
  background-color: #28a745;
  color: white;
}

.follow-btn.pending {
  background-color: #ffc107;
  color: black;
}

.user-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  color: white;
  font-size: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}
</style>
