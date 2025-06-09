<template>
  <div class="container">
    <Navbar />

    <div>
      <div class="follow-requests">
        <h3>Follow Requests</h3>

        <!-- If there are no requests -->
        <div id="nof" v-if="requests.length === 0"><p>No follow requests yet.</p></div>
        <div v-else v-for="request in requests" :key="request.id" class="request-card">
          <p>
            <strong>@{{ request.follower_nickname }}</strong> has requested to follow you.
          </p>
          <button
            class="accept-btn"
            @click="handleFollowRequest(request.id, 'accepted', request.follower_id)"
          >
            ✅ Accept
          </button>
          <button
            class="reject-btn"
            @click="handleFollowRequest(request.id, 'rejected', request.follower_id)"
          >
            ❌ Reject
          </button>
        </div>
      </div>

      <div class="group-join-requests">
        <h3>Group Join Requests</h3>

        <!-- If there are no requests -->
        <div v-if="Grequests.length === 0">
          <p>No pending group join requests.</p>
        </div>

        <div v-else v-for="request in Grequests" :key="request.id" class="request-card">
          <p>
            <strong>@{{ request.nickname }}</strong> has requested to join your group:
            <strong>{{ request.group_name }}</strong>
          </p>
          <button
            class="accept-btn"
            @click="handleMembershipRequest(request, 'approved')"
          >
            ✅ Accept
          </button>
          <button
            class="reject-btn"
            @click="handleMembershipRequest(request, 'rejected')"
          >
            ❌ Reject
          </button>
        </div>
      </div>

      <!-- 🔹 Group Invitations (New) -->
      <div class="group-invitations">
        <h3>Group Invitations</h3>
        <div v-if="invitations.length === 0">
          <p>No invitations to join a group.</p>
        </div>
        <div
          v-else
          v-for="invitation in invitations"
          :key="invitation.id"
          class="request-card"
        >
          <p>
            <strong>@{{ invitation.inviter_nickname }}</strong> invited you to join
            <strong>{{ invitation.group_name }}</strong>
          </p>
          <button class="accept-btn" @click="handleInvitation(invitation, 'accept')">
            ✅ Accept
          </button>
          <button class="reject-btn" @click="handleInvitation(invitation, 'reject')">
            ❌ Reject
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

const requests = ref([]); // Store follow requests
axios.defaults.withCredentials = true;
const invitations = ref([]); // Group invitations

const Grequests = ref([]); // Store group join requests

// Fetch group join requests where the user is the group creator
const fetchGroupJoinRequests = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/groups/pending-requests`);
    Grequests.value = response.data || [];
  } catch (error) {
    console.error("❌ Error fetching group join requests:", error);
  }
};

// ✅ Fetch group invitations
const fetchGroupInvitations = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/groups/invitations`);
    invitations.value = response.data || [];
  } catch (error) {
    console.error("❌ Error fetching group invitations:", error);
  }
};

// ✅ Accept or Reject Group Invitation
const handleInvitation = async (invitation, action) => {
  const endpoint = action === "accept"
    ? `${config.API_URL}/api/groups/accept-invitation?group_id=${invitation.group_id}`
    : `${config.API_URL}/api/groups/reject-invitation?group_id=${invitation.group_id}`;

  try {
    await axios.post(endpoint);
    invitations.value = invitations.value.filter(inv => inv.group_id !== invitation.group_id);
  } catch (error) {
    console.error(`❌ Error ${action}ing invitation:`, error);
  }
};
const handleMembershipRequest = async (request, status) => {
  const endpoint =
    status === "approved"
      ? `${config.API_URL}/api/groups/approve?group_id=${request.group_id}&user_id=${request.id}`
      : `${config.API_URL}/api/groups/reject?group_id=${request.group_id}&user_id=${request.id}`;

  try {
    await axios.post(endpoint);

    // ✅ Debugging: Check before filtering
    console.log("Before Removing:", Grequests.value);

    // ✅ Ensure request is removed in real-time
    Grequests.value =
      Grequests.value.filter(
        (req) => !(req.group_id === request.group_id && req.id === request.id)
      ) || [];

    // ✅ Debugging: Check after filtering
    console.log("After Removing:", Grequests.value);
  } catch (error) {
    console.error(`❌ Error updating membership request (${status}):`, error);
  }
};

// Fetch follow requests
const fetchFollowRequests = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/follow-requests`, {
      withCredentials: true,
    });
    requests.value = response.data || [];
  } catch (error) {
    console.error("❌ Error fetching follow requests:", error);
  }
};

// Accept or Reject a follow request
const handleFollowRequest = async (requestId, status, followerId) => {
  try {
    await axios.post(
      `${config.API_URL}/api/update-follow-request`,
      { requestId, status, followerId },
      { withCredentials: true }
    );

    // ✅ Remove request from UI in real-time
    requests.value = requests.value.filter((request) => request.id !== requestId);
  } catch (error) {
    console.error(`❌ Error updating follow request (${status}):`, error);
  }
};

onMounted(() => {
  fetchFollowRequests();
  fetchGroupJoinRequests();
  fetchGroupInvitations();
});
</script>

<style scoped>
.follow-requests {
  padding: 1rem;
  background-color: #fff;
  /* align-items: center; */
  width: 400px;
  border-radius: 10px;
  /* box-shadow: 0px 2px 10px rgba(0, 0, 0, 0.1); */
  max-width: 800px;
}

.request-card {
  padding: 1rem;
  border: 1px solid #ddd;
  border-radius: 8px;
  margin: 10px 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.3125rem;
}

.accept-btn {
  background-color: #4caf50;
  color: white;
  border: none;
  padding: 2px 5px;
  cursor: pointer;
}

.reject-btn {
  background-color: #ff4c4c;
  color: white;
  border: none;
  padding: 5px 10px;
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
.nof {
  box-shadow: none;
}
.group-join-requests {
  max-width: 600px;
  margin: 20px auto;
  padding: 20px;
  background: #f9f9f9;
  border-radius: 8px;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
}

.request-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  margin-top: 10px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 5px;
}

.accept-btn {
  background-color: #44cc49;
  color: white;
  border: none;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
}

.reject-btn {
  background-color: #e63434;
  color: white;
  border: none;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
}

.accept-btn:hover {
  background-color: #3aa93f;
}

.reject-btn:hover {
  background-color: #c92a2a;
}
</style>
