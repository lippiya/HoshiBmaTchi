import { createRouter, createWebHistory } from "vue-router";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      component: () => import(/* webpackPrefetch: true */ "@/views/HomePage.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/register",
      name: "register",
      component: () => import(/* webpackPrefetch: true */ "@/views/RegisterPage.vue"),
      meta: { guestOnly: true },
    },
    {
      path: "/login",
      name: "login",
      component: () => import(/* webpackPrefetch: true */ "@/views/LoginPage.vue"),
      meta: { guestOnly: true },
    },
    {
      path: "/profile/:userId?",
      name: "profile",
      component: () => import(/* webpackPrefetch: true */ "@/views/ProfilePage.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/settings",
      name: "settings",
      component: () => import(/* webpackPrefetch: true */ "@/views/SettingsPage.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/reels",
      name: "reels",
      component: () => import(/* webpackPrefetch: true */ "@/views/ReelsPage.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/messages",
      name: "messages",
      component: () => import(/* webpackPrefetch: true */ "@/views/MessagesPage.vue"),
      meta: { requiresAuth: true },

    },
    {
      path: "/explore",
      name: "explore",
      component: () => import(/* webpackPrefetch: true */ "@/views/ExplorePage.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/collection/:collectionID",
      name: "collection-detail",
      component: () => import("@/views/CollectionDetailPage.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/notification",
      name: "notification",
      component: () => import("@/views/NotificationsPage.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/archive",
      name: "archive",
      component: () => import("@/views/ArchivePage.vue"),
      meta: { requiresAuth: true },
    },
    {
      path: "/:pathMatch(.*)*",
      name: "not-found",
      component: () => import("@/views/NotFound.vue"),
      meta: { guestOnly: false, requiresAuth: false },
    },
  ],
});


declare const atob: (data: string) => string;

export const isTokenExpired = (token: string) => {
  try {
    const parts = token.split(".");
    if (parts.length !== 3 || !parts[1]) return true;

    const payload = JSON.parse(atob(parts[1]));
    return Date.now() >= payload.exp * 1000;
  } catch {
    return true;
  }
};

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem("token");
  const hasValidToken = token && !isTokenExpired(token);

  if (token && isTokenExpired(token)) {
    localStorage.removeItem("token");
  }

  if (to.meta.requiresAuth && !hasValidToken) {
    if (to.path !== "/login") {
      next("/login");
    } else {
      next();
    }
  } else if (to.meta.guestOnly && hasValidToken) {
    if (to.path !== "/") {
      next("/");
    } else {
      next();
    }
  } else {
    next();
  }
});

export default router;
