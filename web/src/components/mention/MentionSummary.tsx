import React from 'react';
import { Link } from 'react-router-dom';
import styles from './MentionSummary.module.css';
import { MentionSummary as MentionData } from '@src/client/models';
import StatusBadge from '@components/common/StatusBadge';

const MentionSummaryComponent: React.FC<{ mention: MentionData }> = ({ mention }) => {
    const formatDate = (date: Date) => {
        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

        // 如果超过7天，显示英文日期格式
        if (diffDays > 7) {
            const options: Intl.DateTimeFormatOptions = {
                month: 'short',
                day: 'numeric',
                year: date.getFullYear() !== now.getFullYear() ? 'numeric' : undefined
            };
            return date.toLocaleDateString('en-US', options);
        }

        if (diffDays > 0) {
            return `${diffDays}d`;
        }

        const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
        if (diffHours > 0) {
            return `${diffHours}h`;
        }

        const diffMinutes = Math.floor(diffMs / (1000 * 60));
        return `${diffMinutes > 0 ? diffMinutes : 1}m`;
    };

    // 单独处理Twitter用户名点击
    const handleTwitterLinkClick = (e: React.MouseEvent) => {
        e.stopPropagation();
        e.preventDefault();
        window.open(`https://twitter.com/${mention?.threadAuthor?.screenName}`, '_blank', 'noopener,noreferrer');
    };

    return (
        <div className={styles.mentionContainer}>
            <div className={styles.mentionHeader}>
                <img
                    src={mention?.threadAuthor?.profileImageUrl}
                    alt={`${mention?.threadAuthor?.name} profile`}
                    className={styles.profileImage}
                />
                <div className={styles.authorInfo}>
                    <span className={styles.authorName}>{mention?.threadAuthor?.name}</span>
                    <span className={styles.screenName}>
                        <span 
                            onClick={handleTwitterLinkClick}
                            style={{ textDecoration: 'none', color: '#1da1f2', cursor: 'pointer' }}
                        >
                            @{mention?.threadAuthor?.screenName}
                        </span>
                    </span>
                </div>
                <div className={styles.metaInfo}>
                    <StatusBadge status={mention.status as any} size="small" />
                    <span className={styles.createdAt}>{formatDate(mention.createdAt)}</span>
                </div>
            </div>
            <div className={styles.contentPreview}>
                {mention.contentPreview}
            </div>
            <div style={{ fontSize: '0.8em', color: '#888', wordBreak: 'break-all', marginTop: 4 }}>
                CID: {mention.cid}
            </div>
            <div className={styles.footer}>
                <Link 
                    to={`/thread/${mention.threadId}`}
                    state={{ mentionData: mention }}
                    className={styles.readMore}
                    style={{ textDecoration: 'none' }}
                >
                    Read {mention?.numTweets} tweets
                </Link>
            </div>
        </div>
    );
};

export default MentionSummaryComponent;