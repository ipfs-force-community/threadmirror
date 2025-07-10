import React from 'react';
import { Link } from 'react-router-dom';
import styles from './MentionSummary.module.css';
import { MentionSummary as MentionData } from '@src/client/models';
import StatusBadge from '@components/common/StatusBadge';
import defaultProfile from '../../default_profile.png';

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
        window.open(`https://x.com/${mention?.threadAuthor?.screenName}`, '_blank', 'noopener,noreferrer');
    };

    // 根据状态获取状态信息
    const getStatusInfo = (status: string) => {
        switch (status) {
            case 'pending':
                return {
                    icon: '⏱️',
                    title: 'Processing Queue',
                    description: 'Your thread has been added to the processing queue',
                    bgColor: '#f7f9fc',
                    borderColor: '#dbeafe',
                    textColor: '#3b82f6'
                };
            case 'scraping':
                return {
                    icon: '🔄',
                    title: 'Analyzing Content',
                    description: 'We\'re extracting and processing the thread content',
                    bgColor: '#fefce8',
                    borderColor: '#fef08a',
                    textColor: '#ca8a04'
                };
            case 'failed':
                return {
                    icon: '⚠️',
                    title: 'Processing Failed',
                    description: 'Something went wrong. Please try again',
                    bgColor: '#fef2f2',
                    borderColor: '#fecaca',
                    textColor: '#dc2626'
                };
            default:
                return {
                    icon: '📝',
                    title: 'Processing',
                    description: 'Working on your request',
                    bgColor: '#f9fafb',
                    borderColor: '#e5e7eb',
                    textColor: '#6b7280'
                };
        }
    };

    const statusInfo = getStatusInfo(mention.status);

    return (
        <div className={styles.mentionContainer}>
            <div className={styles.mentionHeader}>
                <img
                    src={mention?.threadAuthor?.profileImageUrl || defaultProfile}
                    alt={`${mention?.threadAuthor?.name} profile`}
                    className={styles.profileImage}
                />
                <div className={styles.authorInfo}>
                    <span className={styles.authorName}>{mention?.threadAuthor?.name}</span>
                    {mention?.threadAuthor?.screenName && (
                        <span className={styles.screenName}>
                            <span 
                                onClick={handleTwitterLinkClick}
                                style={{ textDecoration: 'none', color: '#1da1f2', cursor: 'pointer' }}
                            >
                                @{mention?.threadAuthor?.screenName}
                            </span>
                        </span>
                    )}
                </div>
                <div className={styles.metaInfo}>
                    {mention.status !== 'completed' && (
                        <StatusBadge status={mention.status as any} size="small" />
                    )}
                    <span className={styles.createdAt}>{formatDate(mention.createdAt)}</span>
                </div>
            </div>
            <div className={styles.contentPreview}>
                {mention.status === 'completed' ? (
                    mention.contentPreview
                ) : (
                    <div 
                        className={styles.processingState}
                        style={{
                            backgroundColor: statusInfo.bgColor,
                            borderColor: statusInfo.borderColor,
                            color: statusInfo.textColor
                        }}
                    >
                        <div className={styles.statusContent}>
                            
                            <div className={styles.statusText}>
                                <h4 className={styles.statusTitle}>{statusInfo.title}</h4>
                                <p className={styles.statusDescription}>{statusInfo.description}</p>
                            </div>
                        </div>
                        
                        {/* Original tweet link */}
                        { mention.threadId && (
                            <div className={styles.originalLinkContainer}>
                                <a
                                    href={`https://x.com/elonmusk/status/${mention.threadId}`}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    className={styles.originalLink}
                                >
                                    <svg className={styles.linkIcon} viewBox="0 0 24 24" fill="currentColor">
                                        <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
                                    </svg>
                                    View original thread
                                </a>
                            </div>
                        )}
                    </div>
                )}
            </div>
            {/* Only show CID if it is not empty */}
            {mention.cid && (
                <div style={{ fontSize: '0.8em', color: '#888', wordBreak: 'break-all', marginTop: 4 }}>
                    CID: {mention.cid}
                </div>
            )}
            <div className={styles.footer}>
                {mention.numTweets > 0 && (
                    <Link 
                        to={`/thread/${mention.threadId}`}
                        state={{ mentionData: mention }}
                        className={styles.readMore}
                        style={{ textDecoration: 'none' }}
                    >
                        Read {mention?.numTweets} tweets
                    </Link>
                )}
            </div>
        </div>
    );
};

export default MentionSummaryComponent;