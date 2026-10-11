import { ref } from "vue";

const showCreatePost = ref(false);
const createPostContent = ref<any>(null);

export function useCreatePost() {
  const open = (content?: any) => {
    createPostContent.value = content;
    showCreatePost.value = true;
  };

  const close = () => {
    showCreatePost.value = false;
    createPostContent.value = null;
  };

  return {
    showCreatePost,
    createPostContent,
    open,
    close,
  };
}


const showPostContent = ref(false);
const postContent = ref<any>(null);
const passedPostID = ref<string | null>(null);

export function useShowPost() {
  const showPost = (postID: string, content?: any) => {

    postContent.value = content;
    passedPostID.value = postID;
    showPostContent.value = true;

  };

  const close = () => {
    showPostContent.value = false;
    postContent.value = null;
    passedPostID.value = null;
  };

  return {
    showPostContent,
    postContent,
    passedPostID,
    showPost,
    close,
  };
}


const showSearch = ref(false);

export function useSearch() {
  const openSearch = () => {
    showSearch.value = true;
  };

  const closeSearch = () => {
    showSearch.value = false;
  };

  return {
    showSearch,
    openSearch,
    closeSearch,
  };
}