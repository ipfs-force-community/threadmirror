import { useEffect, useState, useRef, useCallback } from 'react';
import { useLocation, useParams } from 'react-router-dom';
import { useApiService } from '@services/api';
import UserProfileComponent from '@components/common/UserProfile';
import { toast } from 'sonner';
import styles from './PostDetail.module.css';
import { Post as PostData } from '@src/client/models';

const PostDetail = () => {
  const { postData } = useLocation().state || {};
  const { id } = useParams<{ id: string }>();
  const [detail, setDetail] = useState<PostData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const toastShownRef = useRef(false);
  const { fetchGetPostsId } = useApiService();

  const loadData = useCallback(async (showToast = true) => {
    if (!id) {
      setError('No ID provided');
      return;
    }

    setLoading(true);
    try {
      const detailData = await fetchGetPostsId({
        id,
      });
      setDetail(detailData?.data || null);
      setError(null);
      toastShownRef.current = false;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Data Loading Failed';
      setError(errorMessage);

      if (showToast && !toastShownRef.current) {
        toast.error("Failed to load post details", { duration: 5000 });
        toastShownRef.current = true;
      }
    } finally {
      setLoading(false);
    }
  }, [id, fetchGetPostsId]);

  const handleRetry = useCallback(() => {
    loadData(false);
  }, [loadData]);

  useEffect(() => {
    if (postData) {
      setDetail(postData);
    } else {
      loadData();
    }
  }, [postData, loadData]);

  if (loading) return <div className={styles.container}>Loading...</div>;
  
  if (error) {
    return (
      <div className={styles.container}>
        <div className={styles.errorContainer}>
          <p className={styles.errorMessage}>Error: {error}</p>
          <button 
            className={styles.retryButton}
            onClick={handleRetry}
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  const author = detail?.threads?.[0]?.author;
  if (!author) {
    return <div className={styles.container}>Lose Author Info</div>;
  }

  const handleClick = (url: string) => {
    const selection = window.getSelection();
    if (selection && selection.toString().length === 0) {
      window.open(url, '_blank');
    }
  };

  return (
    <div className={styles.container}>
      <UserProfileComponent profile={author} sample={true} />
      <div className={styles.tweetContent}>
        {detail?.threads?.map((tweet, index) => (
          <div
            key={tweet.id}
            className={styles.tweetItem}
            onClick={() => handleClick(`https://x.com/${tweet?.author?.screenName}/status/${tweet.id}`)}
          >
            <div className={styles.tweetLink}>
              <p className={styles.tweetText}>{tweet.text}</p>
              {tweet?.entities?.media && tweet.entities.media.length > 0 && (
                <div className={styles.mediaContainer}>
                  {tweet.entities.media.map(media => (
                    <img
                      key={media.id}
                      src={media.url}
                      alt={`${media.type}${index + 1}`}
                      className={styles.mediaImage}
                    />
                  ))}
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default PostDetail;