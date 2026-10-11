<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import SideBar from "@/components/SideBar.vue";
import ChatWindow from "@/components/ChatWindow.vue";
import { messageService, type Conversation, type User } from "@/services/messageService";
import { usePhoto } from "@/composables/usePhoto";
import { useEventBus } from "@/composables/useEventBus";

const { loadPhotoUrl } = usePhoto();
const { on, off } = useEventBus();

const conversations = ref<Conversation[]>([]);
const selectedConversation = ref<Conversation | null>(null);
const searchQuery = ref("");
const isSearching = ref(false);
const currentUserId = ref("");
const showNewChatModal = ref(false);
const newChatQuery = ref("");
const newChatUsers = ref<User[]>([]);
const selectedNewChatUsers = ref<string[]>([]);
const newChatGroupName = ref("");
const isCreatingChat = ref(false);
let searchTimeout: number | null = null;

const filteredConversations = computed(() => {
  if (!searchQuery.value.trim()) {
    return conversations.value;
  }
  return conversations.value;
});

const loadConversations = async () => {
  try {
    const results = await messageService.getConversations();
    const convList = Array.isArray(results) ? results : Object.values(results || {});
    const cachedParticipants = JSON.parse(localStorage.getItem("conversationParticipants") || "{}");
    
    conversations.value = await Promise.all(convList.map(async (conv: any) => {
      const convId = conv.id || conv.ID;
      let participants = conv.participants || conv.Participants || [];
      
      if (!participants || participants.length === 0) {
        if (cachedParticipants[convId]) {
          participants = cachedParticipants[convId];
        } else {
          try {
            const detail = await messageService.getConversation(convId);
            participants = detail.conversation?.participants || detail.Participants || [];
          } catch (err) {
            console.error("Failed to fetch conversation details:", err);
          }
        }
      }
      
      participants = participants.map((p: any) => ({
        id: p.id || p.ID,
        username: p.username || p.Username || p.name || p.Name,
        email: p.email || p.Email,
        full_name: p.full_name || p.FullName || p.name || p.Name,
        profile_picture: p.profile_picture || p.ProfilePicture || p.picture || p.Picture
      }));
      
      if (participants && participants.length > 0) {
        cachedParticipants[convId] = participants;
      }
      
      const lastMsg = conv.last_message || conv.LastMessage;
      const lastMsgObj = lastMsg ? {
        id: lastMsg.id || lastMsg.ID || "",
        content: lastMsg.content || lastMsg.Content || "",
        sender_id: lastMsg.sender_id || lastMsg.SenderId || lastMsg.SenderID || "",
        created_at: lastMsg.created_at || lastMsg.CreatedAt || lastMsg.created || "",
        sender: lastMsg.sender ? {
          id: lastMsg.sender.id || lastMsg.sender.ID || "",
          username: lastMsg.sender.username || lastMsg.sender.Username || "",
          profile_picture: lastMsg.sender.profile_picture || lastMsg.sender.ProfilePicture || ""
        } : participants.find((p: any) => p.id === (lastMsg.sender_id || lastMsg.SenderId || lastMsg.SenderID))
      } : undefined;
      
      const unreadCount = conv.unread_count !== undefined ? conv.unread_count : (conv.UnreadCount !== undefined ? conv.UnreadCount : 0);
      console.log("Conversation", convId, "unread_count:", unreadCount);
      
      return {
        id: convId,
        participants: participants,
        is_group: conv.is_group !== undefined ? conv.is_group : conv.IsGroup,
        group_name: conv.group_name || conv.GroupName || "",
        last_message: lastMsgObj,
        unread_count: unreadCount,
        updated_at: conv.updated_at || conv.UpdatedAt || conv.CreatedAt || new Date().toISOString()
      };
    }));
    
    localStorage.setItem("conversationParticipants", JSON.stringify(cachedParticipants));
  } catch (error) {
    console.error("Failed to load conversations:", error);
  }
};

const handleSearch = (query: string) => {
  searchQuery.value = query;
  
  if (searchTimeout) {
    clearTimeout(searchTimeout);
  }

  if (!query.trim()) {
    loadConversations();
    return;
  }

  isSearching.value = true;
  searchTimeout = window.setTimeout(async () => {
    try {
      conversations.value = await messageService.searchConversations(query);
    } catch (error) {
      console.error("Search failed:", error);
    } finally {
      isSearching.value = false;
    }
  }, 500);
};

const selectConversation = async (conversation: Conversation) => {
  selectedConversation.value = conversation;
  try {
    await messageService.markAsSeen(conversation.id);
    const conv = conversations.value.find(c => c.id === conversation.id);
    if (conv) {
      conv.unread_count = 0;
    }
  } catch (error) {
    console.error("Failed to mark as seen:", error);
  }
};

const handleDeleteConversation = async (conversationId: string) => {
  try {
    await messageService.deleteConversation(conversationId);
    conversations.value = conversations.value.filter(c => c.id !== conversationId);
    if (selectedConversation.value?.id === conversationId) {
      selectedConversation.value = null;
    }
  } catch (error) {
    console.error("Failed to delete conversation:", error);
  }
};

const getConversationName = (conversation: Conversation): string => {
  if (conversation.is_group) {
    return conversation.group_name || "Group Chat";
  }
  const otherParticipant = conversation.participants?.find(p => p.id !== currentUserId.value);
  return otherParticipant?.username || "Unknown User";
};

const getConversationAvatar = (conversation: Conversation): string => {
  if (conversation.is_group || conversation.IsGroup) {
    return "";
  }
  const otherParticipant = conversation.participants?.find(p => p.id !== currentUserId.value);
  const pic = otherParticipant?.profile_picture || (otherParticipant as any)?.ProfilePicture || (otherParticipant as any)?.picture || (otherParticipant as any)?.Picture;
  return loadPhotoUrl(pic) || "";
};

const formatTimestamp = (timestamp: string): string => {
  const date = new Date(timestamp);
  const now = new Date();
  const diffInMs = now.getTime() - date.getTime();
  const diffInHours = diffInMs / (1000 * 60 * 60);

  if (diffInHours < 24) {
    return date.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" });
  } else if (diffInHours < 168) {
    return date.toLocaleDateString("en-US", { weekday: "short" });
  } else {
    return date.toLocaleDateString("en-US", { month: "short", day: "numeric" });
  }
};

const getUserIdFromToken = (): string => {
  const token = localStorage.getItem("token");
  if (!token) return "";
  try {
    const payload = JSON.parse(atob(token.split(".")[1]));
    return payload.user_id || payload.sub || "";
  } catch {
    return "";
  }
};

const searchNewChatUsers = async () => {
  if (!newChatQuery.value.trim()) {
    newChatUsers.value = [];
    return;
  }

  try {
    const results = await messageService.searchUsers(newChatQuery.value);
    
    let userArray: any[] = [];
    
    if (Array.isArray(results)) {
      userArray = results;
    } else if (results && typeof results === "object") {
      userArray = (results as any).users || (results as any).Users || Object.values(results);
    }
    
    newChatUsers.value = userArray
      .filter((user: any) => (user.id ) !== currentUserId.value)
      .map((user: any) => ({
        id: user.id || "",
        username: user.username || user.Username || user.name || user.Name || "",
        email: user.email || user.Email || "",
        full_name: user.full_name || user.FullName || user.name || user.Name || "",
        profile_picture: user.profile_picture || user.ProfilePicture || user.picture || user.Picture || ""
      }));
  } catch (error) {
    console.error("Failed to search users:", error);
    newChatUsers.value = [];
    alert("Search failed: " + (error instanceof Error ? error.message : "Unknown error"));
  }
};

const toggleUserSelection = (userId: string) => {
  const index = selectedNewChatUsers.value.indexOf(userId);
  if (index >= 0) {
    selectedNewChatUsers.value.splice(index, 1);
  } else {
    selectedNewChatUsers.value.push(userId);
  }
};

const createNewChat = async () => {
  if (selectedNewChatUsers.value.length === 0) return;

  isCreatingChat.value = true;
  try {
    const isGroup = selectedNewChatUsers.value.length > 1;
    const conversation = await messageService.createConversation(
      selectedNewChatUsers.value,
      isGroup ? newChatGroupName.value || "New Group" : undefined
    );
    
    const selectedUsersData = newChatUsers.value.filter(u => selectedNewChatUsers.value.includes(u.id));
    const convId = conversation.id || conversation.ID;
    
    const mapped = {
      id: convId,
      participants: selectedUsersData.length > 0 ? selectedUsersData : [],
      is_group: conversation.is_group !== undefined ? conversation.is_group : conversation.IsGroup,
      group_name: conversation.group_name || conversation.GroupName || "",
      last_message: conversation.last_message || conversation.LastMessage,
      unread_count: conversation.unread_count || conversation.UnreadCount || 0,
      updated_at: conversation.updated_at || conversation.UpdatedAt || conversation.CreatedAt || new Date().toISOString()
    };
    
    const cachedParticipants = JSON.parse(localStorage.getItem("conversationParticipants") || "{}");
    cachedParticipants[convId] = selectedUsersData;
    localStorage.setItem("conversationParticipants", JSON.stringify(cachedParticipants));
    
    conversations.value.unshift(mapped);
    selectedConversation.value = mapped;
    
    showNewChatModal.value = false;
    newChatQuery.value = "";
    newChatUsers.value = [];
    selectedNewChatUsers.value = [];
    newChatGroupName.value = "";
  } catch (error) {
    console.error("Failed to create conversation:", error);
    alert("Failed to create conversation");
  } finally {
    isCreatingChat.value = false;
  }
};

const closeNewChatModal = () => {
  showNewChatModal.value = false;
  newChatQuery.value = "";
  newChatUsers.value = [];
  selectedNewChatUsers.value = [];
  newChatGroupName.value = "";
};

const getUserDisplayName = (user: User): string => {
  return user.username || user.full_name || "User";
};

const updateConversationUnreadCount = (conversationId: string, increment: boolean = true) => {
  const conv = conversations.value.find(c => c.id === conversationId);
  if (conv) {
    if (increment) {
      conv.unread_count = (conv.unread_count || 0) + 1;
    } else {
      conv.unread_count = 0;
    }
    console.log("Updated unread count for", conversationId, "to", conv.unread_count);
  }
};

const handleMessageReceived = (conversationId: string, senderId: string) => {
  console.log("Message received in conversation", conversationId, "from", senderId);
};

const handleConversationUpdated = async () => {
  if (selectedConversation.value) {
    try {
      const detail = await messageService.getConversation(selectedConversation.value.id);
      const messages = detail.messages || [];
      
      if (messages.length > 0) {
        const lastMsg = messages[messages.length - 1];
        const conv = conversations.value.find(c => c.id === selectedConversation.value!.id);
        
        if (conv) {
          conv.last_message = {
            id: lastMsg.id || lastMsg.ID,
            content: lastMsg.content || lastMsg.Content || "",
            sender_id: lastMsg.sender_id || lastMsg.SenderID || lastMsg.sender?.id,
            created_at: lastMsg.created_at || lastMsg.CreatedAt || new Date().toISOString(),
            sender: lastMsg.sender
          };
          conv.updated_at = lastMsg.created_at || lastMsg.CreatedAt || new Date().toISOString();
          
          conversations.value = conversations.value.sort((a, b) => 
            new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
          );
        }
      }
    } catch (error) {
      console.error("Failed to update conversation:", error);
    }
  }
};

onMounted(() => {
  currentUserId.value = getUserIdFromToken();
  loadConversations();
  on('conversationUpdated', handleConversationUpdated);
});

onUnmounted(() => {
  if (searchTimeout) {
    clearTimeout(searchTimeout);
  }
  off('conversationUpdated', handleConversationUpdated);
});
</script>

<template>
  <SideBar />
  <div class="messages-container">
    <div class="message-list">
      <div class="message-list-header">
        <h2>Messages</h2>
        <button
          class="new-message-btn"
          @click="showNewChatModal = true"
        >
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
          >
            <path
              d="M21 12H3M12 3v18"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
            />
          </svg>
        </button>
      </div>
      
      <div class="search-box">
        <svg
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
        >
          <circle
            cx="11"
            cy="11"
            r="8"
            stroke="currentColor"
            stroke-width="2"
          />
          <path
            d="m21 21-4.35-4.35"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
          />
        </svg>
        <input
          type="text" 
          placeholder="Search conversations..."
          :value="searchQuery"
          @input="handleSearch(($event.target as HTMLInputElement).value)"
        />
      </div>

      <div class="conversations-list">
        <div
          v-for="conversation in filteredConversations"
          :key="conversation.id"
          class="conversation-item"
          :class="{ active: selectedConversation?.id === conversation.id, 'has-unread': conversation.unread_count > 0 }"
        >
          <div
            class="conversation-main"
            @click="selectConversation(conversation)"
          >
            <div class="avatar-wrapper">
              <div class="avatar">
                <img
                  v-if="getConversationAvatar(conversation)"
                  :src="getConversationAvatar(conversation)"
                  :alt="getConversationName(conversation)"
                />
                <div
                  v-else
                  class="avatar-placeholder"
                >
                  {{ getConversationName(conversation).charAt(0).toUpperCase() }}
                </div>
              </div>
              <span
                v-if="conversation.unread_count > 0"
                class="unread-count-badge"
              >
                {{ conversation.unread_count > 99 ? '99+' : conversation.unread_count }}
              </span>
            </div>
            <div class="conversation-info">
              <div class="conversation-header">
                <span class="username">{{ getConversationName(conversation) }}</span>
                <span class="timestamp">{{ formatTimestamp(conversation.updated_at) }}</span>
              </div>
              <div class="last-message">
                <span :class="{ unread: conversation.unread_count > 0 }">
                  <span v-if="conversation.last_message?.sender_id === currentUserId" class="sender-label">You: </span>
                  {{ conversation.last_message?.content || 'No messages yet' }}
                </span>
                <span
                  v-if="conversation.unread_count > 0"
                  class="unread-badge"
                >
                  {{ conversation.unread_count > 99 ? '99+' : conversation.unread_count }}
                </span>
              </div>
            </div>
          </div>
          <button
            class="delete-conversation-btn"
            title="Delete conversation"
            @click.stop="handleDeleteConversation(conversation.id)"
          >
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
            >
              <path
                d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2M10 11v6M14 11v6"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
              />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <div class="chat-area">
      <ChatWindow
        v-if="selectedConversation"
        :conversation="selectedConversation"
        :current-user-id="currentUserId"
        @conversation-deleted="handleDeleteConversation"
        @conversation-updated="loadConversations"
        @message-received="handleMessageReceived"
      />
      <div
        v-else
        class="no-chat-selected"
      >
        <svg
          width="96"
          height="96"
          viewBox="0 0 24 24"
          fill="none"
        >
          <path
            d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"
            stroke="currentColor"
            stroke-width="2"
          />
        </svg>
        <h3>Your Messages</h3>
        <p>Send private messages to a friend or group.</p>
      </div>
    </div>

    <div
      v-if="showNewChatModal"
      class="modal-overlay"
      @click="closeNewChatModal"
    >
      <div
        class="new-chat-modal"
        @click.stop
      >
        <div class="modal-header">
          <h3>New Message</h3>
          <button
            class="close-btn"
            @click="closeNewChatModal"
          >
            ×
          </button>
        </div>
        <div class="modal-content">
          <input
            v-model="newChatQuery"
            type="text"
            placeholder="Search users..."
            class="search-input"
            @input="searchNewChatUsers"
          />
          
          <div
            v-if="selectedNewChatUsers.length > 1"
            class="group-name-input-wrapper"
          >
            <input
              v-model="newChatGroupName"
              type="text"
              placeholder="Group name (optional)..."
              class="group-name-input"
            />
          </div>

          <div
            v-if="selectedNewChatUsers.length > 0"
            class="selected-users"
          >
            <div class="selected-label">
              Selected ({{ selectedNewChatUsers.length }})
            </div>
            <div class="selected-tags">
              <div
                v-for="userId in selectedNewChatUsers"
                :key="userId"
                class="user-tag"
              >
                <span>{{ Array.isArray(newChatUsers) && newChatUsers.find(u => u.id === userId)?.username || Array.isArray(newChatUsers) && newChatUsers.find(u => u.id === userId)?.full_name || 'User' }}</span>
                <button
                  class="tag-remove"
                  @click="toggleUserSelection(userId)"
                >
                  ×
                </button>
              </div>
            </div>
          </div>

          <div class="users-list">
            <div
              v-for="user in newChatUsers"
              v-if="Array.isArray(newChatUsers) && newChatUsers.length > 0"
              :key="user.id"
              class="user-item"
              :class="{ selected: selectedNewChatUsers.includes(user.id) }"
              @click="toggleUserSelection(user.id)"
            >
              <img
                v-if="user.profile_picture"
                :src="loadPhotoUrl(user.profile_picture)"
                :alt="user.username || 'User'"
                class="user-avatar"
              />
              <div
                v-else
                class="avatar-placeholder-small"
              >
                {{ (user.username || user.full_name || 'U').charAt(0).toUpperCase() }}
              </div>
              <div class="user-info">
                <div class="user-name">
                  {{ user.username || user.full_name || 'Unknown User' }}
                </div>
                <div
                  v-if="user.full_name && user.full_name !== user.username"
                  class="user-full-name"
                >
                  {{ user.full_name }}
                </div>
              </div>
              <div
                v-if="selectedNewChatUsers.includes(user.id)"
                class="check-icon"
              >
                ✓
              </div>
            </div>
            <div
              v-else-if="newChatQuery && !Array.isArray(newChatUsers)"
              class="no-results"
            >
              No users found
            </div>
          </div>

          <button
            :disabled="selectedNewChatUsers.length === 0 || isCreatingChat"
            class="create-chat-btn"
            @click="createNewChat"
          >
            {{ isCreatingChat ? 'Creating...' : 'Start Chat' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: flex-end;
  z-index: 1000;
}

.new-chat-modal {
  background: #000;
  width: 100%;
  max-width: 500px;
  max-height: 80vh;
  border-radius: 16px 16px 0 0;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #262626;
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  color: #fff;
  font-size: 32px;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}

.modal-content {
  padding: 16px 24px;
  overflow-y: auto;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.search-input {
  width: 100%;
  background: #262626;
  border: 1px solid #363636;
  border-radius: 8px;
  padding: 12px;
  color: #fff;
  font-size: 14px;
  outline: none;
}

.group-name-input-wrapper {
  display: flex;
}

.group-name-input {
  width: 100%;
  background: #262626;
  border: 1px solid #363636;
  border-radius: 8px;
  padding: 12px;
  color: #fff;
  font-size: 14px;
  outline: none;
}

.selected-users {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.selected-label {
  font-size: 12px;
  color: #8e8e8e;
  font-weight: 500;
}

.selected-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.user-tag {
  background: #262626;
  border-radius: 20px;
  padding: 6px 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.tag-remove {
  background: none;
  border: none;
  color: #fff;
  cursor: pointer;
  font-size: 16px;
  padding: 0;
  line-height: 1;
}

.users-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.no-results {
  text-align: center;
  padding: 24px;
  color: #8e8e8e;
  font-size: 14px;
}

.user-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #1a1a1a;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.user-item:hover {
  background: #262626;
}

.user-item.selected {
  background: #0095f6;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
}

.avatar-placeholder-small {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(45deg, #f09433 0%, #e6683c 25%, #dc2743 50%, #cc2366 75%, #bc1888 100%);
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  flex-shrink: 0;
}

.user-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
}

.user-full-name {
  font-size: 12px;
  color: #8e8e8e;
}

.check-icon {
  font-size: 18px;
  font-weight: 600;
}

.create-chat-btn {
  background: #0095f6;
  border: none;
  color: #fff;
  padding: 12px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  margin-top: 8px;
}

.create-chat-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.create-chat-btn:not(:disabled):hover {
  background: #0085d4;
}

.messages-container {
  display: flex;
  position: fixed;
  left: 250px;
  right: 0;
  top: 0;
  height: 100vh;
  background-color: #000;
  color: #fff;
}

.message-list {
  width: 400px;
  border-right: 1px solid #262626;
  display: flex;
  flex-direction: column;
  background-color: #000;
}

.message-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 20px;
  border-bottom: 1px solid #262626;
}

.message-list-header h2 {
  font-size: 24px;
  font-weight: 600;
  margin: 0;
}

.new-message-btn {
  background: none;
  border: none;
  color: #fff;
  cursor: pointer;
  padding: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity 0.2s;
}

.new-message-btn:hover {
  opacity: 0.7;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  margin: 16px;
  background-color: #262626;
  border-radius: 8px;
}

.search-box svg {
  color: #8e8e8e;
}

.search-box input {
  flex: 1;
  background: none;
  border: none;
  color: #fff;
  font-size: 14px;
  outline: none;
}

.search-box input::placeholder {
  color: #8e8e8e;
}

.conversations-list {
  flex: 1;
  overflow-y: auto;
}

.conversation-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  transition: background-color 0.2s;
  position: relative;
}

.conversation-main {
  display: flex;
  gap: 12px;
  flex: 1;
  cursor: pointer;
}

.conversation-item:hover {
  background-color: #1a1a1a;
}

.conversation-item.active {
  background-color: #262626;
}

.delete-conversation-btn {
  background: none;
  border: none;
  color: #ed4956;
  cursor: pointer;
  padding: 8px;
  opacity: 0;
  transition: opacity 0.2s;
  flex-shrink: 0;
}

.conversation-item:hover .delete-conversation-btn {
  opacity: 1;
}

.delete-conversation-btn:hover {
  opacity: 1 !important;
  color: #ff1744;
}

.avatar-wrapper {
  position: relative;
  flex-shrink: 0;
}

.avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
}

.unread-count-badge {
  position: absolute;
  top: -2px;
  right: -2px;
  background-color: #0095f6;
  color: #fff;
  border-radius: 12px;
  padding: 2px 6px;
  font-size: 11px;
  font-weight: 600;
  min-width: 20px;
  text-align: center;
  border: 2px solid #000;
  z-index: 10;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(45deg, #f09433 0%, #e6683c 25%, #dc2743 50%, #cc2366 75%, #bc1888 100%);
  font-size: 24px;
  font-weight: 600;
  color: #fff;
}

.conversation-info {
  flex: 1;
  min-width: 0;
}

.conversation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.username {
  font-weight: 600;
  font-size: 14px;
}

.timestamp {
  font-size: 12px;
  color: #8e8e8e;
}

.last-message {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
  color: #8e8e8e;
}

.last-message span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.last-message .unread {
  color: #fff;
  font-weight: 500;
}

.sender-label {
  color: #8e8e8e;
  font-weight: 400;
}

.unread-badge {
  background-color: #0095f6;
  color: #fff;
  border-radius: 12px;
  padding: 2px 8px;
  font-size: 12px;
  font-weight: 600;
  margin-left: 8px;
}

.chat-area {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.no-chat-selected {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #8e8e8e;
}

.no-chat-selected svg {
  margin-bottom: 16px;
  opacity: 0.5;
}

.no-chat-selected h3 {
  font-size: 22px;
  font-weight: 300;
  margin: 0 0 8px 0;
  color: #fff;
}

.no-chat-selected p {
  font-size: 14px;
  margin: 0;
}
</style>