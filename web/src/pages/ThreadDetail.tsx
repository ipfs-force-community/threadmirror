import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { fetchTwitterDetail } from '@services/api';
import UserProfileComponent from '@components/common/UserProfile';
import { TwitterDetailResponse } from '@src/types';
import styles from './ThreadDetail.module.css';

const ThreadDetail = () => {
  const { id } = useParams<{ id: string }>();
  const [detail, setDetail] = useState<TwitterDetailResponse | []>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  useEffect(() => {
    const loadData = async () => {
      try {
        if (!id) throw new Error('No ID provided');
        const detailData = await fetchTwitterDetail(id);
        setDetail(detailData);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Data Loading Failed');
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, [id]);
  if (loading) return <div className={styles.container}>Loading...</div>;
  if (error) return <div className={styles.container}>Error: {error}</div>;
  if (!detail?.length) return <div className={styles.container}>No data found</div>;

  const author = detail[0]?.author;
  if (!author) {
    return <div className={styles.container}>Lose Author Info</div>;
  }

  const handleClick = (tweet: any) => {
    const selection = window.getSelection();
    if (selection && selection.toString().length === 0) {
      window.open(`https://x.com/${tweet.author.screen_name}/status/${tweet.id}`, '_blank');
    }
  };

  return (
    <div className={styles.container}>
      <UserProfileComponent profile={author} sample={true} />
      <div className={styles.tweetContent}>
        {detail?.map((tweet, index) => (
          <div
            key={tweet.id}
            className={styles.tweetItem}
            onClick={() => handleClick(tweet)}
          >
            <div className={styles.tweetLink}>
              <p className={styles.tweetText}>{tweet.text}</p>
              {tweet.entities.media && tweet.entities.media.length > 0 && (
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

export default ThreadDetail;