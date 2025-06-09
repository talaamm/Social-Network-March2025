import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "@/App.vue";
import router from "@/router"; // Import the router
// import './assets/app.css'; // Or the correct path where your CSS file is located

// createApp(App).use(router).mount('#app');

const pinia = createPinia(); // ✅ Create a Pinia instance
const app = createApp(App);

app.use(pinia); // ✅ Register Pinia in Vue
app.use(router); // Use router
app.mount("#app"); // Mount app
