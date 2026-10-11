import { isTokenExpired } from "@/router";

const API_BASE_URL = "https://localhost/api/v1";
const WS_BASE_URL = "wss://localhost/api/v1";

export type NotificationType = "like" | "comment" | "follow" | "mention";

export interface NotificationUser {
  id: string;
  username: string;
  profile_picture?: string;
}

export interface Notification {
  id: string;
  type: NotificationType;
  user: NotificationUser;
  post_id?: string;
  post_thumbnail?: string;
  comment_text?: string;
  created_at: string;
  is_read: boolean;
}

const getAuthHeaders = () => {
  const token = localStorage.getItem("token");
  if (!token || isTokenExpired(token)) {
    throw new Error("Unauthorized");
  }
  return {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  };
};

const generateId = () => {
  return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
};

const NOTIFICATION_STORAGE_KEY = "hoshibmatchi_notifications";

const loadStoredNotifications = (): Notification[] => {
  try {
    const stored = localStorage.getItem(NOTIFICATION_STORAGE_KEY);
    if (stored) {
      return JSON.parse(stored);
    }
  } catch {
    return [];
  }
  return [];
};

const saveNotifications = (notifications: Notification[]) => {
  try {
    localStorage.setItem(NOTIFICATION_STORAGE_KEY, JSON.stringify(notifications.slice(0, 100)));
  } catch {
    return;
  }
};

export const notificationService = {
  async getNotifications(): Promise<Notification[]> {
    const token = localStorage.getItem("token");
    if (!token || isTokenExpired(token)) {
      throw new Error("Unauthorized");
    }

    try {
      const response = await fetch(`${API_BASE_URL}/social/get-notifications`, {
        headers: getAuthHeaders(),
      });

      if (response.ok) {
        const data = await response.json();
        const notifications = data.notifications || [];
        saveNotifications(notifications);
        return notifications;
      }
    } catch {
      return loadStoredNotifications();
    }

    return loadStoredNotifications();
  },

  async markAsRead(notificationId: string): Promise<void> {
    try {
      await fetch(`${API_BASE_URL}/social/mark-notification-read/${notificationId}`, {
        method: "PUT",
        headers: getAuthHeaders(),
      });
    } catch {
      const notifications = loadStoredNotifications();
      const updated = notifications.map(n => 
        n.id === notificationId ? { ...n, is_read: true } : n
      );
      saveNotifications(updated);
    }
  },

  async markAllAsRead(): Promise<void> {
    try {
      await fetch(`${API_BASE_URL}/social/mark-all-notifications-read`, {
        method: "PUT",
        headers: getAuthHeaders(),
      });
    } catch {
      const notifications = loadStoredNotifications();
      const updated = notifications.map(n => ({ ...n, is_read: true }));
      saveNotifications(updated);
    }
  },

  addLocalNotification(notification: Omit<Notification, "id" | "created_at" | "is_read">): Notification {
    const newNotification: Notification = {
      ...notification,
      id: generateId(),
      created_at: new Date().toISOString(),
      is_read: false,
    };
    
    const notifications = loadStoredNotifications();
    notifications.unshift(newNotification);
    saveNotifications(notifications);
    
    return newNotification;
  },

  getUnreadCount(): number {
    const notifications = loadStoredNotifications();
    return notifications.filter(n => !n.is_read).length;
  },

  connectWebSocket(onNotification: (notification: Notification) => void): WebSocket | null {
    const token = localStorage.getItem("token");
    if (!token || isTokenExpired(token)) {
      return null;
    }

    try {
      const ws = new WebSocket(`${WS_BASE_URL}/social/notifications/ws?token=${token}`);

      ws.onmessage = (event) => {
        try {
          const notification = JSON.parse(event.data) as Notification;
          const notifications = loadStoredNotifications();
          notifications.unshift(notification);
          saveNotifications(notifications);
          onNotification(notification);
        } catch {
          return;
        }
      };

      ws.onerror = () => {
        return;
      };

      ws.onclose = () => {
        return;
      };

      return ws;
    } catch {
      return null;
    }
  },

  simulateNotification(type: NotificationType, user: NotificationUser, postId?: string, postThumbnail?: string, commentText?: string): Notification {
    return this.addLocalNotification({
      type,
      user,
      post_id: postId,
      post_thumbnail: postThumbnail,
      comment_text: commentText,
    });
  },

  clearAllNotifications(): void {
    saveNotifications([]);
  },
};
