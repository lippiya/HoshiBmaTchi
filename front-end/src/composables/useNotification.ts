import { ref } from "vue";

export interface Notification {
  ID: string;
  RecipientID: string;
  ActorID: string;
  Type: string;
  PostID?: string;
  CommentText?: string;
  IsRead: boolean;
  CreatedAt: string;
  actor?: {
    ID: string;
    Username: string;
    Name: string;
    ProfilePicture?: string;
  };
}

export function useNotification() {
  const notifications = ref<Notification[]>([]);
  const unreadCount = ref(0);
  const loading = ref(false);

  const getToken = () => localStorage.getItem("token");

  const createNotification = async (
    recipientId: string,
    actorId: string,
    type: "like" | "comment" | "follow" | "mention",
    postId?: string,
    commentText?: string
  ) => {
    const token = getToken();
    if (!token) return;

    
    if (recipientId === actorId) return;

    try {
      const body: Record<string, string> = {
        recipient_id: recipientId,
        actor_id: actorId,
        type: type,
      };

      if (postId) {
        body.post_id = postId;
      }

      if (commentText) {
        body.comment_text = commentText;
      }

      await fetch("https://localhost/api/v1/social/create-notification", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(body),
      });
    } catch (error) {
      console.error("Failed to create notification:", error);
    }
  };

  const fetchNotifications = async (page: number = 1, limit: number = 20) => {
    const token = getToken();
    if (!token) return;

    loading.value = true;
    try {
      const response = await fetch(
        `https://localhost/api/v1/social/get-notifications?page=${page}&limit=${limit}`,
        {
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      if (response.ok) {
        const data = await response.json();
        notifications.value = data.notifications || [];
        unreadCount.value = data.unread_count || 0;
        return data;
      }
    } catch (error) {
      console.error("Failed to fetch notifications:", error);
    } finally {
      loading.value = false;
    }
  };

  const fetchUnreadCount = async () => {
    const token = getToken();
    if (!token) return 0;

    try {
      const response = await fetch(
        "https://localhost/api/v1/social/notifications/unread-count",
        {
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );

      if (response.ok) {
        const data = await response.json();
        unreadCount.value = data.unread_count || 0;
        return unreadCount.value;
      }
    } catch (error) {
      console.error("Failed to fetch unread count:", error);
    }
    return 0;
  };

  const markAsRead = async (notificationId: string) => {
    const token = getToken();
    if (!token) return;

    try {
      await fetch(
        `https://localhost/api/v1/social/mark-notification-read/${notificationId}`,
        {
          method: "PUT",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );
      
      const notif = notifications.value.find((n) => n.ID === notificationId);
      if (notif) {
        notif.IsRead = true;
        unreadCount.value = Math.max(0, unreadCount.value - 1);
      }
    } catch (error) {
      console.error("Failed to mark notification as read:", error);
    }
  };

  const markAllAsRead = async () => {
    const token = getToken();
    if (!token) return;

    try {
      await fetch("https://localhost/api/v1/social/mark-all-notifications-read", {
        method: "PUT",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });
      
      notifications.value.forEach((n) => (n.IsRead = true));
      unreadCount.value = 0;
    } catch (error) {
      console.error("Failed to mark all as read:", error);
    }
  };

  const deleteNotification = async (notificationId: string) => {
    const token = getToken();
    if (!token) return;

    try {
      await fetch(
        `https://localhost/api/v1/social/delete-notification/${notificationId}`,
        {
          method: "DELETE",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }
      );
      
      notifications.value = notifications.value.filter(
        (n) => n.ID !== notificationId
      );
    } catch (error) {
      console.error("Failed to delete notification:", error);
    }
  };

  return {
    notifications,
    unreadCount,
    loading,
    createNotification,
    fetchNotifications,
    fetchUnreadCount,
    markAsRead,
    markAllAsRead,
    deleteNotification,
  };
}
