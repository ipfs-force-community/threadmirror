import ThreadItem from './ThreadItem';
import { Thread } from '@src/types';

interface ThreadListProps {
    threads: Thread[];
}

const ThreadList = ({ threads }: ThreadListProps) => {
    const sortedThreads = [...threads].sort((a, b) => {
        const dateA = new Date(a.date).getTime();
        const dateB = new Date(b.date).getTime();
        if (dateA !== dateB) return dateB - dateA; // Sort by date descending
        return b.id.localeCompare(a.id); // Sort by id descending if dates are equal
    });
    return (
        <div>
            {sortedThreads.map((thread, idx) => (
                <ThreadItem key={`${thread.id}-${idx}`} thread={thread} />
            ))}
        </div>
    );
};

export default ThreadList;