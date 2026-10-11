import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useDebounceSearch } from '../useDebounceSearch'

describe('useDebounceSearch', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  it('should initialize with empty values', () => {
    const mockSearchFn = vi.fn().mockResolvedValue([])
    const { searchQuery, isSearching, results } = useDebounceSearch(mockSearchFn)

    expect(searchQuery.value).toBe('')
    expect(isSearching.value).toBe(false)
    expect(results.value).toEqual([])
  })

  it('should debounce search calls', async () => {
    const mockSearchFn = vi.fn().mockResolvedValue(['result1', 'result2'])
    const { search } = useDebounceSearch(mockSearchFn, 500)

    search('test')
    expect(mockSearchFn).not.toHaveBeenCalled()

    vi.advanceTimersByTime(500)
    await vi.waitFor(() => expect(mockSearchFn).toHaveBeenCalledWith('test'))
  })

  it('should clear results when query is empty', async () => {
    const mockSearchFn = vi.fn().mockResolvedValue(['result1'])
    const { search, results } = useDebounceSearch(mockSearchFn)

    search('')
    vi.advanceTimersByTime(500)

    expect(results.value).toEqual([])
    expect(mockSearchFn).not.toHaveBeenCalled()
  })

  it('should update results after successful search', async () => {
    const mockData = ['item1', 'item2', 'item3']
    const mockSearchFn = vi.fn().mockResolvedValue(mockData)
    const { search, results, isSearching } = useDebounceSearch(mockSearchFn, 300)

    search('query')
    expect(isSearching.value).toBe(false)

    vi.advanceTimersByTime(300)
    await vi.waitFor(() => {
      expect(mockSearchFn).toHaveBeenCalledWith('query')
      expect(results.value).toEqual(mockData)
      expect(isSearching.value).toBe(false)
    })
  })

  it('should clear search and reset values', () => {
    const mockSearchFn = vi.fn().mockResolvedValue([])
    const { search, clearSearch, searchQuery, results } = useDebounceSearch(mockSearchFn)

    search('test query')
    clearSearch()

    expect(searchQuery.value).toBe('')
    expect(results.value).toEqual([])
  })

  it('should cancel previous timeout when new search is triggered', async () => {
    const mockSearchFn = vi.fn().mockResolvedValue(['result'])
    const { search } = useDebounceSearch(mockSearchFn, 500)

    search('first')
    vi.advanceTimersByTime(200)
    search('second')
    vi.advanceTimersByTime(500)

    await vi.waitFor(() => {
      expect(mockSearchFn).toHaveBeenCalledTimes(1)
      expect(mockSearchFn).toHaveBeenCalledWith('second')
    })
  })
})
