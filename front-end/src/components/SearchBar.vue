<script setup lang="ts">
import { ref, computed, onMounted  } from "vue";
import { useRouter } from "vue-router";
import { useSearch } from "@/composables/useOverlay";
import { usePhoto } from "@/composables/usePhoto";
import { useDebounceSearch } from "@/composables/useDebounceSearch";
import { useShare } from "@/composables/useShare";
import defaultProfileIcon from "@/components/icons/user-light.png";

const { showSearch, closeSearch } = useSearch();
const { shareMode, shareContent, selectedShareUsers, openShareMode: exposeShareMode, closeShareMode, toggleUserSelection, sendContent } = useShare();
const router = useRouter();
const { loadPhotoUrl } = usePhoto();

const searchHistory = ref<string[]>([]);
const activeTab = ref("all");
const hashtagResults = ref<any[]>([]);
const isRecording = ref(false);
const mediaRecorder = ref<MediaRecorder | null>(null);
const audioChunks = ref<Blob[]>([]);

onMounted(() => {
  const history = localStorage.getItem("searchHistory");
  if (history) {
    searchHistory.value = JSON.parse(history);
  }
});

const clearHistory = () => {
  searchHistory.value = [];
  localStorage.removeItem("searchHistory");
};

const addToHistory = (query: string) => {
  if (!query.trim()) return;

  const history = [...searchHistory.value];
  const index = history.indexOf(query);
  if (index > -1) {
    history.splice(index, 1);
  }
  history.unshift(query);

  if (history.length > 10) {
    history.pop();
  }

  searchHistory.value = history;
  localStorage.setItem("searchHistory", JSON.stringify(history));
};

const searchUsers = async (query: string) => {
  try {
    const token = localStorage.getItem("token");
    if (!token) return [];

    const response = await fetch(`https://localhost/api/v1/users/search?q=${encodeURIComponent(query)}`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`
      }
    });

    if (response.ok) {
      const data = await response.json();
      return data.users || [];
    } else {
      const errorData = await response.text();
      console.log("API error response:", errorData);
    }
  } catch (error) {
    console.error("Search users error:", error);
  }

  return [];
};

const searchHashtags = async (query: string) => {
  if (!query.startsWith("#")) return [];

  try {
    const token = localStorage.getItem("token");
    if (!token) return [];

    const response = await fetch(`https://localhost/api/v1/posts/search-hashtags?q=${encodeURIComponent(query)}`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`
      }
    });

    if (response.ok) {
      const data = await response.json();
      return data.hashtags || [];
    }
  } catch (error) {
    console.error("Search hashtags error:", error);
  }

  return [];
};

const performSearchLogic = async (query: string) => {
  const [users, hashtags] = await Promise.all([
    searchUsers(query),
    searchHashtags(query)
  ]);

  hashtagResults.value = hashtags;
  
  if (users.length > 0 || hashtags.length > 0) {
    addToHistory(query);
  }

  return users;
};

const { searchQuery, isSearching: loading, results: searchResults, search, clearSearch } = useDebounceSearch(performSearchLogic, 300);

const performSearch = () => {
  if (searchQuery.value.trim()) {
    search(searchQuery.value);
  }
};

const debouncedSearch = () => {
  search(searchQuery.value);
};

const goToProfile = (userId: string) => {
  closeSearch();
  
  router.push(`/profile/${userId}`);
};

const goToHashtag = (hashtag: string) => {
  closeSearch();
  router.push(`/explore?hashtag=${encodeURIComponent(hashtag)}`);
};

const selectHistoryItem = (item: string) => {
  searchQuery.value = item;
  performSearch();
};

const displayResults = computed(() => {
  if (activeTab.value === "users") return searchResults.value;
  if (activeTab.value === "hashtags") return hashtagResults.value;
  return [...searchResults.value, ...hashtagResults.value];
});

const selectedUserCount = computed(() => {
  return selectedShareUsers.value.length;
});

const startRecording = async () => {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    mediaRecorder.value = new MediaRecorder(stream);
    audioChunks.value = [];

    mediaRecorder.value.ondataavailable = (event) => {
      audioChunks.value.push(event.data);
    };

    mediaRecorder.value.onstop = async () => {
      const audioBlob = new Blob(audioChunks.value, { type: "audio/webm" });
      await transcribeAudio(audioBlob);
      stream.getTracks().forEach(track => track.stop());
    };

    mediaRecorder.value.start();
    isRecording.value = true;
  } catch (error) {
    console.error("Error starting recording:", error);
  }
};

const stopRecording = () => {
  if (mediaRecorder.value && isRecording.value) {
    mediaRecorder.value.stop();
    isRecording.value = false;
  }
};

const transcribeAudio = async (audioBlob: Blob) => {
  try {
    loading.value = true;
    const formData = new FormData();
    formData.append("audio", audioBlob, "audio.webm");

    const response = await fetch("https://localhost/api/v1/ai/transcribe", {
      method: "POST",
      body: formData,
    });

    if (response.ok) {
      const data = await response.json();
      searchQuery.value = data.transcription;
      performSearch();
    } else {
      console.error("Transcription failed");
    }
  } catch (error) {
    console.error("Transcription error:", error);
  } finally {
    loading.value = false;
  }
};

const handleClose = () => {
  if (isRecording.value) {
    stopRecording();
  }
  clearSearch();
  hashtagResults.value = [];
  activeTab.value = "all";
  closeSearch();
};

defineExpose({
  exposeShareMode
});
</script>

<template>
  <Teleport to="body">
    <div
      v-if="showSearch"
      class="search-background"
      @click.self="handleClose"
    >
      <div class="search-container">
        <div
          v-if="shareMode"
          class="share-mode"
        >
          <div class="share-header">
            <button
              class="back-btn"
              @click="closeShareMode"
            >
              ‹
            </button>
            <h3>Send to</h3>
            <button
              class="close-btn"
              @click="closeShareMode"
            >
              ×
            </button>
          </div>

          <div
            v-if="shareContent"
            class="share-preview"
          >
            <div class="preview-content">
              <div v-if="shareContent.postMedia && shareContent.postMedia.length > 0" class="preview-media">
                <img
                  v-if="shareContent.postMedia[0].MediaType === 'image'"
                  :src="loadPhotoUrl(shareContent.postMedia[0].FilePath)"
                  alt="preview"
                />
                <video
                  v-else
                  :src="loadPhotoUrl(shareContent.postMedia[0].FilePath)"
                  controls
                />
              </div>
              <div v-if="shareContent.media && shareContent.media.length > 0" class="preview-media">
                <img
                  v-if="shareContent.media[0].type === 'image'"
                  :src="loadPhotoUrl(shareContent.media[0].path)"
                  alt="preview"
                />
                <video
                  v-else
                  :src="loadPhotoUrl(shareContent.media[0].path)"
                  controls
                />
              </div>
              <p v-if="shareContent.caption || shareContent.Caption">
                {{ shareContent.caption || shareContent.Caption }}
              </p>
            </div>
          </div>

          <div class="search-header-small">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search users..."
              class="search-input"
              @input="debouncedSearch"
            />
          </div>

          <div class="search-content">
            <div
              v-if="loading"
              class="skeleton-container"
            >
              <div
                v-for="i in 5"
                :key="i"
                class="skeleton-item"
              >
                <div class="skeleton-avatar" />
                <div class="skeleton-info">
                  <div class="skeleton-line" />
                  <div class="skeleton-line short" />
                </div>
              </div>
            </div>

            <div
              v-else-if="searchResults.length > 0"
              class="search-results"
            >
              <div class="results-list">
                <div
                  v-for="user in searchResults"
                  :key="user.id"
                  class="result-item selectable-user"
                  :class="{ selected: selectedShareUsers.some(u => u.id === user.id) }"
                  @click="toggleUserSelection(user)"
                >
                  <div class="selection-indicator">
                    <input
                      type="checkbox"
                      :checked="selectedShareUsers.some(u => u.id === user.id)"
                      @change="toggleUserSelection(user)"
                    />
                  </div>
                  <img
                    :src="user.profile_picture ? loadPhotoUrl(user.profile_picture) : defaultProfileIcon"
                    width="40"
                    height="40"
                    class="result-avatar"
                    @error="(e: Event) => (e.target as HTMLImageElement).src = defaultProfileIcon"
                  />
                  <div class="result-info">
                    <div class="result-name">
                      {{ user.name }}
                    </div>
                    <div class="result-username">
                      @{{ user.username }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div
              v-else-if="searchQuery.trim() && !loading"
              class="no-results"
            >
              <p>No users found</p>
            </div>
          </div>

          <div v-if="selectedUserCount > 0" class="share-footer">
            <div class="selected-count">
              {{ selectedUserCount }} {{ selectedUserCount === 1 ? 'user' : 'users' }} selected
            </div>
            <button
              class="send-btn"
              @click="sendContent"
            >
              Send
            </button>
          </div>
        </div>

        <div v-else class="search-mode">
          <div class="search-header">
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Search"
              class="search-input"
              @keyup.enter="performSearch"
              @input="debouncedSearch"
            />
            <button
              :class="{ recording: isRecording }"
              class="mic-btn"
              @click="isRecording ? stopRecording() : startRecording()"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z" />
                <path d="M19 10v2a7 7 0 0 1-14 0v-2" />
                <line
                  x1="12"
                  x2="12"
                  y1="19"
                  y2="22"
                />
              </svg>
            </button>
            <button
              class="close-btn"
              @click="handleClose"
            >
              ×
            </button>
          </div>

          <div class="search-content">
          <div
            v-if="loading"
            class="skeleton-container"
          >
            <div
              v-for="i in 5"
              :key="i"
              class="skeleton-item"
            >
              <div class="skeleton-avatar" />
              <div class="skeleton-info">
                <div class="skeleton-line" />
                <div class="skeleton-line short" />
              </div>
            </div>
          </div>

          <div
            v-else-if="!searchQuery.trim() && searchHistory.length > 0"
            class="search-history"
          >
            <div class="history-header">
              <h3>Recent</h3>
              <button
                class="clear-btn"
                @click="clearHistory"
              >
                Clear all
              </button>
            </div>
            <div class="history-list">
              <div
                v-for="item in searchHistory"
                :key="item"
                class="history-item"
                @click="selectHistoryItem(item)"
              >
                <img
                  src="../components/icons/search-light.png"
                  width="16"
                  height="16"
                />
                <span>{{ item }}</span>
              </div>
            </div>
          </div>

          <div
            v-else-if="searchResults.length > 0 || hashtagResults.length > 0"
            class="search-results"
          >
            <div
              v-if="searchResults.length > 0 && hashtagResults.length > 0"
              class="tabs"
            >
              <button
                :class="{ active: activeTab === 'all' }"
                class="tab-btn"
                @click="activeTab = 'all'"
              >
                All
              </button>
              <button
                :class="{ active: activeTab === 'users' }"
                class="tab-btn"
                @click="activeTab = 'users'"
              >
                Users
              </button>
              <button
                :class="{ active: activeTab === 'hashtags' }"
                class="tab-btn"
                @click="activeTab = 'hashtags'"
              >
                Hashtags
              </button>
            </div>

            <div class="results-list">
              <div
                v-for="result in displayResults"
                :key="result.id || result.tag"
                class="result-item"
                @click="result.username ? goToProfile(result.id) : goToHashtag(result.tag)"
              >
                <div
                  v-if="result.username"
                  class="user-result"
                >
                  <img
                    :src="result.profile_picture ? loadPhotoUrl(result.profile_picture) : defaultProfileIcon"
                    width="40"
                    height="40"
                    class="result-avatar"
                    @error="(e: Event) => (e.target as HTMLImageElement).src = defaultProfileIcon"
                  />
                  <div class="result-info">
                    <div class="result-name">
                      {{ result.name }}
                    </div>
                    <div class="result-username">
                      @{{ result.username }}
                    </div>
                  </div>
                </div>
                <div
                  v-else
                  class="hashtag-result"
                >
                  <img
                    src="../components/icons/tag-light.png"
                    width="40"
                    height="40"
                    class="result-icon"
                  />
                  <div class="result-info">
                    <div class="result-name">
                      {{ result.tag }}
                    </div>
                    <div class="result-count">
                      {{ result.count }} posts
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div
            v-else-if="searchQuery.trim() && !loading"
            class="no-results"
          >
            <p>No results found</p>
          </div>

          <div
            v-else
            class="empty-state"
          >
            <img
              src="../components/icons/search-light.png"
              width="60"
              height="60"
            />
            <p>Try searching for people or hashtags</p>
          </div>
        </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.search-background {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  backdrop-filter: blur(3px);
  background-color: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 60px;
  z-index: 1000;
}

.search-container {
  background: rgb(35, 35, 35);
  border-radius: 12px;
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow: hidden;
  color: white;
  display: flex;
  flex-direction: column;
}

.share-mode {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.search-mode {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.share-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid rgb(60, 60, 60);
}

.back-btn {
  background: none;
  border: none;
  color: white;
  font-size: 24px;
  cursor: pointer;
  padding: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
}

.back-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.share-header h3 {
  margin: 0;
  flex: 1;
  text-align: center;
}

.share-preview {
  padding: 16px;
  border-bottom: 1px solid rgb(60, 60, 60);
}

.preview-content {
  display: flex;
  gap: 12px;
  align-items: center;
}

.preview-media {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  overflow: hidden;
  flex-shrink: 0;
}

.preview-media img,
.preview-media video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-content p {
  margin: 0;
  font-size: 13px;
  color: rgb(200, 200, 200);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.search-header-small {
  padding: 12px 16px;
  border-bottom: 1px solid rgb(60, 60, 60);
}

.selection-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border: 2px solid rgb(100, 100, 100);
  border-radius: 50%;
  flex-shrink: 0;
}

.selectable-user {
  display: flex;
  gap: 12px;
  cursor: pointer;
}

.selectable-user.selected {
  background-color: rgba(0, 149, 246, 0.1);
}

.selectable-user.selected .selection-indicator {
  background-color: rgb(0, 149, 246);
  border-color: rgb(0, 149, 246);
}

.selectable-user.selected .selection-indicator input {
  display: none;
}

.share-footer {
  padding: 16px;
  border-top: 1px solid rgb(60, 60, 60);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.selected-count {
  font-size: 14px;
  color: rgb(200, 200, 200);
}

.send-btn {
  background-color: rgb(0, 149, 246);
  border: none;
  color: white;
  padding: 8px 24px;
  border-radius: 20px;
  cursor: pointer;
  font-weight: 600;
  font-size: 14px;
  transition: background-color 0.2s;
}

.send-btn:hover {
  background-color: rgb(0, 129, 212);
}

.search-header {
  display: flex;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid rgb(60, 60, 60);
}

.search-input {
  flex: 1;
  background: rgb(22, 22, 22);
  border: none;
  border-radius: 8px;
  padding: 12px 16px;
  color: white;
  font-size: 16px;
  outline: none;
}

.search-input::placeholder {
  color: rgb(150, 150, 150);
}

.close-btn {
  background: none;
  border: none;
  color: white;
  font-size: 24px;
  cursor: pointer;
  padding: 8px;
  margin-left: 12px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.2s;
}

.close-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.mic-btn {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  padding: 8px;
  margin-left: 8px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.mic-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.mic-btn.recording {
  color: rgb(239, 68, 68);
  animation: pulse-recording 1.5s infinite;
}

@keyframes pulse-recording {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.search-content {
  max-height: 60vh;
  overflow-y: auto;
}

.skeleton-container {
  padding: 16px;
}

.skeleton-item {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}

.skeleton-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: rgb(60, 60, 60);
  animation: pulse 1.5s infinite;
}

.skeleton-info {
  flex: 1;
  margin-left: 12px;
}

.skeleton-line {
  height: 12px;
  background: rgb(60, 60, 60);
  border-radius: 4px;
  animation: pulse 1.5s infinite;
  margin-bottom: 6px;
}

.skeleton-line.short {
  width: 60%;
}

@keyframes pulse {
  0%, 100% {
    opacity: 0.6;
  }
  50% {
    opacity: 1;
  }
}

.search-history {
  padding: 16px;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.history-header h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: rgb(200, 200, 200);
}

.clear-btn {
  background: none;
  border: none;
  color: rgb(150, 150, 150);
  font-size: 12px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s;
}

.clear-btn:hover {
  color: white;
  background-color: rgba(255, 255, 255, 0.1);
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.history-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
  gap: 12px;
}

.history-item:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.history-item img {
  opacity: 0.6;
}

.history-item span {
  color: rgb(200, 200, 200);
  font-size: 14px;
}

.search-results {
  padding: 0;
}

.tabs {
  display: flex;
  border-bottom: 1px solid rgb(60, 60, 60);
  padding: 0 16px;
}

.tab-btn {
  background: none;
  border: none;
  color: rgb(150, 150, 150);
  padding: 12px 16px;
  cursor: pointer;
  font-size: 14px;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.tab-btn.active {
  color: white;
  border-bottom-color: white;
}

.results-list {
  padding: 8px;
}

.result-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.result-item:hover {
  background-color: rgba(255, 255, 255, 0.05);
}

.user-result {
  display: flex;
  align-items: center;
  width: 100%;
}

.hashtag-result {
  display: flex;
  align-items: center;
  width: 100%;
}

.result-avatar {
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgb(60, 60, 60);
}

.result-icon {
  border-radius: 8px;
  opacity: 0.8;
}

.result-info {
  margin-left: 12px;
  flex: 1;
}

.result-name {
  color: white;
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 2px;
}

.result-username {
  color: rgb(150, 150, 150);
  font-size: 12px;
}

.result-count {
  color: rgb(150, 150, 150);
  font-size: 12px;
}

.no-results {
  padding: 40px;
  text-align: center;
  color: rgb(150, 150, 150);
}

.empty-state {
  padding: 40px;
  text-align: center;
  color: rgb(150, 150, 150);
}

.empty-state p {
  margin-top: 16px;
  margin-bottom: 0;
}
</style>