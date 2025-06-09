<template>
  <div class="register-container">
    <div class="register-box">
      <h2>Register</h2>
      <form @submit.prevent="register">
        <input type="text" v-model.lazy="username" placeholder="Username" required />
        <input type="email" v-model.lazy="email" placeholder="Email" required />
        <input type="text" v-model.lazy="fname" placeholder="First Name" required />
        <input type="text" v-model.lazy="lname" placeholder="Last Name" required />
        <input type="text" v-model.lazy="gender" placeholder="Gender" />
        <input type="date" v-model.lazy="bdate" placeholder="Birthdate" required />
        <input type="number" v-model.lazy="age" placeholder="Age" min="13" max="100" />
        <input type="password" v-model.lazy="password" placeholder="Password" required />
        <input
          type="password"
          v-model.lazy="confirmPassword"
          placeholder="Confirm Password"
          required
        />
        <button style="background-color: #007bff" type="submit">Register</button>
      </form>
      <p v-if="errorMessage">{{ errorMessage }}</p>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import axios from "axios";
import config from "@/config";
import { useAuthStore } from "@/authStore";
const auth = useAuthStore();

axios.defaults.withCredentials = true; // ✅ Ensures cookies are sent & received

const router = useRouter();
const username = ref("");
const email = ref("");
const password = ref("");
const confirmPassword = ref("");
const errorMessage = ref("");
const fname = ref("");
const lname = ref("");
const age = ref("");
const bdate = ref("");
const gender = ref("");
console.log("🚀 Registering user...");

const register = async () => {
  if (password.value !== confirmPassword.value) {
    errorMessage.value = "Passwords do not match";
    return;
  }

  try {
    let resp = await axios.post(`${config.API_URL}/register`, {
      nickname: username.value,
      email: email.value,
      password: password.value,
      first_name: fname.value,
      last_name: lname.value,
      dbirth: bdate.value,
      age: age.value ? Number(age.value) : null,
      gender: gender.value ? gender.value : "Not Specified",
    });
    console.log("✅ Registration successful. Logging in...");
    auth.login(); // 🔥 This ensures the state updates properly

    router.push("/home");
  } catch (error) {
    errorMessage.value =
      "Error registering user: " + (error.response?.data?.message || error.message);
  }
};
</script>

<style scoped>
/* Register Styles */
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 70vh;
  background-color: #f4f4f4;
}

.register-box {
  background-color: #fff;
  padding: 30px;
  border-radius: 8px;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
  width: 400px;
  /* Adjust width as needed */
}

.register-box h2 {
  text-align: center;
  margin-bottom: 20px;
  color: #333;
}

.register-box input[type="text"],
.register-box input[type="email"],
.register-box input[type="date"],
.register-box input[type="password"] {
  width: calc(100% - 22px);
  /* Account for padding and border */
  padding: 10px;
  margin-bottom: 15px;
  border: 1px solid #ccc;
  border-radius: 4px;
  box-sizing: border-box;
  font-size: 16px;
}

.register-box input[type="text"]:focus,
.register-box input[type="email"]:focus,
.register-box input[type="date"]:focus,
.register-box input[type="password"]:focus {
  outline: none;
  border-color: #578ac1;
  box-shadow: 0 0 5px rgba(76, 175, 80, 0.2);
}

.register-box button[type="submit"] {
  width: 100%;
  padding: 10px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  transition: background-color 0.3s;
}

.register-box button[type="submit"]:hover {
  background-color: #0364cb;
}

.register-box p {
  margin-top: 15px;
  color: red;
  text-align: center;
}
</style>
