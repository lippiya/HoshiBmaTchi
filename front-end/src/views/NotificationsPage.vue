<script setup lang="ts">
import SideBar from "@/components/SideBar.vue";
import { onMounted, ref, computed } from "vue";
import { useRouter } from "vue-router";
import { useNotifications } from "@/composables/useNotifications";
import { usePhoto } from "@/composables/usePhoto";
import type { Notification, NotificationType } from "@/services/notificationService";
import defaultProfileIcon from "@/components/icons/user-light.png";
import likeIcon from "@/components/icons/notif-light.png";
import commentIcon from "@/components/icons/comment-light.png";

const router = useRouter();
const { notifications, unreadCount, loadNotifications, markAsRead, markAllAsRead, clearAll } = useNotifications();
const { loadPhotoUrl } = usePhoto();

const loading = ref(true);
const activeFilter = ref<"all" | NotificationType>("all");

const filteredNotifications = computed(() => {
  if (activeFilter.value === "all") {
    return notifications.value;
  }
  return notifications.value.filter(n => n.type === activeFilter.value);
});

const groupedNotifications = computed(() => {
  const now = new Date();
  const today: Notification[] = [];
  const thisWeek: Notification[] = [];
  const thisMonth: Notification[] = [];
  const older: Notification[] = [];

  filteredNotifications.value.forEach(notification => {
    const date = new Date(notification.created_at);
    const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));

    if (diffDays === 0) {
      today.push(notification);
    } else if (diffDays < 7) {
      thisWeek.push(notification);
    } else if (diffDays < 30) {
      thisMonth.push(notification);
    } else {
      older.push(notification);
    }
  });

  return { today, thisWeek, thisMonth, older };
});

const getNotificationMessage = (notification: Notification): string => {
  switch (notification.type) {
    case "like":
      return "liked your post";
    case "comment":
      return notification.comment_text 
        ? `commented: "${notification.comment_text.substring(0, 50)}${notification.comment_text.length > 50 ? '...' : ''}"`
        : "commented on your post";
    case "follow":
      return "started following you";
    case "mention":
      return "mentioned you in a post";
    default:
      return "interacted with your content";
  }
};

const getRelativeTime = (dateString: string): string => {
  const date = new Date(dateString);
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffInSeconds < 60) return "just now";
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m`;
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h`;
  if (diffInSeconds < 604800) return `${Math.floor(diffInSeconds / 86400)}d`;
  return `${Math.floor(diffInSeconds / 604800)}w`;
};

const getNotificationIcon = (type: NotificationType): string => {
  switch (type) {
    case "like":
      return likeIcon;
    case "comment":
      return commentIcon;
    case "follow":
      return defaultProfileIcon;
    case "mention":
      return commentIcon;
    default:
      return likeIcon;
  }
};

const handleNotificationClick = async (notification: Notification) => {
  await markAsRead(notification.id);

  if (notification.type === "like" || notification.type === "comment") {
    if (notification.post_id) {
      router.push({ name: "home", query: { post: notification.post_id } });
    }
  } else if (notification.type === "follow" || notification.type === "mention") {
    router.push({ name: "profile", params: { userId: notification.user.id } });
  }
};

const handleMarkAllRead = async () => {
  await markAllAsRead();
};

const handleClearAll = () => {
  clearAll();
};

const pfpError = (e: Event) => {
  const target = e.target as HTMLImageElement;
  target.src = defaultProfileIcon;
};

onMounted(async () => {
  loading.value = true;
  await loadNotifications();
  loading.value = false;
});
</script>

<template>
  <div class="container">
    <SideBar />
    <div class="notifications-page">
      <div class="notifications-container">
        <div class="notifications-header">
          <h1>Notifications</h1>
          <div class="header-actions">
            <button
              v-if="unreadCount > 0"
              class="action-btn"
              @click="handleMarkAllRead"
            >
              Mark all as read
            </button>
            <button
              v-if="notifications.length > 0"
              class="action-btn danger"
              @click="handleClearAll"
            >
              Clear all
            </button>
          </div>
        </div>

        <div class="filter-tabs">
          <button
            :class="['filter-tab', { active: activeFilter === 'all' }]"
            @click="activeFilter = 'all'"
          >
            All
          </button>
          <button
            :class="['filter-tab', { active: activeFilter === 'like' }]"
            @click="activeFilter = 'like'"
          >
            Likes
          </button>
          <button
            :class="['filter-tab', { active: activeFilter === 'comment' }]"
            @click="activeFilter = 'comment'"
          >
            Comments
          </button>
          <button
            :class="['filter-tab', { active: activeFilter === 'follow' }]"
            @click="activeFilter = 'follow'"
          >
            Follows
          </button>
          <button
            :class="['filter-tab', { active: activeFilter === 'mention' }]"
            @click="activeFilter = 'mention'"
          >
            Mentions
          </button>
        </div>

        <div
          v-if="loading"
          class="loading-state"
        >
          <div class="spinner"></div>
          <p>Loading notifications...</p>
        </div>

        <div
          v-else-if="filteredNotifications.length === 0"
          class="empty-state"
        >
          <img
            :src="likeIcon"
            class="empty-icon"
          />
          <h2>No notifications yet</h2>
          <p>When someone likes, comments, follows, or mentions you, you'll see it here.</p>
        </div>

        <div
          v-else
          class="notifications-list"
        >
          <div
            v-if="groupedNotifications.today.length > 0"
            class="notification-group"
          >
            <h3 class="group-title">Today</h3>
            <div
              v-for="notification in groupedNotifications.today"
              :key="notification.id"
              :class="['notification-item', { unread: !notification.is_read }]"
              @click="handleNotificationClick(notification)"
            >
              <div class="notification-avatar-container">
                <img
                  :src="notification.user.profile_picture ? loadPhotoUrl(notification.user.profile_picture) : defaultProfileIcon"
                  class="notification-avatar"
                  @error="pfpError"
                />
                <img
                  :src="getNotificationIcon(notification.type)"
                  class="notification-type-icon"
                />
              </div>
              <div class="notification-content">
                <div class="notification-text">
                  <span class="username">{{ notification.user.username }}</span>
                  <span class="action-text">{{ getNotificationMessage(notification) }}</span>
                </div>
                <div class="notification-time">{{ getRelativeTime(notification.created_at) }}</div>
              </div>
              <div
                v-if="notification.post_thumbnail && (notification.type === 'like' || notification.type === 'comment')"
                class="notification-thumbnail"
              >
                <img
                  :src="notification.post_thumbnail"
                  @error="pfpError"
                />
              </div>
              <div
                v-if="!notification.is_read"
                class="unread-indicator"
              ></div>
            </div>
          </div>

          <div
            v-if="groupedNotifications.thisWeek.length > 0"
            class="notification-group"
          >
            <h3 class="group-title">This Week</h3>
            <div
              v-for="notification in groupedNotifications.thisWeek"
              :key="notification.id"
              :class="['notification-item', { unread: !notification.is_read }]"
              @click="handleNotificationClick(notification)"
            >
              <div class="notification-avatar-container">
                <img
                  :src="notification.user.profile_picture ? loadPhotoUrl(notification.user.profile_picture) : defaultProfileIcon"
                  class="notification-avatar"
                  @error="pfpError"
                />
                <img
                  :src="getNotificationIcon(notification.type)"
                  class="notification-type-icon"
                />
              </div>
              <div class="notification-content">
                <div class="notification-text">
                  <span class="username">{{ notification.user.username }}</span>
                  <span class="action-text">{{ getNotificationMessage(notification) }}</span>
                </div>
                <div class="notification-time">{{ getRelativeTime(notification.created_at) }}</div>
              </div>
              <div
                v-if="notification.post_thumbnail && (notification.type === 'like' || notification.type === 'comment')"
                class="notification-thumbnail"
              >
                <img
                  :src="notification.post_thumbnail"
                  @error="pfpError"
                />
              </div>
              <div
                v-if="!notification.is_read"
                class="unread-indicator"
              ></div>
            </div>
          </div>

          <div
            v-if="groupedNotifications.thisMonth.length > 0"
            class="notification-group"
          >
            <h3 class="group-title">This Month</h3>
            <div
              v-for="notification in groupedNotifications.thisMonth"
              :key="notification.id"
              :class="['notification-item', { unread: !notification.is_read }]"
              @click="handleNotificationClick(notification)"
            >
              <div class="notification-avatar-container">
                <img
                  :src="notification.user.profile_picture ? loadPhotoUrl(notification.user.profile_picture) : defaultProfileIcon"
                  class="notification-avatar"
                  @error="pfpError"
                />
                <img
                  :src="getNotificationIcon(notification.type)"
                  class="notification-type-icon"
                />
              </div>
              <div class="notification-content">
                <div class="notification-text">
                  <span class="username">{{ notification.user.username }}</span>
                  <span class="action-text">{{ getNotificationMessage(notification) }}</span>
                </div>
                <div class="notification-time">{{ getRelativeTime(notification.created_at) }}</div>
              </div>
              <div
                v-if="notification.post_thumbnail && (notification.type === 'like' || notification.type === 'comment')"
                class="notification-thumbnail"
              >
                <img
                  :src="notification.post_thumbnail"
                  @error="pfpError"
                />
              </div>
              <div
                v-if="!notification.is_read"
                class="unread-indicator"
              ></div>
            </div>
          </div>

          <div
            v-if="groupedNotifications.older.length > 0"
            class="notification-group"
          >
            <h3 class="group-title">Older</h3>
            <div
              v-for="notification in groupedNotifications.older"
              :key="notification.id"
              :class="['notification-item', { unread: !notification.is_read }]"
              @click="handleNotificationClick(notification)"
            >
              <div class="notification-avatar-container">
                <img
                  :src="notification.user.profile_picture ? loadPhotoUrl(notification.user.profile_picture) : defaultProfileIcon"
                  class="notification-avatar"
                  @error="pfpError"
                />
                <img
                  :src="getNotificationIcon(notification.type)"
                  class="notification-type-icon"
                />
              </div>
              <div class="notification-content">
                <div class="notification-text">
                  <span class="username">{{ notification.user.username }}</span>
                  <span class="action-text">{{ getNotificationMessage(notification) }}</span>
                </div>
                <div class="notification-time">{{ getRelativeTime(notification.created_at) }}</div>
              </div>
              <div
                v-if="notification.post_thumbnail && (notification.type === 'like' || notification.type === 'comment')"
                class="notification-thumbnail"
              >
                <img
                  :src="notification.post_thumbnail"
                  @error="pfpError"
                />
              </div>
              <div
                v-if="!notification.is_read"
                class="unread-indicator"
              />
            </div>
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
  position: relative;
  min-height: 100vh;
}

.notifications-page {
  flex: 1;
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  overflow-x: hidden;
  margin: 0 auto;
  max-width: 630px;
  padding: 0;
  margin-left: 250px;
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.notifications-page::-webkit-scrollbar {
  display: none;
}

.notifications-container {
  width: 100%;
  background-color: #000;
  padding: 0 16px;
}

.notifications-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 0;
  border-bottom: 1px solid #262626;
}

.notifications-header h1 {
  color: white;
  font-size: 24px;
  font-weight: 600;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.action-btn {
  background: none;
  border: 1px solid #363636;
  color: #0095f6;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: #1a1a1a;
}

.action-btn.danger {
  color: #ed4956;
}

.action-btn.danger:hover {
  background: rgba(237, 73, 86, 0.1);
}

.filter-tabs {
  display: flex;
  gap: 8px;
  padding: 16px 0;
  border-bottom: 1px solid #262626;
  overflow-x: auto;
}

.filter-tab {
  background: #262626;
  border: none;
  color: #a8a8a8;
  padding: 8px 16px;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.filter-tab:hover {
  background: #363636;
  color: white;
}

.filter-tab.active {
  background: white;
  color: black;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #262626;
  border-top-color: #0095f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.loading-state p,
.empty-state p {
  color: #a8a8a8;
  font-size: 14px;
  margin-top: 16px;
}

.empty-icon {
  width: 60px;
  height: 60px;
  opacity: 0.5;
  margin-bottom: 16px;
}

.empty-state h2 {
  color: white;
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 8px 0;
}

.notifications-list {
  display: flex;
  flex-direction: column;
}

.notification-group {
  padding-top: 16px;
}

.group-title {
  color: white;
  font-size: 16px;
  font-weight: 600;
  margin: 0 0 12px 0;
  padding: 0;
}

.notification-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  margin: 0 -16px;
  cursor: pointer;
  transition: background-color 0.2s ease;
  position: relative;
}

.notification-item:hover {
  background-color: #1a1a1a;
}

.notification-item.unread {
  background-color: rgba(0, 149, 246, 0.05);
}

.notification-avatar-container {
  position: relative;
  flex-shrink: 0;
}

.notification-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid #262626;
}

.notification-type-icon {
  position: absolute;
  bottom: -2px;
  right: -2px;
  width: 18px;
  height: 18px;
  background: #262626;
  border-radius: 50%;
  padding: 2px;
}

.notification-content {
  flex: 1;
  min-width: 0;
}

.notification-text {
  color: white;
  font-size: 14px;
  line-height: 1.4;
  word-wrap: break-word;
}

.username {
  font-weight: 600;
  margin-right: 4px;
}

.action-text {
  color: #a8a8a8;
}

.notification-time {
  color: #8e8e8e;
  font-size: 12px;
  margin-top: 2px;
}

.notification-thumbnail {
  flex-shrink: 0;
}

.notification-thumbnail img {
  width: 44px;
  height: 44px;
  object-fit: cover;
  border-radius: 4px;
}

.unread-indicator {
  position: absolute;
  right: 16px;
  width: 8px;
  height: 8px;
  background: #0095f6;
  border-radius: 50%;
}

@media (max-width: 768px) {
  .container {
    flex-direction: column;
  }

  .notifications-page {
    margin-left: 0;
    max-width: 100%;
    padding-bottom: 80px;
  }

  .notifications-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .header-actions {
    width: 100%;
    justify-content: flex-start;
  }

  .filter-tabs {
    padding: 12px 0;
  }

  .filter-tab {
    padding: 6px 12px;
    font-size: 13px;
  }

  .notification-item {
    padding: 10px 12px;
    margin: 0 -12px;
  }

  .notification-avatar {
    width: 40px;
    height: 40px;
  }

  .notification-thumbnail img {
    width: 40px;
    height: 40px;
  }
}
</style>
