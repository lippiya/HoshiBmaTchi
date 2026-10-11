<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { storyService, type Story } from "@/services/storyService";
import { usePhoto } from "@/composables/usePhoto";
import StoryViewer from "@/components/StoryViewer.vue";

const { loadPhotoUrl } = usePhoto();

const stories = ref<Story[]>([]);
const loading = ref(true);
const hasMore = ref(true);
const offset = ref(0);
const limit = 20;
const showStoryViewer = ref(false);
const selectedStoryIndex = ref(0);
const loadingMore = ref(false);

const loadStories = async (append = false) => {
  if (append) {
    loadingMore.value = true;
  } else {
    loading.value = true;
  }

  try {
    const newStories = await storyService.getArchivedStories(limit, offset.value);
    
    if (append) {
      stories.value = [...stories.value, ...newStories];
    } else {
      stories.value = newStories;
    }

    if (newStories.length < limit) {
      hasMore.value = false;
    }

    offset.value += newStories.length;
  } catch (error) {
    console.error("Failed to load archived stories:", error);
  } finally {
    loading.value = false;
    loadingMore.value = false;
  }
};

const handleScroll = () => {
  const scrollTop = window.scrollY || document.documentElement.scrollTop;
  const scrollHeight = document.documentElement.scrollHeight;
  const clientHeight = window.innerHeight;

  if (scrollTop + clientHeight >= scrollHeight - 200 && hasMore.value && !loadingMore.value) {
    loadStories(true);
  }
};

const openStory = (index: number) => {
  selectedStoryIndex.value = index;
  showStoryViewer.value = true;
};

const closeStoryViewer = () => {
  showStoryViewer.value = false;
};

const formatDate = (dateString: string) => {
  const date = new Date(dateString);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));
  
  if (days === 0) {
    return "Today";
  } else if (days === 1) {
    return "Yesterday";
  } else if (days < 7) {
    return `${days} days ago`;
  } else {
    return date.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
  }
};

onMounted(() => {
  loadStories();
  window.addEventListener("scroll", handleScroll);
});

onUnmounted(() => {
  window.removeEventListener("scroll", handleScroll);
});
</script>

<template>
  <div class="archive-page">
    <div class="archive-header">
      <h1>Archive</h1>
      <p>View your past stories</p>
    </div>

    <div v-if="loading" class="loading-grid">
      <div v-for="i in 12" :key="i" class="skeleton-item"></div>
    </div>

    <div v-else-if="stories.length === 0" class="empty-state">
      <svg width="96" height="96" viewBox="0 0 24 24" fill="none">
        <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/>
        <path d="M12 8v4M12 16h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
      </svg>
      <h2>No archived stories</h2>
      <p>Your stories will appear here after they expire</p>
    </div>

    <div v-else class="stories-grid">
      <div
        v-for="(story, index) in stories"
        :key="story.id"
        class="story-item"
        @click="openStory(index)"
      >
        <img
          v-if="story.media_type === 'image'"
          :src="loadPhotoUrl(story.media_url)"
          :alt="`Story ${index + 1}`"
        />
        <video
          v-else
          :src="loadPhotoUrl(story.media_url)"
          muted
        />
        <div class="story-overlay">
          <svg v-if="story.media_type === 'video'" width="48" height="48" viewBox="0 0 24 24" fill="white">
            <path d="M8 5v14l11-7z" />
          </svg>
        </div>
        <div class="story-date">
          {{ formatDate(story.created_at) }}
        </div>
      </div>
    </div>

    <div v-if="loadingMore" class="loading-more">
      <div class="spinner"></div>
    </div>

    <StoryViewer
      v-if="showStoryViewer && stories.length > 0"
      :story-groups="[{ user_id: '', stories: stories, has_viewed: true }]"
      :initial-group-index="0"
      :show="showStoryViewer"
      @close="closeStoryViewer"
    />
  </div>
</template>

<style scoped>
.archive-page {
  min-height: 100vh;
  background-color: #000;
  color: #fff;
  padding: 24px;
  padding-top: 80px;
}

.archive-header {
  max-width: 935px;
  margin: 0 auto 32px;
}

.archive-header h1 {
  font-size: 28px;
  font-weight: 300;
  margin: 0 0 8px;
}

.archive-header p {
  font-size: 16px;
  color: #8e8e8e;
  margin: 0;
}

.loading-grid {
  max-width: 935px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.skeleton-item {
  aspect-ratio: 9/16;
  background: linear-gradient(90deg, #262626 25%, #3a3a3a 50%, #262626 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 8px;
}

@keyframes shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.empty-state {
  max-width: 400px;
  margin: 120px auto;
  text-align: center;
}

.empty-state svg {
  color: #8e8e8e;
  margin-bottom: 24px;
}

.empty-state h2 {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 8px;
}

.empty-state p {
  font-size: 14px;
  color: #8e8e8e;
  margin: 0;
}

.stories-grid {
  max-width: 935px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.story-item {
  aspect-ratio: 9/16;
  position: relative;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s;
}

.story-item:hover {
  transform: scale(1.02);
}

.story-item img,
.story-item video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.story-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.3);
  opacity: 0;
  transition: opacity 0.2s;
}

.story-item:hover .story-overlay {
  opacity: 1;
}

.story-date {
  position: absolute;
  bottom: 12px;
  left: 12px;
  background: rgba(0, 0, 0, 0.7);
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.loading-more {
  display: flex;
  justify-content: center;
  padding: 32px;
}

.spinner {
  width: 24px;
  height: 24px;
  border: 2px solid #262626;
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 935px) {
  .stories-grid,
  .loading-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 735px) {
  .stories-grid,
  .loading-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .archive-page {
    padding: 16px;
    padding-top: 70px;
  }
}

@media (max-width: 480px) {
  .stories-grid,
  .loading-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 8px;
  }
  
  .archive-header h1 {
    font-size: 24px;
  }
  
  .archive-header p {
    font-size: 14px;
  }
}
</style>
