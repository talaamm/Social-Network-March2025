<template>
  <div class="container">
    <Navbar />
<div class="group-container">
    <h1>My Groups</h1>

    <!-- 🏆 Groups Created by the User -->
    <div class="section">
      <h2>Groups I Created</h2>
       <!-- ➕ Create Group Button -->
       <button class="create-group-btn" @click="showCreateModal = true">➕ New Group</button>

      <div v-if="createdGroups.length === 0" class="no-groups">
        You haven't created any groups yet.
      </div>
      <div v-else class="groups-list">
        <div
          v-for="group in createdGroups"
          :key="group.id"
          class="group-card"
          @click="goToGroup(group.id)"
        >
          <h3>{{ group.name }}</h3>
          <p>{{ group.description }}</p>
          <p>👤 Creator: You</p>
          <p>👥 Members: {{ group.member_count }}</p>
        </div>
      </div>
    </div>

    <!-- 👥 Groups the User is a Member In -->
    <div class="section">
      <h2>Groups I'm a Member In</h2>
      <div v-if="memberGroups.length === 0" class="no-groups">
        You are not a member of any groups.
      </div>
      <div v-else class="groups-list">
        <div
          v-for="group in memberGroups"
          :key="group.id"
          class="group-card"
          @click="goToGroup(group.id)"
        >
          <h3>{{ group.name }}</h3>
          <p>{{ group.description }}</p>
          <p>👤 Creator: @{{ group.creator_nickname }}</p>
          <p>👥 Members: {{ group.member_count }}</p>
        </div>
      </div>
    </div>
    </div>
      <!-- 📌 Create Group Modal -->
      <div v-if="showCreateModal" class="modal">
      <div class="modal-content">
        <h3>Create a New Group</h3>
        <input v-model="newGroupName" placeholder="Group Name" />
        <textarea v-model="newGroupDescription" placeholder="Group Description"></textarea>
        <div class="modal-actions">
          <button class="create-btn" @click="createGroup">✅ Create</button>
          <button class="cancel-btn" @click="showCreateModal = false">❌ Cancel</button>
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

const router = useRouter();
const createdGroups = ref([]);
const memberGroups = ref([]);

// ✅ Fetch Groups Created by the User
const fetchCreatedGroups = async () => {
//   if (!userId.value) return; // Ensure user ID is available

  try {
    const response = await axios.get(
      `${config.API_URL}/api/groups-created-by-me`
    );
    createdGroups.value = response.data || [];
  } catch (error) {
    console.error("❌ Error fetching created groups:", error);
  }
};

// ✅ Fetch Groups the User is a Member Of
const fetchMemberGroups = async () => {
  try {
    const response = await axios.get(`${config.API_URL}/api/my-groups`);
    memberGroups.value = response.data || [];
    console.log(response.data)
  } catch (error) {
    console.error("❌ Error fetching member groups:", error);
  }
};

// ✅ Redirect to Group Page
const goToGroup = (groupId) => {
  router.push(`/groups/${groupId}`);
};

const showCreateModal = ref(false);
const newGroupName = ref("");
const newGroupDescription = ref("");

const createGroup = async () => {
  if (!newGroupName.value.trim() || !newGroupDescription.value.trim()) {
    alert("Please fill in both Group Name and Description.");
    return;
  }

  try {
    const response = await axios.post(`${config.API_URL}/api/groups`, {
      name: newGroupName.value,
      description: newGroupDescription.value,
    });

    console.log("Group Created:", response.data);

    let new_group = ({
      id: response.data.group_id, // Assuming backend returns the new group ID
      name: newGroupName.value,
      description: newGroupDescription.value,
      member_count: 1, // Creator is the first member
    });

    createdGroups.value = [new_group , ...createdGroups.value]
    newGroupName.value = "";
    newGroupDescription.value = "";
    showCreateModal.value = false;
  } catch (error) {
    console.error("Error creating group:", error);
  }
};

onMounted(async () => {
  // await fetcCurrUserData();
  await fetchCreatedGroups();
  await fetchMemberGroups();
});
</script>

<style scoped>

h1 {
  text-align: center;
  margin-bottom: 20px;
}

.section {
  margin-bottom: 30px;
}

.groups-list {
  display: flex;
  flex-wrap: wrap;
  gap: 15px;
}

.group-card {
  background: white;
  padding: 15px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: transform 0.2s ease-in-out;
  width: 100%;
}

.group-card:hover {
  transform: scale(1.05);
}

.no-groups {
  color: gray;
  text-align: center;
}
.group-container {
  width: 500px;
  margin: auto;
  padding: 20px;
  /* background: #f9f9f9; */
  border-radius: 10px;
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
/* 🏆 Create Group Button */
.create-group-btn {
  background-color: #28a745;
  color: white;
  padding: 10px 15px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  margin-bottom: 10px;
  font-size: 16px;
}

.create-group-btn:hover {
  background-color: #218838;
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

.create-btn {
  background-color: #007bff;
  color: white;
  padding: 8px 12px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
}

.cancel-btn {
  background-color: #dc3545;
  color: white;
  padding: 8px 12px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
}
</style>
