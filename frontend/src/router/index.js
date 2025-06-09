import axios from "axios";
import config from "@/config";

import { useAuthStore } from "@/authStore";
import { createRouter, createWebHistory } from "vue-router";

import Login from "@/components/Login.vue";
import Register from "@/components/Register.vue";
import Home from "@/components/Home.vue";
import Profile from "@/components/Profile.vue";
import Disc_Ppl from "@/components/Discover-People.vue";
import UserProfile from "@/components/UserProfile.vue";
import ChatList from "@/components/Chat.vue";
import ChatPage from "@/components/Masseges.vue";
import Req from "@/components/Requests.vue";
import GroupPage from "@/components/Group.vue";
import MyGroups from "@/components/MyGroups.vue";
import DiscG from "@/components/Discover-Groups.vue";
import GC from "@/components/Grupchats.vue";

const routes = [
    { path: "/login", component: Login },
    { path: "/register", component: Register },
    {
        path: "/home",
        component: Home,
        meta: { requiresAuth: true }, // ✅ Add meta field to protect this route
    },
    {
        path: "/my-profile",
        component: Profile,
        meta: { requiresAuth: true },
    },
    {
        path: "/discover-people",
        component: Disc_Ppl,
        meta: { requiresAuth: true },
    },
    {
        path: "/requests",
        component: Req,
        meta: { requiresAuth: true },
    },
    {
        path: "/profile/:id",
        name: "UserProfile",
        component: UserProfile,
        meta: { requiresAuth: true },
    },
    { path: "/chats", name: "ChatList", component: ChatList, meta: { requiresAuth: true } },

    // ✅ Chat Page (Opens chat with a specific user)
    {
        path: "/chat/:id/:nickname", // ✅ Add nickname parameter
        name: "ChatPage",
        component: ChatPage,
        props: true,
        meta: { requiresAuth: true },
    },
    {
        path: "/groups/:groupid",
        name: "GroupPage",
        component: GroupPage,
        meta: { requiresAuth: true },
    },
    {
        path: "/my-groups",
        component: MyGroups,
        meta: { requiresAuth: true },
    },
    {
        path: "/discover-groups",
        component: DiscG,
        meta: { requiresAuth: true },
    },
    {
        path: "/group-chat/:groupid/:name",
        name: "GroupChat",
        component: GC,
        meta: { requiresAuth: true },
    }, // ✅ Add GroupChat route
];

// Create the router instance
const router = createRouter({
    history: createWebHistory(),
    routes,
});

router.beforeEach(async(to, from, next) => {
    const auth = useAuthStore();
    if (to.meta.requiresAuth) {
        try {
            await axios.get(`${config.API_URL}/`, { withCredentials: true }); // ✅ Check session
            next(); // ✅ Allow access
        } catch {
            auth.loggedout();
            next("/login"); // 🚨 Redirect to login if session is invalid
        }
    } else {
        next();
    }
});

export default router;