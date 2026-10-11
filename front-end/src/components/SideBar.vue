<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { usePhoto } from "@/composables/usePhoto";
import { useCreatePost } from "@/composables/useOverlay";
import { useSearch } from "@/composables/useOverlay";
import defaultProfileIcon from "@/components/icons/user-light.png";
import MessageShortcut from "./MessageShortcut.vue";
const { open } = useCreatePost();
const { openSearch } = useSearch();
const openCreatePost = () => {
  open();
};

const router = useRouter();

const { loadPhotoUrl } = usePhoto();
const user = ref<any>(null);
const profilePhotoUrl = ref<string>("");
const currentUserId = ref<string>("");

const fetchUser = async () => {
  const token = localStorage.getItem("token");

  if (!token) {
    return;
  }

  try {
    const response = await fetch("https://localhost/api/v1/users/get-current-user", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      const data = await response.json();
      user.value = data.user;
      currentUserId.value = user.value.ID || user.value.id;

      if (user.value.profile_picture) {
        profilePhotoUrl.value = loadPhotoUrl(user.value.profile_picture);
      } else {
        profilePhotoUrl.value = defaultProfileIcon;
      }
    }
  } catch (error) {
    console.error("Failed to fetch user:", error);
  }
};

onMounted(() => {
  fetchUser();
});

const navigate = (path: string) => {
  if (path === "profile") {

    router.push(`/profile/${currentUserId.value}`);
  } else {
    router.push("/" + path);
  }
};

const isMoreOpen = ref(false);

const openMore = () => {
  const moreContainer = document.getElementById("more-container");
  if (moreContainer?.style.display === "none") {
    moreContainer!.style.display = "block";
    isMoreOpen.value = true;
  } else {
    moreContainer!.style.display = "none";
    isMoreOpen.value = false;
  }
};

const logout = async () => {
  const token = localStorage.getItem("token");

  try {
    const response = await fetch("https://localhost/api/v1/users/logout", {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    });

    
  } catch (error) {
    console.error("Logout failed:", error);
  }

  localStorage.removeItem("token");
  router.push("/login");
};

const pfpError = (e: Event) => {
  const target = e.target as HTMLImageElement;
  target.src = defaultProfileIcon;
};
</script>

<template>
  <MessageShortcut />
  <div
    class="side-bar"
    :class="{ 'more-open': isMoreOpen }"
  >
    <div
      class="logo-container"
      width="100%"
      height="100%"
      style="display: flex; justify-content: center; align-items: center"
    >
      <img
        src="../components/icons/logo-dark.png"
        width="150"
        height="50"
        style="margin-bottom: 30px"
      />
    </div>

    <button @click="navigate('')">
      <img
        src="../components/icons/home-light.png"
        width="25"
        height="25"
      />
      <h2>Home</h2>
    </button>

    <button @click="openSearch">
      <img
        src="../components/icons/search-light.png"
        width="25"
        height="25"
      />
      <h2>Search</h2>
    </button>

    <button @click="navigate('explore')">
      <img
        src="../components/icons/explore-light.png"
        width="25"
        height="25"
      />
      <h2>Explore</h2>
    </button>
    <button @click="navigate('reels')">
      <img
        src="../components/icons/reels-light.png"
        width="25"
        height="25"
      />
      <h2>Reels</h2>
    </button>
    <button @click="navigate('messages')">
      <img
        src="../components/icons/messages-light.png"
        width="25"
        height="25"
      />
      <h2>Messages</h2>
    </button>
    <button @click="navigate('notification')">
      <img
        src="../components/icons/notif-light.png"
        width="25"
        height="25"
      />
      <h2>Notification</h2>
    </button>
    <button @click="openCreatePost">
      <img
        src="../components/icons/create-light.png"
        width="25"
        height="25"
      />
      <h2>Create</h2>
    </button>
    <button @click="navigate('dashboard')">
      <img
        src="../components/icons/dashboard-light.png"
        width="25"
        height="25"
      />
      <h2>Dashboard</h2>
    </button>
    <button
      id="profile-picture"
      @click="navigate('profile')"
    >
      <img
        v-if="profilePhotoUrl"
        :src="profilePhotoUrl"
        width="25"
        height="25"
        style="border-radius: 50%; object-fit: cover; border: 2px solid white"
        @error="pfpError"
      />
      <img
        v-else
        src="../components/icons/user-light.png"
        width="25"
        height="25"
      />
      <h2>Profile</h2>
    </button>
    <div id="more-container">
      <button @click="logout">
        Log Out
      </button>
    </div>
    <button
      id="more-button"
      @click="openMore"
    >
      <img
        src="../components/icons/more-light.png"
        width="25"
        height="25"
      />
      <h2>More</h2>
    </button>
    <button @click="navigate('meta')">
      <img
        src="../components/icons/meta-light.png"
        width="25"
        height="25"
      />
      <h2>Meta</h2>
    </button>
  </div>
</template>

<style scoped>
#more-container {
  padding: 10px;
  left: 5%;
  bottom: 30%;
  width: 15%;
  height: 30%;
  position: absolute;
  border-radius: 20px;
  display: none;
  background-color: rgb(48, 48, 48);
}

#more-container button {
  font-size: 18px;
  font-weight: 500;
  align-items: center;
  padding: 10px;
}

.side-bar {
  background-color: black;
  padding: 25px;
  height: 100vh;

  width: 20vw;
  color: white;
  scroll-behavior: smooth;
  overflow-y: auto;
  overflow-x: hidden;
}

.router-link {
  text-decoration: none;
  color: white;
}

h2 {
  font-weight: 500;
  opacity: 0;
  transition: all 0.3s ease;
}

.side-bar:hover h2 {
  opacity: 1;
}

.side-bar.more-open h2 {
  opacity: 1;
}

button {
  width: 100%;
  border: none;
  background-color: transparent;
  color: white;
  display: flex;
  align-items: center;
  gap: 15px;
  cursor: pointer;
  padding: 10px;
  border-radius: 8px;
  transition: background-color 0.3s;
  margin-bottom: 15px;
}

button:hover {
  background-color: rgb(83, 83, 83);
}

button.active h2 {
  font-weight: 700;
}
</style>
