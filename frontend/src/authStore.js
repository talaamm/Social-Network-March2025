import { defineStore } from "pinia";
import { ref } from "vue";

export const useAuthStore = defineStore("auth", () => {
    const isLoggedIn = ref(localStorage.getItem("isLoggedIn") === "true");

    const login = () => {
        isLoggedIn.value = true;
        localStorage.setItem("isLoggedIn", "true");
    };

    const logout = () => {
        isLoggedIn.value = false;
        localStorage.removeItem("isLoggedIn");
    };
    const loggedout = () => {
        isLoggedIn.value = false;
        // localStorage.removeItem("isLoggedIn");
    };

    return { isLoggedIn, login, logout, loggedout };
});