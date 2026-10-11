<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from "vue";
import { storyService, type StoryGroup, type Story } from "@/services/storyService";
import { messageService, type User } from "@/services/messageService";
import { usePhoto } from "@/composables/usePhoto";

const { loadPhotoUrl } = usePhoto();

const props = defineProps<{
  storyGroups: StoryGroup[];
  initialGroupIndex: number;
  show: boolean;
}>();

const emit = defineEmits(["close", "story-viewed"]);

const currentGroupIndex = ref(props.initialGroupIndex);
const currentStoryIndex = ref(0);
const isPaused = ref(false);
const progress = ref(0);
const isLiked = ref(false);
const replyText = ref("");
const showShareModal = ref(false);
const shareSearchQuery = ref("");
const shareSearchResults = ref<User[]>([]);
const loadingShareSearch = ref(false);
const selectedShareUserId = ref<string | null>(null);
const selectedRecipient = ref<User | null>(null);
const showConfirmShare = ref(false);
let shareSearchTimeout: number | null = null;
const videoElement = ref<HTMLVideoElement | null>(null);

let progressInterval: number | null = null;
let storyTimer: number | null = null;

const currentGroup = computed(() => {
  return props.storyGroups[currentGroupIndex.value];
});

const currentStory = computed(() => {
  if (!currentGroup.value) return null;
  return currentGroup.value.stories[currentStoryIndex.value];
});

const isVideo = computed(() => {
  return currentStory.value?.media_type === "video";
});

const storyDuration = computed(() => {
  if (!currentStory.value) return 5000;
  return isVideo.value ? currentStory.value.duration * 1000 : 5000;
});

const canGoPrevious = computed(() => {
  return currentGroupIndex.value > 0 || currentStoryIndex.value > 0;
});

const canGoNext = computed(() => {
  if (!currentGroup.value) return false;
  return (
    currentStoryIndex.value < currentGroup.value.stories.length - 1 ||
    currentGroupIndex.value < props.storyGroups.length - 1
  );
});

const startProgress = () => {
  if (!currentStory.value || isPaused.value) return;
  console.log(currentStory.value);

  progress.value = 0;
  const duration = storyDuration.value;
  const interval = 50;

  if (isVideo.value && videoElement.value) {
    videoElement.value.play();
  }

  progressInterval = window.setInterval(() => {
    if (!isPaused.value) {
      progress.value += (interval / duration) * 100;
      if (progress.value >= 100) {
        nextStory();
      }
    }
  }, interval);
};

const stopProgress = () => {
  if (progressInterval !== null) {
    clearInterval(progressInterval);
    progressInterval = null;
  }
  if (videoElement.value && !videoElement.value.paused) {
    videoElement.value.pause();
  }
};

const markStoryAsViewed = async () => {
  if (currentStory.value && currentStory.value.id) {
    try {
      const storyId = typeof currentStory.value.id === 'string' ? currentStory.value.id.split(':')[0] : currentStory.value.id;
      await storyService.viewStory(storyId);
      emit("story-viewed");
    } catch (error) {
      
    }
  }
};

const checkLikeStatus = async () => {
  if (currentStory.value) {
    try {
      const storyId = typeof currentStory.value.id === 'string' ? currentStory.value.id.split(':')[0] : currentStory.value.id;
      isLiked.value = await storyService.checkStoryLike(storyId);
    } catch (error) {
      
    }
  }
};

const nextStory = () => {
  if (!currentGroup.value) return;

  if (currentStoryIndex.value < currentGroup.value.stories.length - 1) {
    currentStoryIndex.value++;
    resetStory();
  } else if (currentGroupIndex.value < props.storyGroups.length - 1) {
    currentGroupIndex.value++;
    currentStoryIndex.value = 0;
    resetStory();
  } else {
    closeViewer();
  }
};

const previousStory = () => {
  if (currentStoryIndex.value > 0) {
    currentStoryIndex.value--;
    resetStory();
  } else if (currentGroupIndex.value > 0) {
    currentGroupIndex.value--;
    const prevGroup = props.storyGroups[currentGroupIndex.value];
    currentStoryIndex.value = prevGroup.stories.length - 1;
    resetStory();
  }
};

const resetStory = () => {
  stopProgress();
  progress.value = 0;
  replyText.value = "";
  markStoryAsViewed();
  checkLikeStatus();
  setTimeout(() => {
    startProgress();
  }, 100);
};

const togglePause = () => {
  isPaused.value = !isPaused.value;
  if (isPaused.value) {
    stopProgress();
  } else {
    startProgress();
  }
};

const toggleLike = async () => {
  if (!currentStory.value) return;
  try {
    const storyId = typeof currentStory.value.id === 'string' ? currentStory.value.id.split(':')[0] : currentStory.value.id;
    await storyService.likeStory(storyId);
    isLiked.value = !isLiked.value;
  } catch (error) {
    
  }
};

const sendReply = async () => {
  if (!currentStory.value || !replyText.value.trim()) return;
  try {
    const storyId = typeof currentStory.value.id === 'string' ? currentStory.value.id.split(':')[0] : currentStory.value.id;
    await storyService.replyToStory(storyId, replyText.value);
    replyText.value = "";
  } catch (error) {
    
  }
};

const searchShareUsers = () => {
  if (shareSearchTimeout) {
    clearTimeout(shareSearchTimeout);
  }

  const query = shareSearchQuery.value.trim();
  if (!query || query.length < 2) {
    shareSearchResults.value = [];
    loadingShareSearch.value = false;
    return;
  }

  loadingShareSearch.value = true;

  shareSearchTimeout = window.setTimeout(async () => {
    try {
      const token = localStorage.getItem("token");
      const response = await fetch(
        `https://localhost/api/v1/users/search?query=${encodeURIComponent(query)}`,
        {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) throw new Error("Failed to search users");

      const data = await response.json();
      shareSearchResults.value = data.users || [];
    } catch (error) {
      
      shareSearchResults.value = [];
    } finally {
      loadingShareSearch.value = false;
    }
  }, 300);
};

const openShareModal = async () => {
  showShareModal.value = true;
  shareSearchQuery.value = "";
  shareSearchResults.value = [];
  selectedShareUserId.value = null;
  isPaused.value = true;
  stopProgress();
};

const closeShareModal = () => {
  showShareModal.value = false;
  selectedRecipient.value = null;
  showConfirmShare.value = false;
  shareSearchQuery.value = "";
  shareSearchResults.value = [];
  selectedShareUserId.value = null;
  isPaused.value = false;
  startProgress();
};

const shareToUser = async () => {
  if (!selectedShareUserId.value || !currentStory.value) return;

  try {
    const token = localStorage.getItem("token");
    const storyId = typeof currentStory.value.id === 'string' ? currentStory.value.id.split(':')[0] : currentStory.value.id;
    const storyUrl = `${window.location.origin}/story/${currentStory.value.user_id}/${storyId}`;
    
    const conversationPayload = {
      participant_ids: [selectedShareUserId.value],
    };

    const convResponse = await fetch("https://localhost/api/v1/messages/conversations", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(conversationPayload),
    });

    if (!convResponse.ok) throw new Error("Failed to create conversation");

    const convData = await convResponse.json();
    const conversationId = convData.conversation?.id || convData.conversation?.ID;

    if (!conversationId) throw new Error("No conversation ID returned");

    const messageFormData = new FormData();
    messageFormData.append("content", storyUrl);

    const msgResponse = await fetch(
      `https://localhost/api/v1/messages/conversations/${conversationId}/messages`,
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: messageFormData,
      }
    );

    if (!msgResponse.ok) throw new Error("Failed to send message");

    closeShareModal();
  } catch (error) {
    
  }
};

const closeViewer = () => {
  stopProgress();
  emit("close");
};

const handleKeydown = (event: KeyboardEvent) => {
  if (!props.show) return;
  if (event.key === "ArrowRight") {
    nextStory();
  } else if (event.key === "ArrowLeft") {
    previousStory();
  } else if (event.key === "Escape") {
    closeViewer();
  } else if (event.key === " ") {
    event.preventDefault();
    togglePause();
  }
};

watch(() => props.show, (newVal) => {
  if (newVal) {
    currentGroupIndex.value = props.initialGroupIndex;
    currentStoryIndex.value = 0;
    resetStory();
  } else {
    stopProgress();
  }
});

onMounted(() => {
  window.addEventListener("keydown", handleKeydown);
  if (props.show) {
    resetStory();
  }
});

onUnmounted(() => {
  window.removeEventListener("keydown", handleKeydown);
  stopProgress();
});
</script>

<template>
  <Teleport to="body">
    <div
      v-if="show && currentStory"
      class="story-viewer"
      @click.self="closeViewer"
    >
      <div class="story-container">
        <div class="story-header">
          <div class="progress-bars">
            <div
              v-for="(story, index) in currentGroup.stories"
              :key="story.id"
              class="progress-bar-container"
            >
              <div
                class="progress-bar"
                :style="{
                  width:
                    index < currentStoryIndex
                      ? '100%'
                      : index === currentStoryIndex
                      ? `${progress}%`
                      : '0%',
                }"
              />
            </div>
          </div>
          <div class="story-info">
            <div class="user-info">
              <img
                :src="loadPhotoUrl(currentGroup.user?.profile_picture) || 'https://via.placeholder.com/40'"
                :alt="currentGroup.user?.username || 'User'"
                class="user-avatar"
              />
              <span class="username">{{ currentGroup.user?.username || 'User' }}</span>
            </div>
            <button
              class="close-btn"
              @click="closeViewer"
            >
              ×
            </button>
          </div>
        </div>

        <div
          class="story-content"
          @click="togglePause"
        >
          <img
            v-if="!isVideo && currentStory.media_url"
            :src="loadPhotoUrl(currentStory.media_url)"
            alt="Story"
            class="story-media"
          />
          <video
            v-else-if="isVideo && currentStory.media_url"
            ref="videoElement"
            :src="loadPhotoUrl(currentStory.media_url)"
            class="story-media"
            autoplay
            muted
            playsinline
            @ended="nextStory"
            @loadeddata="startProgress"
            @error="nextStory"
          />
          <div
            v-else-if="!currentStory.media_url"
            class="story-media"
            style="display: flex; align-items: center; justify-content: center; background: #1a1a1a; color: white;"
          >
            Loading media...
          </div>

          <button
            v-if="canGoPrevious"
            class="nav-btn prev"
            @click.stop="previousStory"
          >
            ‹
          </button>
          <button
            v-if="canGoNext"
            class="nav-btn next"
            @click.stop="nextStory"
          >
            ›
          </button>

          <div
            v-if="isPaused"
            class="pause-indicator"
          >
            ⏸
          </div>
        </div>

        <div class="story-actions">
          <div class="reply-section">
            <input
              v-model="replyText"
              type="text"
              placeholder="Reply to story..."
              class="reply-input"
              @keyup.enter="sendReply"
              @focus="isPaused = true; stopProgress()"
              @blur="isPaused = false; startProgress()"
            >
          </div>
          <div class="action-buttons">
            <button
              class="action-btn"
              :class="{ liked: isLiked }"
              @click="toggleLike"
            >
              <img 
                src="@/assets/notif-light.png" 
                alt="Like"
                class="action-icon"
              >
            </button>
            <button
              class="action-btn"
              @click="openShareModal"
            >
              <img 
                src="@/assets/messages-light.png" 
                alt="Share"
                class="action-icon"
              >
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="showShareModal"
        class="share-modal"
        @click.self="closeShareModal"
      >
        <div class="share-content">
          <h3>Send Story</h3>
          <input
            v-model="shareSearchQuery"
            type="text"
            placeholder="Search users..."
            class="search-input"
            @input="searchShareUsers"
          >
          <div class="user-results">
            <div
              v-for="user in shareSearchResults"
              :key="user.id"
              class="user-item"
              :class="{ selected: selectedShareUserId === user.id }"
              @click="selectedShareUserId = user.id"
            >
              <img
                :src="loadPhotoUrl(user.profile_picture) || 'https://via.placeholder.com/40'"
                :alt="user.username"
                class="user-avatar"
              />
              <div class="user-info">
                <div class="user-name">{{ user.full_name }}</div>
                <div class="user-username">@{{ user.username }}</div>
              </div>
            </div>
            <div
              v-if="loadingShareSearch"
              class="loading"
            >
              Searching...
            </div>
            <div
              v-if="!loadingShareSearch && shareSearchQuery && shareSearchResults.length === 0"
              class="no-results"
            >
              No users found
            </div>
          </div>
          <div class="modal-actions">
            <button
              class="cancel-btn"
              @click="closeShareModal"
            >
              Cancel
            </button>
            <button
              class="send-btn"
              :disabled="!selectedShareUserId"
              @click="shareToUser"
            >
              Send
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.story-viewer {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.95);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.story-container {
  width: 400px;
  height: 700px;
  background-color: rgb(22, 22, 22);
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  position: relative;
}

.story-header {
  padding: 15px;
  background: linear-gradient(180deg, rgba(0, 0, 0, 0.7) 0%, transparent 100%);
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 10;
}

.progress-bars {
  display: flex;
  gap: 4px;
  margin-bottom: 12px;
}

.progress-bar-container {
  flex: 1;
  height: 3px;
  background-color: rgba(255, 255, 255, 0.3);
  border-radius: 2px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  background-color: white;
  transition: width 0.05s linear;
}

.story-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-avatar {
  width: 35px;
  height: 35px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid white;
  display: block;
  flex-shrink: 0;
}

.username {
  color: white;
  font-weight: 600;
  font-size: 14px;
}

.close-btn {
  background: none;
  border: none;
  color: white;
  font-size: 32px;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: static;
}

.story-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  background-color: black;
}

.story-media {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background-color: rgba(255, 255, 255, 0.3);
  border: none;
  color: white;
  font-size: 32px;
  width: 45px;
  height: 45px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
  padding: 0;
}

.nav-btn:hover {
  background-color: rgba(255, 255, 255, 0.5);
}

.nav-btn.prev {
  left: 15px;
}

.nav-btn.next {
  right: 15px;
}

.pause-indicator {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 60px;
  color: white;
  opacity: 0.7;
}

.story-actions {
  padding: 15px;
  background: linear-gradient(0deg, rgba(0, 0, 0, 0.7) 0%, transparent 100%);
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 10;
}

.reply-section {
  margin-bottom: 10px;
}

.reply-input {
  width: 100%;
  padding: 12px;
  border-radius: 25px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  background-color: rgba(255, 255, 255, 0.1);
  color: white;
  font-size: 14px;
  outline: none;
}

.reply-input::placeholder {
  color: rgba(255, 255, 255, 0.6);
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 20px;
}

.action-btn {
  background: none;
  border: none;
  cursor: pointer;
  transition: transform 0.2s;
  padding: 0;
  position: static;
  width: 28px;
  height: 28px;
}

.action-icon {
  width: 100%;
  height: 100%;
  object-fit: contain;
  filter: brightness(0) invert(1);
}

.action-btn.liked .action-icon {
  filter: none;
}

.action-btn:hover {
  transform: scale(1.2);
}

.action-btn.liked {
  animation: like-pulse 0.3s ease;
}

@keyframes like-pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.3);
  }
}

.share-modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1001;
}

.share-content {
  background-color: rgb(35, 35, 35);
  border-radius: 12px;
  padding: 20px;
  width: 90%;
  max-width: 400px;
  max-height: 600px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.share-content h3 {
  color: white;
  margin-bottom: 15px;
  font-size: 18px;
  text-align: center;
}

.search-input {
  width: 100%;
  padding: 10px;
  border-radius: 8px;
  border: 1px solid rgb(60, 60, 60);
  background-color: rgb(22, 22, 22);
  color: white;
  font-size: 14px;
  margin-bottom: 15px;
  outline: none;
}

.search-input:focus {
  border-color: #a800f6;
}

.user-results {
  max-height: 400px;
  overflow-y: auto;
  margin-bottom: 15px;
}

.user-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.user-item:hover {
  background-color: rgb(50, 50, 50);
}

.user-item.selected {
  background-color: rgba(168, 0, 246, 0.2);
  border: 2px solid #a800f6;
}

.user-item .user-avatar {
  width: 45px;
  height: 45px;
  border-radius: 50%;
  object-fit: cover;
}

.user-info {
  flex: 1;
}

.user-name {
  color: white;
  font-weight: 600;
  font-size: 14px;
}

.user-username {
  color: rgb(150, 150, 150);
  font-size: 12px;
}

.loading,
.no-results {
  text-align: center;
  color: rgb(150, 150, 150);
  padding: 20px;
  font-size: 14px;
}

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 15px;
}

.cancel-btn,
.send-btn {
  flex: 1;
  padding: 12px;
  border-radius: 8px;
  border: none;
  font-weight: 600;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
}

.cancel-btn {
  background-color: rgb(60, 60, 60);
  color: white;
}

.cancel-btn:hover {
  background-color: rgb(80, 80, 80);
}

.send-btn {
  background-color: #a800f6;
  color: white;
}

.send-btn:hover {
  background-color: #8900d8;
}

.send-btn:disabled {
  background-color: rgb(60, 60, 60);
  cursor: not-allowed;
  opacity: 0.5;
}

</style>
