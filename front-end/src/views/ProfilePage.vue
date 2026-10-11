<script setup lang="ts">
import SideBar from "@/components/SideBar.vue";
import { onMounted, onUnmounted, ref, watch, computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import { usePhoto } from "@/composables/usePhoto";
import defaultProfileIcon from "@/components/icons/user-light.png";
import { isTokenExpired } from "@/router";
import { usePostFetcher } from "@/composables/usePostFetcher";
import likeIcon from "@/components/icons/notif-light.png";
import commentIcon from "@/components/icons/comment-light.png";
import { useEventBus } from "@/composables/useEventBus";
import { useShowPost } from "@/composables/useOverlay";
import PostDetail from "@/components/PostDetail.vue";
import { getCollectionsByUserID, createCollection, type CollectionWithPosts } from "@/services/collectionService";
import { useNotification } from "@/composables/useNotification";

const { showPost } = useShowPost();
const { on, off } = useEventBus();
const { createNotification } = useNotification();

const showPostDetail = (postID: string) => {
  showPost(postID);
};

interface PostWithUser {
  post: any
  user: any
  postMedia: any[]
  collaborators?: any[]
}

const posts = ref<PostWithUser[]>([]);
const reels = ref<PostWithUser[]>([]);
const saved = ref<PostWithUser[]>([]);
const tag = ref<PostWithUser[]>([]);
const collections = ref<CollectionWithPosts[]>([]);
const showCreateCollectionModal = ref(false);
const newCollectionName = ref("");

const { fetchPostsWithDetails, fetchPost } = usePostFetcher();
const router = useRouter();
const route = useRoute();
const { loadPhotoUrl } = usePhoto();

const profilePhotoUrl = ref<string>("");
const name = ref<string>("");
const username = ref<string>("");
const postCount = ref<number>(0);
const followingCount = ref<number>(0);
const followersCount = ref<number>(0);
const bio = ref<string>("");
const currentTab = ref<string>("post");
const userID = ref<string>("");
const currUserID = ref<string>("");
const isFollowing = ref<boolean>(false);
const isLoadingFollow = ref<boolean>(false);
const isBlocked = ref<boolean>(false);
const isLoadingBlock = ref<boolean>(false);
const isPrivate = ref<boolean>(false);

const canViewContent = computed(() => {
  return currUserID.value === userID.value || !isPrivate.value || isFollowing.value;
});

const fetchCurrentUser = async () => {
  const token = localStorage.getItem("token");
  if (!token || isTokenExpired(token)) {
    localStorage.removeItem("token");
    router.push("/login");
    return "";
  }

  try {
    const response = await fetch("https://localhost/api/v1/users/get-current-user", {
      method: "GET",
      headers: { Authorization: `Bearer ${token}` },
    });

    if (response.ok) {
      const data = await response.json();
      const user = data.user;
      currUserID.value = user.ID || user.id;
      return user.ID || user.id;
    }
  } catch (error) {
    console.error("Failed to fetch current user:", error);
  }
  return "";
};

const fetchUser = async (passedUserID: string) => {
  const token = localStorage.getItem("token");
  if (!token || isTokenExpired(token)) {
    localStorage.removeItem("token");
    router.push("/login");
    return false;
  }

  userID.value = passedUserID;

  try {
    const response = await fetch(`https://localhost/api/v1/users/get-user-by-id/${passedUserID}`, {
      method: "GET",
      headers: { Authorization: `Bearer ${token}` },
    });

    if (response.ok) {
      const data = await response.json();
      const user = data.user;

      profilePhotoUrl.value = user.ProfilePicture ? loadPhotoUrl(user.ProfilePicture) : defaultProfileIcon;
      username.value = user.Username || "Unknown User";
      name.value = user.Name || "";
      postCount.value = user.PostsCount || 0;
      followingCount.value = user.FollowingCount || 0;
      followersCount.value = user.FollowersCount || 0;
      userID.value = user.ID || user.id || "";
      bio.value = user.Bio || "";
      isPrivate.value = user.is_private || user.IsPrivate || false;

      return true;
    }
  } catch (error) {
    console.error("Failed to fetch user:", error);
  }
  return false;
};

const initProfile = async () => {
  const currentUserId = await fetchCurrentUser();
  const passedUserId = (route.params.userId as string) || currentUserId;

  if (currentUserId && passedUserId) {
    await fetchUser(passedUserId);
    
    if (currentUserId !== passedUserId) {
      await fetchFollowStatus();
      await fetchBlockStatus();
    }

    if (canViewContent.value) {
        await fetchUserPosts();
    } else {
        posts.value = [];
        reels.value = [];
        tag.value = [];
    }
  }
};

watch(() => route.params.userId, () => {
  initProfile();
});

watch(currentTab, (newTab) => {
  if (!canViewContent.value) return;

  if (newTab === "reels" && userID.value) {
    fetchReels();
  } else if (newTab === "saved" && userID.value) {
    fetchCollections();
  }
});

const handlePostLiked = async (data: { postID: string }) => {
  const updatePostInArray = (array: PostWithUser[]) => {
    const postIndex = array.findIndex(p => p.post?.ID === data.postID);
    if (postIndex !== -1) {
      fetchPost(data.postID).then(updatedPost => {
        if (updatedPost) {
          array[postIndex] = updatedPost;
        }
      });
    }
  };

  updatePostInArray(posts.value);
  updatePostInArray(reels.value);
  updatePostInArray(saved.value);
  updatePostInArray(tag.value);
};

const fetchBlockStatus = async () => {
  const token = localStorage.getItem("token");
  if (!token || isTokenExpired(token) || !userID.value) return;

  try {
    const response = await fetch(`https://localhost/api/v1/social/check-block-status/${userID.value}`, {
      method: "GET",
      headers: { Authorization: `Bearer ${token}` },
    });

    if (response.ok) {
      const data = await response.json();
      isBlocked.value = data.is_blocked || false;
    }
  } catch (error) {
    console.error("Failed to fetch block status:", error);
  }
};

const handleBlock = async () => {
  if (isLoadingBlock.value || !userID.value) return;
  const token = localStorage.getItem("token");
  if (!token || isTokenExpired(token)) {
    router.push("/login");
    return;
  }
  isLoadingBlock.value = true;
  try {
    const endpoint = isBlocked.value ? "unblock-user" : "block-user";
    const method = isBlocked.value ? "DELETE" : "POST";
    const response = await fetch(`https://localhost/api/v1/social/${endpoint}/${userID.value}`, {
      method: method,
      headers: { Authorization: `Bearer ${token}` },
    });
    if (response.ok) {
      isBlocked.value = !isBlocked.value;
      if (isBlocked.value) {
          posts.value = [];
      }
    } 
  } catch (error) {
    console.error("Failed to block/unblock user:", error);
  } finally {
    isLoadingBlock.value = false;
  }
};

const handlePostCreated = async (data: { userId: string }) => {
  if (data.userId === userID.value) {
    await fetchUserPosts();
    await fetchReels();
  }
  await fetchUser(userID.value);
};

onMounted(() => {
  setTimeout(() => {
    initProfile();
  }, 100);
  on("post-liked", handlePostLiked);
  on("post-created", handlePostCreated);
});

onUnmounted(() => {
  off("post-liked", handlePostLiked);
  off("post-created", handlePostCreated);
});

const fetchUserPosts = async () => {
  if (!userID.value) return;
  const data = await fetchPostsWithDetails(`https://localhost/api/v1/posts/get-posts-by-user/${userID.value}`);
  posts.value = data;
};

const fetchReels = async () => {
  if (!userID.value) return;
  const data = await fetchPostsWithDetails(`https://localhost/api/v1/posts/get-reels-by-user/${userID.value}`);
  reels.value = data;
};

const pfpError = (e: Event) => {
  const target = e.target as HTMLImageElement;
  target.src = defaultProfileIcon;
};

const fetchFollowStatus = async () => {
  const token = localStorage.getItem("token");
  if (!token || isTokenExpired(token) || !userID.value) return;

  try {
    const response = await fetch(`https://localhost/api/v1/users/follow-status/${userID.value}`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json"
      },
    });

    if (response.ok) {
      const data = await response.json();
      isFollowing.value = data.is_following || false;
      if (data.followers_count !== undefined) {
        followersCount.value = data.followers_count;
      }
    }
  } catch (error) {
    console.error("Failed to fetch follow status:", error);
  }
};

const handleFollow = async () => {
  if (isLoadingFollow.value || !userID.value) return;
  const token = localStorage.getItem("token");
  if (!token || isTokenExpired(token)) {
    router.push("/login");
    return;
  }
  isLoadingFollow.value = true;
  const wasFollowing = isFollowing.value;
  try {
    const endpoint = wasFollowing ? "unfollow" : "follow";
    const response = await fetch(`https://localhost/api/v1/users/${endpoint}`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ user_id: userID.value })
    });

    if (response.ok) {
      const data = await response.json();
      isFollowing.value = data.is_following;
      if (data.followers_count !== undefined) {
        followersCount.value = data.followers_count;
      }
      if (canViewContent.value) {
          fetchUserPosts();
      } else {
          posts.value = [];
          reels.value = [];
      }
      
      
      if (!wasFollowing && data.is_following && currUserID.value && userID.value) {
        await createNotification(
          userID.value,
          currUserID.value,
          "follow"
        );
      }
    }
  } catch (error) {
    console.error("Failed to follow/unfollow user:", error);
  } finally {
    isLoadingFollow.value = false;
  }
};

const handleMessage = () => {
  router.push("/messages");
};

const fetchCollections = async () => {
  if (!userID.value) return;
  try {
    const data = await getCollectionsByUserID(userID.value);
    const token = localStorage.getItem("token");
    
    const enrichedCollections = await Promise.all(
      data.map(async (collection) => {
        const enrichedPosts = await Promise.all(
          collection.recent_posts.map(async (post) => {
            try {
              const mediaRes = await fetch(
                `https://localhost/api/v1/posts/get-post-media-by-id/${post.ID}`,
                { headers: { Authorization: `Bearer ${token}` } }
              );
              const mediaData = mediaRes.ok ? await mediaRes.json() : {};
              return {
                ...post,
                Media: mediaData.media || []
              };
            } catch {
              return post;
            }
          })
        );
        return {
          ...collection,
          recent_posts: enrichedPosts
        };
      })
    );
    collections.value = enrichedCollections;
  } catch (error) {
    console.error("Failed to fetch collections:", error);
    collections.value = [];
  }
};

const handleCreateCollection = async () => {
  if (!newCollectionName.value.trim()) return;
  try {
    await createCollection(newCollectionName.value);
    newCollectionName.value = "";
    showCreateCollectionModal.value = false;
    await fetchCollections();
  } catch (error) {
    console.error("Failed to create collection:", error);
  }
};

const viewCollection = (collectionID: string) => {
  router.push(`/collection/${collectionID}`);
};
</script>

<template>
  <PostDetail />
  <div class="container">
    <SideBar />
    <div class="profile">
      <div class="profile-up">
        <img
          :src="profilePhotoUrl"
          class="profile-pic"
          @error="pfpError"
        />
        <div class="profile-info">
          <div class="upper-profile">
            <h2>{{ username }}</h2>
            <div class="profile-actions">
              <template v-if="currUserID === userID">
                <button
                  class="btn"
                  style="cursor: pointer;"
                  @click="router.push('/settings')"
                >
                  Settings
                </button>
                <button
                  class="btn"
                  style="cursor: pointer;"
                  @click="router.push('/archive')"
                >
                  View Archive
                </button>
              </template>
              <template v-else>
                <button
                  class="btn follow-btn"
                  :class="{ 'following': isFollowing, 'loading': isLoadingFollow }"
                  :disabled="isLoadingFollow"
                  @click="handleFollow"
                >
                  {{ isLoadingFollow ? 'Loading...' : (isFollowing ? 'Following' : 'Follow') }}
                </button>
                <button
                  class="btn message-btn"
                  @click="handleMessage"
                >
                  Message
                </button>
                <button
                  class="btn block-btn"
                  :class="{ 'blocked': isBlocked, 'loading': isLoadingBlock }"
                  :disabled="isLoadingBlock"
                  @click="handleBlock"
                >
                  {{ isLoadingBlock ? 'Loading...' : (isBlocked ? 'Unblock' : 'Block') }}
                </button>
              </template>
            </div>
          </div>
          <h3>{{ name }}</h3>
          <div class="middle-profile">
            <h3>{{ postCount }} posts</h3>
            <h3>{{ followingCount }} following</h3>
            <h3>{{ followersCount }} followers</h3>
          </div>
          <div class="lower-profile">
            <h3>{{ bio }}</h3>
          </div>
        </div>
      </div>

      <div
        v-if="isBlocked"
        class="blocked-message"
      >
        <h2>This user is blocked</h2>
        <p>You have blocked this user. Unblock to view their content.</p>
      </div>

      <div
        v-else-if="!canViewContent" 
        class="private-message"
      >
        <div class="private-icon-container">
            <svg aria-label="Private" class="x1lliihq x1n2onr6" color="rgb(245, 245, 245)" fill="rgb(245, 245, 245)" height="62" role="img" viewBox="0 0 96 96" width="62"><title>Private</title><circle cx="48" cy="48" fill="none" r="47" stroke="currentColor" stroke-miterlimit="10" stroke-width="2"></circle><path d="M60 41.5V36a12 12 0 1 0-24 0v5.5a4.5 4.5 0 0 0-4.5 4.5v19a4.5 4.5 0 0 0 4.5 4.5h24a4.5 4.5 0 0 0 4.5-4.5v-19a4.5 4.5 0 0 0-4.5-4.5Zm-19.5-5.5a7.5 7.5 0 1 1 15 0v5.5h-15Z"></path></svg>
        </div>
        <h2>This Account is Private</h2>
        <p>Follow to see their photos and videos.</p>
      </div>

      <div
        v-else
        class="profile-content"
      >
        <div class="tab-button">
          <img
            src="../components/icons/post-light.png"
            width="30"
            height="30"
            :class="{ active: currentTab === 'post' }"
            @click="currentTab = 'post'"
          />
          <img
            src="../components/icons/reels-light.png"
            width="30"
            height="30"
            :class="{ active: currentTab === 'reels' }"
            @click="currentTab = 'reels'"
          />
          <img 
            v-if="currUserID === userID"
            src="../components/icons/saved-light.png"
            width="30"
            height="30"
            :class="{ active: currentTab === 'saved' }"
            @click="currentTab = 'saved'"
          />
          <img
            src="../components/icons/tag-light.png"
            width="30"
            height="30"
            :class="{ active: currentTab === 'tag' }"
            @click="currentTab = 'tag'"
          />
        </div>

      <div
        v-if="currentTab === 'post'"
        class="tab-container"
      >
        <div class="grid-container">
          <div
            v-for="postWithUser in posts"
            :key="postWithUser.post?.id || postWithUser.post?.ID"
            class="grid-item"
            @click="showPostDetail(postWithUser.post?.ID)"
          >
            <div
              v-if="postWithUser.postMedia.length > 0"
              class="media-container"
            >
              <img
                v-if="postWithUser.postMedia[0].MediaType === 'image'"
                :src="loadPhotoUrl(postWithUser.postMedia[0].FilePath)"
                class="post-image"
              />
              <div
                v-else-if="postWithUser.postMedia[0].MediaType === 'video'"
                class="video-thumbnail"
              >
                <video
                  :src="loadPhotoUrl(postWithUser.postMedia[0].FilePath)"
                  class="post-video"
                  muted
                  preload="metadata"
                />
                <div class="video-overlay">
                  <img
                    src="../components/icons/reels-light.png"
                    width="30"
                    height="30"
                    class="video-icon"
                  />
                </div>
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
                    width="20"
                    height="20"
                    class="stat-icon"
                  />
                  <span class="stat-count">{{ postWithUser.post?.LikesCount || 0 }}</span>
                </div>
                <div class="stat-item">
                  <img
                    :src="commentIcon"
                    width="20"
                    height="20"
                    class="stat-icon"
                  />
                  <span class="stat-count">{{ postWithUser.post?.CommentsCount || 0 }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="currentTab === 'reels'"
        class="tab-container"
      >
        <div class="grid-container">
          <div
            v-for="postWithUser in reels"
            :key="postWithUser.post.ID"
            class="grid-item"
            @click="showPostDetail(postWithUser.post?.ID)"
          >
            <div
              v-if="postWithUser.postMedia.length > 0"
              class="media-container"
            >
              <img
                v-if="postWithUser.postMedia[0].MediaType === 'image'"
                :src="loadPhotoUrl(postWithUser.postMedia[0].FilePath)"
                class="post-image"
              />
              <div
                v-else-if="postWithUser.postMedia[0].MediaType === 'video'"
                class="video-thumbnail"
              >
                <video
                  :src="loadPhotoUrl(postWithUser.postMedia[0].FilePath)"
                  class="post-video"
                  muted
                  preload="metadata"
                />
                <div class="video-overlay">
                  <img
                    src="../components/icons/reels-light.png"
                    width="30"
                    height="30"
                    class="video-icon"
                  />
                </div>
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
                    width="20"
                    height="20"
                    class="stat-icon"
                  />
                  <span class="stat-count">{{ postWithUser.post?.LikesCount || 0 }}</span>
                </div>
                <div class="stat-item">
                  <img
                    :src="commentIcon"
                    width="20"
                    height="20"
                    class="stat-icon"
                  />
                  <span class="stat-count">{{ postWithUser.post?.CommentsCount || 0 }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="currentTab === 'saved' && currUserID === userID"
        class="tab-container"
      >
        <div class="collections-header">
          <button
            v-if="currUserID === userID"
            class="create-collection-btn"
            @click="showCreateCollectionModal = true"
          >
            + New Collection
          </button>
        </div>
        
        <div class="grid-container">
          <div
            v-for="collectionData in collections"
            :key="collectionData.collection.id"
            class="collection-item"
            @click="viewCollection(collectionData.collection.id)"
          >
            <div class="collection-cover">
              <div
                v-for="(post, index) in collectionData.recent_posts.slice(0, 4)"
                :key="post.ID"
                class="cover-item"
                :class="`cover-${index + 1}`"
              >
                <img
                  v-if="post.Media && post.Media.length > 0"
                  :src="loadPhotoUrl(post.Media[0].FilePath)"
                  class="cover-image"
                />
                <div
                  v-else
                  class="no-media-cover"
                />
              </div>
              <div
                v-if="collectionData.recent_posts.length === 0"
                class="empty-collection"
              />
            </div>
            <div class="collection-name">
              {{ collectionData.collection.name }}
            </div>
          </div>
        </div>
        
        <div
          v-if="showCreateCollectionModal"
          class="modal-overlay"
          @click.self="showCreateCollectionModal = false"
        >
          <div class="modal-content">
            <h3>Create New Collection</h3>
            <input
              v-model="newCollectionName"
              type="text"
              placeholder="Collection name"
              class="collection-input"
              @keyup.enter="handleCreateCollection"
            />
            <div class="modal-actions">
              <button
                class="btn-cancel"
                @click="showCreateCollectionModal = false"
              >
                Cancel
              </button>
              <button
                class="btn-create"
                @click="handleCreateCollection"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="currentTab === 'tag'"
        class="tab-container"
      >
        <div class="grid-container">
          <div
            v-for="postWithUser in tag"
            :key="postWithUser.post.ID"
            class="grid-item"
            @click="showPostDetail(postWithUser.post?.ID)"
          >
            <div
              v-if="postWithUser.postMedia.length > 0"
              class="media-container"
            >
              <img
                v-if="postWithUser.postMedia[0].MediaType === 'image'"
                :src="loadPhotoUrl(postWithUser.postMedia[0].FilePath)"
                class="post-image"
              />
              <div
                v-else-if="postWithUser.postMedia[0].MediaType === 'video'"
                class="video-thumbnail"
              >
                <video
                  :src="loadPhotoUrl(postWithUser.postMedia[0].FilePath)"
                  class="post-video"
                  muted
                  preload="metadata"
                />
                <div class="video-overlay">
                  <img
                    src="../components/icons/reels-light.png"
                    width="30"
                    height="30"
                    class="video-icon"
                  />
                </div>
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
                    width="20"
                    height="20"
                    class="stat-icon"
                  />
                  <span class="stat-count">{{ postWithUser.post?.LikesCount || 0 }}</span>
                </div>
                <div class="stat-item">
                  <img
                    :src="commentIcon"
                    width="20"
                    height="20"
                    class="stat-icon"
                  />
                  <span class="stat-count">{{ postWithUser.post?.CommentsCount || 0 }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>
</template>

<style>

body {
  margin: 0;
  padding: 0;
  background-color: black; 
  overflow-x: hidden; 
}

* {
  box-sizing: border-box;
}
</style>

<style scoped>
.container {
  display: flex;
  background-color: black;
  min-height: 100vh;
  color: white;
}

.profile {
  width: 100%;
  padding: 30px;
  background-color: black;
  display: flex;
  flex-direction: column;
  max-width: 935px;
  margin: 0 auto;
}

.profile-up {
  width: 100%;
  margin-bottom: 44px;
  display: flex;
  gap: 50px;
  align-items: center;
  justify-content: flex-start;
}

.profile-pic {
  width: 150px;
  height: 150px;
  border-radius: 50%;
  object-fit: cover;
  border: 1px solid #333;
}

.profile-info {
  flex: 1;
}

.upper-profile {
  display: flex;
  gap: 20px;
  align-items: center;
  margin-bottom: 20px;
}

.upper-profile h2 {
  font-size: 20px;
  font-weight: 400;
}

.middle-profile {
  display: flex;
  gap: 40px;
  margin-bottom: 20px;
}

.middle-profile h3 {
  font-size: 16px;
  font-weight: 400;
}

.lower-profile h3 {
  font-size: 14px;
  font-weight: 400;
  white-space: pre-wrap;
}

.btn {
  border-radius: 8px;
  padding: 7px 16px;
  border: none;
  background-color: #363636;
  font-weight: 600;
  color: white;
  font-size: 14px;
  cursor: pointer;
}

.btn:hover {
  background-color: #262626;
}

.block-btn {
  background-color: #ed4956;
}

.block-btn:hover {
  background-color: #c92a34;
}

.block-btn.blocked {
  background-color: #363636;
}

.block-btn.blocked:hover {
  background-color: #262626;
}

.blocked-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  color: white;
}

.blocked-message h2 {
  font-size: 24px;
  margin-bottom: 10px;
  font-weight: 600;
}

.blocked-message p {
  font-size: 16px;
  color: #a0a0a0;
}

.profile-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

.follow-btn {
  background-color: #0095f6;
  color: white;
  border: none;
  font-weight: 600;
  transition: background-color 0.2s ease;
}

.follow-btn:hover:not(:disabled) {
  background-color: #0081d6;
}

.follow-btn.following {
  background-color: #363636;
  color: white;
  border: 1px solid #444;
}

.follow-btn.following:hover:not(:disabled) {
  background-color: #262626;
  border-color: #555;
}

.follow-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.message-btn {
  background-color: #363636;
  color: white;
  border: 1px solid #444;
}

.message-btn:hover {
  background-color: #262626;
  border-color: #555;
}

.tab-button {
  width: 100%;
  padding: 10px 0;
  border-top: 1px solid #333;
  display: flex;
  justify-content: center;
  gap: 60px;
}

.tab-button img {
  cursor: pointer;
  opacity: 0.5;
  transition: opacity 0.2s;
}

.tab-button img.active {
  opacity: 1;
  border-bottom: 1px solid white;
  padding-bottom: 10px;
}

.tab-button img:hover {
  opacity: 1;
}

.tab-container {
  width: 100%;
  
}

.grid-container {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 4px;

  width: 100%;
}

.grid-item {
  width: 100%;
  aspect-ratio: 1 / 1;
  background-color: #121212;
  position: relative;
  overflow: hidden;
  background-color: black;
  cursor: pointer;
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
  gap: 20px;
  align-items: center;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 5px;
  color: white;
  font-weight: 600;
  font-size: 14px;
}

.stat-icon {
  filter: brightness(0) invert(1);
  opacity: 0.9;
}

.stat-count {
  text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.5);
}

.post-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.no-media {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: white;
}

.media-container {
  width: 100%;
  height: 100%;
  position: relative;
}

.video-thumbnail {
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
}

.post-video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.video-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  background: rgba(0, 0, 0, 0.3);
  pointer-events: none;
}

.video-icon {
  opacity: 0.8;
  filter: brightness(0) invert(1);
}

.collections-header {
  width: 100%;
  padding: 20px 0;
  display: flex;
  justify-content: flex-end;
}

.create-collection-btn {
  background: #0095f6;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: background 0.2s;
}

.create-collection-btn:hover {
  background: #1877f2;
}

.collection-item {
  width: 100%;
  aspect-ratio: 1 / 1;
  cursor: pointer;
  position: relative;
}

.collection-cover {
  width: 100%;
  height: calc(100% - 30px);
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  grid-template-rows: repeat(2, 1fr);
  gap: 2px;
  background: #262626;
}

.cover-item {
  width: 100%;
  height: 100%;
  overflow: hidden;
  background: #121212;
}

.cover-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.no-media-cover {
  width: 100%;
  height: 100%;
  background: #1a1a1a;
}

.empty-collection {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #888;
  font-size: 18px;
}

.collection-name {
  width: 100%;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 14px;
  font-weight: 500;
  background: #1a1a1a;
  padding: 0 8px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
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
  z-index: 1000;
}

.modal-content {
  background: #262626;
  border-radius: 12px;
  padding: 24px;
  width: 400px;
  max-width: 90%;
}

.modal-content h3 {
  color: white;
  margin: 0 0 20px 0;
  text-align: center;
}

.collection-input {
  width: 100%;
  padding: 12px;
  background: #121212;
  border: 1px solid #333;
  border-radius: 8px;
  color: white;
  font-size: 14px;
  margin-bottom: 20px;
}

.collection-input:focus {
  outline: none;
  border-color: #0095f6;
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
}

.btn-cancel:hover {
  background: #333;
}

.btn-create {
  background: #0095f6;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
}

.btn-create:hover {
  background: #1877f2;
}

.private-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  color: white;
  border-top: 1px solid #333;
}

.private-icon-container {
    border: 2px solid white;
    border-radius: 50%;
    padding: 15px;
    margin-bottom: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 90px;
    height: 90px;
}

.private-message h2 {
  font-size: 18px;
  margin-bottom: 10px;
  font-weight: 600;
}

.private-message p {
  font-size: 14px;
  color: #a0a0a0;
}
</style>