<template>
  <div class="container">
    <Navbar />

    <main class="content">
      <!-- User Info Section -->
      <div class="user-info-box">
        <div class="user-avatar">{{ userData.nickname }}</div>
        <div class="user-details">
          <h2>{{ userData.nickname }}</h2>
          <p>{{ userData.first_name }} {{ userData.last_name }}</p>
          <p>Email: {{ userData.email }}</p>
          <p>Age: {{ userData.age || "N/A" }}</p>
          <p>Gender: {{ userData.gender }}</p>
          <p>Birthdate: {{ userData.dbirth }}</p>
          <p v-if="userData.isPrivate || privateP == 'private'">
            Profile: Private
          </p>
          <p v-else>Profile: Public</p>

          <select v-model="privateP">
            <option value="public">Public</option>
            <option value="private">Private</option>
          </select>
          <div class="follow-buttons">
            <button @click="fetchFollowers">
              Followers ({{ followersLen }})
            </button>
            <button @click="fetchFollowing">
              Following ({{ followingLen }})
            </button>
            <button @click="fetchSelectedUsers">Manage Close Friends</button>
          </div>

        <!-- ✅ Close Friends (Selected Users) Modal -->
        <div v-if="showSelectedUsers" class="modal">
        <div class="modal-content">
          <h3>Manage Close Friends</h3>
          <ul v-if="followers.length > 0">
            <li v-for="user in followers" :key="user.id" class="selected-user-item">
              <label>
                <input
                  type="checkbox"
                  :checked="isSelected(user.id)"
                  @change="toggleSelectedUser(user.id)"
                />
                <strong> @{{ user.nickname }} </strong>
              </label>
            </li>
          </ul>
          <p v-else>No followers to select.</p>
          <button @click="saveSelectedUsers">Save Changes</button>
          <button @click="showSelectedUsers = false">Close</button>
        </div>
      </div>
          <!-- Followers List -->
          <div v-if="showFollowers" class="modal">
            <div class="modal-content">
              <h3>Followers</h3>
              <ul v-if="followers.length > 0">
                <li
                  @click="showProfile(user.id)"
                  style="text-align: left; cursor: pointer"
                  v-for="user in followers"
                  :key="user.id"
                >
                  <strong> {{ user.first_name }} {{ user.last_name }} </strong
                  >(@{{ user.nickname }})
                </li>
              </ul>
              <p v-else>No followers yet.</p>
              <!-- ✅ Graceful handling -->
              <button @click="showFollowers = false">Close</button>
            </div>
          </div>

          <!-- Following List -->
          <div v-if="showFollowing" class="modal">
            <div class="modal-content">
              <h3>Following</h3>
              <ul v-if="following.length > 0">
                <li
                  @click="showProfile(user.id)"
                  style="text-align: left; cursor: pointer"
                  v-for="user in following"
                  :key="user.id"
                >
                  <strong> {{ user.first_name }} {{ user.last_name }} </strong>
                  (@{{ user.nickname }})
                  <span v-if="user.status === 'pending'"> - Request Sent</span>
                </li>
              </ul>
              <p v-else>You are not following anyone yet.</p>
              <!-- ✅ Graceful handling -->
              <button @click="showFollowing = false">Close</button>
            </div>
          </div>
        </div>
      </div>

      <!-- User Posts Section -->
      <div class="post-feed">
        <h3>{{ userData.nickname }}'s Posts</h3>
        <div v-for="post in userPosts" :key="post.id" class="post">
          <div class="post-header">
            <!-- <div class="avatar"></div> -->
            <h3>@{{ post.nickname }}</h3>
            <p v-if="post.privacy == 'public'">Privacy: Public</p>
            <p v-else-if="post.privacy == 'selected'">Privacy: Close Friends</p>
            <p v-else-if="post.privacy == 'followers'">
              Privacy: Followers Only
            </p>
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
            <button @click="dislikepost(post.id)">
              👎 {{ post.dislikes }}
            </button>
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
              <input
                v-model="newComment"
                type="text"
                placeholder="Write a comment..."
              />
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
import { useRouter } from "vue-router";

const router = useRouter();

const privateP = ref("");
const userData = ref({});
const userPosts = ref([]);
const selectedPostId = ref(null);
const comments = ref([]);
const loadingComments = ref(false);
const newComment = ref("");
const selectedUsers = ref([]);
const showSelectedUsers = ref(false);

// const followers = ref([]); // ✅ Stores followers list
let loggedin_id = 0;
axios.defaults.withCredentials = true;

watch(privateP, async (newValue, oldValue) => {
  console.log(`Privacy changed from ${oldValue} to ${newValue}`);

  // Convert "public"/"private" to boolean for backend
  const isPrivate = newValue === "private";
  try {
    const response = await axios.post(
      `${config.API_URL}/api/private`,
      { isprivate: isPrivate },
      { withCredentials: true }
    );

    if (response.status === 200) {
      console.log("Privacy setting updated successfully");
    } else if (response.status === 304) {
      console.log("No update needed (same value)");
    }
  } catch (error) {
    console.error("Error updating privacy setting:", error);
  }
});

async function fetcCurrUserData() {
  try {
    const response = await axios.get(`${config.API_URL}/api/myself`, {
      withCredentials: true,
    });
    console.log(response.data);
    loggedin_id = response.data.id;
    privateP.value = response.data.isprivate ? "private" : "public";

    console.log(loggedin_id);
  } catch (error) {
    console.error("Error fetching user data:", error);
  }
}

const fetchUserData = async () => {
  try {
    const response = await axios.get(
      `${config.API_URL}/api/user-data?user_id=${loggedin_id}`,
      {
        withCredentials: true,
      }
    );
    // privateP = response.data.isprivate
    userData.value = response.data;
  } catch (error) {
    console.error("Error fetching user data:", error);
  }
};

const fetchUserPosts = async () => {
  try {
    const response = await axios.get(
      `${config.API_URL}/api/user-posts?user_id=${loggedin_id}`,
      {
        withCredentials: true,
      }
    );
    userPosts.value = response.data;
  } catch (error) {
    console.error("Error fetching user posts:", error);
  }
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
    const response = await axios.get(
      `${config.API_URL}/api/comments?post_id=${postId}`
    );
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
  const post = userPosts.value.find((p) => p.id === postid);
  if (!post) return;
  try {
    const response = await axios.post(`${config.API_URL}/api/like`, {
      postid: post.id,
      islike: true, // Liking the post
    });
    console.log(response.data);
    // ✅ Update the post with the new like/dislike count from API
    post.likes = response.data.likes ?? post.likes;
    post.dislikes = response.data.dislikes ?? post.dislikes;
  } catch (error) {
    console.error("Error liking post:", error);
  }
};
const dislikepost = async (postid) => {
  const post = userPosts.value.find((p) => p.id === postid);
  if (!post) return;
  try {
    const response = await axios.post(`${config.API_URL}/api/like`, {
      postid: post.id,
      islike: false, // Disliking the post
    });

    // ✅ Update the post with the new like/dislike count from API
    post.likes = response.data.likes ?? post.likes;
    post.dislikes = response.data.dislikes ?? post.dislikes;
  } catch (error) {
    console.error("Error disliking post:", error);
  }
};

const followers = ref([]);
const following = ref([]);
const showFollowers = ref(false);
const showFollowing = ref(false);
const followersLen = ref(0);
const followingLen = ref(0);

// ✅ Fetch Followers & Selected Users
const fetchSelectedUsers = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/selected-users`, {
      withCredentials: true,
    });

    selectedUsers.value = response.data? response.data.map(user => user.user_id) : []; // ✅ Store only user_id
    await fetchFollowers(); // ✅ Fetch followers for selection list
    showSelectedUsers.value = true;
    showFollowers.value = false
  } catch (error) {
    console.error("❌ Error fetching selected users:", error);
  }
};

// ✅ Check if a follower is already selected (for pre-checking the box)
const isSelected = (followerID) => {
  return selectedUsers.value.includes(followerID);
};

// ✅ Toggle Selection of a Follower LATEST PUSH WITH IS SELECTED
const toggleSelectedUser = (followerID) => {
  if (isSelected(followerID)) {
    selectedUsers.value = selectedUsers.value.filter((id) => id !== followerID);
  } else {
    selectedUsers.value.push(followerID);
  }
};

const saveSelectedUsers = async () => {
  try {
    await axios.post(
      `${config.API_URL}/api/update-selected-users`,
      { user_ids: selectedUsers.value },  // ✅ Send an array instead of a single user_id
      { withCredentials: true }
    );
    showSelectedUsers.value = false;
    console.log("✅ Selected users updated successfully!");
  } catch (error) {
    console.error("❌ Error saving selected users:", error);
  }
};


// Fetch counts on page load
const fetchFollowCounts = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/follow-counts`, {
      withCredentials: true,
    });
    followersLen.value = response.data.followers || 0;
    followingLen.value = response.data.following || 0;
  } catch (error) {
    console.error("Error fetching follow counts:", error);
    followersLen.value = 0;
    followingLen.value = 0;
  }
};

// Fetch full list when clicking the button
const fetchFollowers = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/followers`, {
      withCredentials: true,
    });
    followers.value = response.data || [];
    showFollowers.value = true;
  } catch (error) {
    console.error("Error fetching followers:", error);
    followers.value = [];
  }
};

const fetchFollowing = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/following`, {
      withCredentials: true,
    });
    following.value = response.data || [];
    showFollowing.value = true;
  } catch (error) {
    console.error("Error fetching following:", error);
    following.value = [];
  }
};
function showProfile(userId) {
  if (userId == loggedin_id) {
    router.push("/my-profile");
  } else {
    router.push({ name: "UserProfile", params: { id: userId } });
  }
}

onMounted(async () => {
  await fetcCurrUserData();
  fetchUserData();
  await fetchUserPosts();
  fetchFollowCounts();
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

.follow-buttons {
  margin-top: 20px;
  display: flex;
  gap: 10px;
}

.follow-buttons button {
  padding: 8px 12px;
  border: none;
  background-color: #007bff;
  color: white;
  border-radius: 5px;
  cursor: pointer;
  transition: 0.3s;
}

.follow-buttons button:hover {
  background-color: #0056b3;
}

.modal {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 4px 10px rgba(0, 0, 0, 0.2);
  width: 300px;
}

.modal-content {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.modal-content h3 {
  margin-bottom: 15px;
}

.modal-content ul {
  list-style: none;
  padding: 0;
}

.modal-content li {
  margin-bottom: 8px;
}

.modal-content button {
  margin-top: 10px;
  padding: 8px 12px;
  border: none;
  background-color: #dc3545;
  color: white;
  border-radius: 5px;
  cursor: pointer;
}

.modal-content button:hover {
  background-color: #c82333;
}
.selected-users-section {
  margin-top: 20px;
  padding: 15px;
  border: 1px solid #ddd;
  border-radius: 5px;
  background: #f9f9f9;
}

.selected-user-item {
  display: flex;
  align-items: center;
  margin: 5px 0;
}

.selected-user-item label {
  cursor: pointer;
}

.selected-user-item input {
  margin-right: 10px;
}
.modal-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.modal button {
  margin-top: 10px;
}

/* ✅ Checkbox List Styling */
.selected-user-item {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
