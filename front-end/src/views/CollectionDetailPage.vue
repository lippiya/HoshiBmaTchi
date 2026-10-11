<script setup lang="ts">
import SideBar from "@/components/SideBar.vue";
import { onMounted, ref, computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import { usePhoto } from "@/composables/usePhoto";
import { isTokenExpired } from "@/router";
import { useShowPost } from "@/composables/useOverlay";
import PostDetail from "@/components/PostDetail.vue";
import likeIcon from "@/components/icons/notif-light.png";
import commentIcon from "@/components/icons/comment-light.png";
import { 
  getPostsByCollectionID, 
  updateCollection, 
  deleteCollection,
  type Post 
} from "@/services/collectionService";
import { usePostFetcher } from "@/composables/usePostFetcher";

const { showPost } = useShowPost();
const router = useRouter();
const route = useRoute();
const { loadPhotoUrl } = usePhoto();
const { fetchPostsWithDetails } = usePostFetcher();

const collectionID = ref<string>("");
const collectionName = ref<string>("");
const posts = ref<Post[]>([]);
const showMenu = ref(false);
const showRenameModal = ref(false);
const newName = ref("");
const currUserID = ref<string>("");

interface PostWithMedia {
  post: Post;
  postMedia: any[];
}

const postsWithMedia = ref<PostWithMedia[]>([]);

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

const fetchCollection = async () => {
  try {
    const response = await fetch(
      `https://localhost/api/v1/collections/posts/${collectionID.value}`,
      {
        method: "GET",
        headers: {
          Authorization: `Bearer ${localStorage.getItem("token")}`,
        },
      }
    );

    if (response.ok) {
      const data = await response.json();
      posts.value = data.posts || [];
      
      const token = localStorage.getItem("token");
      postsWithMedia.value = await Promise.all(
        posts.value.map(async (post) => {
          try {
            const mediaRes = await fetch(
              `https://localhost/api/v1/posts/get-post-media-by-id/${post.ID}`,
              {
                headers: { Authorization: `Bearer ${token}` }
              }
            );
            const mediaData = mediaRes.ok ? await mediaRes.json() : {};
            return {
              post: post,
              postMedia: mediaData.media || []
            };
          } catch (err) {
            console.error(`Failed to fetch media for post ${post.ID}:`, err);
            return {
              post: post,
              postMedia: []
            };
          }
        })
      );
    }
  } catch (error) {
    console.error("Failed to fetch collection:", error);
  }
};

const fetchCollectionDetails = async () => {
  try {
    const response = await fetch(
      "https://localhost/api/v1/users/get-current-user",
      {
        method: "GET",
        headers: {
          Authorization: `Bearer ${localStorage.getItem("token")}`,
        },
      }
    );

    if (response.ok) {
      const userData = await response.json();
      const userId = userData.user.ID || userData.user.id;
      
      const collectionsResponse = await fetch(
        `https://localhost/api/v1/collections/user/${userId}`,
        {
          method: "GET",
          headers: {
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
        }
      );

      if (collectionsResponse.ok) {
        const collectionsData = await collectionsResponse.json();
        const collection = collectionsData.collections?.find(
          (c: any) => c.collection.id === collectionID.value
        );
        if (collection) {
          collectionName.value = collection.collection.name;
        }
      }
    }
  } catch (error) {
    console.error("Failed to fetch collection details:", error);
  }
};

const handleRename = async () => {
  if (!newName.value.trim()) return;
  
  try {
    await updateCollection(collectionID.value, newName.value);
    collectionName.value = newName.value;
    newName.value = "";
    showRenameModal.value = false;
    showMenu.value = false;
  } catch (error) {
    console.error("Failed to rename collection:", error);
  }
};

const handleDelete = async () => {
  if (!confirm("Are you sure you want to delete this collection?")) return;
  
  try {
    await deleteCollection(collectionID.value);
    router.push("/profile");
  } catch (error) {
    console.error("Failed to delete collection:", error);
  }
};

const showPostDetail = (postID: string) => {
  showPost(postID);
};

onMounted(async () => {
  collectionID.value = route.params.collectionID as string;
  await fetchCurrentUser();
  await fetchCollectionDetails();
  await fetchCollection();
});
</script>

<template>
  <PostDetail />
  <div class="container">
    <SideBar />
    <div class="collection-detail">
      <div class="collection-header">
        <h2>{{ collectionName }}</h2>
        <div class="menu-container">
          <button
            class="menu-btn"
            @click="showMenu = !showMenu"
          >
            ⋮
          </button>
          <div
            v-if="showMenu"
            class="menu-dropdown"
          >
            <div
              class="menu-item"
              @click="showRenameModal = true; showMenu = false"
            >
              Rename
            </div>
            <div
              class="menu-item delete"
              @click="handleDelete"
            >
              Delete
            </div>
          </div>
        </div>
      </div>

      <div class="grid-container">
        <div
          v-for="postData in postsWithMedia"
          :key="postData.post.ID"
          class="grid-item"
          @click="showPostDetail(postData.post.ID)"
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
                  width="20"
                  height="20"
                  class="stat-icon"
                />
                <span class="stat-count">{{ postData.post.LikesCount || 0 }}</span>
              </div>
              <div class="stat-item">
                <img
                  :src="commentIcon"
                  width="20"
                  height="20"
                  class="stat-icon"
                />
                <span class="stat-count">{{ postData.post.CommentsCount || 0 }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="showRenameModal"
        class="modal-overlay"
        @click.self="showRenameModal = false"
      >
        <div class="modal-content">
          <h3>Rename Collection</h3>
          <input
            v-model="newName"
            type="text"
            placeholder="New collection name"
            class="collection-input"
            @keyup.enter="handleRename"
          />
          <div class="modal-actions">
            <button
              class="btn-cancel"
              @click="showRenameModal = false"
            >
              Cancel
            </button>
            <button
              class="btn-create"
              @click="handleRename"
            >
              Rename
            </button>
          </div>
        </div>
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

.collection-detail {
  width: calc(100% - 250px);
  padding: 40px;
  overflow-y: auto;
}

.collection-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  padding-bottom: 20px;
  border-bottom: 1px solid #333;
}

.collection-header h2 {
  color: white;
  font-size: 28px;
  margin: 0;
}

.menu-container {
  position: relative;
}

.menu-btn {
  background: transparent;
  border: none;
  color: white;
  font-size: 24px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 4px;
}

.menu-btn:hover {
  background: #262626;
}

.menu-dropdown {
  position: absolute;
  right: 0;
  top: 100%;
  background: #262626;
  border: 1px solid #333;
  border-radius: 8px;
  min-width: 150px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
  z-index: 100;
}

.menu-item {
  padding: 12px 16px;
  color: white;
  cursor: pointer;
  transition: background 0.2s;
}

.menu-item:hover {
  background: #333;
}

.menu-item.delete {
  color: #ff4444;
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
  cursor: pointer;
}

.media-container {
  width: 100%;
  height: 100%;
  position: relative;
}

.post-image {
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

.post-video {
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
</style>
