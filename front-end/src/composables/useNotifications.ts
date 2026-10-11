import { ref, onMounted, onUnmounted } from "vue";
import { type Notification, type NotificationType, type NotificationUser, notificationService } from "@/services/notificationService";

const notifications = ref<Notification[]>([]);
const unreadCount = ref(0);
const toastQueue = ref<Notification[]>([]);
const isConnected = ref(false);
let ws: WebSocket | null = null;
let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
let pollInterval: ReturnType<typeof setInterval> | null = null;

const loadNotifications = async () => {
  try {
    const data = await notificationService.getNotifications();
    notifications.value = data;
    unreadCount.value = data.filter(n => !n.is_read).length;
  } catch {
    return;
  }
};

const connectWebSocket = () => {
  if (ws && ws.readyState === WebSocket.OPEN) return;

  ws = notificationService.connectWebSocket((notification) => {
    notifications.value.unshift(notification);
    unreadCount.value++;
    toastQueue.value.push(notification);
  });

  if (ws) {
    ws.onopen = () => {
      isConnected.value = true;
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
      }
    };

    ws.onclose = () => {
      isConnected.value = false;
      reconnectTimeout = setTimeout(connectWebSocket, 5000);
    };
  }
};

const startPolling = () => {
  if (pollInterval) return;
  
  pollInterval = setInterval(async () => {
    await loadNotifications();
  }, 30000);
};

const stopPolling = () => {
  if (pollInterval) {
    clearInterval(pollInterval);
    pollInterval = null;
  }
};

export function useNotifications() {
  const initialize = async () => {
    await loadNotifications();
    connectWebSocket();
    startPolling();
  };

  const cleanup = () => {
    if (ws) {
      ws.close();
      ws = null;
    }
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout);
      reconnectTimeout = null;
    }
    stopPolling();
  };

  const markAsRead = async (notificationId: string) => {
    await notificationService.markAsRead(notificationId);
    const notification = notifications.value.find(n => n.id === notificationId);
    if (notification && !notification.is_read) {
      notification.is_read = true;
      unreadCount.value = Math.max(0, unreadCount.value - 1);
    }
  };

  const markAllAsRead = async () => {
    await notificationService.markAllAsRead();
    notifications.value.forEach(n => n.is_read = true);
    unreadCount.value = 0;
  };

  const removeToast = (notificationId: string) => {
    const index = toastQueue.value.findIndex(n => n.id === notificationId);
    if (index > -1) {
      toastQueue.value.splice(index, 1);
    }
  };

  const addNotification = (type: NotificationType, user: NotificationUser, postId?: string, postThumbnail?: string, commentText?: string) => {
    const notification = notificationService.simulateNotification(type, user, postId, postThumbnail, commentText);
    notifications.value.unshift(notification);
    unreadCount.value++;
    toastQueue.value.push(notification);
    return notification;
  };

  const clearAll = () => {
    notificationService.clearAllNotifications();
    notifications.value = [];
    unreadCount.value = 0;
  };

  return {
    notifications,
    unreadCount,
    toastQueue,
    isConnected,
    initialize,
    cleanup,
    loadNotifications,
    markAsRead,
    markAllAsRead,
    removeToast,
    addNotification,
    clearAll,
  };
}

export function useNotificationSetup() {
  const { initialize, cleanup } = useNotifications();

  onMounted(() => {
    const token = localStorage.getItem("token");
    if (token) {
      initialize();
    }
  });

  onUnmounted(() => {
    cleanup();
  });
}
