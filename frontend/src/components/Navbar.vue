<template>
  <!-- <h1>WELCOME To NTLink</h1> -->
  <!-- Sidebar -->
  <aside class="sidebar">
    <button @click="logout">Log-out</button>
    <ul>
      <li @click="navigateTo('my-profile')">👤 My Profile</li>
      <li @click="navigateTo('home')">🏠 Home</li>
      <li @click="navigateTo('chats')">💬 My Chats</li>
      <li @click="navigateTo('my-groups')">👥 My Groups</li>
      <li @click="navigateTo('requests')">➕ Requests & Invitations</li>
      <li @click="navigateTo('discover-people')">🔍🙋 Discover People</li>
      <li @click="navigateTo('discover-groups')">🔍👥 Discover Groups</li>
       </ul>
  </aside>
</template>

<script setup>
import { useAuthStore } from "@/authStore";
import { useRouter } from "vue-router";
import axios from "axios";
import config from "@/config";

axios.defaults.withCredentials = true; // ✅ Ensures cookies are sent & received
const auth = useAuthStore();
const router = useRouter();

const logout = async () => {
  try {
    await axios.post(`${config.API_URL}/logout`, { withCredentials: true });
    auth.logout(); // ✅ Updates UI instantly
    router.push("/login");
 } catch (error) {
    throw Error("Logout failed:", error);
  }
};

const navigateTo = (routeName) => {
  router.push(`/${routeName}`); // Navigate using route name
};
</script>

<style scoped>
.sidebar {
  width: 250px;
  background-color: #fff;
  padding: 20px;
  border-right: 1px solid #eee;
  /* Light border */
}

.sidebar ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.sidebar li {
  padding: 10px 0;
  cursor: pointer;
  transition: background-color 0.2s;
  /* Smooth transition on hover */
}

.sidebar li:hover {
  background-color: #f0f0f0;
  /* Light gray background on hover */
}

.sidebar ul {
  list-style: none;
  /* Remove bullet points */
  padding: 0;
  /* Remove default padding */
  margin: 0;
  /* Remove default margin */
}

.sidebar li {
  /* ... other li styles ... */
  text-align: left;
  /* Align text within the li to the left */
  /* OR, if you want the *entire* li to be on the left (including any icons): */
  display: block;
  /* Make the li a block element so it takes full width */
}

/* Optional styling for the icons/emojis */
.sidebar li span {
  /* Or target the specific icon class if you have one */
  margin-right: 5px;
  /* Add some space between icon and text */
}

/* Example with flexbox for more advanced alignment */
.sidebar li {
  display: flex;
  /* Use flexbox for alignment */
  align-items: center;
  /* Vertically center the icon and text */
}

.sidebar li span {
  margin-right: 10px;
}
</style>
