import { ref, computed } from "vue";
import { useSearch } from "./useOverlay";

const shareMode = ref(false);
const shareContent = ref<any>(null);
const selectedShareUsers = ref<any[]>([]);

export function useShare() {
  const { showSearch } = useSearch();

  const openShareMode = (content: any) => {
    shareContent.value = {
      id: content.ID || content.id,
      caption: content.Caption || content.caption || "",
      postMedia: content.postMedia || [],
      type: "post"
    };
    selectedShareUsers.value = [];
    shareMode.value = true;
    showSearch();
  };

  const closeShareMode = () => {
    shareMode.value = false;
    shareContent.value = null;
    selectedShareUsers.value = [];
  };

  const toggleUserSelection = (user: any) => {
    const index = selectedShareUsers.value.findIndex(u => u.id === user.id);
    if (index > -1) {
      selectedShareUsers.value.splice(index, 1);
    } else {
      selectedShareUsers.value.push(user);
    }
  };

  const sendContent = async () => {
    if (selectedShareUsers.value.length === 0 || !shareContent.value) return;

    try {
      const token = localStorage.getItem("token");
      if (!token) return;

      for (const user of selectedShareUsers.value) {
        const payload = {
          content: shareContent.value.caption || "",
          media: shareContent.value.postMedia || [],
          contentType: shareContent.value.type || "post",
          contentId: shareContent.value.id || ""
        };

        await fetch("https://localhost/api/v1/messages/send-content", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`
          },
          body: JSON.stringify({
            userId: user.id,
            ...payload
          })
        });
      }

      closeShareMode();
    } catch (error) {
      console.error("Failed to send content:", error);
    }
  };

  const selectedUserCount = computed(() => {
    return selectedShareUsers.value.length;
  });

  return {
    shareMode,
    shareContent,
    selectedShareUsers,
    selectedUserCount,
    openShareMode,
    closeShareMode,
    toggleUserSelection,
    sendContent
  };
}
