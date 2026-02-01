import { useEffect, useCallback, useRef } from "react";
import { groupAPI, feedAPI } from "@/lib/api";
import { useDataStore } from "@/store";

export function useFeeds() {
  const didInitRef = useRef(false);
  const {
    groups,
    feeds,
    isLoadingGroups,
    isLoadingFeeds,
    groupsError,
    feedsError,
    setGroups,
    setFeeds,
    setLoadingGroups,
    setLoadingFeeds,
    setGroupsError,
    setFeedsError,
    getFeedsByGroup,
  } = useDataStore();

  const fetchGroups = useCallback(async () => {
    setLoadingGroups(true);
    setGroupsError(null);
    try {
      const response = await groupAPI.list();
      setGroups(response.data);
    } catch (error) {
      setGroupsError(error instanceof Error ? error.message : "Failed to load groups");
    } finally {
      setLoadingGroups(false);
    }
  }, [setGroups, setLoadingGroups, setGroupsError]);

  const fetchFeeds = useCallback(async () => {
    setLoadingFeeds(true);
    setFeedsError(null);
    try {
      const response = await feedAPI.list();
      setFeeds(response.data);
    } catch (error) {
      setFeedsError(error instanceof Error ? error.message : "Failed to load feeds");
    } finally {
      setLoadingFeeds(false);
    }
  }, [setFeeds, setLoadingFeeds, setFeedsError]);

  const refresh = useCallback(async () => {
    await Promise.all([fetchGroups(), fetchFeeds()]);
  }, [fetchGroups, fetchFeeds]);

  useEffect(() => {
    // Prevent infinite re-fetch loops on failed requests (or valid empty responses).
    // Only perform the initial load attempt once per mount.
    if (didInitRef.current) return;
    didInitRef.current = true;

    if (groups.length === 0) {
      fetchGroups();
    }
    if (feeds.length === 0) {
      fetchFeeds();
    }
  }, [fetchGroups, fetchFeeds, groups.length, feeds.length]);

  const getUnreadCount = useCallback(
    (feedId: number) => {
      const feed = feeds.find((f) => f.id === feedId);
      return feed?.unread_count ?? 0;
    },
    [feeds]
  );

  const getGroupUnreadCount = useCallback(
    (groupId: number) => {
      return getFeedsByGroup(groupId).reduce(
        (sum, f) => sum + (f.unread_count ?? 0),
        0
      );
    },
    [getFeedsByGroup]
  );

  const getTotalUnreadCount = useCallback(() => {
    return feeds.reduce((sum, f) => sum + (f.unread_count ?? 0), 0);
  }, [feeds]);

  return {
    groups,
    feeds,
    isLoading: isLoadingGroups || isLoadingFeeds,
    error: groupsError || feedsError,
    refresh,
    getFeedsByGroup,
    getUnreadCount,
    getGroupUnreadCount,
    getTotalUnreadCount,
  };
}
