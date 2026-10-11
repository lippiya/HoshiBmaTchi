<script setup lang="ts">
import { RouterView } from "vue-router";
import CreatePost from "@/components/CreatePost.vue";
import Search from "./components/SearchBar.vue";
import NotificationToast from "@/components/NotificationToast.vue";
import { useNotificationSetup } from "@/composables/useNotifications";

useNotificationSetup();
</script>

<template>
  <RouterView />
  <CreatePost />
  <Search />
  <NotificationToast />
</template>

<style>
body {
  margin: 0;
  padding: 0;
  background-color: black;
  overflow-x: hidden;
}

#app {
  background-color: black;
  color: white;
}
</style>
