import React from 'react';
import PostSummary from './PostSummary';
import styles from './PostList.module.css';
import { PostSummary as PostData } from '@src/client/models';

const PostList: React.FC<{ posts: PostData[] }> = ({ posts }) => {
    return (
        <div className={styles.postListContainer}>
            {posts.map((post, index) => (
                <PostSummary key={`${post.id}-${index}`} post={post} />
            ))}
        </div>
    );
};

export default PostList;