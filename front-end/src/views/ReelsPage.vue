<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick } from "vue";
import PostDetail from "@/components/PostDetail.vue";
import { isTokenExpired } from "@/router";
import { usePostFetcher } from "@/composables/usePostFetcher";
import { usePhoto } from "@/composables/usePhoto";
import { useEventBus } from "@/composables/useEventBus";
import { useShare } from "@/composables/useShare";
import SideBar from "@/components/SideBar.vue";
import likeIcon from "@/components/icons/notif-light.png";
import commentIcon from "@/components/icons/comment-light.png";
import sendIcon from "@/components/icons/messages-light.png";
import saveIcon from "@/components/icons/saved-light.png";
import moreIcon from "@/components/icons/more-light.png";
import playIcon from "@/components/icons/play-light.png";
import userLight from "@/components/icons/user-light.png";

interface PostWithUser {
  post: unknown
  user: unknown
  postMedia: unknown[]
  collaborators?: unknown[]
}

import { useShowPost } from "@/composables/useOverlay";
import type PostDetailVue from "@/components/PostDetail.vue";
const { showPost } = useShowPost();
const { openShareMode } = useShare();
const { fetchPostsWithDetails  } = usePostFetcher();
const { loadPhotoUrl } = usePhoto();
const { emit } = useEventBus();

const reels = ref<PostWithUser[]>([]);
const loading = ref(false);
const error = ref("");
const currentPage = ref(1);
const hasMore = ref(true);
const currentReelIndex = ref(0);
const isPaused = ref(false);
const showMoreMenu = ref(false);
const moreMenuReelId = ref("");
const currentUserId = ref("");
const userActions = ref<Map<string, { liked: boolean; saved: boolean }>>(new Map());

const reelsContainer = ref<HTMLElement>();
const currentVideo = ref<HTMLVideoElement>();
const skeletonCount = 3;

const initAuth = async () => {
  const token = typeof localStorage !== "undefined" ? localStorage.getItem("token") : null;
  if (!token || isTokenExpired(token)) {
    if (typeof localStorage !== "undefined") {
      localStorage.removeItem("token");
    }
    return "";
  }

  try {
    const response = await fetch("https://localhost/api/v1/users/get-current-user", {
      method: "GET",
      headers: { Authorization: `Bearer ${token}` },
    });

    if (response.ok) {
      const data = await response.json();
      currentUserId.value = data.user.ID || data.user.id;
      return currentUserId.value;
    }
  } catch (error) {
    console.error("Failed to fetch current user:", error);
  }
  return "";
};

const fetchReels = async (page: number = 1, append: boolean = false) => {
  if (loading.value || !hasMore.value || !currentUserId.value) return;

  loading.value = true;
  error.value = "";

  try {
    const apiUrl = `https://localhost/api/v1/posts/get-reels-by-user/${currentUserId.value}`;

    const data = await fetchPostsWithDetails(apiUrl);


    if (data.length === 0) {
      hasMore.value = false;

    } else {
      if (append) {
        reels.value = [...reels.value, ...data];
      } else {
        reels.value = data;
      }
      currentPage.value = page;
      
      await fetchUserInteractions();

      if (reels.value.length > 0) {
        nextTick(() => {
          setTimeout(() => {
            const video = getCurrentVideo();

            if (video && typeof video.play === "function") {

              video.play().then(() => {

                isPaused.value = false;
              }).catch(err => {
                console.error("Failed to play first video:", err);
                isPaused.value = true;
              });
            } else {
              console.log("First video element not available:", currentVideo.value);
            }
          }, 1000);
        });
      }
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Failed to fetch reels";

  } finally {
    loading.value = false;
  }
};

const handleScroll = () => {
  if (!reelsContainer.value) return;

  const { scrollTop, scrollHeight, clientHeight } = reelsContainer.value;
  const threshold = 100;

  if (scrollTop + clientHeight >= scrollHeight - threshold && !loading.value && hasMore.value) {
    fetchReels(currentPage.value + 1, true);
  }

  const reelHeight = typeof window !== "undefined" ? window.innerHeight : 1000;
  const newIndex = Math.round(scrollTop / reelHeight);
  if (newIndex !== currentReelIndex.value && newIndex >= 0 && newIndex < reels.value.length) {
    currentReelIndex.value = newIndex;
    playCurrentReel();
  }
};

const getCurrentVideo = () => {
  const videoRef = currentVideo.value;
  if (Array.isArray(videoRef) && videoRef.length > 0) {
    return videoRef[0] as HTMLVideoElement;
  }
  return videoRef as HTMLVideoElement;
};

const playCurrentReel = () => {
  nextTick(() => {
    const video = getCurrentVideo();


    if (video && typeof video.play === "function") {

      video.play().then(() => {
        isPaused.value = false;
      }).catch(err => {
        console.error("Failed to play video:", err);
        isPaused.value = true;
      });
    } else {
      console.log("Current video not available or invalid:", currentVideo.value);
    }
  });
};

const togglePlayPause = () => {
  const video = getCurrentVideo();
  if (!video || typeof video.play !== "function" || typeof video.pause !== "function") {
    return;
  }

  if (isPaused.value) {

    video.play().then(() => {
      isPaused.value = false;
    }).catch(err => {
      console.error("Failed to play video:", err);
    });
  } else {

    video.pause();
    isPaused.value = true;
  }
};

const handleVideoEnd = () => {
  if (currentReelIndex.value < reels.value.length - 1) {
    const nextIndex = currentReelIndex.value + 1;
    const nextReelElement = document.getElementById(`reel-${nextIndex}`);
    if (nextReelElement && reelsContainer.value) {
      nextReelElement.scrollIntoView({ behavior: "smooth" });
    }
  }
};

const handleReelLoad = (index: number) => {
  if (index === currentReelIndex.value && index === 0) {
    isPaused.value = false;
  }
};

const fetchUserInteractions = async () => {
  const token = localStorage.getItem("token");
  if (!token) return;

  for (const reel of reels.value) {
    try {
      const [likesRes, savesRes] = await Promise.allSettled([
        fetch(`https://localhost/api/v1/posts/check-like/${reel.post.ID}`, {
          headers: { Authorization: `Bearer ${token}` }
        }),
        fetch(`https://localhost/api/v1/posts/check-save/${reel.post.ID}`, {
          headers: { Authorization: `Bearer ${token}` }
        })
      ]);

      const currentActions = userActions.value.get(reel.post.ID) || { liked: false, saved: false };

      if (likesRes.status === "fulfilled" && likesRes.value.ok) {
        try {
          const data = await likesRes.value.json();
          currentActions.liked = data.liked || false;
        } catch (e) {
          currentActions.liked = false;
        }
      }

      if (savesRes.status === "fulfilled" && savesRes.value.ok) {
        try {
          const data = await savesRes.value.json();
          currentActions.saved = data.saved || false;
        } catch (e) {
          currentActions.saved = false;
        }
      }

      userActions.value.set(reel.post.ID, currentActions);
    } catch (error) {
      console.error("Error fetching interactions:", error);
    }
  }
};

const handleLike = async (reel: PostWithUser) => {
  const postId = reel.post.ID;
  const currentActions = userActions.value.get(postId) || { liked: false, saved: false };

  try {
    const response = await fetch(`https://localhost/api/v1/posts/like-post/${postId}`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${typeof localStorage !== "undefined" ? localStorage.getItem("token") : ""}`,
      },
    });

    if (response.ok) {
      
      if (currentActions.liked) {
        reel.post.LikesCount = Math.max(0, (reel.post.LikesCount || 0) - 1);
        currentActions.liked = false;
      } else {
        reel.post.LikesCount = (reel.post.LikesCount || 0) + 1;
        currentActions.liked = true;
      }

      userActions.value.set(postId, currentActions);
      emit("post-liked", { postID: postId });
    }
  } catch (err) {
    console.error("Failed to toggle like:", err);
  }
};


const handleSave = async (reel: PostWithUser) => {
  const postId = reel.post.ID;
  const currentActions = userActions.value.get(postId) || { liked: false, saved: false };

  try {
    const response = await fetch(`https://localhost/api/v1/posts/save-post/${postId}`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${typeof localStorage !== "undefined" ? localStorage.getItem("token") : ""}`,
      },
    });

    if (response.ok) {
      const data = await response.json();
      currentActions.saved = data.saved !== undefined ? data.saved : !currentActions.saved;
      userActions.value.set(postId, currentActions);
    }
  } catch (err) {
    console.error("Failed to toggle save:", err);
  }
};


const handleComment = (postID: string) => {
  showPost(postID);
};

const handleSend = (reel: any) => {
  openShareMode(reel.post);
};

const handleMore = (reelId: string) => {
  moreMenuReelId.value = reelId;
  showMoreMenu.value = true;
};

const handleReport = () => {

  showMoreMenu.value = false;
};

const handleGoToPost = () => {
  emit("show-post-detail", moreMenuReelId.value);
  showMoreMenu.value = false;
};

const copiedLinkId = ref<string | null>(null);

const handleCopyLink = async () => {
  if (typeof window !== "undefined" && navigator.clipboard) {
    const link = `${window.location.origin}/reel/${moreMenuReelId.value}`;
    try {
      await navigator.clipboard.writeText(link);
      copiedLinkId.value = moreMenuReelId.value;
      setTimeout(() => {
        copiedLinkId.value = null;
      }, 2000);
    } catch (err) {
      console.error("Failed to copy link:", err);
    }
  }
  showMoreMenu.value = false;
};

const handleDeleteReel = async () => {
  const token = localStorage.getItem("token");
  if (!token || !moreMenuReelId.value) return;

  try {
    const response = await fetch(`https://localhost/api/v1/posts/delete-post/${moreMenuReelId.value}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      reels.value = reels.value.filter(r => r.post.ID !== moreMenuReelId.value);
      showMoreMenu.value = false;
      moreMenuReelId.value = "";
    }
  } catch (error) {
    console.error("Failed to delete reel:", error);
  }
};

onMounted(async () => {
  const userId = await initAuth();
  if (userId) {
    await fetchReels();
  }

  if (reelsContainer.value) {
    reelsContainer.value.addEventListener("scroll", handleScroll, { passive: true });
  }
});

onUnmounted(() => {
  if (reelsContainer.value) {
    reelsContainer.value.removeEventListener("scroll", handleScroll);
  }
});


const currentReel = computed(() => reels.value[currentReelIndex.value]);
const showSkeletons = computed(() => loading.value && reels.value.length === 0);
const showLoadMoreSkeletons = computed(() => loading.value && reels.value.length > 0);
</script>

<template>
  <div class="container">
    <SideBar />

    <div
      ref="reelsContainer"
      class="reels-container"
    >
      <div
        v-if="showSkeletons"
        class="skeleton-container"
      >
        <div
          v-for="i in skeletonCount"
          :key="i"
          class="skeleton-reel"
        >
          <div class="skeleton-video" />
          <div class="skeleton-overlay">
            <div class="skeleton-actions">
              <div class="skeleton-icon" />
              <div class="skeleton-icon" />
              <div class="skeleton-icon" />
              <div class="skeleton-icon" />
            </div>
          </div>
        </div>
      </div>

      <div
        v-for="(reel, index) in reels"
        :id="`reel-${index}`"
        :key="reel.post.ID"
        class="reel"
      >
        <div
          v-if="reel.postMedia.length > 0"
          class="reel-media"
        >
          <video
            v-if="reel.postMedia[0].MediaType === 'video'"
            :ref="index === currentReelIndex ? 'currentVideo' : undefined"
            :src="loadPhotoUrl(reel.postMedia[0].FilePath)"
            class="reel-video"
            :loop="index !== currentReelIndex"
            :playsinline="true"
            :autoplay="index === currentReelIndex && index === 0"
            :muted="index !== currentReelIndex"
            @ended="handleVideoEnd"
            @click="togglePlayPause"
            @load="() => handleReelLoad(index)"
          />
          <img
            v-else
            :src="loadPhotoUrl(reel.postMedia[0].FilePath)"
            class="reel-image"
            @click="togglePlayPause"
          />
        </div>


        <div class="reel-overlay">
          <div
            v-if="isPaused"
            class="play-pause-indicator"
          >
            <img
              :src="playIcon"
              width="80"
              height="80"
            />
          </div>

          <div class="reel-user">
            <img
              :src="reel.user.ProfilePicture ? loadPhotoUrl(reel.user.ProfilePicture) : userLight"
              class="user-avatar"
            />
            <span class="username">{{ reel.user.Username }}</span>
          </div>
          <p style="font-size: 18px; font-weight: 600; padding-bottom: 10px;">
            {{ reel.post.Caption }}
          </p>
        </div>


        <div class="reel-actions">
          <button
            class="action-btn"
            @click="handleLike(reel)"
          >
            <img
              :src="likeIcon"
              width="32"
              height="32"
              class="action-icon"
              :class="{ 'liked': userActions.get(reel.post.ID)?.liked }"
            />
            <span class="action-count">{{ reel.post.LikesCount || 0 }}</span>
          </button>

          <button
            class="action-btn"
            @click="handleComment(reel.post.ID)"
          >
            <img
              :src="commentIcon"
              width="32"
              height="32"
              class="action-icon"
            />
            <span class="action-count">{{ reel.post.CommentsCount || 0 }}</span>
          </button>

          <button
            class="action-btn"
            @click="handleSend(reel.post.ID)"
          >
            <img
              :src="sendIcon"
              width="32"
              height="32"
              class="action-icon"
            />
          </button>

          <button
            class="action-btn"
            @click="handleSave(reel)"
          >
            <img
              :src="saveIcon"
              width="32"
              height="32"
              class="action-icon"
              :class="{ 'saved': userActions.get(reel.post.ID)?.saved }"
            />
          </button>

          <button
            class="action-btn"
            @click="handleMore(reel.post.ID)"
          >
            <img
              :src="moreIcon"
              width="32"
              height="32"
              class="action-icon"
            />
          </button>
        </div>
      </div>
    </div>

    <div
      v-if="showLoadMoreSkeletons"
      class="skeleton-container"
    >
      <div
        v-for="i in 2"
        :key="`load-${i}`"
        class="skeleton-reel"
      >
        <div class="skeleton-video" />
      </div>
    </div>

    <div
      v-if="error"
      class="error-message"
    >
      <p>{{ error }}</p>
      <button
        class="retry-btn"
        @click="fetchReels()"
      >
        Retry
      </button>
    </div>
  </div>

  <div
    v-if="showSendModal"
    class="modal-overlay"
    @click.self="showSendModal = false"
  >
    <div class="modal-content">
      <h3>Send Reel</h3>
      <p>Choose a recipient to send this reel to</p>

      <div class="modal-actions">
        <button
          class="cancel-btn"
          @click="showSendModal = false"
        >
          Cancel
        </button>
        <button
          class="send-btn"
          @click="showSendModal = false"
        >
          Send
        </button>
      </div>
    </div>
  </div>

  <div
    v-if="showMoreMenu"
    class="modal-overlay"
    @click.self="showMoreMenu = false"
  >
    <div class="more-menu">
      <button
        class="menu-item"
        @click="handleReport"
      >
        Report
      </button>
      <button
        class="menu-item"
        @click="handleGoToPost"
      >
        Go to Post
      </button>
      <button
        class="menu-item"
        @click="handleCopyLink"
      >
        Copy Link
      </button>
      <button
        v-if="currentReel && currentReel.user.ID === currentUserId"
        class="menu-item delete-option"
        @click="handleDeleteReel"
      >
        Delete
      </button>
    </div>
  </div>

  <div
    v-if="copiedLinkId"
    class="toast-notification"
  >
    Link copied to clipboard!
  </div>

  <div
    v-if="showSendModal"
    class="modal-overlay"
    @click.self="showSendModal = false"
  >
    <div class="share-modal">
      <h3>Send Reel</h3>
      <input
        v-model="shareSearchQuery"
        type="text"
        placeholder="Search users..."
        class="search-input"
        @input="searchShareUsers"
      />
      <div
        v-if="loadingShareSearch"
        class="loading"
      >
        Searching...
      </div>
      <div
        v-else-if="shareSearchResults.length === 0 && shareSearchQuery"
        class="no-data"
      >
        No users found
      </div>
      <div
        v-else-if="shareSearchResults.length > 0"
        class="conversations-list"
      >
        <div
          v-for="user in shareSearchResults"
          :key="user.ID || user.id"
          class="conversation-item user-item"
          :class="{ selected: selectedShareUserId === (user.ID || user.id) }"
          @click="selectedShareUserId = user.ID || user.id"
        >
          <div class="conversation-info">
            <img
              :src="user.ProfilePicture ? loadPhotoUrl(user.ProfilePicture) : userLight"
              class="user-avatar"
            />
            <span class="conversation-name">{{ user.Username }}</span>
          </div>
        </div>
      </div>
      <div class="modal-actions">
        <button
          class="cancel-btn"
          @click="showSendModal = false"
        >
          Cancel
        </button>
        <button
          class="send-btn"
          :disabled="!selectedShareUserId"
          @click="shareReelToUser"
        >
          Send
        </button>
      </div>
    </div>
  </div>
  <PostDetail />
</template>

<style scoped>
.container {
  display: flex;
  background-color: black;
  height: 100vh;
  overflow: hidden;
}

.reels-container {
  flex: 1;
  height: 100vh;
  overflow-y: auto;
  scroll-snap-type: y mandatory;
  scroll-behavior: smooth;
  position: relative;
  display: flex;
  justify-content: center;
}

.reel {
  height: 100vh;
  width: 50%;
  scroll-snap-align: start;
  position: relative;
  background-color: #000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.reel-media {
  width: 100%;
  height: 100%;
  position: relative;
}

.reel-video,
.reel-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
  cursor: pointer;
}

.reel-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(to bottom,
    rgba(0,0,0,0.7) 0%,
    transparent 20%,
    transparent 80%,
    rgba(0,0,0,0.7) 100%
  );
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 20px;
  pointer-events: none;
}

.play-pause-indicator {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 10;
  filter: brightness(0) invert(1);
}

.reel-user {
  display: flex;
  align-items: center;
  gap: 12px;
  color: white;
  font-weight: 600;
  font-size: 16px;
  z-index: 5;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid white;
}

.username {
  text-shadow: 1px 1px 2px rgba(0,0,0,0.8);
}

.reel-caption {
    font-size: 14px;
  color: white;
  max-width: 80%;
  z-index: 5;
}

.reel-caption p {
  margin: 0;
  font-size: 14px;
  text-shadow: 1px 1px 2px rgba(0,0,0,0.8);
  line-height: 1.4;
}

.reel-actions {
  display: flex;
  flex-direction: column;
  gap: 20px;
  align-items: center;
  position: absolute;
  right: 20px;
  bottom: 100px;
  z-index: 5;
}

.action-btn {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  transition: transform 0.2s;
  pointer-events: all;
}

.action-btn:hover {
  transform: scale(1.1);
}

.action-icon {
  filter: brightness(0) invert(1);
  transition: filter 0.2s;
}

.action-icon.liked {
  filter: invert(27%) sepia(98%) saturate(7471%) hue-rotate(3deg) brightness(102%) contrast(115%);
}

.action-icon.saved {
  filter: invert(85%) sepia(89%) saturate(1991%) hue-rotate(358deg) brightness(101%) contrast(103%);
}

.action-count {
  font-size: 12px;
  font-weight: 600;
  text-shadow: 1px 1px 2px rgba(0,0,0,0.8);
}

.skeleton-container {
  height: 100vh;
  width: 100%;
  scroll-snap-align: start;
}

.skeleton-reel {
  height: 100vh;
  width: 100%;
  position: relative;
  background-color: #111;
  display: flex;
  align-items: center;
  justify-content: center;
}

.skeleton-video {
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, #222 25%, #333 50%, #222 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s infinite;
}

.skeleton-overlay {
  position: absolute;
  bottom: 100px;
  right: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.skeleton-actions {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.skeleton-icon {
  width: 32px;
  height: 32px;
  background: linear-gradient(90deg, #333 25%, #444 50%, #333 75%);
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s infinite;
  border-radius: 50%;
}

@keyframes skeleton-loading {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 24px;
  max-width: 400px;
  width: 90%;
  color: white;
}

.modal-content h3 {
  margin: 0 0 12px 0;
  font-size: 18px;
  font-weight: 600;
}

.modal-content p {
  margin: 0 0 20px 0;
  font-size: 14px;
  color: #ccc;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.cancel-btn,
.send-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
}

.cancel-btn {
  background: #333;
  color: white;
}

.cancel-btn:hover {
  background: #444;
}

.send-btn {
  background: #0095f6;
  color: white;
}

.send-btn:hover {
  background: #0078d4;
}

.more-menu {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 8px 0;
  min-width: 200px;
  color: white;
}

.menu-item {
  width: 100%;
  padding: 12px 20px;
  background: none;
  border: none;
  color: white;
  text-align: left;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
}

.menu-item:hover {
  background: #333;
}

.delete-option {
  color: #ed4956;
}

.share-modal {
  background: #1a1a1a;
  border-radius: 12px;
  padding: 24px;
  max-width: 450px;
  width: 90%;
  color: white;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.share-modal h3 {
  margin: 0 0 16px 0;
  font-size: 20px;
  font-weight: 600;
  text-align: center;
  border-bottom: 1px solid #333;
  padding-bottom: 12px;
}

.loading,
.no-data {
  padding: 40px 20px;
  text-align: center;
  color: #888;
  font-size: 14px;
}

.conversations-list {
  max-height: 400px;
  overflow-y: auto;
  margin-bottom: 16px;
  flex: 1;
}

.conversations-list::-webkit-scrollbar {
  width: 6px;
}

.conversations-list::-webkit-scrollbar-track {
  background: #1a1a1a;
}

.conversations-list::-webkit-scrollbar-thumb {
  background: #555;
  border-radius: 3px;
}

.conversation-item {
  padding: 14px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
  border-radius: 8px;
  margin-bottom: 6px;
  border: 1px solid #2a2a2a;
}

.conversation-item:hover {
  background: #2a2a2a;
  border-color: #0095f6;
}

.conversation-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.conversation-name {
  font-size: 15px;
  font-weight: 500;
}

.error-message {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 20px;
  border-radius: 12px;
  text-align: center;
  z-index: 100;
}

.retry-btn {
  margin-top: 12px;
  padding: 8px 16px;
  background: #0095f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
}

.retry-btn:hover {
  background: #0078d4;
}

.search-input {
  width: 100%;
  padding: 12px;
  background: #121212;
  border: 1px solid #333;
  border-radius: 8px;
  color: white;
  font-size: 14px;
  margin-bottom: 16px;
}

.search-input:focus {
  outline: none;
  border-color: #0095f6;
}

.user-item.selected {
  background: #0095f6;
  border-color: #0095f6;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  margin-right: 12px;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 16px;
}

.send-btn:disabled {
  background: #555;
  cursor: not-allowed;
  opacity: 0.6;
}

.toast-notification {
  position: fixed;
  bottom: 30px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.9);
  color: white;
  padding: 12px 24px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  z-index: 2000;
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translate(-50%, 20px);
  }
  to {
    opacity: 1;
    transform: translate(-50%, 0);
  }
}
</style>