import { useEffect, useState, useCallback, useRef } from 'react';
import { useAuth0 } from '@auth0/auth0-react';
import { useApiService } from '@services/api';
import MentionList from '@components/mention/MentionList';
import { MentionSummary as Mention } from '@client/index';
import { isUserLoggedIn } from '@utils/cookie';
import { toast } from 'sonner';
import styles from './UserMentions.module.css';
import LoginPrompt from '@components/common/LoginPrompt';

const UserMentions = () => {
  const { isAuthenticated, isLoading: auth0Loading } = useAuth0();
  const [mentions, setMentions] = useState<Mention[]>([]);
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [apiErrorOccurred, setApiErrorOccurred] = useState(false);
  const [isInitialLoad, setIsInitialLoad] = useState(true);
  const [initialLoadDone, setInitialLoadDone] = useState(false);
  const queryLimit = 10;
  const [pagination, setPagination] = useState({
    offset: 0,
    total: 0,
  });
  const observer = useRef<IntersectionObserver | null>(null);
  const loadingRef = useRef<HTMLDivElement>(null);
  const mentionsContainerRef = useRef<HTMLDivElement>(null);
  const toastShownRef = useRef(false);
  const { fetchGetMentions } = useApiService();

  // 结合Auth0状态和本地cookie状态来确定登录状态
  const isLoggedIn = isAuthenticated && isUserLoggedIn();

  const calculateInitialLoadCount = useCallback(() => {
    const viewportHeight = window.innerHeight;
    const estimatedMentionHeight = 200;
    const mentionCount = Math.ceil(viewportHeight / estimatedMentionHeight) + 2;
    return Math.max(mentionCount, queryLimit);
  }, []);

  const loadMoreMentions = useCallback(async () => {
    if (loading || !hasMore || !isLoggedIn || apiErrorOccurred) return;

    setLoading(true);
    try {
      const currentLimit = isInitialLoad ? calculateInitialLoadCount() : queryLimit;

      const response = await fetchGetMentions({
        limit: currentLimit,
        offset: mentions.length,
      });

      const responseData = response.data || [];

      if (responseData.length > 0) {
        setMentions(prevMentions => {
          const existingIds = new Set(prevMentions.map(mention => mention.id));
          const newMentions = responseData.filter(mention => !existingIds.has(mention.id));
          return [...prevMentions, ...newMentions];
        });
      }

      setPagination(prev => ({
        ...prev,
        offset: mentions.length + responseData.length,
        total: response.meta?.total || prev.total
      }));

      setHasMore(
        responseData.length > 0 &&
        responseData.length === currentLimit &&
        (
          response.meta?.total === undefined ||
          mentions.length + responseData.length < response.meta.total
        )
      );

      setError(null);
      setApiErrorOccurred(false);
      toastShownRef.current = false;
      
      if (isInitialLoad) {
        setIsInitialLoad(false);
        setInitialLoadDone(true);
      }
    } catch (err) {
      console.error('Failed to load mentions:', err);
      setError(err instanceof Error ? err : new Error('unknown error'));
      setApiErrorOccurred(true);

      if (!toastShownRef.current) {
        toast.error("API service is unavailable", { duration: 5000 });
        toastShownRef.current = true;
      }
    } finally {
      setLoading(false);
    }
  }, [fetchGetMentions, loading, hasMore, isLoggedIn, apiErrorOccurred, mentions.length, queryLimit, isInitialLoad, calculateInitialLoadCount]);

  const resetError = useCallback(() => {
    setError(null);
    setApiErrorOccurred(false);
    toastShownRef.current = false;
    setIsInitialLoad(true);
    loadMoreMentions();
  }, [loadMoreMentions]);

  useEffect(() => {
    if (loading || !isLoggedIn || apiErrorOccurred) return;

    const handleObserver = (entries: IntersectionObserverEntry[]) => {
      const [entry] = entries;
      if (entry.isIntersecting && hasMore) {
        loadMoreMentions();
      }
    };

    observer.current = new IntersectionObserver(handleObserver, {
      root: null,
      rootMargin: '100px',
      threshold: 0.1
    });

    if (loadingRef.current) {
      observer.current.observe(loadingRef.current);
    }

    return () => {
      if (observer.current) {
        observer.current.disconnect();
      }
    };
  }, [loading, hasMore, loadMoreMentions, isLoggedIn, apiErrorOccurred]);

  // 重置组件状态当认证状态改变时
  useEffect(() => {
    if (!auth0Loading) {
      setMentions([]);
      setPagination({ offset: 0, total: 0 });
      setHasMore(true);
      setError(null);
      setApiErrorOccurred(false);
      setIsInitialLoad(true);
      setInitialLoadDone(false);
      toastShownRef.current = false;
    }
  }, [isLoggedIn, auth0Loading]);

  useEffect(() => {
    if (isLoggedIn && !apiErrorOccurred && isInitialLoad && !initialLoadDone && !auth0Loading) {
      loadMoreMentions();
    }
  }, [isLoggedIn, loadMoreMentions, apiErrorOccurred, isInitialLoad, initialLoadDone, auth0Loading]);

  // 等待Auth0加载完成
  if (auth0Loading) {
    return <div className="flex justify-center p-4">Loading...</div>;
  }

  if (!isLoggedIn) {
    return <LoginPrompt />;
  }

  return (
    <div className={styles.user_page} ref={mentionsContainerRef}>
      {error ? (
        <div className={styles.fallback_notice}>
          <p>API service is temporarily unavailable</p>
          <button
            className={styles.retry_button}
            onClick={resetError}
          >
            Retry
          </button>
        </div>
      ) : (
        <>
          <MentionList mentions={mentions} />
          <div ref={loadingRef} className={styles.loading}>
            {loading && <p>Loading...</p>}
            {!hasMore && mentions.length > 0 && <p>No more data</p>}
            {!hasMore && mentions.length === 0 && !error && <p>No data available</p>}
            {!hasMore && mentions.length === 0 && error && <p>Failed to load data</p>}
          </div>
        </>
      )}
    </div>
  );
};

export default UserMentions;