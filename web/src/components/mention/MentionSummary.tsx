import React from 'react';
import { Link } from 'react-router-dom';
import styles from './MentionSummary.module.css';
import { MentionSummary as MentionData } from '@src/client/models';

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

    // 单独处理Twitter用户名点击，阻止事件冒泡
    const handleTwitterLinkClick = (e: React.MouseEvent) => {
        e.stopPropagation(); // 阻止事件冒泡到父Link
        e.preventDefault(); // 阻止默认行为
        window.open(`https://twitter.com/${mention?.author?.screenName}`, '_blank', 'noopener,noreferrer');
    };

    return (
        <Link
            to={`/mentions/${mention.id}`}
            state={{ mentionData: mention }}
            className={styles.mentionContainer}
            style={{ textDecoration: 'none', color: 'inherit' }}
        >
            <div className={styles.mentionHeader}>
                <img
                    src={mention?.author?.profileImageUrl}
                    alt={`${mention?.author?.name} profile`}
                    className={styles.profileImage}
                />
                <div className={styles.authorInfo}>
                    <span className={styles.authorName}>{mention?.author?.name}</span>
                    <span className={styles.screenName}>
                        <span 
                            onClick={handleTwitterLinkClick}
                            style={{ textDecoration: 'none', color: '#1da1f2', cursor: 'pointer' }}
                        >
                            @{mention?.author?.screenName}
                        </span>
                    </span>
                </div>
                <span className={styles.createdAt}>{formatDate(mention.createdAt)}</span>
            </div>
            <div className={styles.contentPreview}>
                {mention.contentPreview}
            </div>
            <div style={{ fontSize: '0.8em', color: '#888', wordBreak: 'break-all', marginTop: 4 }}>
                CID: {mention.cid}
            </div>
            <div className={styles.footer}>
                <button className={styles.readMore}>Read {mention?.numTweets} tweets</button>
            </div>
        </Link>
    );
};

export default MentionSummaryComponent;