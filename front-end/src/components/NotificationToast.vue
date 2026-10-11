<script setup lang="ts">
import { onUnmounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useNotifications } from "@/composables/useNotifications";
import { usePhoto } from "@/composables/usePhoto";
import type { Notification } from "@/services/notificationService";
import defaultProfileIcon from "@/components/icons/user-light.png";

const router = useRouter();
const { toastQueue, removeToast, markAsRead } = useNotifications();
const { loadPhotoUrl } = usePhoto();

const visibleToasts = ref<Notification[]>([]);
const maxVisibleToasts = 3;
const toastDuration = 5000;

const timeouts = new Map<string, ReturnType<typeof setTimeout>>();

const getNotificationMessage = (notification: Notification): string => {
  switch (notification.type) {
    case "like":
      return "liked your post";
    case "comment":
      return notification.comment_text 
        ? "commented: \"" + notification.comment_text.substring(0, 30) + (notification.comment_text.length > 30 ? "..." : "") + "\""
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
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
  if (diffInSeconds < 604800) return `${Math.floor(diffInSeconds / 86400)}d ago`;
  return `${Math.floor(diffInSeconds / 604800)}w ago`;
};

const handleToastClick = async (notification: Notification) => {
  await markAsRead(notification.id);
  dismissToast(notification.id);

  if (notification.type === "like" || notification.type === "comment") {
    if (notification.post_id) {
      router.push({ name: "home", query: { post: notification.post_id } });
    }
  } else if (notification.type === "follow" || notification.type === "mention") {
    router.push({ name: "profile", params: { userId: notification.user.id } });
  }
};

const dismissToast = (notificationId: string) => {
  const timeout = timeouts.get(notificationId);
  if (timeout) {
    clearTimeout(timeout);
    timeouts.delete(notificationId);
  }
  
  const index = visibleToasts.value.findIndex(t => t.id === notificationId);
  if (index > -1) {
    visibleToasts.value.splice(index, 1);
  }
  removeToast(notificationId);
};

const scheduleRemoval = (notificationId: string) => {
  const timeout = setTimeout(() => {
    dismissToast(notificationId);
  }, toastDuration);
  timeouts.set(notificationId, timeout);
};

const pfpError = (e: Event) => {
  const target = e.target as HTMLImageElement;
  target.src = defaultProfileIcon;
};

watch(toastQueue, (newQueue) => {
  while (newQueue.length > 0 && visibleToasts.value.length < maxVisibleToasts) {
    const notification = newQueue.shift();
    if (notification && !visibleToasts.value.find(t => t.id === notification.id)) {
      visibleToasts.value.push(notification);
      scheduleRemoval(notification.id);
    }
  }
}, { deep: true });

onUnmounted(() => {
  timeouts.forEach((timeout) => clearTimeout(timeout));
  timeouts.clear();
});
</script>

<template>
  <Teleport to="body">
    <div class="toast-container">
      <TransitionGroup name="toast">
        <div
          v-for="toast in visibleToasts"
          :key="toast.id"
          class="toast"
          @click="handleToastClick(toast)"
        >
          <div class="toast-content">
            <img
              :src="toast.user.profile_picture ? loadPhotoUrl(toast.user.profile_picture) : defaultProfileIcon"
              class="toast-avatar"
              @error="pfpError"
            />
            <div class="toast-info">
              <div class="toast-message">
                <span class="toast-username">{{ toast.user.username }}</span>
                <span class="toast-action">{{ getNotificationMessage(toast) }}</span>
              </div>
              <div class="toast-time">
                {{ getRelativeTime(toast.created_at) }}
              </div>
            </div>
            <button
              class="toast-close"
              @click.stop="dismissToast(toast.id)"
            >
              ×
            </button>
          </div>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-container {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 10000;
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-width: 400px;
}

.toast {
  background: linear-gradient(135deg, #1e1e1e 0%, #2d2d2d 100%);
  border: 1px solid #444;
  border-radius: 12px;
  padding: 12px 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
}

.toast:hover {
  transform: translateX(-5px);
  border-color: #666;
  box-shadow: 0 6px 25px rgba(0, 0, 0, 0.5);
}

.toast-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toast-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid #444;
  flex-shrink: 0;
}

.toast-info {
  flex: 1;
  min-width: 0;
}

.toast-message {
  color: #fff;
  font-size: 14px;
  line-height: 1.4;
  word-wrap: break-word;
}

.toast-username {
  font-weight: 600;
  margin-right: 4px;
}

.toast-action {
  color: #b0b0b0;
}

.toast-time {
  color: #888;
  font-size: 12px;
  margin-top: 4px;
}

.toast-close {
  background: none;
  border: none;
  color: #888;
  font-size: 20px;
  cursor: pointer;
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.toast-close:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(100%);
}

.toast-move {
  transition: transform 0.3s ease;
}

@media (max-width: 480px) {
  .toast-container {
    right: 10px;
    left: 10px;
    max-width: none;
  }

  .toast {
    padding: 10px 12px;
  }

  .toast-avatar {
    width: 36px;
    height: 36px;
  }

  .toast-message {
    font-size: 13px;
  }
}
</style>
