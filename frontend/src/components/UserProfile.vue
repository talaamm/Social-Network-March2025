<template>
  <div class="container">
    <Navbar />

    <main class="content">
      <!-- User Info Section -->
      <div class="user-info-box">
        <div
          class="user-avatar"
          :style="{ backgroundColor: getRandomColor(userData.id) }"
        >
          {{ harf }}
        </div>
        <div class="user-details">
          <h2>{{ userData.nickname }}</h2>
          <p>{{ userData.first_name }} {{ userData.last_name }}</p>
          <p>Email: {{ userData.email }}</p>
          <p>Age: {{ userData.age || "N/A" }}</p>
          <p>Gender: {{ userData.gender }}</p>
          <p>Birthdate: {{ userData.dbirth }}</p>
          <p v-if="userData.isprivate">Profile: Private</p>
          <p v-else>Profile: Public</p>

          <!-- Follow/Unfollow Button -->
          <button
            class="follow-btn"
            :class="getFollowClass(userData.following_status)"
            @click="toggleFollow"
          >
            {{ getFollowText(userData.following_status) }}
          </button>
        </div>
      </div>

      <!-- User Posts Section -->
      <div class="post-feed">
        <h3 v-if="userData.isprivate">THIS ACCOUNT IS PRIVATE 🔒</h3>
        <h3 v-else-if="!userPosts || userPosts == null">No Posts Yet..</h3>
        <h3 v-else="userPosts.length">{{ userData.nickname }}'s Posts</h3>
        <div v-for="post in userPosts" :key="post.id" class="post">
          <div class="post-header">
            <!-- <div class="avatar"></div> -->
            <h3>@{{ post.nickname }}</h3>
            <p v-if="post.privacy == 'public'">Privacy: Public</p>
            <p v-else-if="post.privacy == 'selected'">Privacy: Close Friends</p>
            <p v-else-if="post.privacy == 'followers'">Privacy: Followers Only</p>
            <p>Created At: {{ formatDate(post.created_at) }}</p>
          </div>
          <p class="post-content">{{ post.content }}</p>
          <img
            v-if="post.image"
            :src="`${config.API_URL}/${post.image}`"
            alt="Post Image"
            class="post-image"
          />
          <div class="post-actions">
            <button @click="likepost(post.id)">👍 {{ post.likes }}</button>
            <button @click="toggleComments(post.id)">💬 Comments</button>
            <button @click="dislikepost(post.id)">👎 {{ post.dislikes }}</button>
          </div>

          <div v-if="selectedPostId === post.id" class="comments-section">
            <p v-if="loadingComments">No comments yet.</p>
            <!--Loading comments...-->
            <!-- <p v-else-if="cmntLen === 0">No comments yet.</p> -->
            <p v-else-if="cmntLen === 0">No comments yet.</p>
            <ul v-else>
              <p style="text-align: left; margin-left: 2%" v-if="cmntLen == 1">
                <strong> {{ cmntLen }} Comment</strong>
              </p>
              <p style="text-align: left; margin-left: 2%" v-else>
                <strong>{{ cmntLen }} Comments</strong>
              </p>
              <li
                style="text-align: left; margin-left: 5%"
                v-for="comment in comments"
                :key="comment.id"
              >
                <strong>@{{ comment.nickname }}:</strong> {{ comment.content }}
              </li>
            </ul>

            <!-- ✅ Comment Input & Button -->
            <div class="comment-box">
              <input v-model="newComment" type="text" placeholder="Write a comment..." />
              <button @click="postComment(post.id)">Comment</button>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from "vue";
import axios from "axios";
import config from "@/config";
import Navbar from "@/components/Navbar.vue";
import { useRoute } from "vue-router";

const route = useRoute();
const userId = ref(route.params.id);
const userData = ref({});
const userPosts = ref([]);
const selectedPostId = ref(null);
const comments = ref([]);
const loadingComments = ref(false);
const newComment = ref("");
const harf = ref("");
// let loggedin_id = 0;
axios.defaults.withCredentials = true;

async function fetchUserData() {
  try {
    const response = await axios.get(
      `${config.API_URL}/api/user-data?user_id=${userId.value}`,
      { withCredentials: true }
    );

    if (response.data) {
      userData.value = response.data;
      harf.value = response.data.nickname[0];
      // Fetch the correct following status for the logged-in user
      fetchFollowingStatus();
    }
  } catch (error) {
    console.error("Error fetching user data:", error);
  }
}

const fetchFollowingStatus = async () => {
  try {
    const response = await axios.get(
      `${config.API_URL}/api/follow-status?user_id=${userId.value}`,
      { withCredentials: true }
    );

    if (response.data && response.data.status) {
      userData.value.following_status = response.data.status; // 'accepted', 'pending', or 'not-following'
    } else {
      userData.value.following_status = "not-following"; // Default if no status found
    }
  } catch (error) {
    console.error("Error fetching follow status:", error);
    userData.value.following_status = "not-following"; // Prevent undefined issues
  }
};

const fetchUserPosts = async () => {
  try {
    const response = await axios.get(
      `${config.API_URL}/api/user-posts?user_id=${userId.value}`,
      {
        withCredentials: true,
      }
    );
    userPosts.value = response.data;
  } catch (error) {
    console.error("Error fetching user posts:", error);
  }
};

const toggleFollow = async () => {
  try {
    const userID = parseInt(userId.value, 10); // Convert to integer

    if (
      userData.value.following_status === "accepted" ||
      userData.value.following_status === "pending"
    ) {
      // Unfollow or cancel request
      const response = await axios.post(
        `${config.API_URL}/api/unfollow`,
        { id: userID },
        { withCredentials: true }
      );
      if (response.status === 200) {
        userData.value.following_status = "not-following"; // Update UI immediately
        console.log("Unfollowed:", response.data);
      }
    } else {
      // Follow user
      const response = await axios.post(
        `${config.API_URL}/api/follow`,
        { id: userID },
        { withCredentials: true }
      );
      if (response.status === 200) {
        userData.value.following_status = response.data.status; // Set 'accepted' or 'pending'
        console.log("Followed:", response.data);
      }
    }
  } catch (err) {
    console.error("Error toggling follow status:", err);
  }
};

const getFollowClass = (status) => {
  return status === "accepted"
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

const formatDate = (isoString) => {
  return new Date(isoString).toLocaleString();
};

const toggleComments = async (postId) => {
  if (selectedPostId.value === postId) {
    comments.value = [];
    selectedPostId.value = 0;
    return;
  }
  selectedPostId.value = postId;
  loadingComments.value = true;
  try {
    const response = await axios.get(`${config.API_URL}/api/comments?post_id=${postId}`);
    comments.value = response.data;
  } catch (error) {
    console.error("Error fetching comments:", error);
    comments.value = [];
  }
  loadingComments.value = false;
};

const cmntLen = computed(() =>
  Array.isArray(comments.value) ? comments.value.length : 0
);

const postComment = async (postId) => {
  if (!newComment.value.trim()) return; // Prevent empty comments

  try {
    const response = await axios.post(
      `${config.API_URL}/api/comments`,
      {
        //${postId}

        content: newComment.value,
        post_id: postId,
      },
      { withCredentials: true }
    );

    newComment.value = "";
    if (!Array.isArray(comments.value)) {
      comments.value = [];
    }
    comments.value = [...comments.value, response.data];
    cmntLen.value = comments.value.length;

    console.log("Updated Comments:", comments.value); // Debugging
  } catch (error) {
    console.error("Error posting comment:", error);
    throw Error(error);
  }
};

const likepost = async (postid) => {
  const postIndex = userPosts.value.findIndex((p) => p.id === postid);
  if (postIndex === -1) return;

  try {
    const response = await axios.post(`${config.API_URL}/api/like`, {
      postid: postid,
      islike: true, // Liking the post
    });

    console.log(response.data);

    if (response.data) {
      // ✅ Update the state reactively
      userPosts.value[postIndex] = {
        ...userPosts.value[postIndex],
        likes: response.data.likes,
        dislikes: response.data.dislikes,
      };
    }
  } catch (error) {
    console.error("Error liking post:", error);
  }
};
const dislikepost = async (postid) => {
  const postIndex = userPosts.value.findIndex((p) => p.id === postid);
  if (postIndex === -1) return;

  try {
    const response = await axios.post(`${config.API_URL}/api/like`, {
      postid: postid,
      islike: false, // Disliking the post
    });

    console.log(response.data);

    if (response.data) {
      // ✅ Update the state reactively
      userPosts.value[postIndex] = {
        ...userPosts.value[postIndex],
        likes: response.data.likes,
        dislikes: response.data.dislikes,
      };
    }
  } catch (error) {
    console.error("Error disliking post:", error);
  }
};

onMounted(async () => {
  //   await fetcCurrUserData();
  await fetchUserData();
  await fetchUserPosts();
});
</script>

<style scoped>
.profile-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  max-width: 800px;
  margin: auto;
  padding: 20px;
}

.profile-content {
  width: 100%;
}

.user-info-box {
  display: flex;
  align-items: center;
  padding: 20px;
  background: white;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  border-radius: 10px;
  margin-bottom: 20px;
}

.user-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: #4caf50;
  color: white;
  font-size: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}

.post {
  background: white;
  padding: 20px;
  border-radius: 10px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
}

.post-header {
  display: flex;
  justify-content: space-between;
}

.post-content {
  margin: 10px 0;
}

.post-actions {
  display: flex;
  gap: 10px;
}

.comments-section {
  margin-top: 10px;
  padding: 10px;
  background: #f9f9f9;
  border-radius: 5px;
}

.comment-box {
  display: flex;
  margin-top: 10px;
}

.comment-box input {
  flex: 1;
  padding: 8px;
  margin-right: 5px;
}

body {
  font-family: sans-serif;
  margin: 0;
  background-color: #f4f4f4;
  /* Light gray background */
  color: #333;
  /* Dark gray text color */
  display: flex;
  /* Use flexbox for layout */
  min-height: 100vh;
  /* Ensure full viewport height */
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

/* Main Content */
.content {
  flex: 1;
  /* Take up remaining space */
  background-color: #fff;
  padding: 20px;
}

/* Navbar */
.navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 20px;
  border-bottom: 1px solid #eee;
  margin-bottom: 20px;
}

.nav-links {
  display: flex;
  gap: 10px;
  /* Space between buttons */
}

.nav-links button {
  background-color: #eee;
  border: none;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.nav-links button.active,
.nav-links button:hover {
  background-color: #ddd;
}

.user-info {
  display: flex;
  align-items: center;
}

/* Post Box */
.post-box {
  margin-bottom: 20px;
}

.post-box textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  resize: vertical;
  /* Allow vertical resizing */
  min-height: 100px;
}

.post-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
}

.post-actions button,
.post-actions select {
  padding: 8px 12px;
  border: 1px solid #ccc;
  border-radius: 4px;
  cursor: pointer;
}

.post-actions button.post-btn {
  background-color: #4caf50;
  /* Green */
  color: white;
  border: none;
}

/* Post Feed */
.post {
  border: 1px solid #eee;
  padding: 20px;
  margin-bottom: 20px;
  border-radius: 8px;
  background-color: white;
  /* White background for posts */
}

.post-header {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.post-header .avatar {
  margin-right: 10px;
}

.post-header h3 {
  margin: 0;
}

.post-header p {
  margin: 0;
  font-size: 0.8em;
  color: #777;
}

.post-content {
  margin-bottom: 10px;
}

.post-actions {
  display: flex;
  gap: 10px;
  /* Space between buttons */
}

.post-actions button {
  background-color: #eee;
  border: none;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.post-actions button:hover {
  background-color: #ddd;
}

/* Post Feed */
.post-feed {
  display: flex;
  flex-direction: column;
  gap: 15px; /* Space between posts */
  padding: 20px;
  background-color: #f9f9f9; /* Light background to contrast posts */
  border-radius: 10px;
}

/* Individual Post */
.post {
  background: white;
  padding: 20px;
  border-radius: 10px; /* Rounded corners */
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1); /* Soft shadow for depth */
  transition: transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out;
}

.post:hover {
  transform: translateY(-5px); /* Subtle lift effect */
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.15); /* Enhanced shadow on hover */
}

/* Post Header */
.post-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #eee;
  padding-bottom: 10px;
  margin-bottom: 10px;
}

.post-header .avatar {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background-color: #ddd; /* Placeholder for avatar */
  display: none;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 20px;
  color: #555;
}

/* Post Header Text */
.post-header div {
  flex: 1;
  margin-left: 15px;
}

.post-header h3 {
  margin: 0;
  font-size: 18px;
  color: #333;
}

.post-header p {
  margin: 2px 0;
  font-size: 14px;
  color: #777;
}

/* Post Content */
.post-content {
  font-size: 16px;
  color: #444;
  line-height: 1.5;
  padding: 10px 0;
}

/* Post Actions */
.post-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid #eee;
  padding-top: 10px;
}

.post-actions button {
  background-color: transparent;
  border: 1px solid #ddd;
  padding: 8px 12px;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.2s ease-in-out, color 0.2s ease-in-out;
}

.post-actions button:hover {
  background-color: #4caf50;
  color: white;
  border-color: #4caf50;
}

.post-header {
  display: flex;
  align-items: center;
  justify-content: space-between; /* Keeps avatar on left & text on right */
}

.post-header div {
  text-align: left; /* Aligns text inside the div to the right */
  flex-grow: 1; /* Pushes text to the right */
}

.comments-section {
  margin-top: 10px;
  padding: 10px;
  background: #f9f9f9;
  border-radius: 5px;
}

.comments-section ul {
  list-style: none;
  padding: 0;
}

.comments-section li {
  padding: 5px 0;
  border-bottom: 1px solid #ddd;
}

.post-image {
  max-width: 100%; /* Ensures the image doesn't overflow the container */
  height: auto; /* Maintains the aspect ratio */
  display: block; /* Prevents inline spacing issues */
  margin-top: 10px; /* Adds spacing from content */
  border-radius: 8px; /* Slightly rounded corners for a better look */
  object-fit: contain; /* Ensures the entire image is shown */
  max-height: 500px; /* Limits extremely tall images */
}
</style>
