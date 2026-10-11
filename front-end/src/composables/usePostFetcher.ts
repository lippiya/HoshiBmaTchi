import { ref } from "vue";

export interface PostWithUser {
  post: any
  user: any
  postMedia: any[]
  collaborators?: any[]
}

async function getFollowingList(token: string): Promise<Set<string>> {
  try {
    const response = await fetch("https://localhost/api/v1/users/get-current-user", {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (!response.ok) return new Set();
    const data = await response.json();
    const currentUserId = data.user?.ID || data.user?.id;
    if (!currentUserId) return new Set();
    
    const followingRes = await fetch(`https://localhost/api/v1/users/get-following/${currentUserId}`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (!followingRes.ok) return new Set();
    const followingData = await followingRes.json();
    const followingIds = (followingData.following || []).map((f: any) => f.ID || f.id);
    return new Set(followingIds);
  } catch {
    return new Set();
  }
}

export function usePostFetcher() {
  const postsResult = ref<PostWithUser[]>([]);
  const loading = ref(false);
  const error = ref<string>("");

  const fetchPostsWithDetails = async (apiUrl: string): Promise<PostWithUser[]> => {
    const token = localStorage.getItem("token");

    if (!token) {
      error.value = "Not authenticated";
      return [];
    }

    loading.value = true;
    error.value = "";

    try {
      const [postsResponse, followingSet] = await Promise.all([
        fetch(apiUrl, {
          method: "GET",
          headers: {
            Authorization: `Bearer ${token}`,
          },
        }),
        getFollowingList(token)
      ]);

      if (!postsResponse.ok) {
        throw new Error(`Failed to fetch posts: ${postsResponse.statusText}`);
      }

      const data = await postsResponse.json();

      if (!data.posts || !Array.isArray(data.posts)) {
        console.error("Invalid response format - data.posts not found or not array:", data);
        throw new Error("Invalid response format: posts array expected");
      }
      const processedPosts = await Promise.all(data.posts.map(async (post: any) => {
        const postData: PostWithUser = {
          post,
          user: null,
          postMedia: [],
          collaborators: [],
          isFollowing: false
        } as PostWithUser & { isFollowing: boolean };

        if (!post.UserID || post.UserID === "undefined") {
          return postData;
        }

        const [userRes, mediaRes, collabRes, commentCountRes] = await Promise.allSettled([
          fetch(`https://localhost/api/v1/users/get-user-by-id/${post.UserID}`, {
            headers: { Authorization: `Bearer ${token}` }
          }),
          fetch(`https://localhost/api/v1/posts/get-post-media-by-id/${post.ID}`, {
            headers: { Authorization: `Bearer ${token}` }
          }),
          fetch(`https://localhost/api/v1/posts/get-post-collaborators-by-id/${post.ID}`, {
            headers: { Authorization: `Bearer ${token}` }
          }),
          fetch(`https://localhost/api/v1/social/get-comments-count/${post.ID}`, {
            headers: { Authorization: `Bearer ${token}` }
          })
        ]);

        if (userRes.status === "fulfilled" && userRes.value.ok) {
          const userData = await userRes.value.json();
          postData.user = userData.user;
          (postData as any).isFollowing = followingSet.has(postData.user?.ID || postData.user?.id);
        }

        if (mediaRes.status === "fulfilled" && mediaRes.value.ok) {
          const mediaData = await mediaRes.value.json();
          postData.postMedia = mediaData.media || [];
        }

        if (collabRes.status === "fulfilled" && collabRes.value.ok) {
          const collabData = await collabRes.value.json();
          postData.collaborators = collabData.collaborators || [];
        }

        if (commentCountRes.status === "fulfilled" && commentCountRes.value.ok) {
          const countData = await commentCountRes.value.json();
          postData.post.CommentsCount = countData.count || 0;
        }

        return postData;
      }));

      const currentUserId = localStorage.getItem("user_id");
      
      const sortedPosts = processedPosts.sort((a, b) => {
        const dateA = new Date(a.post.CreatedAt);
        const dateB = new Date(b.post.CreatedAt);
        return dateB.getTime() - dateA.getTime();
      });

      const filteredPosts = sortedPosts.filter(postData => {
        if (!postData.user) return false;
        const userId = postData.user.ID || postData.user.id;
        if (userId === currentUserId) return true;
        if (postData.user.IsPrivate === true || postData.user.is_private === true) {
          return (postData as any).isFollowing === true;
        }
        return true;
      });

      postsResult.value = filteredPosts;
      return filteredPosts;

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Unknown error occurred";
      error.value = errorMessage;
      console.error("Failed to fetch posts:", err);
      return [];
    } finally {
      loading.value = false;
    }
  };

  const postRes = ref<PostWithUser | null>(null);
  const fetchPost = async (postID: string): Promise<PostWithUser | null> => {
    const token = localStorage.getItem("token");

    if (!token) {
      error.value = "Not authenticated";
      return null;
    }

    loading.value = true;
    error.value = "";

    try {
      const response = await fetch(`https://localhost/api/v1/posts/get-post-by-id/${postID}`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch post: ${response.statusText}`);
      }

      const r = await response.json();


      const postsArray = r.posts || r; 

      if (!postsArray || !Array.isArray(postsArray) || postsArray.length === 0) {
        throw new Error("Invalid response format: \"posts\" array expected but got zonk");
      }

      const post = postsArray[0];


      const postData: PostWithUser = {
        post,
        user: null,
        postMedia: [],
        collaborators: []
      };

      if (!post.UserID || post.UserID === "undefined") {
        postRes.value = postData;
        return postData;
      }

      const [userRes, mediaRes, collabRes, commentCountRes] = await Promise.allSettled([
        fetch(`https://localhost/api/v1/users/get-user-by-id/${post.UserID}`, {
          headers: { Authorization: `Bearer ${token}` }
        }),
        fetch(`https://localhost/api/v1/posts/get-post-media-by-id/${post.ID}`, {
          headers: { Authorization: `Bearer ${token}` }
        }),
        fetch(`https://localhost/api/v1/posts/get-post-collaborators-by-id/${post.ID}`, {
          headers: { Authorization: `Bearer ${token}` }
        }),
        fetch(`https://localhost/api/v1/social/get-comments-count/${post.ID}`, {
          headers: { Authorization: `Bearer ${token}` }
        })
      ]);

      if (userRes.status === "fulfilled" && userRes.value.ok) {
        const userData = await userRes.value.json();
        postData.user = userData.user;
      }

      if (mediaRes.status === "fulfilled" && mediaRes.value.ok) {
        const mediaData = await mediaRes.value.json();
        postData.postMedia = mediaData.media || [];
      }

      if (collabRes.status === "fulfilled" && collabRes.value.ok) {
        const collabData = await collabRes.value.json();
        postData.collaborators = collabData.collaborators || [];
      }

      if (commentCountRes.status === "fulfilled" && commentCountRes.value.ok) {
        const countData = await commentCountRes.value.json();
        postData.post.CommentsCount = countData.count || 0;
      }

      postRes.value = postData;
      return postData;

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Unknown error occurred";
      error.value = errorMessage;
      console.error("Failed to fetch post:", err);
      return null;
    } finally {
      loading.value = false;
    }
  };

  return {
    fetchPost,
    postRes,
    postsResult,
    loading,
    error,
    fetchPostsWithDetails
  };
}

