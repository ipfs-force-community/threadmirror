import { Link } from 'react-router-dom';
import { Thread } from '@src/types';
import { formatDate } from '@utils/formatDate';
import './ThreadItem.css';

interface ThreadItemProps {
    thread: Thread;
}

const ThreadItem = ({ thread }: ThreadItemProps) => {
    const formattedDate = formatDate(thread.date);

    const truncatedContent = thread.content.length > 150
        ? thread.content.substring(0, 150) + '...'
        : thread.content;

    return (
        <Link to={`/thread/${thread.id}`} className="thread-item" style={{ textDecoration: 'none', color: 'inherit' }}>
            <div className="thread-item-content">
                <div className="thread-meta">
                    <span>{formattedDate}</span> • <span>{thread.tweetCount} tweets</span> • <span>{thread.readingTime} read</span>
                </div>
                <div className="thread-summary" dangerouslySetInnerHTML={{ __html: truncatedContent }} />
            </div>
        </Link>
    );
};

export default ThreadItem;