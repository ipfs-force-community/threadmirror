import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useApiService } from '@services/api';
import { ThreadDetail } from '@client/index';
import { toast } from 'sonner';
import styles from './TwitterScraper.module.css';

interface TwitterScraperProps {
  onThreadScraped?: (thread: ThreadDetail) => void;
}

const TwitterScraper: React.FC<TwitterScraperProps> = ({ onThreadScraped }) => {
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState<{
    thread?: ThreadDetail;
    tweetId?: string;
    message: string;
    isExisting: boolean;
  } | null>(null);

  const navigate = useNavigate();
  const { postThreadsScrape } = useApiService();

  const validateTwitterUrl = (url: string): boolean => {
    const twitterUrlPattern = /^https?:\/\/(www\.)?(twitter\.com|x\.com|mobile\.twitter\.com|mobile\.x\.com)\/[^/]+\/status(es)?\/\d+/i;
    return twitterUrlPattern.test(url);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!url.trim()) {
      setError('Please enter a Twitter URL');
      return;
    }

    if (!validateTwitterUrl(url.trim())) {
      setError('Please enter a valid Twitter/X URL (e.g., https://twitter.com/user/status/123456789)');
      return;
    }

    setError('');
    setLoading(true);
    setResult(null);

    try {
      const response = await postThreadsScrape({
        url: url.trim()
      });

      // For successful responses (200), the thread scraping job was queued
      setResult({
        tweetId: response.tweetId,
        message: response.message || 'Thread scraping job queued successfully',
        isExisting: false
      });

      // Show success toast
      toast.success('Thread scraped successfully');
      
    } catch (error: any) {
      let errorMessage = 'Failed to scrape thread';
      
      // Check if it's a 409 error (thread already exists)
      if (error.response?.status === 409) {
        // For 409 errors, the thread data is still returned
        try {
          const existingData = error.response.data;
          setResult({
            thread: existingData.data,
            message: existingData.message || 'Thread already exists',
            isExisting: true
          });

          // Call callback if provided
          if (onThreadScraped) {
            onThreadScraped(existingData.data);
          }

          toast.success('Thread already exists in our database');
          return; // Don't show error for 409
        } catch (parseError) {
          errorMessage = 'Thread already exists but could not retrieve details';
        }
      } else if (error.response?.status === 400) {
        errorMessage = 'Invalid Twitter URL format';
      } else if (error.response?.status === 404) {
        errorMessage = 'Tweet not found or not accessible';
      } else if (error.response?.status === 500) {
        errorMessage = 'Server error, please try again later';
      } else if (error.response?.data?.message) {
        errorMessage = error.response.data.message;
      } else if (error.message) {
        errorMessage = error.message;
      }
      
      setError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleExampleClick = (exampleUrl: string) => {
    setUrl(exampleUrl);
    setError('');
    setResult(null);
  };

  const handleViewThread = () => {
    if (result?.thread) {
      navigate(`/thread/${result.thread.id}`);
    }
  };

  const exampleUrls = [
    'https://twitter.com/elonmusk/status/1234567890123456789',
    'https://x.com/username/status/9876543210987654321',
    'https://twitter.com/user/status/1111111111111111111'
  ];

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>Scrape Twitter Thread</h1>
        <p className={styles.subtitle}>
          Enter a Twitter/X URL to extract and save the complete thread
        </p>
      </div>

      <form onSubmit={handleSubmit} className={styles.form}>
        <div className={styles.inputGroup}>
          <label htmlFor="twitter-url" className={styles.label}>
            Twitter/X URL
          </label>
          <div className={styles.inputWrapper}>
            <input
              id="twitter-url"
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://twitter.com/user/status/123456789"
              className={`${styles.input} ${error ? styles.inputError : ''}`}
              disabled={loading}
            />
          </div>
          {error && (
            <div className={styles.errorMessage}>
              <span>⚠️</span>
              {error}
            </div>
          )}
          <div className={styles.helpText}>
            Paste a link to any tweet to extract the complete thread conversation
          </div>
        </div>

        <button
          type="submit"
          disabled={loading || !url.trim()}
          className={styles.button}
        >
          {loading ? (
            <div className={styles.loading}>
              <div className={styles.spinner} />
              Scraping Thread...
            </div>
          ) : (
            <>
              🧵 Scrape Thread
            </>
          )}
        </button>
      </form>

      {result && (
        <div className={styles.resultContainer}>
          <h3 className={styles.resultTitle}>
            {result.isExisting ? '✅' : '🎉'} Success!
          </h3>
          <p className={styles.resultMessage}>
            {result.message}
          </p>
          {result.thread ? (
            <div>
              <strong>Thread ID:</strong> {result.thread.id}
              <br />
              <strong>Tweets:</strong> {result.thread.numTweets}
              <br />
              <strong>Content Preview:</strong> {result.thread.contentPreview}
              <br />
              <br />
              <button
                onClick={handleViewThread}
                className={styles.button}
                style={{ background: '#1da1f2' }}
              >
                📖 View Thread
              </button>
            </div>
          ) : (
            <div>
              <strong>Tweet ID:</strong> {result.tweetId}
              <br />
              <p style={{ marginTop: '12px', color: '#6b6b6b' }}>
                The thread is being processed. You can check back later or use the tweet ID to view the progress.
              </p>
            </div>
          )}
        </div>
      )}

      <div className={styles.examples}>
        <h3 className={styles.examplesTitle}>Example URLs:</h3>
        <ul className={styles.examplesList}>
          {exampleUrls.map((exampleUrl, index) => (
            <li key={index} className={styles.exampleItem}>
              <span
                className={styles.exampleUrl}
                onClick={() => handleExampleClick(exampleUrl)}
                role="button"
                tabIndex={0}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    handleExampleClick(exampleUrl);
                  }
                }}
              >
                {exampleUrl}
              </span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
};

export default TwitterScraper; 