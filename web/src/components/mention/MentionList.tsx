import React from 'react';
import MentionSummary from './MentionSummary';
import styles from './MentionList.module.css';
import { MentionSummary as MentionData } from '@src/client/models';

const MentionList: React.FC<{ mentions: MentionData[] }> = ({ mentions }) => {
    return (
        <div className={styles.mentionListContainer}>
            {mentions.map((mention, index) => (
                <MentionSummary key={`${mention.id}-${index}`} mention={mention} />
            ))}
        </div>
    );
};

export default MentionList;