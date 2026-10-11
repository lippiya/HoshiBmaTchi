import { isTokenExpired } from "@/router";

const API_BASE_URL = "https://localhost/api/v1";
const WS_BASE_URL = "wss://localhost/api/v1";

export interface User {
  id: string;
  username: string;
  email: string;
  full_name: string;
  profile_picture?: string;
  bio?: string;
}

export interface Message {
  id: string;
  conversation_id: string;
  sender_id: string;
  sender?: User;
  content: string;
  media_url?: string;
  media_type?: "image" | "gif" | "video";
  created_at: string;
  status: "sent" | "seen";
  can_unsend: boolean;
}

export interface Conversation {
  id: string;
  participants: User[];
  is_group: boolean;
  group_name?: string;
  last_message?: Message;
  unread_count: number;
  updated_at: string;
}

export interface ConversationDetail {
  conversation: Conversation;
  messages: Message[];
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

export const messageService = {
  async getConversations(): Promise<Conversation[]> {
    const response = await fetch(`${API_BASE_URL}/messages/conversations`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error("Failed to fetch conversations");
    }

    return response.json();
  },

  async searchConversations(query: string): Promise<Conversation[]> {
    const response = await fetch(
      `${API_BASE_URL}/messages/conversations/search?q=${encodeURIComponent(query)}`,
      {
        headers: getAuthHeaders(),
      }
    );

    if (!response.ok) {
      throw new Error("Failed to search conversations");
    }

    return response.json();
  },

  async getConversation(conversationId: string): Promise<ConversationDetail> {
    const response = await fetch(
      `${API_BASE_URL}/messages/conversations/${conversationId}`,
      {
        headers: getAuthHeaders(),
      }
    );

    if (!response.ok) {
      throw new Error("Failed to fetch conversation");
    }

    return response.json();
  },

  async sendMessage(
    conversationId: string,
    content: string,
    mediaFile?: File,
    mediaType?: "image" | "gif" | "video"
  ): Promise<Message> {
    const formData = new FormData();
    formData.append("content", content);
    if (mediaFile) {
      formData.append("media", mediaFile);
      formData.append("media_type", mediaType || "image");
    }

    const token = localStorage.getItem("token");
    if (!token || isTokenExpired(token)) {
      throw new Error("Unauthorized");
    }

    const response = await fetch(
      `${API_BASE_URL}/messages/conversations/${conversationId}/messages`,
      {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      }
    );

    if (!response.ok) {
      throw new Error("Failed to send message");
    }

    return response.json();
  },

  async unsendMessage(conversationId: string, messageId: string): Promise<void> {
    const response = await fetch(
      `${API_BASE_URL}/messages/conversations/${conversationId}/messages/${messageId}`,
      {
        method: "DELETE",
        headers: getAuthHeaders(),
      }
    );

    if (!response.ok) {
      throw new Error("Failed to unsend message");
    }
  },

  async createConversation(participantIds: string[], groupName?: string): Promise<Conversation> {
    const response = await fetch(`${API_BASE_URL}/messages/conversations`, {
      method: "POST",
      headers: getAuthHeaders(),
      body: JSON.stringify({
        participant_ids: participantIds,
        is_group: participantIds.length > 1,
        group_name: groupName,
      }),
    });

    if (!response.ok) {
      throw new Error("Failed to create conversation");
    }

    return response.json();
  },

  async deleteConversation(conversationId: string): Promise<void> {
    const response = await fetch(
      `${API_BASE_URL}/messages/conversations/${conversationId}`,
      {
        method: "DELETE",
        headers: getAuthHeaders(),
      }
    );

    if (!response.ok) {
      throw new Error("Failed to delete conversation");
    }
  },

  async addParticipants(conversationId: string, participantIds: string[]): Promise<void> {
    const response = await fetch(
      `${API_BASE_URL}/messages/conversations/${conversationId}/participants`,
      {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({ participant_ids: participantIds }),
      }
    );

    if (!response.ok) {
      throw new Error("Failed to add participants");
    }
  },

  async removeParticipant(conversationId: string, participantId: string): Promise<void> {
    const response = await fetch(
      `${API_BASE_URL}/messages/conversations/${conversationId}/participants/${participantId}`,
      {
        method: "DELETE",
        headers: getAuthHeaders(),
      }
    );

    if (!response.ok) {
      throw new Error("Failed to remove participant");
    }
  },

  async markAsSeen(conversationId: string): Promise<void> {
    const response = await fetch(
      `${API_BASE_URL}/messages/conversations/${conversationId}/seen`,
      {
        method: "PUT",
        headers: getAuthHeaders(),
      }
    );

    if (!response.ok) {
      throw new Error("Failed to mark as seen");
    }
  },

  async searchUsers(query: string): Promise<User[]> {
    const response = await fetch(
      `${API_BASE_URL}/users/search?q=${encodeURIComponent(query)}`,
      {
        headers: getAuthHeaders(),
      }
    );

    if (!response.ok) {
      throw new Error("Failed to search users");
    }

    return response.json();
  },

  async searchMessages(conversationId: string, query: string): Promise<Message[]> {
    const response = await fetch(
      `${API_BASE_URL}/messages/conversations/${conversationId}/search?q=${encodeURIComponent(query)}`,
      {
        headers: getAuthHeaders(),
      }
    );

    if (!response.ok) {
      throw new Error("Failed to search messages");
    }

    return response.json();
  },

  connectWebSocket(conversationId: string, onMessage: (message: Message) => void): WebSocket {
    const token = localStorage.getItem("token");
    console.log("WebSocket connecting with token:", token ? "exists" : "null");
    const ws = new WebSocket(
      `${WS_BASE_URL}/messages/ws?conversation_id=${conversationId}&token=${token}`
    );

    ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      onMessage(message);
    };

    return ws;
  },
};
