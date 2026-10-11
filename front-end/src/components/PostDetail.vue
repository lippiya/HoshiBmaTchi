<script setup lang="ts">
import { useShowPost } from "@/composables/useOverlay";
import { isTokenExpired } from "@/router";
import { usePhoto } from "@/composables/usePhoto";
import { onMounted, onUnmounted, ref, watchEffect } from "vue";
import { usePostFetcher } from "@/composables/usePostFetcher";
import { useEventBus } from "@/composables/useEventBus";
import { useRouter } from "vue-router";
import { getCollectionsByUserID, addPostToCollection, removePostFromCollection } from "@/services/collectionService";
import { useHashtag } from "@/composables/useHashtag";
import { useNotification } from "@/composables/useNotification";

import likeLight from "@/components/icons/notif-light.png";
import commentLight from "@/components/icons/comment-light.png";
import shareLight from "@/components/icons/messages-light.png";
import savedLight from "@/components/icons/saved-light.png";
import savedDark from "@/components/icons/saved-dark.png";
import userLight from "@/components/icons/user-light.png";

const router = useRouter();
const { fetchPost } = usePostFetcher();
const { loadPhotoUrl} = usePhoto();
const { emit } = useEventBus();
const { formatTextWithHashtags, navigateToHashtag } = useHashtag();
const { createNotification } = useNotification();

const currUserID = ref<string>("");
const currUser = ref<any>({ Username: "", ProfilePicture: "" });
const { showPostContent, close, passedPostID } = useShowPost();

const p = ref<any>({
  post: {},
  user: {},
  postMedia: [],
  collaborators: []
});
const currentSlides = ref<Record<string, number>>({});
const userLikes = ref<Record<string, boolean>>({});
const userComments = ref<Record<string, boolean>>({});
const userSaves = ref<Record<string, boolean>>({});

const showSaveCollectionModal = ref(false);
const userCollections = ref<any[]>([]);
const selectedCollectionForSave = ref<string>("");
const savingPostId = ref<string>("");
const loadingCollections = ref(false);

const replies = ref<any[]>([]);
const replyText = ref("");
const loadingReplies = ref(false);
const hasMoreReplies = ref(true);
const repliesPage = ref(1);
const repliesLimit = 20;
const showGifPicker = ref(false);
const selectedGif = ref("");
const replyToReplyId = ref<string | null>(null);
const showReplyReplies = ref<Record<string, boolean>>({});
const replyReplies = ref<Record<string, unknown[]>>({});
const loadingReplyReplies = ref<Record<string, boolean>>({});
const replyLikes = ref<Record<string, boolean>>({});

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

const fetchUser = async () => {
  const token = localStorage.getItem("token");
  if (!token || isTokenExpired(token)) {
    localStorage.removeItem("token");
    router.push("/login");
    return false;
  }
  try {
    const response = await fetch("https://localhost/api/v1/users/get-current-user", {
      method: "GET",
      headers: { Authorization: `Bearer ${token}` },
    });

    if (response.ok) {
      const data = await response.json();
      const user = data.user;
      currUserID.value = user.ID;
      currUser.value = { 
        Username: user.Username || user.username || "", 
        ProfilePicture: user.ProfilePicture || user.profile_picture || "" 
      };
      return true;
    }
  } catch (error) {
    console.error("Failed to fetch user:", error);
  }
};

const fetchPostInit = async () => {
  const token = localStorage.getItem("token");
  if (!token || isTokenExpired(token)) {
    localStorage.removeItem("token");
    router.push("/login");
    return;
  }

  if (!passedPostID.value) {
    return;
  }
  const data = await fetchPost(passedPostID.value);
  if (data) {
    p.value = data;
    await fetchUserInteractions();
  } else {
    console.error("Failed to load post");
  }
};

const fetchUserInteractions = async () => {
  const token = localStorage.getItem("token");
  if (!token || !p.value.post.ID) return;

  try {
      const [likesRes, commentsRes, savesRes] = await Promise.allSettled([
        fetch(`https://localhost/api/v1/posts/check-like/${p.value.post.ID}`, {
          headers: { Authorization: `Bearer ${token}` }
        }),
        fetch(`https://localhost/api/v1/social/check-user-commented/${p.value.post.ID}`, {
          headers: { Authorization: `Bearer ${token}` }
        }),
        fetch(`https://localhost/api/v1/posts/check-save/${p.value.post.ID}`, {
          headers: { Authorization: `Bearer ${token}` }
        })
      ]);

    if (likesRes.status === "fulfilled" && likesRes.value.ok) {
      try {
        const data = await likesRes.value.json();
        userLikes.value[p.value.post.ID] = data.liked || false;
      } catch {
        userLikes.value[p.value.post.ID] = false;
      }
    } else if (likesRes.status === "fulfilled") {
      console.error("Check like failed. Status:", likesRes.value.status);
    }

    if (commentsRes.status === "fulfilled" && commentsRes.value.ok) {
      try {
        const data = await commentsRes.value.json();
        userComments.value[p.value.post.ID] = data.commented || false;
      } catch {
        userComments.value[p.value.post.ID] = false;
      }
    } else if (commentsRes.status === "fulfilled") {
      console.error("Check comment failed. Status:", commentsRes.value.status);
    }

    if (savesRes.status === "fulfilled" && savesRes.value.ok) {
      try {
        const data = await savesRes.value.json();
        userSaves.value[p.value.post.ID] = data.saved || false;
      } catch {
        userSaves.value[p.value.post.ID] = false;
      }
    } else if (savesRes.status === "fulfilled") {
      console.error("Check save failed. Status:", savesRes.value.status);
    }
  } catch (error) {
    console.error("Error fetching interactions:", error);
  }
};

const fetchReplies = async (reset = false) => {
  if (!passedPostID.value || loadingReplies.value) return;

  if (reset) {
    repliesPage.value = 1;
    replies.value = [];
    hasMoreReplies.value = true;
  }

  loadingReplies.value = true;
  const token = localStorage.getItem("token");
  
  try {
    const response = await fetch(
      `https://localhost/api/v1/social/get-comments/${passedPostID.value}?page=${repliesPage.value}&limit=${repliesLimit}&sort=likes`,
      {
        headers: { Authorization: `Bearer ${token}` }
      }
    );

    if (response.ok) {
      try {
        const data = await response.json();
        const newReplies = data.comments || [];
        
        
        for (const reply of newReplies) {
          if (reply.UserID && !reply.user) {
            try {
              const userResp = await fetch(`https://localhost/api/v1/users/get-user-by-id/${reply.UserID}`, {
                headers: { Authorization: `Bearer ${token}` }
              });
              if (userResp.ok) {
                const userData = await userResp.json();
                reply.user = { 
                  ...userData.user, 
                  ID: reply.UserID,
                  Username: userData.user?.Username || userData.user?.username || "Unknown", 
                  ProfilePicture: userData.user?.ProfilePicture || userData.user?.profile_picture || "" 
                };
              } else {
                reply.user = { Username: "Unknown", ProfilePicture: "", ID: reply.UserID };
              }
            } catch (userError) {
              console.error("Failed to fetch user data for reply:", userError);
              reply.user = { Username: "Unknown", ProfilePicture: "", ID: reply.UserID };
            }
          }
        }
        
        if (reset) {
          replies.value = newReplies;
        } else {
          replies.value = [...replies.value, ...newReplies];
        }

        hasMoreReplies.value = newReplies.length === repliesLimit;
        repliesPage.value++;
      } catch (parseError) {
        console.error("Failed to parse JSON response:", parseError, "Status:", response.status);
        replies.value = [];
      }
    } else {
      console.error("Failed to fetch replies. Status:", response.status);
      const contentType = response.headers.get("content-type");
      if (!contentType?.includes("application/json")) {
        const text = await response.text();
        console.error("Response is not JSON. Content-Type:", contentType, "Body:", text.substring(0, 200));
      }
    }
  } catch (error) {
    console.error("Failed to fetch replies:", error);
  } finally {
    loadingReplies.value = false;
  }
};

onMounted(() => {
  fetchUser();
  
  const handleOpenShareModal = () => {
    if (showPostContent.value && passedPostID.value) {
      openShareModal();
    }
  };
  
  window.addEventListener('open-share-modal', handleOpenShareModal);
  
  onUnmounted(() => {
    window.removeEventListener('open-share-modal', handleOpenShareModal);
  });
});

watchEffect(() => {
  if (passedPostID.value) {
    fetchPostInit();
    fetchReplies();
  }
});

const handleScroll = (e: Event) => {
  const target = e.target as HTMLElement;
  const scrollBottom = target.scrollHeight - target.scrollTop - target.clientHeight;
  
  if (scrollBottom < 100 && hasMoreReplies.value && !loadingReplies.value) {
    fetchReplies();
  }
};

const submitReply = async () => {
  if ((!replyText.value.trim() && !selectedGif.value) || !passedPostID.value) return;

  const token = localStorage.getItem("token");
  const formData = new FormData();
  
  
  formData.append("post_id", passedPostID.value);
  
  if (replyToReplyId.value) {
    formData.append("parent_comment_id", replyToReplyId.value);
  }
  
  if (selectedGif.value) {
    formData.append("gif_url", selectedGif.value);
  } else {
    formData.append("text", replyText.value);
  }

  const tempReply = {
    ID: `temp-${Date.now()}`,
    Text: replyText.value || selectedGif.value,
    GifUrl: selectedGif.value || null,
    user: { 
      Username: currUser.value.Username, 
      ProfilePicture: currUser.value.ProfilePicture, 
      ID: currUserID.value 
    },
    UserID: currUserID.value,
    LikesCount: 0,
    CreatedAt: new Date().toISOString(),
    ParentCommentID: replyToReplyId.value || null
  };

  replies.value.unshift(tempReply);
  replyText.value = "";
  selectedGif.value = "";
  showGifPicker.value = false;
  replyToReplyId.value = null;

  try {
    const response = await fetch("https://localhost/api/v1/social/create-comment", {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` },
      body: formData
    });

    if (response.ok) {
      await fetchReplies(true);
      await fetchPostInit();
      userComments.value[passedPostID.value] = true;
      emit("post-liked", { postID: passedPostID.value });
      
      const postOwnerID = p.value.user?.ID || p.value.post?.UserID;
      if (postOwnerID && currUserID.value && passedPostID.value) {
        const commentText = tempReply.GifUrl ? "[GIF]" : tempReply.Text;
        await createNotification(
          postOwnerID,
          currUserID.value,
          "comment",
          passedPostID.value,
          commentText?.substring(0, 100)
        );
      }
    } else {
      replies.value = replies.value.filter(r => r.ID !== tempReply.ID);
      console.error("Failed to submit reply. Status:", response.status);
      const contentType = response.headers.get("content-type");
      if (!contentType?.includes("application/json")) {
        const text = await response.text();
        console.error("Response is not JSON. Content-Type:", contentType, "Body:", text.substring(0, 200));
      }
    }
  } catch (error) {
    console.error("Failed to submit reply:", error);
  }
};

const deleteReply = async (replyId: string) => {
  const token = localStorage.getItem("token");
  
  const replyIndex = replies.value.findIndex(r => r.ID === replyId);
  if (replyIndex === -1) return;
  
  const removedReply = replies.value.splice(replyIndex, 1)[0];
  
  try {
    const response = await fetch(`https://localhost/api/v1/social/delete-comment/${replyId}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` }
    });

    if (!response.ok) {
      replies.value.splice(replyIndex, 0, removedReply);
    }
  } catch (error) {
    replies.value.splice(replyIndex, 0, removedReply);
    console.error("Failed to delete reply:", error);
  }
};

const likeReply = async (replyId: string) => {
  const token = localStorage.getItem("token");
  
  const replyIndex = replies.value.findIndex(r => r.ID === replyId);
  if (replyIndex === -1) return;
  
  const reply = replies.value[replyIndex];
  const wasLiked = replyLikes.value[replyId];
  
  if (wasLiked) {
    replyLikes.value[replyId] = false;
    reply.LikesCount = Math.max(0, (reply.LikesCount || 1) - 1);
  } else {
    replyLikes.value[replyId] = true;
    reply.LikesCount = (reply.LikesCount || 0) + 1;
  }
  
  try {
    const response = await fetch(`https://localhost/api/v1/social/like-comment/${replyId}`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}` }
    });

    if (!response.ok) {
      if (wasLiked) {
        replyLikes.value[replyId] = true;
        reply.LikesCount = (reply.LikesCount || 0) + 1;
      } else {
        replyLikes.value[replyId] = false;
        reply.LikesCount = Math.max(0, (reply.LikesCount || 1) - 1);
      }
    }
  } catch (error) {
    if (wasLiked) {
      replyLikes.value[replyId] = true;
      reply.LikesCount = (reply.LikesCount || 0) + 1;
    } else {
      replyLikes.value[replyId] = false;
      reply.LikesCount = Math.max(0, (reply.LikesCount || 1) - 1);
    }
    console.error("Failed to like reply:", error);
  }
};

const replyToReply = (replyId: string) => {
  replyToReplyId.value = replyId;
};

const toggleReplyReplies = async (replyId: string) => {
  if (showReplyReplies.value[replyId]) {
    showReplyReplies.value[replyId] = false;
    return;
  }

  if (replyReplies.value[replyId] && replyReplies.value[replyId].length > 0) {
    showReplyReplies.value[replyId] = true;
    return;
  }

  loadingReplyReplies.value[replyId] = true;
  const token = localStorage.getItem("token");
  
  try {
    const response = await fetch(
      `https://localhost/api/v1/social/get-comment-replies/${replyId}`,
      {
        headers: { Authorization: `Bearer ${token}` }
      }
    );

    if (response.ok) {
      try {
        const data = await response.json();
        const replies = data.replies || [];
        
        
        for (const reply of replies) {
          if (reply.UserID && !reply.user) {
            try {
              const userResp = await fetch(`https://localhost/api/v1/users/get-user-by-id/${reply.UserID}`, {
                headers: { Authorization: `Bearer ${token}` }
              });
              if (userResp.ok) {
                const userData = await userResp.json();
                reply.user = { 
                  ...userData.user, 
                  ID: reply.UserID,
                  Username: userData.user?.Username || userData.user?.username || "Unknown", 
                  ProfilePicture: userData.user?.ProfilePicture || userData.user?.profile_picture || "" 
                };
              } else {
                reply.user = { Username: "Unknown", ProfilePicture: "", ID: reply.UserID };
              }
            } catch (userError) {
              console.error("Failed to fetch user data for nested reply:", userError);
              reply.user = { Username: "Unknown", ProfilePicture: "", ID: reply.UserID };
            }
          }
        }
        
        replyReplies.value[replyId] = replies;
        showReplyReplies.value[replyId] = true;
      } catch (parseError) {
        console.error("Failed to parse reply replies JSON:", parseError);
        replyReplies.value[replyId] = [];
      }
    } else {
      console.error("Failed to fetch reply replies. Status:", response.status);
    }
  } catch (error) {
    console.error("Failed to fetch reply replies:", error);
  } finally {
    loadingReplyReplies.value[replyId] = false;
  }
};




const pfpError = (e: Event) => {
  const target = e.target as HTMLImageElement;
  target.src = userLight;
};

const dateToString = (dateStr: string) => {
  const date = new Date(dateStr);
  return date.toLocaleString();
};

const deletePost = async () => {
  const token = localStorage.getItem("token");
  if (!passedPostID.value) return;
  
  try {
    const response = await fetch(`https://localhost/api/v1/posts/delete-post/${passedPostID.value}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` }
    });
    if (response.ok) {
      emit("post-deleted", { postID: passedPostID.value });
      close();
    }
  } catch (error) {
    console.error("Failed to delete post:", error);
  }
};

const showShareModal = ref(false);
const shareConversations = ref<any[]>([]);
const shareSearchQuery = ref("");
const loadingConversations = ref(false);
const shareSearchResults = ref<any[]>([]);
const loadingShareSearch = ref(false);
const selectedShareUserId = ref("");
let shareSearchTimeout: number | null = null;

const openShareModal = async () => {
  showShareModal.value = true;
  shareSearchQuery.value = "";
  shareSearchResults.value = [];
  selectedShareUserId.value = "";
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
    const token = localStorage.getItem("token");
    
    try {
      const response = await fetch(
        `https://localhost/api/v1/users/search?query=${encodeURIComponent(query)}`,
        { headers: { Authorization: `Bearer ${token}` } }
      );
      if (response.ok) {
        const data = await response.json();
        shareSearchResults.value = data.users || [];
      }
    } catch (error) {
      console.error("Failed to search users:", error);
    } finally {
      loadingShareSearch.value = false;
    }
  }, 300);
};

const shareToUser = async () => {
  if (!selectedShareUserId.value || !passedPostID.value) return;
  
  const token = localStorage.getItem("token");
  const postUrl = `${window.location.origin}/post/${passedPostID.value}`;
  
  try {
    const convResponse = await fetch(
      "https://localhost/api/v1/messages/conversations",
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          participant_ids: [selectedShareUserId.value],
          is_group: false
        })
      }
    );
    
    if (convResponse.ok) {
      const convData = await convResponse.json();
      const conversationId = convData.id || convData.ID || convData.conversation_id;
      
      const formData = new FormData();
      formData.append("content", `Check out this post: ${postUrl}`);
      
      const msgResponse = await fetch(
        `https://localhost/api/v1/messages/conversations/${conversationId}/messages`,
        {
          method: "POST",
          headers: { Authorization: `Bearer ${token}` },
          body: formData
        }
      );
      
      if (msgResponse.ok) {
        showShareModal.value = false;
        shareSearchQuery.value = "";
        shareSearchResults.value = [];
        selectedShareUserId.value = "";
      }
    }
  } catch (error) {
    console.error("Failed to share post:", error);
  }
};

const likeClicked = async (postID: string) => {
  try {
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
      const wasLiked = userLikes.value[postID];
      userLikes.value[postID] = !wasLiked;
      fetchPostInit();
      emit("post-liked", { postID });
      
      
      if (!wasLiked && data.liked && p.value.post.UserID && currUserID.value) {
        await createNotification(
          p.value.post.UserID,
          currUserID.value,
          "like",
          postID
        );
      }
    }
  } catch{
    console.error("Failed to like post");
  }
};

const saveClicked = async (postID: string) => {
  const token = localStorage.getItem("token");
  if (!token) return;

  if (!currUserID.value) {
    await fetchUser();
  }

  try {
    const userId = currUserID.value;
    if (!userId) {
      return;
    }
    
    if (userSaves.value[postID]) {
      const data = await getCollectionsByUserID(userId);
      for (const collection of data) {
        try {
          await removePostFromCollection(collection.collection.id, postID);
        } catch (error) {
          
        }
      }
      userSaves.value[postID] = false;
      return;
    }
    
    
    const data = await getCollectionsByUserID(userId);
    
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

const selectGif = (gifUrl: string) => {
  selectedGif.value = gifUrl;
  showGifPicker.value = false;
};

const cancelGif = () => {
  selectedGif.value = "";
  showGifPicker.value = false;
};

</script>

<template>
  <Teleport to="body">
    <div
      v-if="showPostContent"
      class="show-post-background"
      @click.self="close"
    >
      <div class="show-post-container">
        <div class="post-left">
          <div class="user-profile">
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

            <button
              v-if="p.user.ID != currUserID"
              style="text-decoration: none; background-color: transparent; border: none; color:blue; font-weight: 600; font-size: 18px;"
            >
              Follow
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
                    width="250"
                    height="250"
                  />
                  <video
                    v-else-if="media.MediaType === 'video'"
                    :src="loadPhotoUrl(media.FilePath)"
                    width="250"
                    height="250"
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
                  :src=" commentLight"
                  alt="comment"
                  width="30"
                  height="30"
                  :style="{ cursor: 'pointer' }"
                />
                <img
                  :src="shareLight"
                  alt="share"
                  width="30"
                  height="30"
                  style="cursor: pointer;"
                  @click="openShareModal"
                />
                <button
                  v-if="p.user.ID === currUserID"
                  class="delete-post-btn"
                  @click="deletePost"
                >
                  Delete
                </button>
              </div>
              <div class="interaction-right">
                <img
                  :src="userSaves[p.post.ID] ? savedDark : savedLight"
                  alt="save"
                  width="30"
                  height="30"
                  :style="{ cursor: 'pointer', filter: userSaves[p.post.ID] ? 'invert(85%) sepia(89%) saturate(1991%) hue-rotate(358deg) brightness(101%) contrast(103%)' : 'none' }"
                  @click.stop="saveClicked(p.post.ID)"
                />
              </div>
            </div>
            
            <span
              v-if="!p.post.IsHideLikes"
              style="font-weight: 500; cursor: pointer;"
            >{{ p.post.LikesCount }} Likes</span>
            
            <h3 style="font-weight: 600; margin-bottom: 10px;">
              <template
                v-for="part in formatTextWithHashtags(p.post.Caption)"
                :key="part.key"
              >
                <span
                  v-if="part.isHashtag"
                  class="hashtag"
                  @click.stop="navigateToHashtag(part.text); close();"
                >{{ part.text }}</span>
                <span v-else>{{ part.text }}</span>
              </template>
            </h3>
            <p style="font-size: 12px;">
              {{ dateToString(p.post.CreatedAt) }}
            </p>
          </div>
        </div>

        <div class="post-right">
          <div
            class="replies-section"
            @scroll="handleScroll"
          >
            <div
              v-for="reply in replies"
              :key="reply.ID"
              class="reply-item"
            >
              <div class="reply-header">
                <img
                  v-if="reply.user && reply.user.ProfilePicture"
                  :src="loadPhotoUrl(reply.user.ProfilePicture)"
                  width="32"
                  height="32"
                  style="border-radius: 50%; object-fit: cover;"
                  @error="pfpError"
                />
                <img
                  v-else
                  :src="userLight"
                  width="32"
                  height="32"
                  style="border-radius: 50%; object-fit: cover;"
                />
                <div class="reply-content">
                  <span class="reply-username">{{ reply.user?.Username || 'Unknown' }}</span>
                  <p
                    v-if="reply.Text"
                    class="reply-text"
                  >
                    {{ reply.Text }}
                  </p>
                  <img
                    v-if="reply.GifUrl"
                    :src="reply.GifUrl"
                    class="reply-gif"
                  />
                  <div class="reply-actions">
                    <span class="reply-time">{{ dateToString(reply.CreatedAt) }}</span>
                    <span
                      class="reply-likes"
                      @click="likeReply(reply.ID)"
                    >{{ reply.LikesCount }} likes</span>
                    <span
                      class="reply-action-btn"
                      @click="replyToReply(reply.ID)"
                    >Reply</span>
                    <span
                      v-if="reply.RepliesCount > 0"
                      class="reply-action-btn"
                      @click="toggleReplyReplies(reply.ID)"
                    >
                      {{ showReplyReplies[reply.ID] ? 'Hide' : 'View' }} {{ reply.RepliesCount }} replies
                    </span>
                  </div>
                  
                  <div
                    v-if="showReplyReplies[reply.ID] && replyReplies[reply.ID]"
                    class="nested-replies"
                  >
                    <div
                      v-for="nestedReply in replyReplies[reply.ID]"
                      :key="nestedReply.ID"
                      class="reply-item nested"
                    >
                      <div class="reply-header">
                        <img
                          v-if="nestedReply.user && nestedReply.user.ProfilePicture"
                          :src="loadPhotoUrl(nestedReply.user.ProfilePicture)"
                          width="28"
                          height="28"
                          style="border-radius: 50%; object-fit: cover;"
                          @error="pfpError"
                        />
                        <img
                          v-else
                          :src="userLight"
                          width="28"
                          height="28"
                          style="border-radius: 50%; object-fit: cover;"
                        />
                        <div class="reply-content">
                          <span class="reply-username">{{ nestedReply.user?.Username || 'Unknown' }}</span>
                          <p
                            v-if="nestedReply.Text"
                            class="reply-text"
                          >
                            {{ nestedReply.Text }}
                          </p>
                          <img
                            v-if="nestedReply.GifUrl"
                            :src="nestedReply.GifUrl"
                            class="reply-gif"
                          />
                          <div class="reply-actions">
                            <span class="reply-time">{{ dateToString(nestedReply.CreatedAt) }}</span>
                            <span
                              class="reply-likes"
                              @click="likeReply(nestedReply.ID)"
                            >{{ nestedReply.LikesCount }} likes</span>
                          </div>
                        </div>
                        <div
                          v-if="nestedReply.UserID === currUserID || nestedReply.user?.ID === currUserID"
                          class="reply-menu"
                        >
                          <button
                            class="menu-btn"
                            @click="deleteReply(nestedReply.ID)"
                          >
                            ...
                          </button>
                        </div>
                      </div>
                    </div>
                    <div
                      v-if="loadingReplyReplies[reply.ID]"
                      class="loading-indicator"
                    >
                      Loading...
                    </div>
                  </div>
                </div>
                <div class="reply-actions">
                  <button
                    class="like-reply-btn"
                    :class="{ liked: replyLikes[reply.ID] }"
                    @click="likeReply(reply.ID)"
                  >
                    ♥
                  </button>
                  <span class="reply-likes">{{ reply.LikesCount || 0 }}</span>
                  <button
                    v-if="reply.UserID === currUserID || reply.user?.ID === currUserID"
                    class="delete-reply-btn"
                    @click="deleteReply(reply.ID)"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
            
            <div
              v-if="loadingReplies && p.post.IsCommentOff === false"
              class="loading-indicator"
            >
              Loading replies...
            </div>
            
            <div 
              v-if="p.post.IsCommentOff === true" 
              class="loading-indicator"
            >
              Comments for this post are disabled.
            </div>
          </div>

          <div class="reply-input-section">
            <div
              v-if="selectedGif"
              class="selected-gif-preview"
            >
              <img
                :src="selectedGif"
                alt="Selected GIF"
              />
              <button
                class="cancel-gif-btn"
                @click="cancelGif"
              >
                ×
              </button>
            </div>
            <div
              v-if="replyToReplyId"
              class="replying-to-indicator"
            >
              Replying to a comment
              <button @click="replyToReplyId = null">
                Cancel
              </button>
            </div>
            <div
              v-if="p.post.IsCommentOff === false"
              class="input-container"
            >
              <textarea
                v-model="replyText"
                placeholder="Add a reply..."
                class="reply-input"
                :disabled="!!selectedGif"
                @keydown.enter.prevent="submitReply"
              />
              <div class="input-actions">
                <button
                  class="gif-btn"
                  @click="showGifPicker = !showGifPicker"
                >
                  GIF
                </button>
                <button
                  class="submit-btn"
                  :disabled="!replyText.trim() && !selectedGif"
                  @click="submitReply"
                >
                  Post
                </button>
              </div>
            </div>
            <div
              v-if="showGifPicker"
              class="gif-picker"
            >
              <div class="gif-grid">
                <img
                  src="https://media.giphy.com/media/3o7btPCcdNniyf0ArS/giphy.gif"
                  @click="selectGif('https://media.giphy.com/media/3o7btPCcdNniyf0ArS/giphy.gif')"
                />
                <img
                  src="https://media.giphy.com/media/l0MYt5jPR6QX5pnqM/giphy.gif"
                  @click="selectGif('https://media.giphy.com/media/l0MYt5jPR6QX5pnqM/giphy.gif')"
                />
                <img
                  src="https://media.giphy.com/media/kyLYXonQYYfwYDIeZl/giphy.gif"
                  @click="selectGif('https://media.giphy.com/media/kyLYXonQYYfwYDIeZl/giphy.gif')"
                />
              </div>
            </div>
          </div>
        </div>
      </div>

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

      <div
        v-if="showShareModal"
        class="modal-overlay"
        @click.self="showShareModal = false"
      >
        <div class="save-modal-content">
          <h3>Send to User</h3>
          <input
            v-model="shareSearchQuery"
            type="text"
            placeholder="Search users..."
            class="collection-input"
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
            class="no-conversations"
          >
            No users found
          </div>
          <div
            v-else-if="shareSearchResults.length > 0"
            class="collections-list"
          >
            <div
              v-for="user in shareSearchResults"
              :key="user.ID || user.id"
              class="collection-option user-option"
              :class="{ selected: selectedShareUserId === (user.ID || user.id) }"
              @click="selectedShareUserId = user.ID || user.id"
            >
              <img
                :src="user.ProfilePicture ? loadPhotoUrl(user.ProfilePicture) : userLight"
                class="user-avatar-small"
                @error="pfpError"
              />
              <span>{{ user.Username }}</span>
            </div>
          </div>
          <div class="modal-actions">
            <button
              class="btn-cancel"
              @click="showShareModal = false"
            >
              Cancel
            </button>
            <button
              class="btn-save"
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
.show-post-background {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  backdrop-filter: blur(3px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  background-color: rgba(0, 0, 0, 0.8);
}

.show-post-container {
  background-color: black;
  color: white;
  display: flex;
  width: 80%;
  max-width: 1200px;
  height: 85vh;
  border: 0.2px solid white;
  border-radius: 8px;
  overflow: hidden;
}

.post-left {
  max-height: 100%;
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 25px;
  border-right: 1px solid #333;
  
}

.post-right {
  flex: 1;
  display: flex;
  flex-direction: column;
  background-color: #0a0a0a;
}

.user-profile {
  align-items: center;
  justify-content: flex-start;
  display: flex;
  margin-bottom: 20px;
  gap: 20px;
  height: 5%;
}

.post-content {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 80%;
}

.media-slider {
  position: relative;
  width: 100%;
  max-width: 500px;
  display: flex;
  justify-content: center;
  flex-direction: column;
  height: 500px;
  margin-bottom: 15px;
  margin-left: auto;
  margin-right: auto;
}

.slider-container {
  display: flex;
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  border-radius: 8px;
  background-color: black; 
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

  display: flex;
  align-items: center;    
  justify-content: center;
}

.slide.active {
  opacity: 1;
  visibility: visible;
  height: 80%;
  z-index: 2;
}

.slide img,
.slide video {
  width: 100%; 
  height: 100%;
  object-fit: contain; 
  
  border-radius: 8px;
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
  margin-bottom: 12px;
  width: 100%;
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

.replies-section {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.reply-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.reply-item.nested {
  margin-left: 10px;
  padding-left: 12px;
}

.reply-header {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.reply-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.reply-username {
  font-weight: 600;
  font-size: 14px;
  color: white;
}

.reply-text {
  font-size: 14px;
  color: #e0e0e0;
  margin: 0;
  word-wrap: break-word;
}

.reply-gif {
  max-width: 200px;
  border-radius: 8px;
  margin-top: 8px;
}

.reply-actions {
  display: flex;
  gap: 12px;
  margin-top: 4px;
  font-size: 12px;
  color: #888;
}

.reply-time {
  color: #666;
}

.reply-likes {
  cursor: pointer;
  color: #888;
}

.reply-likes:hover {
  color: white;
}

.reply-action-btn {
  cursor: pointer;
  color: #888;
  font-weight: 500;
}

.reply-action-btn:hover {
  color: white;
}

.like-reply-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  color: #888;
  font-size: 12px;
  font-weight: 500;
  transition: color 0.2s;
}

.like-reply-btn:hover {
  color: white;
}

.like-reply-btn.liked {
  color: #ed4956;
}

.delete-reply-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  color: #888;
  font-size: 12px;
  font-weight: 500;
  transition: color 0.2s;
}

.delete-reply-btn:hover {
  color: #ed4956;
}

.delete-post-btn {
  background: #ed4956;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  font-size: 14px;
  transition: background 0.2s;
}

.delete-post-btn:hover {
  background: #c13944;
}

.reply-menu {
  margin-left: auto;
}

.menu-btn {
  background: transparent;
  border: none;
  color: #888;
  cursor: pointer;
  font-size: 20px;
  padding: 0 8px;
}

.menu-btn:hover {
  color: white;
}

.nested-replies {
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.loading-indicator {
  text-align: center;
  padding: 20px;
  color: #888;
  font-size: 14px;
}

.reply-input-section {
  border-top: 1px solid #333;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background-color: #0a0a0a;
}

.selected-gif-preview {
  position: relative;
  display: inline-block;
  max-width: 150px;
}

.selected-gif-preview img {
  width: 100%;
  border-radius: 8px;
}

.cancel-gif-btn {
  position: absolute;
  top: -8px;
  right: -8px;
  background: #ff4444;
  border: none;
  border-radius: 50%;
  width: 24px;
  height: 24px;
  color: white;
  cursor: pointer;
  font-size: 18px;
  line-height: 1;
}

.replying-to-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #888;
}

.replying-to-indicator button {
  background: transparent;
  border: none;
  color: #4a9eff;
  cursor: pointer;
  font-size: 13px;
}

.input-container {
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.reply-input {
  flex: 1;
  background: #1a1a1a;
  border: 1px solid #333;
  border-radius: 20px;
  padding: 12px 16px;
  color: white;
  font-size: 14px;
  resize: none;
  min-height: 44px;
  max-height: 120px;
  font-family: inherit;
}

.reply-input:focus {
  outline: none;
  border-color: #555;
}

.reply-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.input-actions {
  display: flex;
  gap: 8px;
}

.gif-btn,
.submit-btn {
  padding: 10px 20px;
  border-radius: 20px;
  border: none;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.2s;
}

.gif-btn {
  background: #2a2a2a;
  color: white;
}

.gif-btn:hover {
  background: #3a3a3a;
}

.submit-btn {
  background: #0095f6;
  color: white;
}

.submit-btn:hover:not(:disabled) {
  background: #1aa1f7;
}

.submit-btn:disabled {
  background: #2a4a5a;
  cursor: not-allowed;
  opacity: 0.5;
}

.gif-picker {
  background: #1a1a1a;
  border: 1px solid #333;
  border-radius: 12px;
  padding: 12px;
  max-height: 300px;
  overflow-y: auto;
}

.gif-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.gif-grid img {
  width: 100%;
  height: 100px;
  object-fit: cover;
  border-radius: 8px;
  cursor: pointer;
  transition: transform 0.2s;
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

.gif-grid img:hover {
  transform: scale(1.05);
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

.user-option {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-avatar-small {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
}
</style>