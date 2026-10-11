<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from "vue";
import { messageService, type Conversation, type Message, type User } from "@/services/messageService";
import { usePhoto } from "@/composables/usePhoto";

const { loadPhotoUrl } = usePhoto();

interface Props {
  conversation: Conversation;
  currentUserId: string;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  conversationDeleted: [conversationId: string];
  conversationUpdated: [];
  messageReceived: [conversationId: string, senderId: string];
}>();

const messages = ref<Message[]>([]);
const messageText = ref("");
const messagesContainer = ref<HTMLElement | null>(null);
const showOptions = ref(false);
const showParticipantsModal = ref(false);
const showSearchModal = ref(false);
const searchQuery = ref("");
const searchResults = ref<Message[]>([]);
const selectedFile = ref<File | null>(null);
const filePreview = ref<string>("");
const showUserSearch = ref(false);
const userSearchQuery = ref("");
const searchedUsers = ref<User[]>([]);
const selectedUsers = ref<string[]>([]);
const ws = ref<WebSocket | null>(null);
let reconnectTimeout: number | null = null;

const otherParticipants = computed(() => {
  if (!props.conversation.participants) return [];
  return props.conversation.participants.filter(p => p.id !== props.currentUserId);
});

const conversationName = computed(() => {
  if (props.conversation.is_group || props.conversation.IsGroup) {
    return props.conversation.group_name || props.conversation.GroupName || "Group Chat";
  }
  return otherParticipants.value[0]?.username || "Unknown User";
});

const conversationAvatar = computed(() => {
  if (props.conversation.is_group || props.conversation.IsGroup) {
    return "";
  }
  return loadPhotoUrl(otherParticipants.value[0]?.profile_picture) || "";
});

const loadMessages = async () => {
  try {
    const data = await messageService.getConversation(props.conversation.id);
    messages.value = (data.messages || []).map((msg: any) => {
      const senderId = msg.sender_id || msg.SenderID || msg.sender?.id || msg.sender?.ID;
      let senderData = msg.sender || msg.Sender;
      
      if (!senderData || !senderData.username) {
        const participant = props.conversation.participants?.find(p => p.id === senderId);
        if (participant) {
          senderData = participant;
        } else {
          senderData = {
            id: senderId,
            username: "Unknown User",
            email: "",
            full_name: "",
            profile_picture: ""
          };
        }
      }
      
      return {
        id: msg.id || msg.ID,
        conversation_id: msg.conversation_id || msg.ConversationID || props.conversation.id,
        sender_id: senderId,
        sender: {
          id: senderData.id || senderData.ID,
          username: senderData.username || senderData.Username,
          email: senderData.email || senderData.Email || "",
          full_name: senderData.full_name || senderData.FullName || senderData.username || senderData.Username || "",
          profile_picture: senderData.profile_picture || senderData.ProfilePicture || ""
        },
        content: msg.content || msg.Content || "",
        media_url: msg.media_url || msg.MediaURL,
        media_type: msg.media_type || msg.MediaType,
        created_at: msg.created_at || msg.CreatedAt || new Date().toISOString(),
        status: msg.status || msg.Status || "sent",
        can_unsend: msg.can_unsend !== undefined ? msg.can_unsend : true
      };
    }).sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
    await nextTick();
    scrollToBottom();
  } catch (error) {
    console.error("Failed to load messages:", error);
  }
};

const connectWebSocket = () => {
  if (ws.value) {
    ws.value.close();
  }

  try {
    ws.value = messageService.connectWebSocket(props.conversation.id, handleNewMessage);

    ws.value.onerror = () => {
      console.error("WebSocket error");
      scheduleReconnect();
    };

    ws.value.onclose = () => {
      scheduleReconnect();
    };
  } catch (error) {
    console.error("Failed to connect WebSocket:", error);
  }
};

const scheduleReconnect = () => {
  if (reconnectTimeout) {
    clearTimeout(reconnectTimeout);
  }
  reconnectTimeout = window.setTimeout(() => {
    connectWebSocket();
  }, 3000);
};

const handleNewMessage = (message: Message) => {
  const senderId = message.sender_id || message.SenderID || message.sender?.id || message.sender?.ID;
  let senderData = message.sender || message.Sender;
  
  if (!senderData || !senderData.username) {
    const participant = props.conversation.participants?.find(p => p.id === senderId);
    if (participant) {
      senderData = participant;
    } else {
      senderData = {
        id: senderId,
        username: "Unknown User",
        email: "",
        full_name: "",
        profile_picture: ""
      };
    }
  }

  const mappedMessage = {
    id: message.id || message.ID,
    conversation_id: message.conversation_id || message.ConversationID || props.conversation.id,
    sender_id: senderId,
    sender: {
      id: senderData.id || senderData.ID,
      username: senderData.username || senderData.Username,
      email: senderData.email || senderData.Email || "",
      full_name: senderData.full_name || senderData.FullName || senderData.username || senderData.Username || "",
      profile_picture: loadPhotoUrl(senderData.profile_picture) || loadPhotoUrl(senderData.ProfilePicture) || ""
    },
    content: message.content || message.Content || "",
    media_url: message.media_url || message.MediaURL || message.media_url,
    media_type: message.media_type || message.MediaType,
    created_at: message.created_at || message.CreatedAt || new Date().toISOString(),
    status: message.status || message.Status || "sent",
    can_unsend: message.can_unsend !== undefined ? message.can_unsend : true
  };

  const existingIndex = messages.value.findIndex(m => m.id === mappedMessage.id);
  if (existingIndex >= 0) {
    messages.value[existingIndex] = mappedMessage;
  } else {
    messages.value.push(mappedMessage);
  }
  nextTick(() => scrollToBottom());
  
  if (senderId !== props.currentUserId) {
    emit("messageReceived", props.conversation.id, senderId);
  }
  
  emit("conversationUpdated");
};

const sendMessage = async () => {
  if (!messageText.value.trim() && !selectedFile.value) return;

  try {
    const mediaType = selectedFile.value 
      ? (selectedFile.value.type.startsWith("image/") 
        ? (selectedFile.value.type === "image/gif" ? "gif" : "image")
        : "video")
      : undefined;

    await messageService.sendMessage(
      props.conversation.id,
      messageText.value,
      selectedFile.value || undefined,
      mediaType
    );

    messageText.value = "";
    selectedFile.value = null;
    filePreview.value = "";
    
    emit('conversationUpdated');
  } catch (error) {
    console.error("Failed to send message:", error);
  }
};

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file) return;

  const validTypes = ["image/jpeg", "image/png", "image/gif", "video/mp4", "video/webm"];
  if (!validTypes.includes(file.type)) {
    alert("Please select a valid image or video file");
    return;
  }

  selectedFile.value = file;
  const reader = new FileReader();
  reader.onload = (e) => {
    filePreview.value = e.target?.result as string;
  };
  reader.readAsDataURL(file);
};

const removeFile = () => {
  selectedFile.value = null;
  filePreview.value = "";
};

const unsendMessage = async (messageId: string) => {
  try {
    await messageService.unsendMessage(props.conversation.id, messageId);
    messages.value = messages.value.filter(m => m.id !== messageId);
  } catch (error) {
    console.error("Failed to unsend message:", error);
  }
};

const deleteConversation = async () => {
  if (!confirm("Are you sure you want to delete this conversation?")) return;
  emit("conversationDeleted", props.conversation.id);
};

const searchInChat = async () => {
  if (!searchQuery.value.trim()) {
    searchResults.value = [];
    return;
  }

  try {
    searchResults.value = await messageService.searchMessages(
      props.conversation.id,
      searchQuery.value
    );
  } catch (error) {
    console.error("Failed to search messages:", error);
  }
};

const scrollToMessage = (messageId: string) => {
  const element = document.getElementById(`message-${messageId}`);
  if (element) {
    element.scrollIntoView({ behavior: "smooth", block: "center" });
    element.classList.add("highlight");
    setTimeout(() => element.classList.remove("highlight"), 2000);
  }
  showSearchModal.value = false;
};

const searchUsers = async () => {
  if (!userSearchQuery.value.trim()) {
    searchedUsers.value = [];
    return;
  }

  try {
    searchedUsers.value = await messageService.searchUsers(userSearchQuery.value);
  } catch (error) {
    console.error("Failed to search users:", error);
  }
};

const toggleUserSelection = (userId: string) => {
  const index = selectedUsers.value.indexOf(userId);
  if (index >= 0) {
    selectedUsers.value.splice(index, 1);
  } else {
    selectedUsers.value.push(userId);
  }
};

const addParticipants = async () => {
  if (selectedUsers.value.length === 0) return;

  try {
    await messageService.addParticipants(props.conversation.id, selectedUsers.value);
    emit("conversationUpdated");
    showUserSearch.value = false;
    selectedUsers.value = [];
    userSearchQuery.value = "";
  } catch (error) {
    console.error("Failed to add participants:", error);
  }
};

const removeParticipant = async (participantId: string) => {
  if (!confirm("Remove this participant from the group?")) return;

  try {
    await messageService.removeParticipant(props.conversation.id, participantId);
    emit("conversationUpdated");
  } catch (error) {
    console.error("Failed to remove participant:", error);
  }
};

const scrollToBottom = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
  }
};

const formatMessageTime = (timestamp: string): string => {
  if (!timestamp) return "";
  const date = new Date(timestamp);
  if (isNaN(date.getTime())) return "";
  return date.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit" });
};

const canUnsend = (message: Message): boolean => {
  const sentTime = new Date(message.created_at).getTime();
  const now = Date.now();
  return message.sender_id === props.currentUserId && (now - sentTime) < 60000;
};

const initiateCall = (type: "audio" | "video") => {
  
  alert(`${type === "audio" ? "Audio" : "Video"} call feature coming soon!`);
};

watch(() => props.conversation.id, () => {
  loadMessages();
  connectWebSocket();
});

onMounted(() => {
  loadMessages();
  connectWebSocket();
});

onUnmounted(() => {
  if (ws.value) {
    ws.value.close();
  }
  if (reconnectTimeout) {
    clearTimeout(reconnectTimeout);
  }
});
</script>

<template>
  <div class="chat-window">
    <div class="chat-header">
      <div class="header-left">
        <div class="avatar">
          <img
            v-if="conversationAvatar"
            :src="conversationAvatar"
            :alt="conversationName"
          />
          <div
            v-else
            class="avatar-placeholder"
          >
            {{ conversationName.charAt(0).toUpperCase() }}
          </div>
        </div>
        <div class="header-info">
          <h3>{{ conversationName }}</h3>
          <span
            v-if="conversation.is_group"
            class="participants-count"
          >
            {{ conversation.participants.length }} participants
          </span>
        </div>
      </div>
      <div class="header-actions">
        <button
          class="icon-btn"
          title="Audio Call"
          @click="initiateCall('audio')"
        >
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
          >
            <path
              d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"
              stroke="currentColor"
              stroke-width="2"
            />
          </svg>
        </button>
        <button
          class="icon-btn"
          title="Video Call"
          @click="initiateCall('video')"
        >
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
          >
            <path
              d="M23 7l-7 5 7 5V7z"
              stroke="currentColor"
              stroke-width="2"
              stroke-linejoin="round"
            />
            <rect
              x="1"
              y="5"
              width="15"
              height="14"
              rx="2"
              stroke="currentColor"
              stroke-width="2"
            />
          </svg>
        </button>
        <button
          class="icon-btn"
          title="Search in chat"
          @click="showSearchModal = true"
        >
          <svg
            width="24"
            height="24"
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
        </button>
        <button
          v-if="conversation.is_group"
          class="icon-btn"
          title="Edit group"
          @click="showUserSearch = true"
        >
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
          >
            <path
              d="M17 3a2.83 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"
              stroke="currentColor"
              stroke-width="2"
            />
          </svg>
        </button>
        <button
          class="icon-btn"
          title="Options"
          @click="showOptions = !showOptions"
        >
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
          >
            <circle
              cx="12"
              cy="12"
              r="1"
              fill="currentColor"
            />
            <circle
              cx="12"
              cy="5"
              r="1"
              fill="currentColor"
            />
            <circle
              cx="12"
              cy="19"
              r="1"
              fill="currentColor"
            />
          </svg>
        </button>
        <div
          v-if="showOptions"
          class="options-menu"
        >
          <button @click="showParticipantsModal = true; showOptions = false">
            View Participants
          </button>
          <button
            class="danger"
            @click="deleteConversation(); showOptions = false"
          >
            Delete Conversation
          </button>
        </div>
      </div>
    </div>

    <div
      ref="messagesContainer"
      class="messages-container"
    >
      <div
        v-for="message in messages"
        :id="`message-${message.id}`"
        :key="message.id"
        class="message"
        :class="{ sent: message.sender_id === currentUserId, received: message.sender_id !== currentUserId }"
      >
        <div
          v-if="message.sender_id !== currentUserId"
          class="message-avatar"
        >
          <img
            v-if="message.sender?.profile_picture"
            :src="loadPhotoUrl(message.sender.profile_picture)"
            :alt="message.sender.username"
          />
          <div
            v-else
            class="avatar-placeholder-small"
          >
            {{ message.sender?.username?.charAt(0).toUpperCase() || 'U' }}
          </div>
        </div>
        <div class="message-content">
          <div
            v-if="message.sender_id !== currentUserId && conversation.is_group"
            class="sender-name"
          >
            {{ message.sender?.username }}
          </div>
          <div
            v-if="message.media_url"
            class="message-media"
          >
            <img
              v-if="message.media_type === 'image' || message.media_type === 'gif'"
              :src="loadPhotoUrl(message.media_url)"
              alt="media"
            />
            <video
              v-else-if="message.media_type === 'video'"
              :src="message.media_url"
              controls
            />
          </div>
          <div
            v-if="message.content"
            class="message-bubble"
          >
            {{ message.content }}
          </div>
          <div class="message-meta">
            <span class="message-time">{{ formatMessageTime(message.created_at) }}</span>
            <span
              v-if="message.sender_id === currentUserId"
              class="message-status"
            >
              {{ message.status === 'seen' ? 'Seen' : 'Sent' }}
            </span>
          </div>
          <button
            v-if="canUnsend(message)"
            class="unsend-btn"
            title="Unsend"
            @click="unsendMessage(message.id)"
          >
            Unsend
          </button>
        </div>
      </div>
    </div>

    <div class="message-input-area">
      <div
        v-if="filePreview"
        class="file-preview"
      >
        <img
          v-if="selectedFile?.type.startsWith('image/')"
          :src="filePreview"
          alt="preview"
        />
        <video
          v-else
          :src="filePreview"
        />
        <button
          class="remove-preview"
          @click="removeFile"
        >
          ×
        </button>
      </div>
      <div class="message-input">
        <label class="file-input-btn">
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
          >
            <rect
              x="3"
              y="3"
              width="18"
              height="18"
              rx="2"
              stroke="currentColor"
              stroke-width="2"
            />
            <circle
              cx="8.5"
              cy="8.5"
              r="1.5"
              fill="currentColor"
            />
            <path
              d="M21 15l-5-5L5 21"
              stroke="currentColor"
              stroke-width="2"
            />
          </svg>
          <input
            type="file"
            accept="image/*,video/*,.gif"
            hidden
            @change="handleFileSelect"
          />
        </label>
        <input
          v-model="messageText"
          type="text"
          placeholder="Message..."
          @keypress.enter="sendMessage"
        />
        <button
          :disabled="!messageText.trim() && !selectedFile"
          class="send-btn"
          @click="sendMessage"
        >
          Send
        </button>
      </div>
    </div>

    <div
      v-if="showParticipantsModal"
      class="modal-overlay"
      @click="showParticipantsModal = false"
    >
      <div
        class="modal"
        @click.stop
      >
        <div class="modal-header">
          <h3>Participants</h3>
          <button
            class="close-btn"
            @click="showParticipantsModal = false"
          >
            ×
          </button>
        </div>
        <div class="modal-content">
          <div
            v-for="participant in conversation.participants"
            :key="participant.id"
            class="participant-item"
          >
            <div class="participant-info">
              <img
                v-if="participant.profile_picture"
                :src="loadPhotoUrl(participant.profile_picture)"
                :alt="participant.username"
              />
              <div
                v-else
                class="avatar-placeholder-small"
              >
                {{ participant.username.charAt(0).toUpperCase() }}
              </div>
              <span>{{ participant.username }}</span>
            </div>
            <button
              v-if="conversation.is_group && participant.id !== currentUserId"
              class="remove-btn"
              @click="removeParticipant(participant.id)"
            >
              Remove
            </button>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="showSearchModal"
      class="modal-overlay"
      @click="showSearchModal = false"
    >
      <div
        class="modal"
        @click.stop
      >
        <div class="modal-header">
          <h3>Search Messages</h3>
          <button
            class="close-btn"
            @click="showSearchModal = false"
          >
            ×
          </button>
        </div>
        <div class="modal-content">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search in conversation..."
            class="search-input"
            @input="searchInChat"
          />
          <div class="search-results">
            <div
              v-for="result in searchResults"
              :key="result.id"
              class="search-result-item"
              @click="scrollToMessage(result.id)"
            >
              <div class="result-sender">
                {{ result.sender?.username }}
              </div>
              <div class="result-content">
                {{ result.content }}
              </div>
              <div class="result-time">
                {{ formatMessageTime(result.created_at) }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="showUserSearch"
      class="modal-overlay"
      @click="showUserSearch = false"
    >
      <div
        class="modal"
        @click.stop
      >
        <div class="modal-header">
          <h3>Add Participants</h3>
          <button
            class="close-btn"
            @click="showUserSearch = false"
          >
            ×
          </button>
        </div>
        <div class="modal-content">
          <input
            v-model="userSearchQuery"
            type="text"
            placeholder="Search users..."
            class="search-input"
            @input="searchUsers"
          />
          <div class="user-results">
            <div
              v-for="user in searchedUsers"
              :key="user.id"
              class="user-item"
              :class="{ selected: selectedUsers.includes(user.id) }"
              @click="toggleUserSelection(user.id)"
            >
              <img
                v-if="user.profile_picture"
                :src="loadPhotoUrl(user.profile_picture)"
                :alt="user.username"
              />
              <div
                v-else
                class="avatar-placeholder-small"
              >
                {{ user.username }}
              </div>
              <span>{{ user.username }}</span>
            </div>
          </div>
          <button
            :disabled="selectedUsers.length === 0"
            class="add-btn"
            @click="addParticipants"
          >
            Add Selected ({{ selectedUsers.length }})
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-window {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: #000;
}

.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #262626;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  overflow: hidden;
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
  font-size: 18px;
  font-weight: 600;
  color: #fff;
}

.avatar-placeholder-small {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(45deg, #f09433 0%, #e6683c 25%, #dc2743 50%, #cc2366 75%, #bc1888 100%);
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}

.header-info h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.participants-count {
  font-size: 12px;
  color: #8e8e8e;
}

.header-actions {
  display: flex;
  gap: 16px;
  position: relative;
}

.icon-btn {
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

.icon-btn:hover {
  opacity: 0.7;
}

.options-menu {
  position: absolute;
  top: 100%;
  right: 0;
  background: #262626;
  border-radius: 8px;
  padding: 8px;
  min-width: 200px;
  z-index: 10;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
}

.options-menu button {
  width: 100%;
  background: none;
  border: none;
  color: #fff;
  padding: 12px 16px;
  text-align: left;
  cursor: pointer;
  border-radius: 4px;
  font-size: 14px;
}

.options-menu button:hover {
  background: #363636;
}

.options-menu button.danger {
  color: #ed4956;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 16px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message {
  display: flex;
  gap: 8px;
  max-width: 70%;
}

.message.sent {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.message.received {
  align-self: flex-start;
}

.message-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
}

.message-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.message-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
  position: relative;
}

.sender-name {
  font-size: 12px;
  color: #8e8e8e;
  font-weight: 500;
}

.message-media {
  max-width: 300px;
  border-radius: 12px;
  overflow: hidden;
}

.message-media img,
.message-media video {
  width: 100%;
  display: block;
}

.message-bubble {
  background: #262626;
  padding: 12px 16px;
  border-radius: 18px;
  font-size: 14px;
  word-wrap: break-word;
}

.message.sent .message-bubble {
  background: #0095f6;
}

.message-meta {
  display: flex;
  gap: 8px;
  font-size: 11px;
  color: #8e8e8e;
  padding: 0 4px;
}

.unsend-btn {
  background: none;
  border: none;
  color: #ed4956;
  font-size: 11px;
  cursor: pointer;
  padding: 4px 8px;
  margin-top: 4px;
  align-self: flex-start;
}

.unsend-btn:hover {
  text-decoration: underline;
}

.message.highlight {
  animation: highlight 1s ease;
}

@keyframes highlight {
  0%, 100% { background-color: transparent; }
  50% { background-color: rgba(255, 255, 255, 0.1); }
}

.message-input-area {
  border-top: 1px solid #262626;
  padding: 16px 24px;
}

.file-preview {
  position: relative;
  width: 100px;
  height: 100px;
  margin-bottom: 12px;
  border-radius: 8px;
  overflow: hidden;
}

.file-preview img,
.file-preview video {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.remove-preview {
  position: absolute;
  top: 4px;
  right: 4px;
  background: rgba(0, 0, 0, 0.7);
  border: none;
  color: #fff;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  cursor: pointer;
  font-size: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.message-input {
  display: flex;
  align-items: center;
  gap: 12px;
  background: #262626;
  border-radius: 24px;
  padding: 8px 16px;
}

.file-input-btn {
  cursor: pointer;
  display: flex;
  align-items: center;
  color: #fff;
}

.file-input-btn:hover {
  opacity: 0.7;
}

.message-input input {
  flex: 1;
  background: none;
  border: none;
  color: #fff;
  font-size: 14px;
  outline: none;
}

.message-input input::placeholder {
  color: #8e8e8e;
}

.send-btn {
  background: none;
  border: none;
  color: #0095f6;
  font-weight: 600;
  font-size: 14px;
  cursor: pointer;
}

.send-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.send-btn:not(:disabled):hover {
  color: #fff;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: #262626;
  border-radius: 12px;
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #363636;
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
}

.search-input {
  width: 100%;
  background: #000;
  border: 1px solid #363636;
  border-radius: 8px;
  padding: 12px;
  color: #fff;
  font-size: 14px;
  outline: none;
  margin-bottom: 16px;
}

.search-results,
.user-results {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.search-result-item {
  padding: 12px;
  background: #000;
  border-radius: 8px;
  cursor: pointer;
}

.search-result-item:hover {
  background: #1a1a1a;
}

.result-sender {
  font-size: 12px;
  color: #8e8e8e;
  font-weight: 500;
  margin-bottom: 4px;
}

.result-content {
  font-size: 14px;
  margin-bottom: 4px;
}

.result-time {
  font-size: 11px;
  color: #8e8e8e;
}

.participant-item,
.user-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: #000;
  border-radius: 8px;
  margin-bottom: 8px;
  cursor: pointer;
}

.user-item:hover {
  background: #1a1a1a;
}

.user-item.selected {
  background: #0095f6;
}

.participant-info,
.user-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.participant-info img,
.user-item img {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
}

.remove-btn,
.add-btn {
  background: #ed4956;
  border: none;
  color: #fff;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
}

.add-btn {
  background: #0095f6;
  width: 100%;
  margin-top: 16px;
}

.add-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.remove-btn:hover,
.add-btn:not(:disabled):hover {
  opacity: 0.8;
}
</style>