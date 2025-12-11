/**
 * useSearch Hook Tests
 * Validates debounced search behavior, cancellation, and state management
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useSearch } from './useSearch';
import { apiClient } from '../lib/api-client';

// Mock the API client
vi.mock('../lib/api-client', () => ({
  apiClient: {
    searchScenes: vi.fn(),
    searchEvents: vi.fn(),
    searchPosts: vi.fn(),
  },
}));

describe('useSearch', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
    
    // Default mock responses
    vi.mocked(apiClient.searchScenes).mockResolvedValue([]);
    vi.mocked(apiClient.searchEvents).mockResolvedValue([]);
    vi.mocked(apiClient.searchPosts).mockResolvedValue([]);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('Initial State', () => {
    it('returns empty results initially', () => {
      const { result } = renderHook(() => useSearch());
      
      expect(result.current.results).toEqual({
        scenes: [],
        events: [],
        posts: [],
      });
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBe(null);
    });
  });

  describe('Debounce Behavior', () => {
    it('debounces search calls with default delay', async () => {
      const { result } = renderHook(() => useSearch());
      
      // Start search
      result.current.search('test');
      
      // Should not call API immediately
      expect(apiClient.searchScenes).not.toHaveBeenCalled();
      
      // Fast-forward time
      vi.advanceTimersByTime(300);
      
      // Should call API after debounce
      await waitFor(() => {
        expect(apiClient.searchScenes).toHaveBeenCalledWith('test', 5, expect.any(AbortSignal));
      });
    });

    it('uses custom debounce delay', async () => {
      const { result } = renderHook(() => useSearch({ debounceMs: 500 }));
      
      result.current.search('test');
      
      // Should not call after default delay
      vi.advanceTimersByTime(300);
      expect(apiClient.searchScenes).not.toHaveBeenCalled();
      
      // Should call after custom delay
      vi.advanceTimersByTime(200);
      await waitFor(() => {
        expect(apiClient.searchScenes).toHaveBeenCalled();
      });
    });

    it('cancels previous search when new query is typed', async () => {
      const { result } = renderHook(() => useSearch());
      
      // Start first search
      result.current.search('first');
      vi.advanceTimersByTime(100);
      
      // Start second search before first completes
      result.current.search('second');
      vi.advanceTimersByTime(300);
      
      // Should only call API once with second query
      await waitFor(() => {
        expect(apiClient.searchScenes).toHaveBeenCalledTimes(1);
        expect(apiClient.searchScenes).toHaveBeenCalledWith('second', 5, expect.any(AbortSignal));
      });
    });
  });

  describe('Parallel Search Execution', () => {
    it('executes all searches in parallel', async () => {
      const { result } = renderHook(() => useSearch());
      
      result.current.search('test');
      vi.advanceTimersByTime(300);
      
      await waitFor(() => {
        expect(apiClient.searchScenes).toHaveBeenCalledWith('test', 5, expect.any(AbortSignal));
        expect(apiClient.searchEvents).toHaveBeenCalledWith('test', 5, expect.any(AbortSignal));
        expect(apiClient.searchPosts).toHaveBeenCalledWith('test', 5, expect.any(AbortSignal));
      });
    });

    it('uses custom limit', async () => {
      const { result } = renderHook(() => useSearch({ limit: 10 }));
      
      result.current.search('test');
      vi.advanceTimersByTime(300);
      
      await waitFor(() => {
        expect(apiClient.searchScenes).toHaveBeenCalledWith('test', 10, expect.any(AbortSignal));
        expect(apiClient.searchEvents).toHaveBeenCalledWith('test', 10, expect.any(AbortSignal));
        expect(apiClient.searchPosts).toHaveBeenCalledWith('test', 10, expect.any(AbortSignal));
      });
    });
  });

  describe('Results Handling', () => {
    it('updates results when searches complete', async () => {
      const mockScenes = [{ id: '1', name: 'Scene 1', allow_precise: true, coarse_geohash: 'abc' }];
      const mockEvents = [{ id: '2', scene_id: 's1', name: 'Event 1', allow_precise: true }];
      const mockPosts = [{ id: '3', content: 'Post 1' }];

      vi.mocked(apiClient.searchScenes).mockResolvedValue(mockScenes);
      vi.mocked(apiClient.searchEvents).mockResolvedValue(mockEvents);
      vi.mocked(apiClient.searchPosts).mockResolvedValue(mockPosts);

      const { result } = renderHook(() => useSearch());
      
      result.current.search('test');
      vi.advanceTimersByTime(300);
      
      await waitFor(() => {
        expect(result.current.results).toEqual({
          scenes: mockScenes,
          events: mockEvents,
          posts: mockPosts,
        });
      });
    });

    it('sets loading state during search', async () => {
      let resolveSearch: (value: any) => void;
      const searchPromise = new Promise((resolve) => {
        resolveSearch = resolve;
      });
      
      vi.mocked(apiClient.searchScenes).mockReturnValue(searchPromise as any);
      vi.mocked(apiClient.searchEvents).mockReturnValue(searchPromise as any);
      vi.mocked(apiClient.searchPosts).mockReturnValue(searchPromise as any);

      const { result } = renderHook(() => useSearch());
      
      result.current.search('test');
      vi.advanceTimersByTime(300);
      
      await waitFor(() => {
        expect(result.current.loading).toBe(true);
      });
      
      // Resolve search
      resolveSearch!([]);
      
      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });
    });

    it('handles individual endpoint failures gracefully', async () => {
      const mockScenes = [{ id: '1', name: 'Scene 1', allow_precise: true, coarse_geohash: 'abc' }];
      
      vi.mocked(apiClient.searchScenes).mockResolvedValue(mockScenes);
      vi.mocked(apiClient.searchEvents).mockRejectedValue(new Error('API error'));
      vi.mocked(apiClient.searchPosts).mockResolvedValue([]);

      const { result } = renderHook(() => useSearch());
      
      result.current.search('test');
      vi.advanceTimersByTime(300);
      
      // Should still return successful results
      await waitFor(() => {
        expect(result.current.results.scenes).toEqual(mockScenes);
        expect(result.current.results.events).toEqual([]);
        expect(result.current.results.posts).toEqual([]);
        expect(result.current.loading).toBe(false);
      });
    });
  });

  describe('Request Cancellation', () => {
    it('aborts in-flight request when new search starts', async () => {
      let abortSignal: AbortSignal | undefined;
      
      vi.mocked(apiClient.searchScenes).mockImplementation(async (query, limit, signal) => {
        abortSignal = signal;
        // Never resolve
        return new Promise(() => {});
      });

      const { result } = renderHook(() => useSearch());
      
      // Start first search
      result.current.search('first');
      vi.advanceTimersByTime(300);
      
      await waitFor(() => {
        expect(apiClient.searchScenes).toHaveBeenCalled();
      });
      
      const firstSignal = abortSignal;
      
      // Start second search
      result.current.search('second');
      vi.advanceTimersByTime(300);
      
      // First signal should be aborted
      await waitFor(() => {
        expect(firstSignal?.aborted).toBe(true);
      });
    });

    it('does not update state if request was cancelled', async () => {
      let resolveFirst: (value: any) => void;
      const firstPromise = new Promise((resolve) => {
        resolveFirst = resolve;
      });
      
      vi.mocked(apiClient.searchScenes).mockReturnValueOnce(firstPromise as any);
      vi.mocked(apiClient.searchEvents).mockReturnValueOnce(firstPromise as any);
      vi.mocked(apiClient.searchPosts).mockReturnValueOnce(firstPromise as any);

      const mockScenes = [{ id: '2', name: 'Scene 2', allow_precise: true, coarse_geohash: 'def' }];
      vi.mocked(apiClient.searchScenes).mockResolvedValue(mockScenes);
      vi.mocked(apiClient.searchEvents).mockResolvedValue([]);
      vi.mocked(apiClient.searchPosts).mockResolvedValue([]);

      const { result } = renderHook(() => useSearch());
      
      // Start first search
      result.current.search('first');
      vi.advanceTimersByTime(300);
      
      await waitFor(() => {
        expect(result.current.loading).toBe(true);
      });
      
      // Start second search (cancels first)
      result.current.search('second');
      vi.advanceTimersByTime(300);
      
      // Resolve first search (should be ignored)
      resolveFirst!([{ id: '1', name: 'Scene 1', allow_precise: true, coarse_geohash: 'abc' }]);
      
      // Wait for second search
      await waitFor(() => {
        expect(result.current.results.scenes).toEqual(mockScenes);
      });
    });
  });

  describe('Clear Functionality', () => {
    it('clears results and state', async () => {
      const mockScenes = [{ id: '1', name: 'Scene 1', allow_precise: true, coarse_geohash: 'abc' }];
      vi.mocked(apiClient.searchScenes).mockResolvedValue(mockScenes);
      vi.mocked(apiClient.searchEvents).mockResolvedValue([]);
      vi.mocked(apiClient.searchPosts).mockResolvedValue([]);

      const { result } = renderHook(() => useSearch());
      
      result.current.search('test');
      vi.advanceTimersByTime(300);
      
      await waitFor(() => {
        expect(result.current.results.scenes.length).toBeGreaterThan(0);
      });
      
      result.current.clear();
      
      expect(result.current.results).toEqual({
        scenes: [],
        events: [],
        posts: [],
      });
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBe(null);
    });

    it('cancels in-flight request when clearing', async () => {
      let abortSignal: AbortSignal | undefined;
      
      vi.mocked(apiClient.searchScenes).mockImplementation(async (query, limit, signal) => {
        abortSignal = signal;
        return new Promise(() => {}); // Never resolve
      });

      const { result } = renderHook(() => useSearch());
      
      result.current.search('test');
      vi.advanceTimersByTime(300);
      
      await waitFor(() => {
        expect(apiClient.searchScenes).toHaveBeenCalled();
      });
      
      result.current.clear();
      
      expect(abortSignal?.aborted).toBe(true);
    });

    it('clears pending debounced search', () => {
      const { result } = renderHook(() => useSearch());
      
      result.current.search('test');
      
      // Clear before debounce completes
      result.current.clear();
      vi.advanceTimersByTime(300);
      
      // Should not have called API
      expect(apiClient.searchScenes).not.toHaveBeenCalled();
    });
  });

  describe('Empty Query Handling', () => {
    it('clears results when query is empty', async () => {
      const mockScenes = [{ id: '1', name: 'Scene 1', allow_precise: true, coarse_geohash: 'abc' }];
      vi.mocked(apiClient.searchScenes).mockResolvedValue(mockScenes);
      vi.mocked(apiClient.searchEvents).mockResolvedValue([]);
      vi.mocked(apiClient.searchPosts).mockResolvedValue([]);

      const { result } = renderHook(() => useSearch());
      
      // Search with query
      result.current.search('test');
      vi.advanceTimersByTime(300);
      
      await waitFor(() => {
        expect(result.current.results.scenes.length).toBeGreaterThan(0);
      });
      
      // Search with empty query
      result.current.search('');
      vi.advanceTimersByTime(300);
      
      expect(result.current.results).toEqual({
        scenes: [],
        events: [],
        posts: [],
      });
    });

    it('does not call API for whitespace-only query', () => {
      const { result } = renderHook(() => useSearch());
      
      result.current.search('   ');
      vi.advanceTimersByTime(300);
      
      expect(apiClient.searchScenes).not.toHaveBeenCalled();
    });
  });

  describe('Cleanup', () => {
    it('cancels in-flight requests on unmount', async () => {
      let abortSignal: AbortSignal | undefined;
      
      vi.mocked(apiClient.searchScenes).mockImplementation(async (query, limit, signal) => {
        abortSignal = signal;
        return new Promise(() => {}); // Never resolve
      });

      const { result, unmount } = renderHook(() => useSearch());
      
      result.current.search('test');
      vi.advanceTimersByTime(300);
      
      await waitFor(() => {
        expect(apiClient.searchScenes).toHaveBeenCalled();
      });
      
      unmount();
      
      expect(abortSignal?.aborted).toBe(true);
    });

    it('clears debounce timeout on unmount', () => {
      const { result, unmount } = renderHook(() => useSearch());
      
      result.current.search('test');
      unmount();
      
      vi.advanceTimersByTime(300);
      
      // Should not have called API after unmount
      expect(apiClient.searchScenes).not.toHaveBeenCalled();
    });
  });
});
