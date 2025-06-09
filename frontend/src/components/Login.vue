<template>
  <!-- vue formatter: ctr+alt+f -->
  <div class="login-container">
    <div class="login-box">
      <h2>Login</h2>
      <form @submit.prevent="sentdata">
        <input type="text" v-model.lazy="username" placeholder="username" required />
        <input type="password" v-model.lazy="Password" placeholder="Password" required />
        <button style="background-color: #007bff" type="submit">Login</button>
      </form>
      <p v-if="errorMessage">{{ errorMessage }}</p>
    </div>
  </div>
</template>

<script setup>
import { useAuthStore } from "@/authStore";

import { ref } from "vue";
import { useRouter } from "vue-router";
import axios from "axios";
import config from "@/config";

axios.defaults.withCredentials = true; // ✅ Ensures cookies are sent & received
const router = useRouter();
const auth = useAuthStore();
const isLoggedIn = ref(localStorage.getItem("isLoggedIn") === "true");
let username = ref("");
let Password = ref("");
let errorMessage = ref("");

const sentdata = async () => {
  try {
    let resp = await axios.post(
      `${config.API_URL}/login`,
      {
        email: username.value,
        password: Password.value,
      },
      { withCredentials: true }
    );
    console.log(resp.data);
    // isLoggedIn.value = true; // ✅ Update UI instantly
    auth.login(); // ✅ Updates UI instantly
    // localStorage.setItem("isLoggedIn", "true"); // ✅ Store login state
    router.push("/home");
  } catch (error) {
    errorMessage.value = "Invalid login credentials.";
    throw Error(error);
    //throw New.Error("klb")
  }
};
</script>

<style scoped>
/* Login Styles */
.login-container {
  display: flex;
  justify-content: center;
  /* Center horizontally */
  align-items: center;
  /* Center vertically */
  min-height: 70vh;
  /* Ensure full viewport height */
  background-color: #f4f4f4;
  /* Light gray background */
}

.login-box {
  background-color: #fff;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
  width: 350px;
  /* Adjust width as needed */
}

.login-box h2 {
  text-align: center;
  margin-bottom: 20px;
  color: #333;
}

.login-box input[type="text"],
.login-box input[type="password"] {
  width: calc(100% - 22px);
  /* Account for padding and border */
  padding: 10px;
  margin-bottom: 15px;
  border: 1px solid #ccc;
  border-radius: 4px;
  box-sizing: border-box;
  /* Include padding in width calculation */
  font-size: 16px;
}

.login-box input[type="text"]:focus,
.login-box input[type="password"]:focus {
  outline: none;
  border-color: #3f7fc3;
  /* Green highlight on focus */
  box-shadow: 0 0 5px rgba(76, 175, 80, 0.2);
  /* Subtle green shadow */
}

.login-box button[type="submit"] {
  width: 100%;
  padding: 10px;
  background-color: #007bff;
  /* Green */
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  transition: background-color 0.3s;
  /* Smooth transition */
}

.login-box button[type="submit"]:hover {
  background-color: #0269d7;
  /* Darker green on hover */
}

.login-box p {
  margin-top: 15px;
  color: red;
  /* Error message color */
  text-align: center;
}
</style>
