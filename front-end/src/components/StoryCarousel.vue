<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { storyService, type StoryGroup } from "@/services/storyService";
import { useEventBus } from "@/composables/useEventBus";
import { usePhoto } from "@/composables/usePhoto";

const { loadPhotoUrl } = usePhoto();

const emit = defineEmits(["open-story"]);
const { on } = useEventBus();
const storyGroups = ref<StoryGroup[]>([]);
const loading = ref(false);
const currentUserId = ref("");

const fetchStories = async () => {
  loading.value = true;
  try {
    const groups = await storyService.getAllActiveStories();
    storyGroups.value = groups;
    currentUserId.value = await storyService.getCurrentUserId();
  } catch (error) {
    console.error("Failed to fetch stories:", error);
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  fetchStories();
  on("story-created", () => {
    fetchStories();
  });
  on("post-created", () => {
    fetchStories();
  });
});

const openStoryViewer = (groupIndex: number) => {
  emit("open-story", groupIndex);
};

const getProfilePicture = (group: StoryGroup) => {
  if (group.user?.profile_picture) {
    return group.user.profile_picture;
  }
  return "";
};

const getUsername = (group: StoryGroup) => {
  if (group.user?.username) {
    return group.user.username;
  }
  return "User";
};
</script>

<template>
  <div class="story-carousel">
    <div
      v-if="loading"
      class="loading"
    >
      Loading stories...
    </div>
    <div
      v-else-if="storyGroups.length === 0"
      class="no-stories"
    />
    <div
      v-else
      class="story-list"
    >
      <div
        v-for="(group, index) in storyGroups"
        :key="group.user_id"
        class="story-item"
        @click="openStoryViewer(index)"
      >
        <div
          class="story-avatar"
          :class="{ viewed: group.has_viewed, 'my-story': group.user_id === currentUserId }"
        >
          <img
            :src="getProfilePicture(group) ? loadPhotoUrl(getProfilePicture(group)) : 'https://via.placeholder.com/60'"
            :alt="getUsername(group)"
          />
        </div>
        <div class="story-username">
          {{ group.user_id === currentUserId ? 'Your Story' : getUsername(group) }}
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.story-carousel {
  width: 100%;
  padding: 15px 0;
  background-color: rgb(0, 0, 0);
  border-bottom: 1px solid rgb(60, 60, 60);
}

.loading,
.no-stories {
  text-align: center;
  color: rgb(0, 0, 0);
  padding: 20px;
  font-size: 14px;
}

.story-list {
  display: flex;
  gap: 15px;
  overflow-x: auto;
  padding: 0 20px;
  scrollbar-width: thin;
  scrollbar-color: rgb(0, 0, 0) transparent;
}

.story-list::-webkit-scrollbar {
  height: 6px;
}

.story-list::-webkit-scrollbar-track {
  background: transparent;
}

.story-list::-webkit-scrollbar-thumb {
  background-color: rgb(0, 0, 0);
  border-radius: 3px;
}

.story-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
  flex-shrink: 0;
  gap: 8px;
}

.story-avatar {
  width: 70px;
  height: 70px;
  border-radius: 50%;
  padding: 3px;
  background: linear-gradient(45deg, #f09433 0%, #e6683c 25%, #dc2743 50%, #cc2366 75%, #bc1888 100%);
  position: relative;
}

.story-avatar.viewed {
  background: rgb(100, 100, 100);
}

.story-avatar.my-story {
  background: linear-gradient(45deg, #a800f6 0%, #6a00f4 100%);
}

.story-avatar img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid rgb(22, 22, 22);
  display: block;
}

.story-username {
  color: rgb(245, 245, 245);
  font-size: 12px;
  max-width: 70px;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
