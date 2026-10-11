<script setup lang="ts">
import SideBar from "@/components/SideBar.vue";
import { onMounted, ref } from "vue";
import { usePhoto } from "@/composables/usePhoto";
import userLight from "@/components/icons/user-light.png";

const activeTab = ref("editProfile");
const selectedFile = ref<File | null>(null);
const uploadedPhotoUrl = ref("");
const loading = ref(false);
const { loadPhotoUrl, uploadPhoto } = usePhoto();
const profilePhotoUrl = ref<string>("");
const error = ref("");
const success = ref("");

const form = ref({
  username: "",
  name: "",
  bio: "",
  gender: "",
  profile_picture: "",
});

const formErrors = ref({
  username: "",
  name: "",
  bio: "",
  gender: "",
});


const notifications = ref({
  push_notification: false,
  email_notification: false,
});


const accountPrivacy = ref({
  is_private: false,
});

interface UserInfo {
  ID: string;
  id: string;
  username: string;
  name: string;
  full_name?: string;
  profile_picture: string;
}

const closeFriends = ref<UserInfo[]>([]);
const closeFriendSearchQuery = ref("");
const closeFriendSearchResults = ref<UserInfo[]>([]);
const loadingCloseFriendSearch = ref(false);


const blockedUsers = ref<UserInfo[]>([]);

const hiddenStoryViewers = ref<UserInfo[]>([]);
const hideStorySearchQuery = ref("");
const hideStorySearchResults = ref<UserInfo[]>([]);
const loadingHideStorySearch = ref(false);


const verificationForm = ref({
  national_id: "",
  reason: "",
  face_picture: "",
});

const verificationFile = ref<File | null>(null);
const verificationErrors = ref({
  national_id: "",
  reason: "",
  face_picture: "",
});

onMounted(() => {
  fetchUser();
  fetchNotificationSettings();
  fetchPrivacySettings();
  fetchCloseFriends();
  fetchBlockedUsers();
  fetchHiddenStoryViewers();
});



const validateName = (name: string) => {
  return /^[a-zA-Z\s]+$/.test(name);
};

const validateField = (field: string) => {
  switch (field) {
    case "username":
      if (form.value.username.length < 4) {
        formErrors.value.username = "Username must be at least 4 characters";
      } else if (!/^[a-zA-Z0-9_]+$/.test(form.value.username)) {
        formErrors.value.username = "Username can only contain letters, numbers, and underscores";
      } else {
        formErrors.value.username = "";
      }
      break;
    case "name":
      if (!form.value.name) {
        formErrors.value.name = "Name is required";
      } else if (!validateName(form.value.name)) {
        formErrors.value.name = "Name must NOT contain numbers and symbols";
      } else {
        formErrors.value.name = "";
      }
      break;
    case "gender":
      if (form.value.gender !== "Male" && form.value.gender !== "Female") {
        formErrors.value.gender = "Gender must be either Male or Female";
      } else {
        formErrors.value.gender = "";
      }
      break;
    case "bio":
      if (form.value.bio.length > 500) {
        formErrors.value.bio = "Bio must be less than 500 characters";
      } else {
        formErrors.value.bio = "";
      }
      break;
  }
};

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files[0]) {
    selectedFile.value = target.files[0];

    const allowedTypes = ["image/jpeg", "image/jpg", "image/png"];
    if (!allowedTypes.includes(target.files[0].type)) {
      error.value = "Please select a valid image file (JPEG or PNG)";
      selectedFile.value = null;
      return;
    }

    if (target.files[0].size > 5 * 1024 * 1024) {
      error.value = "File size must be less than 5MB";
      selectedFile.value = null;
      return;
    }

    error.value = "";
  }
};

const uploadPhotoHere = async () => {
  if (!selectedFile.value) {
    error.value = "Please select a photo to upload";
    return false;
  }

  const url = await uploadPhoto(selectedFile.value);
  if (!url) {
    error.value = "Photo upload failed. Please try again.";
    return false;
  }

  uploadedPhotoUrl.value = url;
  form.value.profile_picture = url;
  return true;
};

const fetchUser = async () => {
  const token = localStorage.getItem("token");

  if (!token) return "";

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

      if (user.profile_picture) {
        profilePhotoUrl.value = loadPhotoUrl(user.profile_picture);
      }
      
      form.value.username = user.username || "";
      form.value.name = user.name || "";
      form.value.bio = user.bio || "";
      form.value.gender = user.gender || "";
    }
  } catch {

  }
};

const fetchNotificationSettings = async () => {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch("https://localhost/api/v1/users/get-notification-settings", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      const data = await response.json();
      notifications.value = data;
    }
  } catch {

  }
};

const fetchPrivacySettings = async () => {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch("https://localhost/api/v1/users/get-privacy-settings", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      const data = await response.json();
      accountPrivacy.value.is_private = data.is_private;
    }
  } catch {

  }
};

const fetchCloseFriends = async () => {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch("https://localhost/api/v1/social/get-close-friends", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      const data = await response.json();
      closeFriends.value = data.close_friends || [];
    }
  } catch {

  }
};

const fetchBlockedUsers = async () => {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch("https://localhost/api/v1/social/get-blocked-users", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      const data = await response.json();
      blockedUsers.value = data.blocked_users || [];
    }
  } catch {

  }
};

const fetchHiddenStoryViewers = async () => {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch("https://localhost/api/v1/social/get-hidden-story-viewers", {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      const data = await response.json();
      hiddenStoryViewers.value = data.hidden_viewers || [];
    }
  } catch {

  }
};


const submit = async () => {
  loading.value = true;
  error.value = "";
  success.value = "";

  validateField("username");
  validateField("name");
  validateField("gender");
  validateField("bio");

  if (Object.values(formErrors.value).some(err => err !== "")) {
    error.value = "Please fix all validation errors";
    loading.value = false;
    return;
  }

  let photoUploaded = true;
  if (selectedFile.value) {
    photoUploaded = await uploadPhotoHere();
  }

  if (!photoUploaded) {
    loading.value = false;
    return;
  }

  const token = localStorage.getItem("token");
  if (!token) {
    error.value = "User not authenticated.";
    loading.value = false;
    return;
  }

  try {
    const response = await fetch("https://localhost/api/v1/users/edit-profile", {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(form.value),
    });

    if (response.ok) {
      success.value = "Profile updated successfully!";
      setTimeout(() => {
        success.value = "";
      }, 3000);
    } else {
      const data = await response.json();
      error.value = data.message || "Failed to update profile.";
    }
  } catch {
    error.value = "An error occurred. Please try again.";
  } finally {
    loading.value = false;
  }
};

const saveNotificationSettings = async () => {
  loading.value = true;
  error.value = "";
  success.value = "";

  const token = localStorage.getItem("token");
  if (!token) {
    error.value = "User not authenticated.";
    loading.value = false;
    return;
  }

  try {
    const response = await fetch("https://localhost/api/v1/users/update-notification-settings", {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(notifications.value),
    });

    if (response.ok) {
      success.value = "Notification settings updated successfully!";
      setTimeout(() => {
        success.value = "";
      }, 3000);
    } else {
      const data = await response.json();
      error.value = data.message || "Failed to update settings.";
    }
  } catch {
    error.value = "An error occurred. Please try again.";
  } finally {
    loading.value = false;
  }
};

const savePrivacySettings = async () => {
  loading.value = true;
  error.value = "";
  success.value = "";

  const token = localStorage.getItem("token");
  if (!token) {
    error.value = "User not authenticated.";
    loading.value = false;
    return;
  }

  try {
    const response = await fetch("https://localhost/api/v1/users/update-privacy-settings", {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(accountPrivacy.value),
    });

    if (response.ok) {
      success.value = "Privacy settings updated successfully!";
      setTimeout(() => {
        success.value = "";
      }, 3000);
    } else {
      const data = await response.json();
      error.value = data.message || "Failed to update settings.";
    }
  } catch {
    error.value = "An error occurred. Please try again.";
  } finally {
    loading.value = false;
  }
};

const removeCloseFriend = async (userID: string) => {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch(`https://localhost/api/v1/social/remove-close-friend/${userID}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      closeFriends.value = closeFriends.value.filter(friend => friend.ID !== userID);
      success.value = "Removed from close friends!";
      setTimeout(() => {
        success.value = "";
      }, 3000);
    }
  } catch {

  }
};

const searchUsersForCloseFriend = async () => {
  const query = closeFriendSearchQuery.value.trim();
  if (!query || query.length < 2) {
    closeFriendSearchResults.value = [];
    return;
  }

  loadingCloseFriendSearch.value = true;
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch(
      `https://localhost/api/v1/users/search?q=${encodeURIComponent(query)}`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );

    if (response.ok) {
      const data = await response.json();
      closeFriendSearchResults.value = (data.users || []).filter(
        (user: UserInfo) => !closeFriends.value.some(friend => friend.ID === user.id)
      );
    }
  } catch {

  } finally {
    loadingCloseFriendSearch.value = false;
  }
};

const addCloseFriend = async (userID: string) => {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch(`https://localhost/api/v1/social/add-close-friend/${userID}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      await fetchCloseFriends();
      closeFriendSearchQuery.value = "";
      closeFriendSearchResults.value = [];
      success.value = "Added to close friends!";
      setTimeout(() => {
        success.value = "";
      }, 3000);
    }
  } catch {
    error.value = "Failed to add close friend. Please try again.";
  }
};

const unblockUser = async (userID: string) => {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch(`https://localhost/api/v1/social/unblock-user/${userID}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      blockedUsers.value = blockedUsers.value.filter(user => user.ID !== userID);
      success.value = "User unblocked successfully!";
      setTimeout(() => {
        success.value = "";
      }, 3000);
    }
  } catch {

  }
};

const unhideStoryViewer = async (userID: string) => {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch(`https://localhost/api/v1/social/unhide-story-viewer/${userID}`, {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      hiddenStoryViewers.value = hiddenStoryViewers.value.filter(viewer => viewer.ID !== userID);
      success.value = "Story viewer unhidden successfully!";
      setTimeout(() => {
        success.value = "";
      }, 3000);
    }
  } catch {

  }
};

const searchUsersForHiding = async () => {
  const query = hideStorySearchQuery.value.trim();
  if (!query || query.length < 2) {
    hideStorySearchResults.value = [];
    return;
  }

  loadingHideStorySearch.value = true;
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch(
      `https://localhost/api/v1/users/search?q=${encodeURIComponent(query)}`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );

    if (response.ok) {
      const data = await response.json();
      hideStorySearchResults.value = (data.users || []).filter(
        (user: UserInfo) => !hiddenStoryViewers.value.some(viewer => viewer.ID === user.id)
      );
    }
  } catch {

  } finally {
    loadingHideStorySearch.value = false;
  }
};

const hideStoryFromUser = async (userID: string) => {
  const token = localStorage.getItem("token");
  if (!token) return;

  try {
    const response = await fetch(`https://localhost/api/v1/social/hide-story-viewer/${userID}`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
    });

    if (response.ok) {
      await fetchHiddenStoryViewers();
      hideStorySearchQuery.value = "";
      hideStorySearchResults.value = [];
      success.value = "User hidden from viewing your stories successfully!";
      setTimeout(() => {
        success.value = "";
      }, 3000);
    }
  } catch {
    error.value = "Failed to hide user. Please try again.";
  }
};

const handleVerificationFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files[0]) {
    verificationFile.value = target.files[0];

    const allowedTypes = ["image/jpeg", "image/jpg", "image/png"];
    if (!allowedTypes.includes(target.files[0].type)) {
      verificationErrors.value.face_picture = "Please select a valid image file (JPEG or PNG)";
      verificationFile.value = null;
      return;
    }

    if (target.files[0].size > 5 * 1024 * 1024) {
      verificationErrors.value.face_picture = "File size must be less than 5MB";
      verificationFile.value = null;
      return;
    }

    verificationErrors.value.face_picture = "";
  }
};

const validateVerificationForm = () => {
  let isValid = true;

  if (!verificationForm.value.national_id || verificationForm.value.national_id.length < 8) {
    verificationErrors.value.national_id = "National ID must be at least 8 characters";
    isValid = false;
  } else {
    verificationErrors.value.national_id = "";
  }

  if (!verificationForm.value.reason || verificationForm.value.reason.length < 20) {
    verificationErrors.value.reason = "Reason must be at least 20 characters";
    isValid = false;
  } else {
    verificationErrors.value.reason = "";
  }

  if (!verificationFile.value) {
    verificationErrors.value.face_picture = "Please upload your face picture";
    isValid = false;
  } else {
    verificationErrors.value.face_picture = "";
  }

  return isValid;
};

const submitVerificationRequest = async () => {
  loading.value = true;
  error.value = "";
  success.value = "";

  if (!validateVerificationForm()) {
    error.value = "Please fix all validation errors";
    loading.value = false;
    return;
  }

  if (verificationFile.value) {
    const url = await uploadPhoto(verificationFile.value);
    if (!url) {
      error.value = "Face picture upload failed. Please try again.";
      loading.value = false;
      return;
    }
    verificationForm.value.face_picture = url;
  }

  const token = localStorage.getItem("token");
  if (!token) {
    error.value = "User not authenticated.";
    loading.value = false;
    return;
  }

  try {
    const response = await fetch("https://localhost/api/v1/users/request-verification", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify(verificationForm.value),
    });

    if (response.ok) {
      success.value = "Verification request submitted successfully!";
      verificationForm.value = {
        national_id: "",
        reason: "",
        face_picture: "",
      };
      verificationFile.value = null;
    } else {
      const data = await response.json();
      error.value = data.message || "Failed to submit verification request.";
    }
  } catch {
    error.value = "An error occurred. Please try again.";
  } finally {
    loading.value = false;
  }
};

const pfpError = (e: Event) => {
  const target = e.target as HTMLImageElement;
  target.src = userLight;
};
</script>


<template>
  <div class="container">
    <SideBar />
    <div class="settings-page">
      <h1 class="page-title">
        Settings
      </h1>

      <div class="tabs-nav">
        <button
          :class="{ active: activeTab === 'editProfile' }"
          @click="activeTab = 'editProfile'"
        >
          Edit Profile
        </button>
        <button
          :class="{ active: activeTab === 'notifications' }"
          @click="activeTab = 'notifications'"
        >
          Notifications
        </button>
        <button
          :class="{ active: activeTab === 'privacy' }"
          @click="activeTab = 'privacy'"
        >
          Account Privacy
        </button>
        <button
          :class="{ active: activeTab === 'closeFriends' }"
          @click="activeTab = 'closeFriends'"
        >
          Close Friends
        </button>
        <button
          :class="{ active: activeTab === 'blocked' }"
          @click="activeTab = 'blocked'"
        >
          Blocked
        </button>
        <button
          :class="{ active: activeTab === 'hideStory' }"
          @click="activeTab = 'hideStory'"
        >
          Hide Story
        </button>
        <button
          :class="{ active: activeTab === 'verification' }"
          @click="activeTab = 'verification'"
        >
          Request Verified
        </button>
      </div>

 
      <div
        v-if="error"
        class="error-message"
      >
        {{ error }}
      </div>
      <div
        v-if="success"
        class="success-message"
      >
        {{ success }}
      </div>

      
      <div class="tab-content">
        <div
          v-if="activeTab === 'editProfile'"
          class="edit-profile"
        >
          <h2>Edit Profile</h2>
          <div class="edit-profile-pic">
            <img
              v-if="profilePhotoUrl"
              :src="profilePhotoUrl"
              width="150"
              height="150"
              style="border-radius: 50%; object-fit: cover; border: 2px solid white"
              @error="pfpError"
            />
            <img
              v-else
              :src="userLight"
              width="150"
              height="150"
              style="border-radius: 50%; object-fit: cover; border: 2px solid white"
            />
            <div class="form-group">
              <label
                for="profile_picture"
                class="file-label"
              >
                {{ selectedFile ? selectedFile.name : 'Choose Profile Picture' }}
              </label>
              <input
                id="profile_picture"
                type="file"
                accept="image/*"
                :disabled="loading"
                class="file-input"
                @change="handleFileSelect"
              />
            </div>
          </div>
          <div class="form-group">
            <label
              for="username"
              style="color: white; font-weight: 500"
            >Username</label>
            <br />
            <input
              id="username"
              v-model="form.username"
              type="text"
              placeholder="Enter your username"
              :disabled="loading"
              style="background-color: gray; border: none; color: white; padding: 10px; border-radius: 5px; width: 100%; max-width: 300px"
              @blur="validateField('username')"
            />
            <span
              v-if="formErrors.username"
              class="field-error"
            >{{ formErrors.username }}</span>
          </div>
          <div class="form-group">
            <label
              for="name"
              style="color: white; font-weight: 500"
            >Name</label>
            <br />
            <input
              id="name"
              v-model="form.name"
              type="text"
              placeholder="Enter your name"
              :disabled="loading"
              style="background-color: gray; border: none; color: white; padding: 10px; border-radius: 5px; width: 100%; max-width: 300px"
              @blur="validateField('name')"
            />
            <span
              v-if="formErrors.name"
              class="field-error"
            >{{ formErrors.name }}</span>
          </div>
          <div class="form-group">
            <label
              for="bio"
              style="color: white; font-weight: 500"
            >Bio</label>
            <br />
            <textarea
              id="bio"
              v-model="form.bio"
              rows="4"
              cols="50"
              :disabled="loading"
              style="background-color: gray; border: none; color: white"
              @blur="validateField('bio')"
            />
            <span
              v-if="formErrors.bio"
              class="field-error"
            >{{ formErrors.bio }}</span>
          </div>
          <label
            for="gender"
            style="color: white; font-weight: 500"
          >Gender</label>
          <div class="form-group">
            <select
              id="gender"
              v-model="form.gender"
              required
              :disabled="loading"
              style="font-size: 14px; background-color: gray; border: none; color: white"
              @blur="validateField('gender')"
            >
              <option
                value=""
                disabled
              >
                Select gender
              </option>
              <option value="Male">
                Male
              </option>
              <option value="Female">
                Female
              </option>
            </select>
            <span
              v-if="formErrors.gender"
              class="field-error"
            >{{ formErrors.gender }}</span>
          </div>
          <div class="form-group">
            <button
              :disabled="loading"
              @click="submit"
            >
              {{ loading ? 'Updating...' : 'Update Profile' }}
            </button>
          </div>
        </div>

        
        <div
          v-if="activeTab === 'notifications'"
          class="notifications-settings"
        >
          <h2>Notification Settings</h2>
          <div class="setting-item">
            <label class="switch">
              <input
                v-model="notifications.push_notification"
                type="checkbox"
                :disabled="loading"
              />
              <span class="slider" />
            </label>
            <span class="setting-label">Push Notifications</span>
          </div>
          <div class="setting-item">
            <label class="switch">
              <input
                v-model="notifications.email_notification"
                type="checkbox"
                :disabled="loading"
              />
              <span class="slider" />
            </label>
            <span class="setting-label">Email Notifications</span>
          </div>
          <div class="form-group">
            <button
              :disabled="loading"
              @click="saveNotificationSettings"
            >
              {{ loading ? 'Saving...' : 'Save Settings' }}
            </button>
          </div>
        </div>

        
        <div
          v-if="activeTab === 'privacy'"
          class="privacy-settings"
        >
          <h2>Account Privacy</h2>
          <div class="setting-item">
            <label class="switch">
              <input
                v-model="accountPrivacy.is_private"
                type="checkbox"
                :disabled="loading"
              />
              <span class="slider" />
            </label>
            <span class="setting-label">Private Account</span>
          </div>
          <p class="setting-description">
            When your account is private, only people you approve can see your photos and videos. Your existing followers won't be affected.
          </p>
          <div class="form-group">
            <button
              :disabled="loading"
              @click="savePrivacySettings"
            >
              {{ loading ? 'Saving...' : 'Save Settings' }}
            </button>
          </div>
        </div>

        
        <div
          v-if="activeTab === 'closeFriends'"
          class="close-friends-list"
        >
          <h2>Close Friends</h2>
          <p class="tab-description">
            Manage your close friends list. You can add or remove users at any time.
          </p>
          
          <div class="search-section">
            <h3>Add Close Friend</h3>
            <input
              v-model="closeFriendSearchQuery"
              type="text"
              placeholder="Search users to add as close friend..."
              class="search-input"
              @input="searchUsersForCloseFriend"
            />
            
            <div
              v-if="loadingCloseFriendSearch"
              class="loading-text"
            >
              Searching...
            </div>
            
            <div
              v-if="closeFriendSearchResults.length > 0"
              class="search-results"
            >
              <div
                v-for="user in closeFriendSearchResults"
                :key="user.id"
                class="user-item"
              >
                <img
                  :src="user.profile_picture ? loadPhotoUrl(user.profile_picture) : userLight"
                  class="user-avatar"
                  @error="pfpError"
                />
                <div class="user-info">
                  <span class="username">{{ user.username }}</span>
                  <span class="name">{{ user.full_name }}</span>
                </div>
                <button
                  class="add-btn"
                  @click="addCloseFriend(user.id)"
                >
                  Add
                </button>
              </div>
            </div>
          </div>

          <h3 style="margin-top: 30px;">
            Your Close Friends
          </h3>
          <div
            v-if="closeFriends.length === 0"
            class="empty-state"
          >
            No close friends yet.
          </div>
          <div
            v-for="friend in closeFriends"
            :key="friend.ID"
            class="user-item"
          >
            <img
              :src="friend.profile_picture ? loadPhotoUrl(friend.profile_picture) : userLight"
              class="user-avatar"
              @error="pfpError"
            />
            <div class="user-info">
              <span class="username">{{ friend.username }}</span>
              <span class="name">{{ friend.name }}</span>
            </div>
            <button
              class="remove-btn"
              @click="removeCloseFriend(friend.ID)"
            >
              Remove
            </button>
          </div>
        </div>

        
        <div
          v-if="activeTab === 'blocked'"
          class="blocked-users-list"
        >
          <h2>Blocked Accounts</h2>
          <p class="tab-description">
            Manage your blocked accounts. You can unblock users at any time.
          </p>
          <div
            v-if="blockedUsers.length === 0"
            class="empty-state"
          >
            No blocked users.
          </div>
          <div
            v-for="user in blockedUsers"
            :key="user.ID"
            class="user-item"
          >
            <img
              :src="user.profile_picture ? loadPhotoUrl(user.profile_picture) : userLight"
              class="user-avatar"
              @error="pfpError"
            />
            <div class="user-info">
              <span class="username">{{ user.username }}</span>
              <span class="name">{{ user.name }}</span>
            </div>
            <button
              class="unblock-btn"
              @click="unblockUser(user.ID)"
            >
              Unblock
            </button>
          </div>
        </div>

        
        <div
          v-if="activeTab === 'hideStory'"
          class="hidden-story-viewers-list"
        >
          <h2>Hidden Story Viewers</h2>
          <p class="tab-description">
            Manage users hidden from viewing your stories. You can unhide them at any time.
          </p>
          
          <div class="search-section">
            <h3>Add User to Hide List</h3>
            <input
              v-model="hideStorySearchQuery"
              type="text"
              placeholder="Search users to hide..."
              class="search-input"
              @input="searchUsersForHiding"
            />
            
            <div
              v-if="loadingHideStorySearch"
              class="loading-text"
            >
              Searching...
            </div>
            
            <div
              v-if="hideStorySearchResults.length > 0"
              class="search-results"
            >
              <div
                v-for="user in hideStorySearchResults"
                :key="user.id"
                class="user-item"
              >
                <img
                  :src="user.profile_picture ? loadPhotoUrl(user.profile_picture) : userLight"
                  class="user-avatar"
                  @error="pfpError"
                />
                <div class="user-info">
                  <span class="username">{{ user.username }}</span>
                  <span class="name">{{ user.full_name }}</span>
                </div>
                <button
                  class="hide-btn"
                  @click="hideStoryFromUser(user.id)"
                >
                  Hide
                </button>
              </div>
            </div>
          </div>

          <h3 style="margin-top: 30px;">
            Hidden Users
          </h3>
          <div
            v-if="hiddenStoryViewers.length === 0"
            class="empty-state"
          >
            No hidden story viewers.
          </div>
          <div
            v-for="viewer in hiddenStoryViewers"
            :key="viewer.ID"
            class="user-item"
          >
            <img
              :src="viewer.profile_picture ? loadPhotoUrl(viewer.profile_picture) : userLight"
              class="user-avatar"
              @error="pfpError"
            />
            <div class="user-info">
              <span class="username">{{ viewer.username }}</span>
              <span class="name">{{ viewer.name }}</span>
            </div>
            <button
              class="unhide-btn"
              @click="unhideStoryViewer(viewer.ID)"
            >
              Unhide
            </button>
          </div>
        </div>

        
        <div
          v-if="activeTab === 'verification'"
          class="verification-request"
        >
          <h2>Request Verification</h2>
          <p class="tab-description">
            Submit your verification request with the required documents.
          </p>
          <div class="form-group">
            <label for="national_id">National Identity Card Number</label>
            <input
              id="national_id"
              v-model="verificationForm.national_id"
              type="text"
              placeholder="Enter your national ID number"
              :disabled="loading"
            />
            <span
              v-if="verificationErrors.national_id"
              class="field-error"
            >{{ verificationErrors.national_id }}</span>
          </div>
          <div class="form-group">
            <label for="reason">Reason for Verification</label>
            <textarea
              id="reason"
              v-model="verificationForm.reason"
              rows="5"
              placeholder="Explain why you want to be verified (minimum 20 characters)"
              :disabled="loading"
            />
            <span
              v-if="verificationErrors.reason"
              class="field-error"
            >{{ verificationErrors.reason }}</span>
          </div>
          <div class="form-group">
            <label
              for="face_picture"
              class="file-label"
            >
              {{ verificationFile ? verificationFile.name : 'Upload Face Picture' }}
            </label>
            <input
              id="face_picture"
              type="file"
              accept="image/*"
              :disabled="loading"
              class="file-input"
              @change="handleVerificationFileSelect"
            />
            <span
              v-if="verificationErrors.face_picture"
              class="field-error"
            >{{ verificationErrors.face_picture }}</span>
          </div>
          <div class="form-group">
            <button
              :disabled="loading"
              @click="submitVerificationRequest"
            >
              {{ loading ? 'Submitting...' : 'Submit Verification Request' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
<style scoped>
.container {
  display: flex;
  background-color: black;
  min-height: 100vh;
}

.settings-page {
  padding: 30px 50px;
  flex: 1;
  color: white;
  overflow-y: auto;
}

.page-title {
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 30px;
  color: white;
}

.tabs-nav {
  display: flex;
  gap: 10px;
  margin-bottom: 30px;
  border-bottom: 2px solid #333;
  overflow-x: auto;
  flex-wrap: wrap;
}

.tabs-nav button {
  padding: 12px 20px;
  background: transparent;
  border: none;
  color: #888;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
  border-bottom: 3px solid transparent;
  white-space: nowrap;
}

.tabs-nav button:hover {
  color: white;
}



.tab-content {
  padding: 20px 0;
}

.error-message {
  background-color: #ff4444;
  color: white;
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.success-message {
  background-color: #00aa00;
  color: white;
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.field-error {
  color: #ff4444;
  font-size: 12px;
  display: block;
  margin-top: 5px;
}

.form-group {
  margin-bottom: 20px;
  border: none;
  color: white;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: white;
  font-weight: 500;
}

.form-group input[type="text"],
.form-group textarea {
  width: 100%;
  max-width: 500px;
  padding: 12px;
  background-color: #333;
  border: 1px solid #555;
  border-radius: 8px;
  color: white;
  font-size: 14px;
}

.form-group input[type="text"]:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #9500ff;
}

.edit-profile {
  max-width: 600px;
}

.edit-profile-pic {
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 50px;
}

.file-input {
  display: none;
}

.file-label {
  display: inline-block;
  padding: 12px 24px;
  background-color: #9500ff;
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: center;
  font-weight: 500;
  margin: 10px 0;
}

.file-label:hover {
  background-color: blueviolet;
}

button {
  padding: 12px 24px;
  font-weight: 700;
  background-color: #9500ff;
  border: none;
  border-radius: 10px;
  color: white;
  cursor: pointer;
  transition: all 0.3s ease;
}

button:hover:not(:disabled) {
  background-color: rgb(45, 45, 45);
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}


.notifications-settings,
.privacy-settings {
  max-width: 600px;
}

.notifications-settings h2,
.privacy-settings h2 {
  font-size: 24px;
  margin-bottom: 20px;
  color: white;
}

.setting-item {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
  padding: 15px;
  background-color: #1a1a1a;
  border-radius: 10px;
}

.setting-label {
  font-size: 16px;
  color: white;
}

.setting-description {
  color: #888;
  font-size: 14px;
  margin-bottom: 20px;
  line-height: 1.6;
}


.switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 24px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #555;
  transition: 0.4s;
  border-radius: 24px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.4s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: #9500ff;
}

input:checked + .slider:before {
  transform: translateX(26px);
}

.close-friends-list,
.blocked-users-list,
.hidden-story-viewers-list {
  max-width: 700px;
}

.close-friends-list h2,
.blocked-users-list h2,
.hidden-story-viewers-list h2 {
  font-size: 24px;
  margin-bottom: 10px;
  color: white;
}

.tab-description {
  color: #888;
  font-size: 14px;
  margin-bottom: 20px;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: #888;
  font-size: 16px;
}

.user-item {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px;
  background-color: #1a1a1a;
  border-radius: 10px;
  margin-bottom: 10px;
  transition: background-color 0.3s ease;
}

.user-item:hover {
  background-color: #252525;
}

.user-avatar {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid #333;
}

.user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.username {
  font-weight: 600;
  font-size: 15px;
  color: white;
}

.name {
  font-size: 13px;
  color: #888;
}

.remove-btn,
.unblock-btn,
.unhide-btn {
  padding: 8px 20px;
  font-size: 13px;
  background-color: #ff4444;
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  transition: all 0.3s ease;
  font-weight: 600;
}

.remove-btn:hover,
.unblock-btn:hover,
.unhide-btn:hover {
  background-color: #cc0000;
}

.verification-request {
  max-width: 600px;
}

.verification-request h2 {
  font-size: 24px;
  margin-bottom: 10px;
  color: white;
}

.search-section {
  background-color: #1a1a1a;
  padding: 20px;
  border-radius: 10px;
  margin-bottom: 20px;
}

.search-section h3 {
  color: white;
  font-size: 18px;
  margin-bottom: 15px;
}

.search-input {
  width: 100%;
  padding: 12px;
  background-color: #333;
  border: 1px solid #555;
  border-radius: 8px;
  color: white;
  font-size: 14px;
  margin-bottom: 15px;
}

.search-input:focus {
  outline: none;
  border-color: #9500ff;
}

.loading-text {
  text-align: center;
  color: #888;
  padding: 10px;
  font-size: 14px;
}

.search-results {
  max-height: 300px;
  overflow-y: auto;
}

.hide-btn {
  padding: 8px 20px;
  font-size: 13px;
  background-color: #9500ff;
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  transition: all 0.3s ease;
  font-weight: 600;
}

.hide-btn:hover {
  background-color: #7000cc;
}

.add-btn {
  padding: 8px 20px;
  font-size: 13px;
  background-color: #00aa00;
  border: none;
  border-radius: 8px;
  color: white;
  cursor: pointer;
  transition: all 0.3s ease;
  font-weight: 600;
}

.add-btn:hover {
  background-color: #008800;
}


@media (max-width: 768px) {
  .settings-page {
    padding: 20px;
  }

  .tabs-nav {
    flex-direction: column;
  }

  .edit-profile-pic {
    flex-direction: column;
    gap: 20px;
  }

  .user-item {
    flex-wrap: wrap;
  }
}
</style>
