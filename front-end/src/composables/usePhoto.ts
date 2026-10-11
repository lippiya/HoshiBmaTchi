import { ref } from "vue";

export function usePost(){
  const photoUrl = ref<string>("");
  const loading = ref(false);
  const error = ref<string>("");

  const loadPhotoUrl = (url: string | null | undefined): string => {
    if (!url) {
      return "https://via.placeholder.com/150?text=No+Photo";
    }

    const photoPath = encodeURIComponent(url);
    return `https://localhost/api/v1/posts/photo?path=${photoPath}`;
  };

  const uploadPhoto = async (file: File): Promise<string | null> => {
    const token = localStorage.getItem("token");
    if (!token) {
      error.value = "Not authenticated";
      return null;
    }

    loading.value = true;
    error.value = "";

    try {
      const formData = new FormData();
      formData.append("photo", file);

      const response = await fetch("https://localhost/api/v1/users/upload-picture", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Upload failed");
      }

      photoUrl.value = data.url;
      return data.url;
    } catch {
      return null;
    } finally {
      loading.value = false;
    }
  };

  return {
    photoUrl,
    loading,
    error,
    loadPhotoUrl,
    uploadPhoto,
  };
}

export function usePhoto() {
  const photoUrl = ref<string>("");
  const loading = ref(false);
  const error = ref<string>("");

  const loadPhotoUrl = (url: string | null | undefined): string => {
    if (!url) {
      console.log("No URL provided for photo.");
      return "https://via.placeholder.com/150?text=No+Photo";
    }

    const photoPath = encodeURIComponent(url);
    return `https://localhost/api/v1/users/photo?path=${photoPath}`;
  };

  const uploadPhoto = async (file: File): Promise<string | null> => {
    const token = localStorage.getItem("token");
    if (!token) {
      error.value = "Not authenticated";
      return null;
    }

    loading.value = true;
    error.value = "";

    try {
      const formData = new FormData();
      formData.append("photo", file);

      const response = await fetch("https://localhost/api/v1/users/upload-picture", {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
        body: formData,
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Upload failed");
      }

      photoUrl.value = data.url;
      return data.url;
    } catch {
      return null;
    } finally {
      loading.value = false;
    }
  };

  return {
    photoUrl,
    loading,
    error,
    loadPhotoUrl,
    uploadPhoto,
  };
}
