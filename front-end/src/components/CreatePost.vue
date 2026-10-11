<script setup lang="ts">
import { useCreatePost } from "@/composables/useOverlay";
import { onMounted, ref, computed } from "vue";
import { useEventBus } from "@/composables/useEventBus";
import { storyService } from "@/services/storyService";

const { showCreatePost, close } = useCreatePost();
const { emit } = useEventBus();
const createPost = ref(1);
const selectedFiles = ref<File[]>([]);
const currentSlide = ref(0);
const loading = ref(false);
const error = ref("");
const caption = ref("");
const userId = ref("");
const isCommentOff = ref(false);
const postAsStory = ref(false);

const showStoryToggle = computed(() => {
  return selectedFiles.value.length === 1;
});

const getPostType = () => {
  if (selectedFiles.value.length === 1) {
    const file = selectedFiles.value[0];
    if (postAsStory.value) {
      return 'story';
    }
    if (file.type.startsWith('video/')) {
      return 'reels';
    }
  }
  return 'post';
};


const wordCount = computed(() => {
  if (!caption.value.trim()) return 0;
  return caption.value.trim().split(/\s+/).length;
});

const wordProgress = computed(() => {
  return Math.min((wordCount.value / 1000) * 100, 100);
});

const circumference = 2 * Math.PI * 18;

const strokeDashoffset = computed(() => {
  return circumference - (wordProgress.value / 100) * circumference;
});
onMounted(() => {
  fetchUser();
});


const fetchUser = async () => {
  const token = localStorage.getItem("token");

  if (!token) {
    return;
  }

  try {
    const response = await fetch("https://localhost/api/v1/users/get-current-user", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      const data = await response.json();
      const user = data.user;

      if (user && user.id) {
        userId.value = user.id;

      }
    }
  } catch (error) {
    console.error("Failed to fetch user:", error);
  }
};

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files.length > 0) {
    const files = Array.from(target.files);
    

    if (files.length > 10) {
      error.value = "Maximum 10 files allowed";
      return;
    }


    const allowedTypes = ["image/jpeg", "image/jpg", "image/png", "image/gif", "video/mp4", "video/quicktime"];
    for (const file of files) {
      if (!allowedTypes.includes(file.type)) {
        error.value = "Please select valid image or video files";
        selectedFiles.value = [];
        return;
      }

      if (file.size > 10 * 1024 * 1024) {
        error.value = `File ${file.name} exceeds 10MB`;
        selectedFiles.value = [];
        return;
      }
    }

    selectedFiles.value = files;
    currentSlide.value = 0;
    error.value = "";
    
    setTimeout(() => {
      createPost.value = 2;
    }, 500);
  }
};

const removeFile = (index: number) => {
  selectedFiles.value = selectedFiles.value.filter((_, i) => i !== index);
  if (selectedFiles.value.length === 0) {
    createPost.value = 1;
  } else if (currentSlide.value >= selectedFiles.value.length) {
    currentSlide.value = selectedFiles.value.length - 1;
  }
};

const nextSlide = () => {
  if (currentSlide.value < selectedFiles.value.length - 1) {
    currentSlide.value++;
  }
};

const prevSlide = () => {
  if (currentSlide.value > 0) {
    currentSlide.value--;
  }
};

const goToSlide = (index: number) => {
  currentSlide.value = index;
};

const getFilePreview = (file: File) => {
  return window.URL.createObjectURL(file);
};

const handleClose = () => {
  createPost.value = 1;
  selectedFiles.value = [];
  caption.value = "";
  currentSlide.value = 0;
  error.value = "";
  postAsStory.value = false;
  isCommentOff.value = false;
  close();
};

const handleStorySubmit = async () => {
  try {
    const mediaFile = selectedFiles.value[0];
    await storyService.createStory(mediaFile);

    selectedFiles.value = [];
    caption.value = "";
    currentSlide.value = 0;
    createPost.value = 1;
    postAsStory.value = false;

    emit("story-created");
    handleClose();
  } catch (err) {
    error.value = err instanceof Error ? err.message : "Failed to create story";
    console.error("Error creating story:", err);
  }
};

const handleSubmit = async () => {
  if (selectedFiles.value.length === 0) {
    error.value = "Please select at least one file";
    return;
  }

  loading.value = true;
  error.value = "";

  try {
    if (!userId.value) {
      error.value = "User not logged in. Please refresh and try again.";
      loading.value = false;
      return;
    }

    if (selectedFiles.value.length === 1 && postAsStory.value) {
      await handleStorySubmit();
      return;
    }

    const formData = new FormData();
    formData.append("user_id", userId.value);
    formData.append("caption", caption.value);
    formData.append("is_comment_off", isCommentOff.value ? "true" : "false");
    
    selectedFiles.value.forEach((file) => {
      formData.append("media", file);
    });
    
    const token = localStorage.getItem("token");

    const response = await fetch("https://localhost/api/v1/posts/create-post", {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${token}`
      },
      body: formData
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || "Failed to create post");
    }

    const data = await response.json();



    selectedFiles.value = [];
    caption.value = "";
    currentSlide.value = 0;
    createPost.value = 1;

    try {
      await fetch(`https://localhost/api/v1/users/add-post-count/${userId.value}`, {
        method: "POST",
        headers: {
          "Authorization": `Bearer ${token}`
        }
      });
    } catch (err) {
      console.error("Failed to update post count:", err);
    }

    emit("post-created", { userId: userId.value });

    handleClose();

  } catch (err) {
    error.value = err instanceof Error ? err.message : "Failed to create post";
    console.error("Error creating post:", err);
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <Teleport to="body">
    <div
      v-if="showCreatePost"
      class="create-post-background"
      @click.self="handleClose"
    >
      <div class="create-post-container">
        <div
          v-if="createPost === 1"
          class="create-post"
        >
          <h2>Create new post</h2>
          <button @click="handleClose">
            X
          </button>
          <div
            class="form-group"
            style="margin-top: 100px;"
          >
            <p
              v-if="error"
              class="error-message"
            >
              {{ error }}
            </p>
            <img
              src="../components/icons/image-light.png"
              width="300"
              height="300"
              style="margin-bottom: 20px;"
            />
            <label
              for="media_files"
              class="file-label"
            >
              {{ selectedFiles.length > 0 ? `${selectedFiles.length} file(s) selected` : 'Choose File' }}
            </label>
            <input
              id="media_files"
              type="file"
              accept="image/jpeg, image/jpg, image/png, image/gif, video/mp4, video/quicktime"
              :disabled="loading"
              class="file-input"
              multiple
              @change="handleFileSelect"
            />
          </div>
        </div>
        <div
          v-if="createPost === 2"
          class="create-post"
        >
          <h2>Preview</h2>
          <button @click="handleClose">
            X
          </button>
          <div class="preview-container">
            <div class="slider-container">
              <div class="slider-wrapper">
                <div 
                  v-for="(file, index) in selectedFiles" 
                  :key="index" 
                  class="slide"
                  :class="{ active: currentSlide === index }"
                >
                  <img 
                    v-if="file.type.startsWith('image/')" 
                    :src="getFilePreview(file)" 
                    width="300"
                    height="300"
                    :alt="file.name" 
                  />
                  <video 
                    v-else-if="file.type.startsWith('video/')" 
                    :src="getFilePreview(file)" 
                    controls
                  />
                  <button
                    class="remove-btn"
                    @click="removeFile(index)"
                  >
                    ×
                  </button>
                </div>
              </div>

              <button 
                v-if="selectedFiles.length > 1 && currentSlide > 0"
                class="nav-btn prev-btn" 
                @click="prevSlide"
              >
                ‹
              </button>
              <button 
                v-if="selectedFiles.length > 1 && currentSlide < selectedFiles.length - 1"
                class="nav-btn next-btn" 
                @click="nextSlide"
              >
                ›
              </button>

              <div
                v-if="selectedFiles.length > 1"
                class="slide-indicators"
              >
                <span 
                  v-for="(_, index) in selectedFiles" 
                  :key="index"
                  class="indicator"
                  :class="{ active: currentSlide === index }"
                  @click="goToSlide(index)"
                />
              </div>
              <div class="file-counter">
                {{ currentSlide + 1 }} / {{ selectedFiles.length }}
              </div>
            </div>
            <button
              class="submit-btn"
              @click="createPost = 3"
            >
              Next
            </button>
          </div>
        </div>
        <div
          v-if="createPost === 3"
          class="create-post"
        >
          <h2>Create Post</h2>
          <button @click="handleClose">
            X
          </button>
          <div class="post-form">
            <p
              v-if="error"
              class="error-message"
            >
              {{ error }}
            </p>
            <div class="caption-container">
              <textarea 
                v-model="caption" 
                placeholder="Write a caption" 
                class="caption-input"
                :disabled="loading"
              />
              <div class="word-counter">
                <svg
                  width="45"
                  height="45"
                  class="progress-ring"
                >
                  <circle
                    class="progress-ring-circle-bg"
                    stroke="rgb(60, 60, 60)"
                    stroke-width="3"
                    fill="transparent"
                    r="12"
                    cx="22.5"
                    cy="22.5"
                  />
                  <circle
                    class="progress-ring-circle"
                    :stroke="wordCount > 1000 ? '#ff4444' : '#a800f6'"
                    stroke-width="3"
                    fill="transparent"
                    r="18"
                    cx="22.5"
                    cy="22.5"
                    :style="{
                      strokeDasharray: circumference,
                      strokeDashoffset: strokeDashoffset
                    }"
                  />
                </svg>
                <span 
                  class="word-count-text"
                  :class="{ 'over-limit': wordCount > 1000 }"
                >
                  {{ wordCount }}
                </span>
              </div>
            </div>
            <div class="form">
              <input
                id="turn-off-comments"
                v-model="isCommentOff"
                type="checkbox"
                :disabled="loading || postAsStory"
              />
              <label for="turn-off-comments">Turn Off Comments</label>
            </div>
            <div
              v-if="showStoryToggle"
              class="form"
            >
              <input
                id="post-as-story"
                v-model="postAsStory"
                type="checkbox"
                :disabled="loading"
              />
              <label for="post-as-story">Post as Story</label>
            </div>
            <button 
              class="submit-btn" 
              :disabled="loading"
              @click="handleSubmit"
            >
              {{ loading ? 'Creating...' : postAsStory ? 'Create Story' : getPostType() === 'reels' ? 'Create Reels' : 'Create Post' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
  

<style scoped>
.create-post-background {
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
}

.create-post-container {
  background: white;
  border-radius: 20px;
  width: 35%;
  height: 70%;
  overflow-y: auto;
  background-color: rgb(35, 35, 35);
}

button {
  position: fixed;
  background-color: transparent;
  border: none;
  font-size: 28px;
  cursor: pointer;
  font-weight: 700;
  color: white;
  padding: 10px;
  top: 0;
  right: 0;
}

h2 {
  display: flex;
  justify-content: center;
  width: 100%;
  background-color: rgb(22, 22, 22);
  color: white;
  padding: 5px;
  font-size: 16px;
  font-weight: 600;
}

.form-group {
  justify-content: center;
  align-items: center;
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
}

.form {
  width: 100%;
  display: flex;
  justify-content: flex-start;
  align-items: center;
  margin-top: 10px;
  margin-bottom: 10px;
}

.form input[type="checkbox"] {
  margin-right: 10px;
}

.form label {
  color: white;
  font-size: 14px;
  cursor: pointer;
}

.error-message {
  color: #ff4444;
  font-size: 14px;
  margin-bottom: 10px;
}

.file-label {
  padding: 15px 30px;
  background-color: #a800f6;
  color: white;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: background-color 0.2s;
}

.file-label:hover {
  background-color: #8900d8;
}

.file-input {
  display: none;
}

.preview-container {
  padding: 20px;
  width: 100%;
  height: 100%;
  color: white;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.slider-container {
  position: relative;
  width: 100%;
  height: 400px;
  margin: 0 auto 20px;
  aspect-ratio: 1 / 1;
  overflow: hidden;
  border-radius: 8px;
  background-color: rgb(22, 22, 22);
}

.slider-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
}

.slide {
  position: absolute;
  width: 100%;
  height: 100%;
  opacity: 0;
  transition: opacity 0.3s ease-in-out;
  display: flex;
  align-items: center;
  justify-content: center;
}

.slide.active {
  opacity: 1;
  z-index: 1;
}

.slide img,
.slide video {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  display: block;
}

.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background-color: rgba(255, 255, 255, 0.8);
  color: rgb(22, 22, 22);
  border: none;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  font-size: 24px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  transition: background-color 0.2s;
  padding: 0;
}

.nav-btn:hover {
  background-color: rgba(255, 255, 255, 1);
}

.prev-btn {
  left: 10px;
}

.next-btn {
  right: 10px;
}

.slide-indicators {
  position: absolute;
  bottom: 15px;
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
  background-color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  transition: background-color 0.2s;
}

.indicator.active {
  background-color: rgba(255, 255, 255, 1);
}

.indicator:hover {
  background-color: rgba(255, 255, 255, 0.8);
}

.file-counter {
  position: absolute;
  top: 15px;
  right: 15px;
  background-color: rgba(0, 0, 0, 0.7);
  color: white;
  padding: 5px 12px;
  border-radius: 15px;
  font-size: 13px;
  z-index: 10;
}container {
  width: 100%;
  position: relative;
}

.caption-input {
  width: 100%;
  min-height: 100px;
  padding: 15px;
  padding-right: 60px;
  border-radius: 8px;
  border: 1px solid rgb(60, 60, 60);
  background-color: rgb(22, 22, 22);
  color: white;
  font-size: 14px;
  resize: vertical;
  font-family: inherit;
}

.caption-input:focus {
  outline: none;
  border-color: #b400f6;
}

.word-counter {
  position: absolute;
  bottom: 10px;
  right: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.progress-ring {
  transform: rotate(-90deg);
}

.progress-ring-circle {
  transition: stroke-dashoffset 0.3s ease, stroke 0.3s ease;
  stroke-linecap: round;
}

.word-count-text {
  position: absolute;
  font-size: 11px;
  font-weight: 600;
  color: white;
  pointer-events: none;
}

.word-count-text.over-limit {
  color: #ff4444;
}

.remove-btn:hover {
  background-color: rgba(255, 0, 0, 0.8);
}

.post-form {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 90%;
}

.caption-container {
  width: 100%;
  position: relative;
}

.caption-input {
  width: 100%;
  min-height: 100px;
  padding: 15px;
  padding-right: 60px;
  border-radius: 8px;
  border: 1px solid rgb(60, 60, 60);
  background-color: rgb(22, 22, 22);
  color: white;
  font-size: 14px;
  resize: vertical;
  font-family: inherit;
  box-sizing: border-box;
}

.caption-input:focus {
  outline: none;
  border-color: #b400f6;
}

.submit-btn {
  position: static;
  margin-top: 15px;
  width: 60%;
  padding: 10px;
  background-color: #9800f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
}

.submit-btn:hover {
  background-color: #8200d8;
}

.create-post{
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

</style>
