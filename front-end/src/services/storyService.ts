import { isTokenExpired } from "@/router";

const API_BASE_URL = "https://localhost/api/v1";

export interface Story {
  id: string;
  user_id: string;
  media_url: string;
  media_type: string;
  duration: number;
  created_at: string;
  expires_at: string;
  user?: {
    id: string;
    username: string;
    profile_picture?: string;
    full_name?: string;
    is_private?: boolean;
  };
  likes_count?: number;
  views_count?: number;
  replies_count?: number;
}

export interface StoryGroup {
  user_id: string;
  user?: {
    id: string;
    username: string;
    profile_picture?: string;
    full_name?: string;
    is_private?: boolean;
  };
  stories: Story[];
  has_viewed: boolean;
}

export interface StoryReply {
  id: string;
  story_id: string;
  user_id: string;
  reply: string;
  created_at: string;
  user?: {
    id: string;
    username: string;
    profile_picture?: string;
    full_name?: string;
  };
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

export const storyService = {
  async createStory(mediaFile: File): Promise<Story> {
    const token = localStorage.getItem("token");
    if (!token || isTokenExpired(token)) {
      throw new Error("Unauthorized");
    }

    let duration = 5;
    if (mediaFile.type.startsWith("video/")) {
      const video = document.createElement("video");
      video.preload = "metadata";
      await new Promise((resolve) => {
        video.onloadedmetadata = () => {
          duration = Math.ceil(video.duration);
          resolve(duration);
        };
        video.src = URL.createObjectURL(mediaFile);
      });
    }

    const formData = new FormData();
    formData.append("media", mediaFile);
    formData.append("duration", duration.toString());

    const response = await fetch(`${API_BASE_URL}/stories/create`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
      },
      body: formData,
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || "Failed to create story");
    }

    const data = await response.json();
    return data.story;
  },

  async getAllActiveStories(): Promise<StoryGroup[]> {
    let response;
    try {
      response = await fetch(`${API_BASE_URL}/stories/get-all-stories`, {
        headers: getAuthHeaders(),
      });
    } catch {
      return [];
    }

    if (!response.ok) {
      return [];
    }

    let data;
    try {
      data = await response.json();
    } catch {
      return [];
    }
    
    let storyGroups = data.story_groups || [];

    const currentUserId = await this.getCurrentUserId();

    let followingIds: Set<string> = new Set();
    try {
      const followingRes = await fetch(`${API_BASE_URL}/users/get-following/${currentUserId}`, {
        headers: getAuthHeaders(),
      });
      if (followingRes.ok) {
        const followingData = await followingRes.json();
        followingIds = new Set((followingData.following || []).map((f: { ID?: string; id?: string }) => f.ID || f.id || ""));
      }
    } catch {
      followingIds = new Set();
    }

    storyGroups = storyGroups.filter((group: StoryGroup) => {
      if (group.user_id === currentUserId) return true;
      const user = group.user;
      if (user && user.is_private === true) {
        return followingIds.has(group.user_id);
      }
      return true;
    });

    const hiddenResponse = await fetch(`${API_BASE_URL}/social/get-hidden-story-viewers`, {
      headers: getAuthHeaders(),
    });

    if (hiddenResponse.ok) {
      const hiddenData = await hiddenResponse.json();
      const hiddenUserIds = (hiddenData.hidden_viewers || []).map((viewer: { ID?: string; id?: string }) => viewer.ID || viewer.id);
      storyGroups = storyGroups.filter((group: StoryGroup) => !hiddenUserIds.includes(group.user_id));
    }

    for (const group of storyGroups) {
      let hasViewed = true;
      for (const story of group.stories) {
        const viewResponse = await fetch(
          `${API_BASE_URL}/stories/check-view/${story.id}`,
          {
            headers: getAuthHeaders(),
          }
        );
        if (viewResponse.ok) {
          const viewData = await viewResponse.json();
          if (!viewData.viewed) {
            hasViewed = false;
            break;
          }
        }
      }
      group.has_viewed = hasViewed;
      
      if (!group.user && group.user_id === currentUserId) {
        const userResponse = await fetch(`${API_BASE_URL}/users/get-current-user`, {
          headers: getAuthHeaders(),
        });
        if (userResponse.ok) {
          const userData = await userResponse.json();
          group.user = userData.user;
        }
      }
    }

    return storyGroups;
  },

  async getStoryById(storyId: string): Promise<Story> {
    const response = await fetch(`${API_BASE_URL}/stories/get/${storyId}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error("Failed to fetch story");
    }

    const data = await response.json();
    return data.story;
  },

  async getUserStories(userId: string): Promise<Story[]> {
    const response = await fetch(`${API_BASE_URL}/stories/user/${userId}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error("Failed to fetch user stories");
    }

    const data = await response.json();
    return data.stories || [];
  },

  async deleteStory(storyId: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/stories/delete/${storyId}`, {
      method: "DELETE",
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error("Failed to delete story");
    }
  },

  async likeStory(storyId: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/stories/like-story/${storyId}`, {
      method: "POST",
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error("Failed to like story");
    }
  },

  async checkStoryLike(storyId: string): Promise<boolean> {
    const response = await fetch(`${API_BASE_URL}/stories/check-like/${storyId}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      return false;
    }

    const data = await response.json();
    return data.liked;
  },

  async replyToStory(storyId: string, reply: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/stories/reply/${storyId}`, {
      method: "POST",
      headers: getAuthHeaders(),
      body: JSON.stringify({ reply }),
    });

    if (!response.ok) {
      throw new Error("Failed to reply to story");
    }
  },

  async getStoryReplies(storyId: string): Promise<StoryReply[]> {
    const response = await fetch(`${API_BASE_URL}/stories/replies/${storyId}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error("Failed to fetch story replies");
    }

    const data = await response.json();
    return data.replies || [];
  },

  async viewStory(storyId: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/stories/view/${storyId}`, {
      method: "POST",
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error("Failed to mark story as viewed");
    }
  },

  async checkStoryView(storyId: string): Promise<boolean> {
    const response = await fetch(`${API_BASE_URL}/stories/check-view/${storyId}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      return false;
    }

    const data = await response.json();
    return data.viewed;
  },

  async getStoryViews(storyId: string): Promise<{ id: string; story_id: string; user_id: string; created_at: string }[]> {
    const response = await fetch(`${API_BASE_URL}/stories/views/${storyId}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error("Failed to fetch story views");
    }

    const data = await response.json();
    return data.views || [];
  },

  async shareStory(storyId: string, recipientId: string): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/stories/share/${storyId}`, {
      method: "POST",
      headers: getAuthHeaders(),
      body: JSON.stringify({ recipient_id: recipientId }),
    });

    if (!response.ok) {
      throw new Error("Failed to share story");
    }
  },

  async getCurrentUserId(): Promise<string> {
    const response = await fetch(`${API_BASE_URL}/users/get-current-user`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error("Failed to fetch current user");
    }

    const data = await response.json();
    return data.user.id;
  },

  async getArchivedStories(limit: number = 20, offset: number = 0): Promise<Story[]> {
    const response = await fetch(`${API_BASE_URL}/stories/archived?limit=${limit}&offset=${offset}`, {
      headers: getAuthHeaders(),
    });

    if (!response.ok) {
      throw new Error("Failed to fetch archived stories");
    }

    const data = await response.json();
    return data.stories || [];
  },
};
