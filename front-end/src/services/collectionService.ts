import { isTokenExpired } from "@/router";

const API_BASE_URL = "https://localhost/api/v1";

export interface Collection {
  id: string;
  user_id: string;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface CollectionWithPosts {
  collection: Collection;
  recent_posts: Post[];
}

export interface Post {
  ID: string;
  UserID: string;
  Caption: string;
  IsCommentOff: boolean;
  IsHideLikes: boolean;
  IsReel: boolean;
  LikesCount: number;
  CommentsCount: number;
  Media: PostMedia[];
  CreatedAt: string;
  UpdatedAt: string;
}

export interface PostMedia {
  ID: string;
  PostID: string;
  FilePath: string;
  MediaType: string;
  OrderNum: number;
  Duration: number;
  Width: number;
  Height: number;
}

const getAuthHeaders = () => {
  const token = localStorage.getItem("token");
  if (!token || isTokenExpired(token)) {
    localStorage.removeItem("token");
    window.location.href = "/login";
    throw new Error("Unauthorized");
  }
  return {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  };
};

export const createCollection = async (name: string): Promise<Collection> => {
  const response = await fetch(`${API_BASE_URL}/collections/create`, {
    method: "POST",
    headers: getAuthHeaders(),
    body: JSON.stringify({ name }),
  });

  if (!response.ok) {
    throw new Error("Failed to create collection");
  }

  const data = await response.json();
  return data.collection;
};

export const getCollectionsByUserID = async (userID: string): Promise<CollectionWithPosts[]> => {
  const response = await fetch(`${API_BASE_URL}/collections/user/${userID}`, {
    method: "GET",
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw new Error("Failed to fetch collections");
  }

  const data = await response.json();
  return data.collections || [];
};

export const updateCollection = async (collectionID: string, name: string): Promise<Collection> => {
  const response = await fetch(`${API_BASE_URL}/collections/update/${collectionID}`, {
    method: "PUT",
    headers: getAuthHeaders(),
    body: JSON.stringify({ name }),
  });

  if (!response.ok) {
    throw new Error("Failed to update collection");
  }

  const data = await response.json();
  return data.collection;
};

export const deleteCollection = async (collectionID: string): Promise<void> => {
  const response = await fetch(`${API_BASE_URL}/collections/delete/${collectionID}`, {
    method: "DELETE",
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw new Error("Failed to delete collection");
  }
};

export const addPostToCollection = async (collectionID: string, postID: string): Promise<void> => {
  const response = await fetch(`${API_BASE_URL}/collections/add-post`, {
    method: "POST",
    headers: getAuthHeaders(),
    body: JSON.stringify({ collection_id: collectionID, post_id: postID }),
  });

  if (!response.ok) {
    throw new Error("Failed to add post to collection");
  }
};

export const removePostFromCollection = async (collectionID: string, postID: string): Promise<void> => {
  const response = await fetch(`${API_BASE_URL}/collections/remove-post/${collectionID}/${postID}`, {
    method: "DELETE",
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw new Error("Failed to remove post from collection");
  }
};

export const getPostsByCollectionID = async (collectionID: string): Promise<Post[]> => {
  const response = await fetch(`${API_BASE_URL}/collections/posts/${collectionID}`, {
    method: "GET",
    headers: getAuthHeaders(),
  });

  if (!response.ok) {
    throw new Error("Failed to fetch posts");
  }

  const data = await response.json();
  return data.posts || [];
};
