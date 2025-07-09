import React from 'react';
import styles from './StatusBadge.module.css';

export type StatusType = 'pending' | 'scraping' | 'completed' | 'failed';

interface StatusBadgeProps {
  status: StatusType;
  size?: 'small' | 'medium';
  className?: string;
}

const StatusBadge: React.FC<StatusBadgeProps> = ({ 
  status, 
  size = 'medium', 
  className = '' 
}) => {
  const getStatusText = (status: StatusType) => {
    switch (status) {
      case 'pending':
        return 'Pending';
      case 'scraping':
        return 'Scraping';
      case 'completed':
        return 'Completed';
      case 'failed':
        return 'Failed';
      default:
        return 'Unknown';
    }
  };

  const getStatusIcon = (status: StatusType) => {
    switch (status) {
      case 'pending':
        return '⏳';
      case 'scraping':
        return '🔄';
      case 'completed':
        return '✅';
      case 'failed':
        return '❌';
      default:
        return '❓';
    }
  };

  return (
    <span 
      className={`${styles.statusBadge} ${styles[status]} ${styles[size]} ${className}`}
      title={`Status: ${getStatusText(status)}`}
    >
      <span className={styles.icon}>{getStatusIcon(status)}</span>
      <span className={styles.text}>{getStatusText(status)}</span>
    </span>
  );
};

export default StatusBadge; 