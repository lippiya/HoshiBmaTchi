import { ref } from "vue";

export function useDebounceSearch<T>(
  searchFn: (query: string) => Promise<T[]>,
  delay: number = 500
) {
  const searchQuery = ref("");
  const isSearching = ref(false);
  const results = ref<T[]>([]);
  let searchTimeout: number | null = null;

  const search = async (query: string) => {
    searchQuery.value = query;

    if (searchTimeout) {
      clearTimeout(searchTimeout);
    }

    if (!query.trim()) {
      results.value = [];
      return;
    }

    isSearching.value = true;
    searchTimeout = window.setTimeout(async () => {
      try {
        results.value = await searchFn(query);
      } finally {
        isSearching.value = false;
      }
    }, delay);
  };

  const clearSearch = () => {
    searchQuery.value = "";
    results.value = [];
    if (searchTimeout) {
      clearTimeout(searchTimeout);
    }
  };

  return {
    searchQuery,
    isSearching,
    results,
    search,
    clearSearch
  };
}
