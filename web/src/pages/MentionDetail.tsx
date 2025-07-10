import { useEffect, useState, useRef, useCallback } from 'react';
import { useLocation, useParams } from 'react-router-dom';
import { useApiService } from '@services/api';
import UserProfileComponent from '@components/common/UserProfile';
import { toast } from 'sonner';
import styles from './MentionDetail.module.css';
import { ThreadDetail as ThreadData, Tweet as TweetModel } from '@src/client/models';
import { renderTweetContent } from '@utils/richText';
import StatusBadge from '@components/common/StatusBadge';

const MentionDetail = () => {
  const { mentionData } = useLocation().state || {};
  const { id } = useParams<{ id: string }>();
  const [detail, setDetail] = useState<ThreadData | null>(null);
  const [loading, setLoading] = useState(false);
  const [shareLoading, setShareLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const toastShownRef = useRef(false);
  const { fetchGetThreadId, fetchGetShare } = useApiService();

  const loadData = useCallback(async (showToast = true) => {
    if (!id) {
      setError('No ID provided');
      return;
    }

    setLoading(true);
    try {
      const detailData = await fetchGetThreadId({
        id,
      });
      setDetail(detailData?.data || null);
      setError(null);
      toastShownRef.current = false;
    } catch (err) {
      // Keep a short message for UI while keeping full trace in console
      console.error('Failed to load mention details:', err);
      const errorMessage = err instanceof Error ? (err.message || 'Data Loading Failed') : 'Data Loading Failed';
      setError(errorMessage);

      if (showToast && !toastShownRef.current) {
        toast.error('Failed to load mention details', { duration: 5000 });
        toastShownRef.current = true;
      }
    } finally {
      setLoading(false);
    }
  }, [id, fetchGetThreadId]);

  const handleRetry = useCallback(() => {
    loadData(false);
  }, [loadData]);

  const handleShare = useCallback(async () => {
    if (!id || shareLoading) return;
    setShareLoading(true);
    try {
      const blob = await fetchGetShare(id);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = `thread_${id}.png`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);
    } catch (err) {
      console.error('Failed to download share image:', err);
      toast.error("Failed to download share image", { duration: 5000 });
    } finally {
      setShareLoading(false);
    }
  }, [id, fetchGetShare, shareLoading]);

  useEffect(() => {
    if (mentionData?.tweets?.length > 0) {
      setDetail(mentionData);
    } else {
      loadData();
    }
  }, [mentionData, loadData]);

  // Update document title based on detail data
  useEffect(() => {
    const defaultTitle = 'Thread Mirror';
    if (detail?.tweets?.length) {
      const firstTweet = detail.tweets[0];
      const authorName = firstTweet?.author?.name || firstTweet?.author?.screenName;
      document.title = authorName ? `${authorName} | Thread Mirror` : `Detail | Thread Mirror`;
    } else {
      document.title = `Detail | Thread Mirror`;
    }
    // Cleanup: reset title on unmount
    return () => {
      document.title = defaultTitle;
    };
  }, [detail]);

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

  const author = detail?.tweets?.[0]?.author;
  if (!author) {
    return <div className={styles.container}>Lose Author Info</div>;
  }

  return (
    <div className={styles.container}>
      <UserProfileComponent profile={author} sample={true} />
      
      {detail?.status !== 'completed' && (
        <div className={styles.statusBar}>
          <StatusBadge status={detail?.status as any} size="medium" />
          <span className={styles.statusText}>
            Thread processing status: {detail?.status}
          </span>
        </div>
      )}

      <div className={styles.actionBar}>
        <button 
          className={styles.shareButton} 
          onClick={handleShare}
          disabled={shareLoading}
        >
          {shareLoading ? 'Sharing...' : 'Share Image'}
        </button>
      </div>

      <div className={styles.metaBar}>
        {detail?.tweets?.[0] && (
          <a
            href={`https://x.com/${author.screenName}/status/${detail.tweets[detail.tweets.length - 1].id}`}
            target="_blank"
            rel="noopener noreferrer"
            className={styles.sourceLink}
          >
            原文链接
          </a>
        )}
        {detail?.cid && (
          <span className={styles.cid}>CID: {detail.cid}</span>
        )}
      </div>

      <div className={styles.tweetContent}>
        {detail?.tweets?.map((tweet, index) => (
          <article
            key={tweet.id}
            className={styles.tweetItem}
          >
            <p className={styles.tweetText}>{renderTweetContent(tweet.text, tweet.entities || undefined, tweet.richtext)}</p>
            {tweet?.entities?.media && tweet.entities.media.length > 0 && (
              <figure className={styles.mediaContainer}>
                {tweet.entities.media.map(media => (
                  <img
                    key={media.idStr}
                    src={media.mediaUrlHttps}
                    alt={`${media.type}${index + 1}`}
                    className={styles.mediaImage}
                  />
                ))}
              </figure>
            )}
          </article>
        ))}
      </div>
    </div>
  );
};

export default MentionDetail;