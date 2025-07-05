import React from 'react';
import styles from './UsageGuide.module.css';

const UsageGuide: React.FC = () => {
  return (
    <div className={styles.container}>
      <h2>How to use</h2>
      <ol className={styles.steps}>
        <li>
          Reply <strong>@threadmirror</strong> under any X (Twitter) thread.
        </li>
        <li>
          The bot sends back a permanent link, AI summary and long image — no signup needed.
        </li>
      </ol>
    </div>
  );
};

export default UsageGuide; 