import { useEffect, useState, useCallback, useRef } from 'react';
import { useApiService } from '@services/api';
import PostList from '@components/post/PostList';
import { Post } from '@client/index';
import { isUserLoggedIn } from '@utils/cookie';
import { toast } from 'sonner';
import styles from './UserPosts.module.css';

const UserPosts = () => {
  const [posts, setPosts] = useState<Post[]>([]);
  const [loading, setLoading] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [apiErrorOccurred, setApiErrorOccurred] = useState(false);
  const queryLimit = 3;
  const [pagination, setPagination] = useState({
    offset: 0,
    total: 0,
  });
  const observer = useRef<IntersectionObserver | null>(null);
  const loadingRef = useRef<HTMLDivElement>(null);
  const toastShownRef = useRef(false);
  const { fetchGetPosts } = useApiService();

  // 检查用户是否已登录
  const isLoggedIn = isUserLoggedIn();

  const loadMorePosts = useCallback(async () => {
    // 如果已经发生API错误，不再尝试加载
    if (loading || !hasMore || !isLoggedIn || apiErrorOccurred) return;

    setLoading(true);
    try {
      const response = await fetchGetPosts({
        limit: queryLimit,
        offset: pagination.offset / queryLimit,
      });

      const responseData = response.data || [];

      if (responseData.length > 0) {
        setPosts(prevPosts => {
          const existingIds = new Set(prevPosts.map(post => post.id));
          const newPosts = responseData.filter(post => !existingIds.has(post.id));
          return [...prevPosts, ...newPosts];
        });
      }

      setPagination(prev => ({
        ...prev,
        offset: posts?.length || 0,
        total: response.meta?.total || prev.total
      }));

      setHasMore(
        responseData.length > 0 &&
        responseData.length === queryLimit &&
        (
          response.meta?.total === undefined ||
          pagination.offset + responseData.length < response.meta.total
        )
      );

      // 成功加载数据后重置错误状态
      setError(null);
      setApiErrorOccurred(false);
      toastShownRef.current = false;
    } catch (err) {
      console.error('Failed to load posts:', err);
      setError(err instanceof Error ? err : new Error('未知错误'));
      setApiErrorOccurred(true);

      // 确保只显示一次toast错误提示
      if (!toastShownRef.current) {
        toast.error("API service is unavailable", { duration: 5000 });
        toastShownRef.current = true;
      }
    } finally {
      setLoading(false);
    }
  }, [fetchGetPosts, loading, hasMore, pagination, isLoggedIn, apiErrorOccurred]);

  const resetError = useCallback(() => {
    setError(null);
    setApiErrorOccurred(false);
    toastShownRef.current = false;
    loadMorePosts();
  }, [loadMorePosts]);

  useEffect(() => {
    if (loading || !isLoggedIn || apiErrorOccurred) return;

    const handleObserver = (entries: IntersectionObserverEntry[]) => {
      const [entry] = entries;
      if (entry.isIntersecting && hasMore) {
        loadMorePosts();
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
  }, [loading, hasMore, loadMorePosts, isLoggedIn, apiErrorOccurred]);

  useEffect(() => {
    if (isLoggedIn && !apiErrorOccurred) {
      loadMorePosts();
    }
  }, [isLoggedIn, loadMorePosts, apiErrorOccurred]);

  // 如果用户未登录，返回空白页面
  if (!isLoggedIn) {
    return <div className={styles.user_page}></div>;
  }

  return (
    <div className={styles.user_page}>
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
          <PostList posts={posts} />
          <div ref={loadingRef} className={styles.loading}>
            {loading && <p>Loading...</p>}
            {!hasMore && posts.length > 0 && <p>No more data</p>}
            {!hasMore && posts.length === 0 && !error && <p>No data available</p>}
            {!hasMore && posts.length === 0 && error && <p>Failed to load data</p>}
          </div>
        </>
      )}
    </div>
  );
};

export default UserPosts;