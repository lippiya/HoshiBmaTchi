<script setup lang="ts">
import SideBar from "@/components/SideBar.vue";
import PostDetail from "@/components/PostDetail.vue";
import { onMounted, onUnmounted, ref, nextTick, watch, computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import { isTokenExpired } from "@/router";
import { usePostFetcher } from "@/composables/usePostFetcher";
import { usePhoto } from "@/composables/usePhoto";
import { useShowPost } from "@/composables/useOverlay";
import { useHashtag } from "@/composables/useHashtag";
import likeIcon from "@/components/icons/notif-light.png";
import commentIcon from "@/components/icons/comment-light.png";

const { fetchPostsWithDetails } = usePostFetcher();
const { loadPhotoUrl } = usePhoto();
const { showPost } = useShowPost();
const { extractHashtags, navigateToHashtag } = useHashtag();
const router = useRouter();
const route = useRoute();

interface PostWithUser {
  post: any;
  user: any;
  postMedia: any[];
  collaborators?: any[];
}

interface HashtagCount {
  tag: string;
  count: number;
}

const postsWithUsers = ref<PostWithUser[]>([]);
const allPosts = ref<PostWithUser[]>([]);
const currentPage = ref(1);
const loading = ref(false);
const initialLoading = ref(true);
const hasMore = ref(true);
const exploreContainer = ref<HTMLElement>();
const searchQuery = ref("");
const activeHashtag = ref("");
const hashtagCounts = ref<HashtagCount[]>([]);

const calculateHashtagCounts = () => {
  const counts: Record<string, number> = {};
  
  for (const postData of allPosts.value) {
    const caption = postData.post.Caption || "";
    const tags = extractHashtags(caption);
    
    for (const tag of tags) {
      const normalizedTag = tag.toLowerCase().replace("#", "");
      counts[normalizedTag] = (counts[normalizedTag] || 0) + 1;
    }
  }
  
  hashtagCounts.value = Object.entries(counts)
    .map(([tag, count]) => ({ tag, count }))
    .sort((a, b) => b.count - a.count);
};

const filteredPosts = computed(() => {
  if (!activeHashtag.value) {
    return allPosts.value;
  }
  
  const searchTag = activeHashtag.value.toLowerCase().replace("#", "");
  
  return allPosts.value.filter(postData => {
    const caption = postData.post.Caption || "";
    const tags = extractHashtags(caption).map(t => t.toLowerCase().replace("#", ""));
    return tags.includes(searchTag);
  });
});

const filteredHashtags = computed(() => {
  if (activeHashtag.value) {
    return [];
  }
  
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase().replace("#", "");
    return hashtagCounts.value.filter(hc => 
      hc.tag.toLowerCase().includes(query)
    ).slice(0, 10);
  }
  
  return hashtagCounts.value.slice(0, 10);
});

const fetchPosts = async (page: number = 1, append: boolean = false) => {
  if (loading.value) return;

  const token = localStorage.getItem("token");
  if (!token || isTokenExpired(token)) {
    localStorage.removeItem("token");
    router.push("/login");
    return;
  }

  loading.value = true;

  try {
    const limit = 20;

    if (allPosts.value.length === 0) {
      const data = await fetchPostsWithDetails("https://localhost/api/v1/posts/get-all-posts");
      allPosts.value = data;
      calculateHashtagCounts();
    }

    const postsToDisplay = filteredPosts.value;
    const startIndex = (page - 1) * limit;
    const endIndex = startIndex + limit;
    const paginatedData = postsToDisplay.slice(startIndex, endIndex);
    
    if (paginatedData.length === 0) {
      hasMore.value = false;
    } else {
      if (append) {
        postsWithUsers.value = [...postsWithUsers.value, ...paginatedData];
      } else {
        postsWithUsers.value = paginatedData;
      }
      currentPage.value = page;
      

      if (endIndex >= postsToDisplay.length) {
        hasMore.value = false;
      } else {
        hasMore.value = true;
      }
    }
  } catch (error) {
    console.error("Failed to fetch posts:", error);
  } finally {
    loading.value = false;
    initialLoading.value = false;
  }
};

const selectHashtag = (tag: string) => {
  activeHashtag.value = tag;
  searchQuery.value = "";
  router.push({ name: "explore", query: { hashtag: tag } });
  currentPage.value = 1;
  hasMore.value = true;
  fetchPosts(1, false);
};

const clearSearch = () => {
  activeHashtag.value = "";
  searchQuery.value = "";
  router.push({ name: "explore" });
  currentPage.value = 1;
  hasMore.value = true;
  fetchPosts(1, false);
};

const handleScroll = () => {
  if (!loading.value && hasMore.value) {
    const container = exploreContainer.value;
    if (!container) return;
    
    const scrollTop = container.scrollTop;
    const scrollHeight = container.scrollHeight;
    const clientHeight = container.clientHeight;
    
    if (scrollTop + clientHeight >= scrollHeight - 100) {
      fetchPosts();
    }
  }
};

onUnmounted(() => {
  if (exploreContainer.value) {
    exploreContainer.value.removeEventListener("scroll", handleScroll);
  }
});

watch(() => route.query.hashtag, (newHashtag) => {
  if (newHashtag) {
    activeHashtag.value = newHashtag as string;
    currentPage.value = 1;
    hasMore.value = true;
    fetchPosts(1, false);
  } else if (activeHashtag.value) {
    clearSearch();
  }
});

const handlePostClick = (postID: string) => {
  showPost(postID);
};

const getGridClass = (post: any) => {
  return post.IsReel ? "grid-item-tall" : "grid-item-normal";
};

onMounted(async () => {
  await fetchPosts();
  await nextTick();
  if (exploreContainer.value) {
    exploreContainer.value.addEventListener("scroll", handleScroll, { passive: true });
  }
});

onUnmounted(() => {
  if (exploreContainer.value) {
    exploreContainer.value.removeEventListener("scroll", handleScroll);
  }
});
</script>

<template>
  <PostDetail />
  <div class="container">
    <SideBar />
    <div
      ref="exploreContainer"
      class="explore"
    >
      <h1 class="explore-title">
        Explore
      </h1>

      <div class="search-section">
        <div class="search-bar">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search hashtags..."
            class="search-input"
            @keyup.enter="searchQuery && selectHashtag(searchQuery)"
          />
          <button
            v-if="activeHashtag || searchQuery"
            class="clear-btn"
            @click="clearSearch"
          >
            ×
          </button>
        </div>

        <div
          v-if="activeHashtag"
          class="active-hashtag"
        >
          <span class="hashtag-label">Active Filter:</span>
          <span class="hashtag-value">#{{ activeHashtag.replace('#', '') }}</span>
          <span class="post-count">{{ filteredPosts.length }} posts</span>
        </div>

        <div
          v-if="(searchQuery || !activeHashtag) && filteredHashtags.length > 0"
          class="hashtag-suggestions"
        >
          <h3 class="suggestions-title">
            {{ searchQuery ? "Related Hashtags" : "Popular Hashtags" }}
          </h3>
          <div class="hashtag-list">
            <div
              v-for="hc in filteredHashtags"
              :key="hc.tag"
              class="hashtag-item"
              @click="selectHashtag(hc.tag)"
            >
              <span class="hashtag-name">#{{ hc.tag }}</span>
              <span class="hashtag-count">{{ hc.count }} posts</span>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="initialLoading"
        class="grid-container"
      >
        <div
          v-for="i in 12"
          :key="i"
          class="skeleton-item grid-item-normal"
        >
          <div class="skeleton-shimmer" />
        </div>
      </div>

      <div
        v-else
        class="grid-container"
      >
        <div
          v-for="postData in postsWithUsers"
          :key="postData.post.ID"
          class="grid-item"
          :class="getGridClass(postData.post)"
          @click="handlePostClick(postData.post.ID)"
        >
          <div
            v-if="postData.postMedia.length > 0"
            class="media-container"
          >
            <img
              v-if="postData.postMedia[0].MediaType === 'image'"
              :src="loadPhotoUrl(postData.postMedia[0].FilePath)"
              class="post-image"
            />
            <div
              v-else-if="postData.postMedia[0].MediaType === 'video'"
              class="video-thumbnail"
            >
              <video
                :src="loadPhotoUrl(postData.postMedia[0].FilePath)"
                class="post-video"
                muted
                preload="metadata"
              />
            </div>
          </div>
          <div
            v-else
            class="no-media"
          >
            No Media
          </div>

          <div class="hover-overlay">
            <div class="engagement-stats">
              <div class="stat-item">
                <img
                  :src="likeIcon"
                  width="24"
                  height="24"
                  class="stat-icon"
                />
                <span class="stat-count">{{ postData.post.LikesCount || 0 }}</span>
              </div>
              <div class="stat-item">
                <img
                  :src="commentIcon"
                  width="24"
                  height="24"
                  class="stat-icon"
                />
                <span class="stat-count">{{ postData.post.CommentsCount || 0 }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="loading && !initialLoading"
        class="loading-more"
      >
        <div class="spinner" />
      </div>

      <div
        v-if="!hasMore && postsWithUsers.length > 0"
        class="end-message"
      >
        You have reached the end
      </div>
    </div>
  </div>
</template>

<style scoped>
.container {
  width: 100%;
  height: 100vh;
  display: flex;
  background-color: black;
}

.explore {
  width: calc(100% - 250px);
  padding: 40px;
  overflow-y: auto;
  background-color: black;
}

.explore-title {
  color: white;
  font-size: 32px;
  font-weight: 600;
  margin-bottom: 30px;
}

.search-section {
  margin-bottom: 30px;
}

.search-bar {
  position: relative;
  margin-bottom: 20px;
}

.search-input {
  width: 100%;
  padding: 12px 40px 12px 16px;
  background-color: #1a1a1a;
  border: 1px solid #333;
  border-radius: 8px;
  color: white;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.search-input:focus {
  border-color: #0095f6;
}

.search-input::placeholder {
  color: #666;
}

.clear-btn {
  position: absolute;
  right: 10px;
  top: 50%;
  transform: translateY(-50%);
  background: transparent;
  border: none;
  color: #666;
  font-size: 24px;
  cursor: pointer;
  padding: 0 8px;
  transition: color 0.2s;
}

.clear-btn:hover {
  color: white;
}

.active-hashtag {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background-color: #1a1a1a;
  border-radius: 8px;
  margin-bottom: 20px;
}

.hashtag-label {
  color: #999;
  font-size: 14px;
}

.hashtag-value {
  color: #0095f6;
  font-size: 16px;
  font-weight: 600;
}

.post-count {
  color: #666;
  font-size: 14px;
  margin-left: auto;
}

.hashtag-suggestions {
  margin-bottom: 30px;
}

.suggestions-title {
  color: white;
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
}

.hashtag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.hashtag-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background-color: #1a1a1a;
  border: 1px solid #333;
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.2s;
}

.hashtag-item:hover {
  background-color: #262626;
  border-color: #0095f6;
  transform: translateY(-2px);
}

.hashtag-name {
  color: #0095f6;
  font-size: 14px;
  font-weight: 500;
}

.hashtag-count {
  color: #666;
  font-size: 12px;
}

.grid-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  grid-auto-flow: dense;
  gap: 8px;
  width: 100%;
}

@media (max-width: 768px) {
  .grid-container {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 769px) and (max-width: 1200px) {
  .grid-container {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (min-width: 1201px) {
  .grid-container {
    grid-template-columns: repeat(4, 1fr);
  }
}

.grid-item {
  position: relative;
  overflow: hidden;
  cursor: pointer;
  background-color: #121212;
  border-radius: 4px;
}

.grid-item-normal {
  grid-column: span 1;
  grid-row: span 1;
  aspect-ratio: 1 / 1;
}

.grid-item-tall {
  grid-column: span 1;
  grid-row: span 2;
  aspect-ratio: 1 / 2;
}

.media-container {
  width: 100%;
  height: 100%;
  position: relative;
}

.post-image,
.post-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.video-thumbnail {
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
}

.no-media {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #666;
  font-size: 14px;
}

.hover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  opacity: 0;
  transition: opacity 0.3s ease;
  z-index: 10;
}

.grid-item:hover .hover-overlay {
  opacity: 1;
}

.engagement-stats {
  display: flex;
  gap: 30px;
  align-items: center;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: white;
  font-weight: 700;
  font-size: 16px;
}

.stat-icon {
  filter: brightness(0) invert(1);
  opacity: 0.9;
}

.stat-count {
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
}

.skeleton-item {
  background: #1a1a1a;
  position: relative;
  overflow: hidden;
  border-radius: 4px;
}

.skeleton-shimmer {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(255, 255, 255, 0.05),
    transparent
  );
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}

.loading-more {
  display: flex;
  justify-content: center;
  padding: 40px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #333;
  border-top-color: #0095f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.end-message {
  text-align: center;
  padding: 40px;
  color: #666;
  font-size: 14px;
}
</style>
