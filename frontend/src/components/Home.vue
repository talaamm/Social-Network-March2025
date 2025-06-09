<template>
  <div class="container">
    <Navbar />

    <main class="content">
      <div class="post-box">
        <form>
          <textarea
            placeholder="What’s on your mind?"
            v-model="inputpost"
            required
          ></textarea>
          <p v-if="er">{{ er }}</p>
          <div class="post-actions">
          <label for="image">
            📷 Add Image
            <input name="image" type="file" @change="handleFileUpload" accept="image/*" />
          </label>
            <div class="privacy-options">
              <select v-model="privacypost">
                <option value="public">Public</option>
                <option value="followers">Followers</option>
                <option value="selected">Close Friends</option>
              </select>
              <button
                class="post-btn"
                type="submit"
                :disabled="inputpost.trim() === ''"
                @click.prevent="submitPost"
              >
                Post
              </button>
            </div>
          </div>
        </form>
      </div>
      <!-- 🔄 Refresh Feed Button -->
      <div class="refresh-container">
        <button class="refresh-btn" @click="fetchPosts">🔄 Refresh Feed</button>
      </div>
      <p v-if="error">{{ error }}</p>

      <!-- Posts Feed -->
      <div class="post-feed">
        <div v-for="post in posts" :key="post.id" class="post">
          <!--  @click="goToPost(post.id)" -->
          <div class="post-header">
            <!-- <div class="avatar"></div> -->
            <h3 style="cursor: pointer" @click="showProfile(post.user_id)">
              @{{ post.nickname }}
            </h3>
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
          <!-- Comments Section (Shows when selectedPostId matches post.id) -->
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
// import { useAuthStore } from "@/authStore";
// import { useRouter } from "vue-router";
import { ref, onMounted, computed } from "vue";
import axios from "axios";
import config from "@/config";
import Navbar from "@/components/Navbar.vue";
import { useRouter } from "vue-router";
import { eventBus } from "@/eventBus"; // Import the Event Bus

axios.defaults.withCredentials = true; // ✅ Ensures cookies are sent & received
const router = useRouter();
function showProfile(userId) {
  if (userId == loggedin_id){
    router.push('/my-profile')
  }else{
    router.push({ name: "UserProfile", params: { id: userId } });
  }
}
let loggedin_id = 0
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

let inputpost = ref("");
let er = ref("");
const privacypost = ref("public");
const posts = ref([]);
const selectedPostId = ref(0);
const comments = ref([]);
// const cmntLen = computed(() => comments.value.length || 0);
const cmntLen = computed(() =>
  Array.isArray(comments.value) ? comments.value.length : 0
);

const loadingComments = ref(false);
const newComment = ref("");

const likepost = async (postid) => {
  const post = posts.value.find((p) => p.id === postid);
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
  const post = posts.value.find((p) => p.id === postid);
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

const formatDate = (isoString) => {
  return new Date(isoString).toLocaleString();
};

const selectedFile = ref(null); // Store selected image file

const handleFileUpload = (event) => {
  selectedFile.value = event.target.files[0]; // Store the file
};

const submitPost = async () => {
  try {
    const formData = new FormData();
    formData.append("content", inputpost.value);
    formData.append("privacy", privacypost.value);
    if (selectedFile.value) {
      formData.append("image", selectedFile.value); // Only attach if user selected an image
    }
    let resp = await axios.post(`${config.API_URL}/api/posts`, formData, {
      "Content-Type": "multipart/form-data",
    });
    console.log(resp.data);
    // posts.value.push(resp.data)
    posts.value = [resp.data, ...posts.value];
    inputpost.value = "";
    selectedFile.value = "";
  } catch (error) {
    throw Error(error);
  }
};

const loadingPosts = ref(false);
const error = ref("");

axios.defaults.withCredentials = true;

// ✅ Fetch posts from the backend
const fetchPosts = async () => {
  try {
    loadingPosts.value = true;
    const response = await axios.get(`${config.API_URL}/all-posts`, {
      withCredentials: true,
    });
    posts.value = response.data ? response.data : [];
  } catch (err) {
    error.value = "Failed to load posts.";
    console.error("Error fetching posts:", err);
  } finally {
    loadingPosts.value = false;
  }
};

const toggleComments = async (postId) => {
  if (selectedPostId.value === postId) {
    comments.value = [];

    // If comments are already shown, hide them
    selectedPostId.value = 0;
    return;
  }

  selectedPostId.value = postId;
  loadingComments.value = true;

  try {
    const response = await axios.get(`${config.API_URL}/api/comments?post_id=${postId}`);
    comments.value = response.data;
    console.log(response.data);
  } catch (error) {
    console.error("Error fetching comments:", error);
    comments.value = [];
  }

  loadingComments.value = false;
};

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

    // comments.value.push(response.data); // ✅ Instantly add the new comment
    comments.value = [...comments.value, response.data];
    cmntLen.value = comments.value.length;
    console.log("Updated Comments:", comments.value); // Debugging
  } catch (error) {
    console.error("Error posting comment:", error);
    throw Error(error);
  }
};

onMounted(async () => {
  await fetcCurrUserData()
  await fetchPosts();

    // ✅ Listen for real-time post updates
    eventBus.on("postUpdate", (data) => {
    console.log("🔄 Received Post Update:", data);

    const postIndex = posts.value.findIndex((p) => p.id === data.post_id);
    if (postIndex !== -1) {
      posts.value[postIndex] = {
        ...posts.value[postIndex],
        likes: data.likes,
        dislikes: data.dislikes,
      };
    }
  });
});
</script>

<style scoped>

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
  width: 95%;
  padding: 10px;
  border: 1px solid #ccc;
  border-radius: 4px;
  resize: vertical;
  /* Allow vertical resizing */
  min-height: 100px;
  
  /* width: 100%; */
  padding: 10px;

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

.refresh-container {
  text-align: center;
  margin-bottom: 20px;
}

.refresh-btn {
  background-color: #007bff;
  color: white;
  font-size: 16px;
  padding: 10px 15px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  transition: background 0.3s;
}

.refresh-btn:hover {
  background-color: #0056b3;
}

.loading {
  text-align: center;
  font-size: 16px;
  color: gray;
  margin-top: 10px;
}

textarea {
  width: 100%;
  height: 80px;
  padding: 10px;
}

</style>
