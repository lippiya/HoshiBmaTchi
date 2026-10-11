import { useRouter } from "vue-router";

export function useHashtag() {
  const router = useRouter();

  const extractHashtags = (text: string): string[] => {
    const hashtagRegex = /#[a-zA-Z0-9_\-]+/g;
    const matches = text.match(hashtagRegex);
    return matches ? matches.map(tag => tag.substring(1)) : [];
  };

  const formatTextWithHashtags = (text: string) => {
    if (!text) return [{ text: '', isHashtag: false, key: 0 }];
    const hashtagRegex = /(#[a-zA-Z0-9_\-]+)/g;
    return text.split(hashtagRegex).map((part, index) => ({
      text: part,
      isHashtag: part.startsWith("#"),
      key: index
    }));
  };

  const navigateToHashtag = (hashtag: string) => {
    const cleanTag = hashtag.startsWith("#") ? hashtag.substring(1) : hashtag;
    router.push({
      name: "explore",
      query: { hashtag: cleanTag }
    });
  };

  return {
    extractHashtags,
    formatTextWithHashtags,
    navigateToHashtag
  };
}
