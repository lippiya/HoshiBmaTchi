<script setup lang="ts">
import SideBar from "@/components/SideBar.vue";
import PostDetail from "@/components/PostDetail.vue";
import StoryCarousel from "@/components/StoryCarousel.vue";
import StoryViewer from "@/components/StoryViewer.vue";
import { onMounted, onUnmounted, ref } from "vue";
import { useRouter } from "vue-router";
import { isTokenExpired } from "@/router";
import { usePostFetcher } from "@/composables/usePostFetcher";
const { fetchPostsWithDetails } = usePostFetcher();

import { usePhoto } from "@/composables/usePhoto";
const { loadPhotoUrl } = usePhoto();

import { useEventBus } from "@/composables/useEventBus";
const { on, off } = useEventBus();

import { useShowPost } from "@/composables/useOverlay";
const { showPost } = useShowPost();

import { useShare } from "@/composables/useShare";
const { openShareMode } = useShare();

const { formatTextWithHashtags, navigateToHashtag } = useHashtag();
const { createNotification } = useNotification();

import { getCollectionsByUserID, addPostToCollection, removePostFromCollection } from "@/services/collectionService";
import { storyService, type StoryGroup } from "@/services/storyService";
import { useHashtag } from "@/composables/useHashtag";
import { useNotification } from "@/composables/useNotification";

import likeLight from "@/components/icons/notif-light.png";
import commentLight from "@/components/icons/comment-light.png";
import sendLight from "@/components/icons/messages-light.png";
import savedLight from "@/components/icons/saved-light.png";
import savedDark from "@/components/icons/saved-dark.png";
import userLight from "@/components/icons/user-light.png";

interface PostWithUser {
  post: any;
  user: any;
  postMedia: any[];
  collaborators?: any[];
}

const currUserID = ref<string>("");
const router = useRouter();
const userLikes = ref<Record<string, boolean>>({});
const userComments = ref<Record<string, boolean>>({});
const userSaves = ref<Record<string, boolean>>({});

const showSaveCollectionModal = ref(false);
const userCollections = ref<any[]>([]);
const selectedCollectionForSave = ref<string>("");
const savingPostId = ref<string>("");
const loadingCollections = ref(false);

const storyGroups = ref<StoryGroup[]>([]);
const showStoryViewer = ref(false);
const selectedStoryGroupIndex = ref(0);

const postsWithUsers = ref<PostWithUser[]>([]);
const currentSlides = ref<Record<string, number>>({});
const currentPage = ref(1);
const loading = ref(false);
const initialLoading = ref(true);
const hasMore = ref(true);
const postsContainer = ref<HTMLElement>();
const showMoreOverlay = ref<string | null>(null);
const copiedPostId = ref<string | null>(null);
const loadedPostIds = ref<Set<string>>(new Set());
const suggestedUsers = ref<any[]>([]);

const fetchPost = async (page: number = 1, append: boolean = false) => {
  if (loading.value || (!append && !hasMore.value)) return;

  loading.value = true;
  const token = localStorage.getItem("token");
  if (!token || isTokenExpired(token)) {
    localStorage.removeItem("token");
    router.push("/login");
    return;
  }

  try {
    const response = await fetch("https://localhost/api/v1/users/get-current-user", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    if (!response.ok) throw new Error("Failed to fetch current user");
    const userData = await response.json();
    currUserID.value = userData.user.ID || userData.user.id;
  } catch (error) {
    console.error("Fatal error fetching current user:", error);
  }

  try {
    const followedResponse = await fetch("https://localhost/api/v1/social/get-following", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    let followedPosts: PostWithUser[] = [];
    let otherPosts: PostWithUser[] = [];

    if (followedResponse.ok) {
      const followedData = await followedResponse.json();
      const followedUserIds = followedData.following?.map((u: any) => u.ID || u.id) || [];

      const allPosts = await fetchPostsWithDetails("https://localhost/api/v1/posts/get-all-posts");
      
      followedPosts = allPosts.filter((p: PostWithUser) => 
        followedUserIds.includes(p.user.ID || p.user.id)
      );
      otherPosts = allPosts.filter((p: PostWithUser) => 
        !followedUserIds.includes(p.user.ID || p.user.id)
      );

      followedPosts.sort((a, b) => new Date(b.post.CreatedAt).getTime() - new Date(a.post.CreatedAt).getTime());
      otherPosts.sort((a, b) => new Date(b.post.CreatedAt).getTime() - new Date(a.post.CreatedAt).getTime());
    } else {
      otherPosts = await fetchPostsWithDetails("https://localhost/api/v1/posts/get-all-posts");
      otherPosts.sort((a, b) => new Date(b.post.CreatedAt).getTime() - new Date(a.post.CreatedAt).getTime());
    }

    const sortedPosts = [...followedPosts, ...otherPosts];

    const uniquePosts = sortedPosts.filter(p => {
      const postId = p.post.ID;
      if (loadedPostIds.value.has(postId)) {
        return false;
      }
      loadedPostIds.value.add(postId);
      return true;
    });

    if (uniquePosts.length === 0) {
      hasMore.value = false;
    } else {
      if (append) {
        postsWithUsers.value = [...postsWithUsers.value, ...uniquePosts];
      } else {
        loadedPostIds.value.clear();
        uniquePosts.forEach(p => loadedPostIds.value.add(p.post.ID));
        postsWithUsers.value = uniquePosts;
      }
      currentPage.value = page;
    }
  } catch (error) {
    console.error("Failed to fetch posts:", error);
  }

  await fetchUserInteractions();
  loading.value = false;
  initialLoading.value = false;
};

const fetchStories = async () => {
  try {
    const groups = await storyService.getAllActiveStories();
    storyGroups.value = groups;
  } catch (error) {
    console.error("Failed to fetch stories:", error);
  }
};

const openStoryViewer = (groupIndex: number) => {
  selectedStoryGroupIndex.value = groupIndex;
  showStoryViewer.value = true;
};

const closeStoryViewer = () => {
  showStoryViewer.value = false;
  fetchStories();
};

const fetchSuggestedUsers = async () => {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch("https://localhost/api/v1/users/get-suggested-users", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      const data = await response.json();
      const users = data.users || data.data?.users || [];
      suggestedUsers.value = users.slice(0, 5);
      if (suggestedUsers.value.length === 0) {
        const allUsersResponse = await fetch("https://localhost/api/v1/users/get-all-users", {
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        if (allUsersResponse.ok) {
          const allUsersData = await allUsersResponse.json();
          const allUsers = allUsersData.users || [];
          suggestedUsers.value = allUsers.filter((u: any) => (u.ID || u.id) !== currUserID.value).slice(0, 5);
        }
      }
    }
  } catch (error) {
    console.error("Failed to fetch suggested users:", error);
  }
};

const handleScroll = () => {
  if (!postsContainer.value) return;
  const { scrollTop, scrollHeight, clientHeight } = postsContainer.value;
  const threshold = 200;
  if (scrollTop + clientHeight >= scrollHeight - threshold && !loading.value && hasMore.value) {
    fetchPost(currentPage.value + 1, true);
  }
};

const fetchUserInteractions = async () => {
  const token = localStorage.getItem("token");
  if (!token) return;

  for (const p of postsWithUsers.value) {
    try {
      const [likesRes, commentsRes, savesRes] = await Promise.allSettled([
        fetch(`https://localhost/api/v1/posts/check-like/${p.post.ID}`, {
          headers: { Authorization: `Bearer ${token}` }
        }),
        fetch(`https://localhost/api/v1/social/check-user-commented/${p.post.ID}`, {
          headers: { Authorization: `Bearer ${token}` }
        }),
        fetch(`https://localhost/api/v1/posts/check-save/${p.post.ID}`, {
          headers: { Authorization: `Bearer ${token}` }
        })
      ]);

      if (likesRes.status === "fulfilled" && likesRes.value.ok) {
        try {
          const data = await likesRes.value.json();
          userLikes.value[p.post.ID] = data.liked || false;
        } catch (e) {
          userLikes.value[p.post.ID] = false;
        }
      }

      if (commentsRes.status === "fulfilled" && commentsRes.value.ok) {
        try {
          const data = await commentsRes.value.json();
          userComments.value[p.post.ID] = data.commented || false;
        } catch (e) {
          userComments.value[p.post.ID] = false;
        }
      }

      if (savesRes.status === "fulfilled" && savesRes.value.ok) {
        try {
          const data = await savesRes.value.json();
          userSaves.value[p.post.ID] = data.saved || false;
        } catch (e) {
          userSaves.value[p.post.ID] = false;
        }
      }
    } catch (error) {
      console.error("Error fetching interactions:", error);
    }
  }
};

const likeClicked = async (postID: string) => {
  try {
    const wasLiked = userLikes.value[postID];
    const response = await fetch(`https://localhost/api/v1/posts/like-post/${postID}`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${localStorage.getItem("token")}`,
      },
    });
    if(!response.ok){
      throw new Error("Failed to like post");
    } else{
      const data = await response.json();
      userLikes.value[postID] = !wasLiked;
      
      const post = postsWithUsers.value.find(p => p.post.ID === postID);
      if (!wasLiked && data.liked && post && post.post.UserID && currUserID.value) {
        await createNotification(
          post.post.UserID,
          currUserID.value,
          "like",
          postID
        );
      }
      
      fetchPost();
    }
  } catch{
    console.error("Failed to like post");
  }
};

const saveClicked = async (postID: string) => {
  const token = localStorage.getItem("token");
  if (!token || !currUserID.value) return;

  try {
    
    if (userSaves.value[postID]) {
      
      const data = await getCollectionsByUserID(currUserID.value);
      for (const collection of data) {
        try {
          await removePostFromCollection(collection.collection.id, postID);
        } catch (error) {
          console.error("Failed to remove post from collection:", error);
        }
      }
      userSaves.value[postID] = false;
      return;
    }
    
    const data = await getCollectionsByUserID(currUserID.value);
    
    if (data.length === 0) {
    
      userSaves.value[postID] = true;
    } else {
      savingPostId.value = postID;
      selectedCollectionForSave.value = "";
      userCollections.value = data;
      showSaveCollectionModal.value = true;
    }
  } catch (error) {
    console.error("Failed to process save:", error);
  }
};

const confirmSaveToCollection = async () => {
  if (!savingPostId.value) return;

  try {
    if (selectedCollectionForSave.value) {
      await addPostToCollection(selectedCollectionForSave.value, savingPostId.value);
    }
    userSaves.value[savingPostId.value] = true;
    showSaveCollectionModal.value = false;
    selectedCollectionForSave.value = "";
    savingPostId.value = "";
  } catch (error) {
    console.error("Failed to save post:", error);
  }
};

const openPostDetail = (postID: string) => {
  showPost(postID);
};

const sharePost = (postID: string) => {
  const post = postsWithUsers.value.find(p => p.post.ID === postID);
  if (post) {
    openShareMode(post.post);
  }
};

const handlePostCreated = async (data: { userId: string }) => {
  await fetchPost();
};

const handlePostLiked = async () => {
  await fetchPost();
};

onMounted(() => {
  fetchPost();
  fetchStories();
  fetchSuggestedUsers();
  on("post-created", handlePostCreated);
  on("post-liked", handlePostLiked);
  on("story-created", fetchStories);
  if (postsContainer.value) {
    postsContainer.value.addEventListener("scroll", handleScroll, { passive: true });
  }
});

onUnmounted(() => {
  off("post-created", handlePostCreated);
  off("post-liked", handlePostLiked);
  off("story-created", fetchStories);
  if (postsContainer.value) {
    postsContainer.value.removeEventListener("scroll", handleScroll);
  }
});

const pfpError = (e: Event) => {
  const target = e.target as HTMLImageElement;
  target.src = userLight;
};


const prevSlide = (postId: string) => {
  const current = currentSlides.value[postId] || 0;
  currentSlides.value[postId] = current === 0 ? 0 : current - 1;
};

const nextSlide = (postId: string, totalSlides: number) => {
  const current = currentSlides.value[postId] || 0;
  currentSlides.value[postId] = current >= totalSlides - 1 ? totalSlides - 1 : current + 1;
};

const goToSlide = (postId: string, slideIndex: number) => {
  currentSlides.value[postId] = slideIndex;
};

const goToProfile = (userId: string) => {
  router.push(`/profile/${userId}`);
};

const initializeSlides = () => {
  postsWithUsers.value.forEach(post => {
    if (post.postMedia && post.postMedia.length > 0) {
      currentSlides.value[post.post.ID] = 0;
    }
  });
};



const openMore = (postId: string) => {
  if (showMoreOverlay.value === postId) {
    showMoreOverlay.value = null;
  } else {
    showMoreOverlay.value = postId;
  }
};

const copyLink = async (postId: string) => {
  try {
    const link = `${window.location.origin}/post/${postId}`;
    await navigator.clipboard.writeText(link);
    copiedPostId.value = postId;
    setTimeout(() => {
      copiedPostId.value = null;
    }, 2000);
    showMoreOverlay.value = null;
  } catch (err) {
    console.error("Failed to copy link:", err);
  }
};

const dateToString = (dateStr: string) => {
  const date = new Date(dateStr);
  return date.toLocaleString();
};

import { watch } from "vue";
watch(postsWithUsers, initializeSlides, { immediate: true });

</script>

<template>
  <div class="container">
    <SideBar />
    <div 
      ref="postsContainer"
      class="home" 
    >
      <StoryCarousel
        :story-groups="storyGroups"
        @open-story="openStoryViewer"
      />
      
      <StoryViewer
        :story-groups="storyGroups"
        :initial-group-index="selectedStoryGroupIndex"
        :show="showStoryViewer"
        @close="closeStoryViewer"
        @story-viewed="fetchStories"
      />
      
      <div
        v-if="initialLoading"
        class="post-container"
      >
        <div
          v-for="i in 3"
          :key="i"
          class="skeleton-post"
        >
          <div class="skeleton-header">
            <div class="skeleton-avatar" />
            <div class="skeleton-text-line short" />
          </div>
          <div class="skeleton-media" />
          <div class="skeleton-actions">
            <div class="skeleton-icon" />
            <div class="skeleton-icon" />
            <div class="skeleton-icon" />
          </div>
          <div class="skeleton-text-line" />
          <div class="skeleton-text-line short" />
        </div>
      </div>
      
      <div
        v-else
        class="post-container"
      >
        <div
          v-for="p in postsWithUsers"
          :key="p.post.ID"
          class="post-item"
        >
          <div class="upper-content">
            <div
              class="user-profile"
              @click="goToProfile(p.user.ID || p.user.id)"
            >
              <img
                v-if="p.user && p.user.ProfilePicture"
                :src="loadPhotoUrl(p.user.ProfilePicture)"
                width="50"
                height="50"
                style="border-radius: 50%; object-fit: cover; border: 2px solid white"
                @error="pfpError"
              />
              <img
                v-else
                :src="userLight"
                width="50"
                height="50"
                style="border-radius: 50%; object-fit: cover; border: 2px solid white"
              />

              <span
                v-if="p.user && p.user.Username"
                style="font-weight: 600; font-size: 20px;"
              >{{ p.user.Username }}</span>
              <span v-else>User not found</span>
              
              <button v-if="p.collaborators && p.collaborators.length > 0">
                Collaborators ({{ p.collaborators.length }})
              </button>
            </div>
            <button 
              style="display: flex; background-color: transparent; font-size: 36px; font-weight: bold; color: white; border: none; cursor: pointer; justify-content: center; text-align: center;"
              @click="openMore(p.post.ID)"
            >
              ...
            </button>
          </div>
          <div 
            v-if="showMoreOverlay === p.post.ID" 
            class="more-overlay"
          >
            <button class="more-option">
              Report
            </button>
            <button
              class="more-option"
              @click="openPostDetail(p.post.ID)"
            >
              Go to Post Detail
            </button>
            <button 
              class="more-option" 
              @click="copyLink(p.post.ID)"
            >
              {{ 'Copy Link' }}
            </button>
            <button 
              class="more-option" 
              @click="openMore(p.post.ID)"
            >
              Cancel
            </button>
          </div>
          <div class="post-content">
            <div
              v-if="p.postMedia && p.postMedia.length > 0"
              class="media-slider"
            >
              <div class="slider-container">
                <div
                  v-for="(media, index) in p.postMedia"
                  :key="index"
                  class="slide"
                  :class="{ active: (currentSlides[p.post.ID] || 0) === index }"
                >
                  <img
                    v-if="media.MediaType === 'image'"
                    :src="loadPhotoUrl(media.FilePath)"
                    width="400"
                    height="400"
                  />
                  <video
                    v-else-if="media.MediaType === 'video'"
                    :src="loadPhotoUrl(media.FilePath)"
                    width="400"
                    height="400"
                    controls
                  />
                </div>
              </div>            
              <button
                v-if="p.postMedia.length > 1"
                class="slider-btn prev-btn"
                @click="prevSlide(p.post.ID)"
              >
                ‹
              </button>
              <button
                v-if="p.postMedia.length > 1"
                class="slider-btn next-btn"
                @click="nextSlide(p.post.ID, p.postMedia.length)"
              >
                ›
              </button>

              <div
                v-if="p.postMedia.length > 1"
                class="slide-indicators"
              >
                <span
                  v-for="(media, index) in p.postMedia"
                  :key="index"
                  class="indicator"
                  :class="{ active: currentSlides[p.post.ID] === index }"
                  @click="goToSlide(p.post.ID, index)"
                />
              </div>
            </div>
            <div class="post-interaction">
              <div class="interaction-left">
                <img
                  :src="likeLight"
                  alt="like"
                  width="30"
                  height="30"
                  :style="{ cursor: 'pointer', filter: userLikes[p.post.ID] ? 'invert(27%) sepia(98%) saturate(7471%) hue-rotate(3deg) brightness(102%) contrast(115%)' : 'none' }"
                  @click="likeClicked(p.post.ID)"
                />
                <img
                  :src="commentLight"
                  alt="comment"
                  width="30"
                  height="30"
                  :style="{ cursor: 'pointer'}"
                  @click="openPostDetail(p.post.ID)"
                />
                <img
                  :src="sendLight"
                  alt="share"
                  width="30"
                  height="30"
                  style="cursor: pointer;"
                  @click="sharePost(p.post.ID)"
                />
              </div>
              <div class="interaction-right">
                <img
                  :src="userSaves[p.post.ID] ? savedDark : savedLight"
                  alt="save"
                  width="30"
                  height="30"
                  :style="{ cursor: 'pointer', filter: userSaves[p.post.ID] ? 'invert(85%) sepia(89%) saturate(1991%) hue-rotate(358deg) brightness(101%) contrast(103%)' : 'none' }"
                  @click="saveClicked(p.post.ID)"
                />
              </div>
            </div>
            <span
              v-if="!p.post.IsHideLikes"
              style="font-weight: 500; cursor: pointer;"
            >{{ p.post.LikesCount }} Likes</span>
            <span
              v-if="!p.post.IsCommentOff"
              style="font-weight: 400; color: #a0a0a0; cursor: pointer; font-size: 14px;"
              @click="openPostDetail(p.post.ID)"
            >View all {{ p.post.CommentsCount }} comments</span>
            <template
                v-for="part in formatTextWithHashtags(p.post.Caption)"
                :key="part.key"
              >
                <span
                  v-if="part.isHashtag"
                  class="hashtag"
                  @click.stop="navigateToHashtag(part.text)"
                >{{ part.text }}</span>
                <span v-else>{{ part.text }}</span>
              </template>
              {{ p.post.Caption }}
            
            <p style="font-size: 12px;">
              {{ dateToString(p.post.CreatedAt) }}
            </p>
          </div>
        </div>
        <div
          v-if="loading && !initialLoading"
          class="loading-container"
        >
          <div class="loading-spinner" />
        </div>
      </div>
    </div>

    <div class="right-sidebar">
      <div class="suggested-section">
        <h3 class="suggested-title">
          Suggested for you
        </h3>
        <div
          v-for="user in suggestedUsers"
          :key="user.ID"
          class="suggested-user"
        >
          <img
            v-if="user.ProfilePicture"
            :src="loadPhotoUrl(user.ProfilePicture)"
            class="suggested-avatar"
            @error="pfpError"
          />
          <img
            v-else
            :src="userLight"
            class="suggested-avatar"
          />
          <div class="suggested-info">
            <span class="suggested-username">{{ user.Username }}</span>
            <span class="suggested-followers">{{ user.FollowerCount || 0 }} followers</span>
          </div>
          <button
            class="follow-btn"
            @click="goToProfile(user.ID)"
          >
            View
          </button>
        </div>
      </div>
    </div>

    <PostDetail />
    <div
      v-if="showSaveCollectionModal"
      class="modal-overlay"
      @click.self="showSaveCollectionModal = false"
    >
      <div class="save-modal-content">
        <h3>Save to Collection</h3>
        <div
          v-if="loadingCollections"
          class="loading"
        >
          Loading collections...
        </div>
        <div
          v-else-if="userCollections.length === 0"
          class="no-collections"
        >
          No collections yet
        </div>
        <div
          v-else
          class="collections-list"
        >
          <div
            v-for="collection in userCollections"
            :key="collection.collection.id"
            class="collection-option"
            :class="{ selected: selectedCollectionForSave === collection.collection.id }"
            @click="selectedCollectionForSave = collection.collection.id"
          >
            {{ collection.collection.name }}
          </div>
        </div>
        <div class="modal-actions">
          <button
            class="btn-cancel"
            @click="showSaveCollectionModal = false"
          >
            Cancel
          </button>
          <button
            class="btn-save"
            :disabled="loadingCollections"
            @click="confirmSaveToCollection"
          >
            {{ selectedCollectionForSave ? 'Save to Collection' : 'Save to All Posts' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.container {
  display: flex;
  background-color: black;
  position: relative;
  min-height: 100vh;
}

.home {
  flex: 1;
  height: 100vh;
  display: flex;
  align-items: center;
  flex-direction: column;
  scroll-behavior: smooth;
  overflow-y: auto;
  overflow-x: hidden;
  margin: 0 auto;
  max-width: 630px;
  padding: 0;
  margin-left: 250px;
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.home::-webkit-scrollbar {
  display: none;
}

.right-sidebar {
  width: 320px;
  height: 100vh;
  padding: 30px 20px;
  background-color: black;
  border-left: 1px solid #333;
  overflow-y: auto;
  position: fixed;
  right: 0;
  top: 0;
}

@media (max-width: 1400px) {
  .right-sidebar {
    display: none;
  }
}

@media (max-width: 900px) {
  .home {
    margin-left: 80px;
    max-width: calc(100vw - 100px);
  }
}

@media (max-width: 640px) {
  .home {
    margin-left: 0;
    max-width: 100vw;
    padding: 0 10px;
  }
}

.suggested-section {
  color: white;
}

.suggested-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 20px;
  color: #888;
}

.suggested-user {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 15px;
  padding: 8px;
  border-radius: 8px;
  transition: background 0.2s;
}

.suggested-user:hover {
  background: #1a1a1a;
}

.suggested-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid #333;
}

.suggested-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.suggested-username {
  font-size: 14px;
  font-weight: 600;
  color: white;
}

.suggested-followers {
  font-size: 12px;
  color: #888;
}

.follow-btn {
  background: transparent;
  border: 1px solid #333;
  color: white;
  padding: 6px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.follow-btn:hover {
  background: #1a1a1a;
  border-color: white;
}

.skeleton-post {
  width: 100%;
  max-width: 600px;
  background: #1a1a1a;
  border: 1px solid #333;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}

.skeleton-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 15px;
}

.skeleton-avatar {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: linear-gradient(90deg, #2a2a2a 25%, #333 50%, #2a2a2a 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

.skeleton-media {
  width: 100%;
  height: 400px;
  background: linear-gradient(90deg, #2a2a2a 25%, #333 50%, #2a2a2a 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 8px;
  margin-bottom: 15px;
}

.skeleton-actions {
  display: flex;
  gap: 15px;
  margin-bottom: 12px;
}

.skeleton-icon {
  width: 30px;
  height: 30px;
  background: linear-gradient(90deg, #2a2a2a 25%, #333 50%, #2a2a2a 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 4px;
}

.skeleton-text-line {
  height: 16px;
  background: linear-gradient(90deg, #2a2a2a 25%, #333 50%, #2a2a2a 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 4px;
  margin-bottom: 8px;
}

.skeleton-text-line.short {
  width: 60%;
}

@keyframes shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}

.story-container {
  display: flex;
  gap: 15px;
  padding: 20px;
  overflow-x: auto;
  overflow-y: hidden;
  border-bottom: 1px solid #333;
  margin-bottom: 20px;
  width: 100%;
}

.story-container::-webkit-scrollbar {
  height: 6px;
}

.story-container::-webkit-scrollbar-track {
  background: #1a1a1a;
}

.story-container::-webkit-scrollbar-thumb {
  background: #555;
  border-radius: 3px;
}

.story-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  min-width: 70px;
}

.story-avatar {
  width: 70px;
  height: 70px;
  border-radius: 50%;
  padding: 3px;
  background: linear-gradient(45deg, #f09433 0%, #e6683c 25%, #dc2743 50%, #cc2366 75%, #bc1888 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.story-avatar.viewed {
  background: #555;
}

.story-avatar img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid black;
}

.story-username {
  font-size: 12px;
  color: white;
  max-width: 70px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.story-viewer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.95);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2000;
}

.story-viewer-container {
  position: relative;
  width: 420px;
  height: 100vh;
  max-height: 90vh;
  background: black;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.story-close-btn {
  position: absolute;
  top: 20px;
  right: 20px;
  background: rgba(0, 0, 0, 0.5);
  border: none;
  color: white;
  font-size: 32px;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  cursor: pointer;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.story-nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(0, 0, 0, 0.5);
  border: none;
  color: white;
  font-size: 32px;
  width: 50px;
  height: 50px;
  border-radius: 50%;
  cursor: pointer;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
}

.story-nav-btn.prev {
  left: 20px;
}

.story-nav-btn.next {
  right: 20px;
}

.story-progress-bars {
  position: absolute;
  top: 10px;
  left: 10px;
  right: 10px;
  display: flex;
  gap: 4px;
  z-index: 10;
}

.progress-bar-container {
  flex: 1;
  height: 3px;
  background: rgba(255, 255, 255, 0.3);
  border-radius: 2px;
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  background: white;
  transition: width 0.1s linear;
}

.story-header {
  position: absolute;
  top: 30px;
  left: 10px;
  right: 10px;
  display: flex;
  align-items: center;
  gap: 12px;
  z-index: 10;
  padding: 10px;
}

.story-user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid white;
}

.story-user-name {
  color: white;
  font-weight: 600;
  font-size: 16px;
  flex: 1;
}

.story-pause-btn {
  background: rgba(0, 0, 0, 0.5);
  border: none;
  color: white;
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 16px;
}

.story-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: black;
  cursor: pointer;
}

.story-content img,
.story-content video {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.story-footer {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  background: linear-gradient(to top, rgba(0, 0, 0, 0.7), transparent);
  z-index: 10;
}

.story-action-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0;
}

.story-action-btn img {
  width: 30px;
  height: 30px;
}

.story-reply-input {
  flex: 1;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 20px;
  padding: 10px 16px;
  color: white;
  font-size: 14px;
}

.story-reply-input::placeholder {
  color: rgba(255, 255, 255, 0.6);
}

.story-reply-input:focus {
  outline: none;
  border-color: white;
}

.recipient-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  margin-right: 10px;
}

.collection-option {
  display: flex;
  align-items: center;
}

.post-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  padding: 50px;
}

.story-container {
  width: 100%;
  margin-bottom: 20px;
}

.post-content {
  display: flex;
  flex-direction: column;
  width: 100%;
  gap: 5px;
}
  
.post-item {
  color: white;
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 600px;
  background-color: black;
  border: 1px solid #262626;
  border-radius: 8px;
  padding: 30px;
  padding-bottom: 40px;
  margin: 0 auto;
  overflow: hidden;
}

.user-profile {
  align-items: center;
  justify-content: space-between;
  display: flex;
  padding: 12px 16px;
  gap: 12px;
}

@media (max-width: 640px) {
  .post-item {
    border-radius: 0;
    border-left: none;
    border-right: none;
  }
}


.media-slider {
  position: relative;
  width: 100%;
  display: flex;
  justify-content: center;
  flex-direction: column;
  aspect-ratio: 1 / 1;
  background: black;
}

.slider-container {
  display: flex;
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.slide {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  visibility: hidden;
  transition: opacity 0.3s ease-in-out;
  z-index: 1;
}

.slide.active {
  opacity: 1;
  visibility: visible;
  z-index: 2;
}

.slide img,
.slide video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

@media (max-width: 640px) {
  .media-slider {
    aspect-ratio: 4 / 5;
  }
}


.slider-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(0, 0, 0, 0.5);
  color: white;
  border: none;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  font-size: 18px;
  cursor: pointer;
  z-index: 10;
  transition: background 0.3s ease;
}

.slider-btn:hover {
  background: rgba(0, 0, 0, 0.8);
}

.upper-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.prev-btn {
  left: 10px;
}

.next-btn {
  right: 10px;
}

.slide-indicators {
  position: absolute;
  bottom: 10px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 8px;
  z-index: 10;
}

.indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  transition: background 0.3s ease;
}

.indicator.active {
  background: white;
}

.indicator:hover {
  background: rgba(255, 255, 255, 0.8);
}

.post-interaction {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  width: 100%;
}

.post-item > span,
.post-item > h3,
.post-item > p {
  padding: 0 16px;
}

.interaction-left {
  display: flex;
  gap: 15px;
  align-items: center;
}

.interaction-right {
  display: flex;
  align-items: center;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 40px;
  width: 100%;
}

.loading-spinner {
  border: 4px solid rgba(255, 255, 255, 0.1);
  border-top: 4px solid white;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  animation: spin 1s linear infinite;
}

.more-overlay {
  position: absolute;
  top: 60px;
  right: 20px;
  background-color: rgb(38, 38, 38);
  border-radius: 12px;
  padding: 8px;
  min-width: 200px;
  z-index: 100;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
}

.more-option {
  background: none;
  border: none;
  color: white;
  padding: 12px 16px;
  text-align: left;
  cursor: pointer;
  font-size: 16px;
  font-weight: 400;
  transition: background-color 0.2s;
}

.more-option:hover {
  background-color: rgba(255, 255, 255, 0.1);
  border-radius: 8px;
}


@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1500;
}

.save-modal-content {
  background: #262626;
  border-radius: 12px;
  padding: 24px;
  width: 400px;
  max-width: 90%;
  color: white;
}

.save-modal-content h3 {
  margin: 0 0 20px 0;
  font-size: 16px;
}

.loading,
.no-collections {
  text-align: center;
  padding: 20px;
  color: #999;
}

.collections-list {
  max-height: 300px;
  overflow-y: auto;
  margin-bottom: 20px;
  border: 1px solid #333;
  border-radius: 8px;
}

.collection-option {
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid #333;
  transition: background 0.2s;
}

.collection-option:last-child {
  border-bottom: none;
}

.collection-option:hover {
  background: #333;
}

.collection-option.selected {
  background: #0095f6;
  font-weight: 600;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.btn-cancel {
  background: transparent;
  color: white;
  border: 1px solid #333;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: background 0.2s;
}

.btn-cancel:hover {
  background: #333;
}

.btn-save {
  background: #0095f6;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: background 0.2s;
}

.btn-save:hover:not(:disabled) {
  background: #1877f2;
}

.btn-save:disabled {
  background: #555;
  cursor: not-allowed;
  opacity: 0.6;
}

.hashtag {
  color: #0095f6;
  cursor: pointer;
  transition: color 0.2s;
}

.hashtag:hover {
  color: #0081d5;
  text-decoration: underline;
}

</style>
