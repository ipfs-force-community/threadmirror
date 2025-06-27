import React from 'react';
import { Link } from 'react-router-dom';
import styles from './PostSummary.module.css';
import { Post as PostData } from '@src/client/models';

const PostSummaryComponent: React.FC<{ post: PostData }> = ({ post }) => {
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
        window.open(`https://twitter.com/${post?.author?.screenName}`, '_blank', 'noopener,noreferrer');
    };

    return (
        <Link
            to={`/posts/${post.id}`}
            state={{ postData: post }}
            className={styles.postContainer}
            style={{ textDecoration: 'none', color: 'inherit' }}
        >
            <div className={styles.postHeader}>
                <img
                    src={post?.author?.profileImageUrl}
                    alt={`${post?.author?.name} profile`}
                    className={styles.profileImage}
                />
                <div className={styles.authorInfo}>
                    <span className={styles.authorName}>{post?.author?.name}</span>
                    <span className={styles.screenName}>
                        <span 
                            onClick={handleTwitterLinkClick}
                            style={{ textDecoration: 'none', color: '#1da1f2', cursor: 'pointer' }}
                        >
                            @{post?.author?.screenName}
                        </span>
                    </span>
                </div>
                <span className={styles.createdAt}>{formatDate(post.createdAt)}</span>
            </div>
            <div className={styles.contentPreview}>
                {post.contentPreview}
            </div>
            <div className={styles.footer}>
                <button className={styles.readMore}>Read {post?.threads?.length ? post.threads.length : 0} tweets</button>
            </div>
        </Link>
    );
};

export default PostSummaryComponent;