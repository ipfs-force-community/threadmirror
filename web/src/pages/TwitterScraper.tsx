import React from 'react';
import { useNavigate } from 'react-router-dom';
import TwitterScraperComponent from '@components/TwitterScraper';
import { ThreadDetail } from '@client/index';
import styles from './TwitterScraper.module.css';

const TwitterScraperPage: React.FC = () => {
  const navigate = useNavigate();

  const handleThreadScraped = (thread: ThreadDetail) => {
    // Navigate to the thread detail page after successful scraping
    navigate(`/thread/${thread.id}`);
  };

  return (
    <div className={styles.pageContainer}>
      <div className={styles.breadcrumb}>
        <button
          onClick={() => navigate('/')}
          className={styles.backButton}
        >
          ← Back to Mentions
        </button>
      </div>
      
      <TwitterScraperComponent onThreadScraped={handleThreadScraped} />
      
      <div className={styles.infoSection}>
        <h3>How it works</h3>
        <ul>
          <li>Paste any Twitter/X URL from a tweet</li>
          <li>Our system will extract the entire thread conversation</li>
          <li>The thread will be saved and available for viewing</li>
          <li>You can share, download, or generate QR codes for any thread</li>
        </ul>
      </div>
    </div>
  );
};

export default TwitterScraperPage; 