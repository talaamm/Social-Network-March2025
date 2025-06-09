<template>
  <div class="container">
    <Navbar />
    <div class="group-container">
      <header class="group-header">
        <h1>{{ groupName }}</h1>
        <h5 style="text-align: left">Group Description: {{ description }}</h5>
        <h5 style="text-align: left">Creator: @{{ creatorName }}</h5>
      </header>

      <!-- Post Creation Box -->
      <div class="post-box">
        <textarea
          v-model="newPost"
          placeholder="What’s on your mind?"
        ></textarea>
        <label class="add-image">
          📷 Add Image
          <input type="file" @change="handleFileUpload" accept="image/*" />
        </label>
        <button class="post-btn" @click.prevent="submitPost">Post</button>
      </div>

      <!-- Group Actions -->
      <div class="group-actions">
        <button
          style="background-color: #44cc49"
          @click="fetchFollowersToInvite"
        >
          ➕ Invite a Friend
        </button>
        <button @click="showCreateEventModal = true">📅 Make an Event</button>

        <!-- Members Button -->
        <button @click="fetchGroupMembers">
          👥 Members ({{ member_count }})
        </button>

        <button
          v-if="loggedin_id == creatorid"
          style="background-color: #e63434"
          class="leave-btn"
          @click="leaveGroup"
        >
          ❌ Leave & Delete Group
        </button>
        <button
          v-else
          style="background-color: #e63434"
          class="leave-btn"
          @click="leaveGroup"
        >
          🚪 Leave Group
        </button>
      </div>

      <!-- 🔵 Create Event Modal -->
      <div v-if="showCreateEventModal" class="events-modal">
        <div class="modal-content">
          <h3>Create a New Event</h3>
          <input v-model="newEvent.title" placeholder="Event Title" />
          <textarea
            v-model="newEvent.description"
            placeholder="Event Description"
          ></textarea>
          <input type="datetime-local" v-model="newEvent.event_date" />
          <div class="modal-actions">
            <button class="create-btn" @click="createEvent">✅ Create</button>
            <button class="cancel-btn" @click="showCreateEventModal = false">
              ❌ Cancel
            </button>
          </div>
        </div>
      </div>

      <!-- 📌 Members Modal -->
      <div v-if="showMembersModal" class="modal">
        <div class="modal-content">
          <h3>Group Members</h3>
          <ul v-if="members.length > 0">
            <li
              v-for="member in members"
              :key="member.id"
              @click="showProfile(member.user_id)"
              class="member-item"
            >
              <strong v-if="creatorName == member.nickname">
                👑 (@{{ member.nickname }})</strong
              >
              <strong v-else> (@{{ member.nickname }})</strong>
            </li>
          </ul>
          <p v-else>No members yet.</p>
          <button @click="showMembersModal = false">Close</button>
        </div>
      </div>

      <!-- 📌 Invite Friends Modal -->
      <div v-if="showInviteModal" class="modal">
        <div class="modal-content">
          <h3>Invite a Friend to Group</h3>
          <ul v-if="followersToInvite.length > 0">
            <li v-for="user in followersToInvite" :key="user.id">
              <strong>@{{ user.nickname }}</strong>
              <button @click="inviteUser(user.id)" :disabled="user.invited">
                {{ user.invited ? "✅ Invited" : "Invite" }}
              </button>
            </li>
          </ul>
          <p v-else>No followers available for invitation.</p>
          <button @click="showInviteModal = false">Close</button>
        </div>
      </div>

      <!-- 🟠 Events Section -->
      <div class="events-section">
        <h2>Upcoming Events</h2>
        <div v-if="events.length === 0">No events yet.</div>
        <div v-for="event in events" :key="event.id" class="event-card">
          <h3>@{{ event.creator }}: {{ event.title }}</h3>
          <p>{{ event.description }}</p>
          <p>📅 Date: {{ formatDate(event.event_date) }}</p>
          <!-- <p>👥 Going: {{ event.going }}</p> -->
          <button
            @click="rsvpToEvent(event.id, 'going')"
            :class="{ active: event.user_rsvp === 'going' }"
          >
            ✅ Going {{ event.going }}
          </button>
          <button
            @click="rsvpToEvent(event.id, 'not going')"
            :class="{ active: event.user_rsvp === 'not going' }"
          >
            ❌ Not Going {{ event.not_going }}
          </button>
        </div>
      </div>

      <!-- Posts Feed -->
      <div class="post-feed">
        <div v-for="post in posts" :key="post.id" class="post">
          <div class="post-header">
            <h3
              style="cursor: pointer"
              @click.prevent="showProfile(post.member_id)"
            >
              @{{ post.nickname }}
            </h3>
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
            <button @click="likePost(post.id, true)">
              👍 {{ post.likes }}
            </button>
            <button @click="toggleComments(post.id)">💬 Comments</button>
            <button @click="likePost(post.id, false)">
              👎 {{ post.dislikes }}
            </button>
          </div>
          <div
            v-if="selectedPostId !== null && selectedPostId === post.id"
            class="comments-section"
          >
            <!-- <div v-if="selectedPostId === post.id" class="comments-section"> -->
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
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from "vue";
import axios from "axios";
import config from "@/config";
import Navbar from "@/components/Navbar.vue";
import { useRouter, useRoute } from "vue-router";
import { eventBus } from "@/eventBus";

axios.defaults.withCredentials = true;
const router = useRouter();
const route = useRoute();
let member_count = ref(0);
const groupId = ref(parseInt(route.params.groupid));
const groupName = ref("");
const newPost = ref("");
const newComment = ref("");
const posts = ref([]);
const comments = ref([]);
const selectedPostId = ref(0);
const selectedFile = ref(null);
const loggedin_id = ref(0);
const loadingPosts = ref(false);
const loadingComments = ref(false);
const cmntLen = computed(() =>
  Array.isArray(comments.value) ? comments.value.length : 0
);
const formatDate = (isoString) => new Date(isoString).toLocaleString();
const members = ref([]); // Store group members
const showMembersModal = ref(false); // Controls modal visibility

const handleFileUpload = (event) => {
  selectedFile.value = event.target.files[0]; // Store the file
};

async function fetchCurrUserData() {
  try {
    const response = await axios.get(`${config.API_URL}/api/myself`);
    loggedin_id.value = response.data.id;
  } catch (error) {
    console.error("Error fetching user data:", error);
  }
}
let description = ref("");
let creatorid = 0;
let creatorName = ref("");

async function fetchGroupDetails() {
  try {
    const response = await axios.get(
      `${config.API_URL}/api/groups?group_id=${groupId.value}`
    );
    groupName.value = response.data.name;
    description.value = response.data.description;
    creatorid = response.data.creator_id;
    creatorName = response.data.creator_nickname;
    member_count.value = response.data.member_count || 0;
  } catch (error) {
    console.error("Error fetching group details:", error);
  }
}

async function fetchPosts() {
  try {
    loadingPosts.value = true;
    const response = await axios.get(
      `${config.API_URL}/api/groups/posts?group_id=${groupId.value}`
    );
    posts.value = response.data ? response.data : [];
  } catch (error) {
    console.error("Error fetching posts:", error);
  } finally {
    loadingPosts.value = false;
  }
}

async function submitPost() {
  try {
    const formData = new FormData();
    formData.append("content", newPost.value);
    formData.append("group_id", groupId.value);
    if (selectedFile.value) {
      formData.append("image", selectedFile.value);
      console.log("📸 Selected File:", selectedFile.value);
    }
    let resp = await axios.post(`${config.API_URL}/api/groups/posts`, formData); // Let Axios auto-set Content-Type
    console.log("new post:", resp.data);

    // posts.value = [resp.data, ...posts.value];
    posts.value = [resp.data, ...posts.value];
    newPost.value = "";
    selectedFile.value = null; // ✅ Ensure the file input resets properly
    selectedPostId.value = null; // ✅ Make sure no comments are shown
    comments.value = []; // ✅ Remove all old comments
    cmntLen.value = 0;
    // fetchPosts();
  } catch (error) {
    console.error("Error submitting post:", error);
  }
}

async function likePost(postId, isLike) {
  const post = posts.value.find((p) => p.id === postId);
  if (!post) return;
  try {
    let response = await axios.post(`${config.API_URL}/api/groups/like`, {
      postid: postId,
      islike: isLike,
    });
    post.likes = response.data.likes ?? post.likes;
    post.dislikes = response.data.dislikes ?? post.dislikes;
  } catch (error) {
    console.error("Error liking post:", error);
  }
}

async function toggleComments(postId) {
  if (selectedPostId.value === postId) {
    comments.value = [];
    cmntLen.value = 0;
    selectedPostId.value = null;
    return;
  }
  selectedPostId.value = postId;
  loadingComments.value = true;
  try {
    const response = await axios.get(
      `${config.API_URL}/api/groups/comments?post_id=${postId}`
    );
    comments.value = response.data ? response.data : [];
    cmntLen.value = response.data ? response.data.length : 0;
  } catch (error) {
    console.error("Error fetching comments:", error);
    comments.value = [];
  }
  loadingComments.value = false;
}

async function postComment(postId) {
  if (!newComment.value.trim()) return;
  try {
    let response = await axios.post(`${config.API_URL}/api/groups/comments`, {
      g_post_id: postId,
      content: newComment.value,
    });
    newComment.value = "";
    // newComment.value = "";
    if (!Array.isArray(comments.value)) {
      comments.value = [];
    }
    comments.value = [...comments.value, response.data];
    cmntLen.value = comments.value.length;
  } catch (error) {
    console.error("Error posting comment:", error);
  }
}

const showInviteModal = ref(false);
const followersToInvite = ref([]);

// ✅ Fetch Followers Who Can Be Invited
const fetchFollowersToInvite = async () => {
  try {
    const response = await axios.get(
      `${config.API_URL}/api/followers-to-invite?group_id=${groupId.value}`
    );
    followersToInvite.value = response.data
      ? response.data.map((user) => ({
          ...user,
          invited: false, // Track if invited
        }))
      : [];
    showInviteModal.value = true;
  } catch (error) {
    console.error("Error fetching followers:", error);
  }
};

// ✅ Send Invitation
const inviteUser = async (userId) => {
  try {
    await axios.post(
      `${config.API_URL}/api/groups/invite?group_id=${groupId.value}&invited_user_id=${userId}`
    );
    const user = followersToInvite.value.find((user) => user.id === userId);
    if (user) user.invited = true; // Mark user as invited in UI
  } catch (error) {
    console.error("Error inviting user:", error);
  }
};

const leaveGroup = async () => {
  console.log("Leave Group Clicked");

  try {
    let response = await axios.post(
      `${config.API_URL}/api/groups/leave?group_id=${groupId.value}`
    );
    ///api/groups/leave?group_id=
    console.log("user left the group", response.data);
    router.push("/my-groups");
    // router.push("/home");
  } catch (error) {
    throw Error("error ehile leaving ", error);
  }
};

// ✅ Fetch Group Members
const fetchGroupMembers = async () => {
  try {
    const response = await axios.get(
      `${config.API_URL}/api/groups/members?group_id=${groupId.value}`
    );
    members.value = response.data || [];
    showMembersModal.value = true; // Show modal
  } catch (error) {
    console.error("Error fetching group members:", error);
  }
};

function showProfile(userId) {
  console.log(
    "current logged in",
    loggedin_id.value,
    "userid to fetch:",
    userId
  );
  if (userId === loggedin_id.value) {
    router.push("/my-profile");
  } else {
    router.push({ name: "UserProfile", params: { id: userId } });
  }
}

const showCreateEventModal = ref(false);
const newEvent = ref({ title: "", description: "", event_date: "" });
const events = ref([]);

// ✅ Fetch Group Events
const fetchEvents = async () => {
  try {
    const response = await axios.get(
      `${config.API_URL}/api/groups/events?group_id=${groupId.value}`
    );
    events.value = response.data || [];
  } catch (error) {
    console.error("Error fetching events:", error);
  }
};

// ✅ Create a New Event
const createEvent = async () => {
  if (
    !newEvent.value.title ||
    !newEvent.value.description ||
    !newEvent.value.event_date
  ) {
    alert("Please fill in all event details!");
    return;
  }
  try {
    await axios.post(`${config.API_URL}/api/groups/events`, {
      group_id: Number(groupId.value),
      title: newEvent.value.title,
      description: newEvent.value.description,
      event_date: newEvent.value.event_date,
    });
    showCreateEventModal.value = false;
    newEvent.value = { title: "", description: "", event_date: "" };
    fetchEvents();
  } catch (error) {
    console.error("Error creating event:", error);
  }
};

// ✅ RSVP to Event
const rsvpToEvent = async (eventID, status) => {
  try {
    await axios.post(
      `${config.API_URL}/api/groups/events/rsvp?event_id=${eventID}`,
      { status }
    );
    fetchEvents(); // Refresh UI
  } catch (error) {
    console.error("Error RSVPing:", error);
  }
};

onMounted(async () => {
  // await fetcCurrUserData();
  if (route.params.groupid) {
    groupId.value = route.params.groupid; // ✅ Set groupID properly
  } else {
    console.error("❌ No group ID provided in route.");
  }
  await fetchCurrUserData();
  await fetchGroupDetails();
  await fetchPosts();
  await fetchEvents();

  // ✅ Listen for group post updates
  // eventBus.on("groupPostUpdate", (data) => {
  //   if (data.group_id === route.params.groupid) {
  //     console.log("📢 New Group Post:", data);
  //     let newPost = ({
  //       id: data.post_id,
  //       // group_id: data.group_id,
  //       nickname: data.author,
  //       content: data.content,
  //       created_at: data.timestamp,
  //       likes: 0,
  //       dislike: 0,
  //       member_id: data.member_id,
  //     });
  //     posts.value = [newPost , ...posts.value]
  //   }
  // });

  eventBus.on("groupPostUpdate", (data) => {
    if (data.group_id === Number(route.params.groupid)) {
      console.log("📢 New Group Post:", data);
      let newPost = {
        id: data.post_id,
        nickname: data.nickname,
        content: data.content,
        created_at: data.timestamp,
        likes: 0,
        dislike: 0,
        member_id: data.member_id,
      };
      if (loggedin_id !== Number(data.member_id)) posts.value.unshift(newPost); // ✅ This maintains reactivity
    }
  });

  eventBus.on("groupEvent", async (data) => {
    if (data.group_id === Number(route.params.groupid)) {
      console.log("📢 New Group Post:", data);
      // await fetchEvents();
      console.log("🔄 Calling fetchEvents()...");
      await fetchEvents();
      console.log("✅ fetchEvents() completed!");
    }
  });
});
</script>

<style scoped>
.group-container {
  width: auto;
  min-width: 600px;
  margin: auto;
  padding: 20px;
  /* background: #f9f9f9; */
  border-radius: 10px;
}

.group-header {
  text-align: center;
  font-size: 24px;
  font-weight: bold;
}

.post-box {
  background: white;
  padding: 15px;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.post-box textarea {
  width: 100%;
  height: 80px;
  padding: 10px;
}

.add-image {
  display: flex;
  align-items: center;
  cursor: pointer;
}

.group-actions {
  display: flex;
  justify-content: space-between;
  margin: 20px 0;
}

.leave-btn {
  background-color: red;
  color: white;
}

.posts-feed {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.post {
  background: white;
  padding: 15px;
  border-radius: 8px;
}

.post-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
}

.post-image {
  width: 100%;
  border-radius: 8px;
  margin-top: 10px;
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
  width: 600px;
  /* Ensure full viewport height */
}

.container {
  display: flex;
  /* width: 90%; */
  width: auto;
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

/* .modal {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0px 0px 10px rgba(0, 0, 0, 0.2);
  z-index: 1000;
  width: 300px;
}

.modal-content {
  text-align: center;
} */

.member-item {
  cursor: pointer;
  padding: 5px;
}

.member-item:hover {
  background: #f0f0f0;
}
button {
  cursor: pointer;
}

/* 🎨 Modal Styling */
.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-content {
  background: white;
  padding: 20px;
  border-radius: 8px;
  width: 300px;
  text-align: center;
}

.modal-content input,
.modal-content textarea {
  width: 100%;
  padding: 8px;
  margin: 8px 0;
  border: 1px solid #ccc;
  border-radius: 5px;
}

.modal-actions {
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
}

/* 🎨 General Modal Styling */
.events-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  /* position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: white;
  padding: 20px;
  border-radius: 12px;
  background: rgba(0, 0, 0, 0.5);
  box-shadow: 0px 4px 10px rgba(0, 0, 0, 0.2);
  z-index: 1000;
  width: 90%;
  max-width: 400px;
  text-align: center; */
}

/*.modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
} */

.events-modal input,
.events-modal textarea {
  width: 95%;
  margin: 10px 0;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 8px;
  font-size: 16px;
}

/* .modal-actions {
  display: flex;
  justify-content: space-around;
  margin-top: 15px;
} */

.create-btn,
.cancel-btn {
  padding: 10px 15px;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 16px;
}

.create-btn {
  background-color: #44cc49;
  color: white;
}

.cancel-btn {
  background-color: #e63434;
  color: white;
}

/* 🔵 Events Section Styling */
.events-section {
  margin-top: 30px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 12px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.events-section h2 {
  text-align: center;
  color: #333;
  margin-bottom: 15px;
}

/* 🎟️ Event Cards */
.event-card {
  background: white;
  padding: 15px;
  margin: 15px 0;
  border-radius: 12px;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s, box-shadow 0.2s;
}

.event-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 6px 12px rgba(0, 0, 0, 0.15);
}

.event-card h3 {
  color: #007bff;
  margin-bottom: 8px;
}

.event-card p {
  color: #555;
  font-size: 14px;
  margin-bottom: 5px;
}

/* ✅ RSVP Buttons */
.event-card button {
  padding: 8px 15px;
  margin: 5px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  cursor: pointer;
  transition: background 0.3s, transform 0.2s;
}

.event-card button:hover {
  transform: scale(1.05);
}

/* 🟢 Going Button */
.event-card button:nth-child(4) {
  background-color: #28a745;
  color: white;
}

/* ❌ Not Going Button */
.event-card button:nth-child(5) {
  background-color: #dc3545;
  color: white;
}

/* 🎯 Selected RSVP Button */
.event-card button.active {
  background-color: #ffcc00;
  color: black;
  font-weight: bold;
}
</style>
