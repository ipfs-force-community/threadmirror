
import { Thread, Tweet, TweetData } from './index';


export const convTweetData2Thread = (tweetData: TweetData): Thread => {
    if (!tweetData || !Array.isArray(tweetData) || !tweetData[0]) {
        throw new Error('Invalid tweet data');
    }

    const firstTweet = tweetData[0] as Tweet;
    if (!firstTweet.author) {
        console.warn('Missing author in tweet data:', tweetData);
        throw new Error('Missing author in tweet data');
    }

    const allTextLengths = tweetData.reduce((sum: number, tweet: Tweet) => sum + tweet.text.length, 0);

    const thread: Thread = {
        id: firstTweet.id,
        title: firstTweet.text.slice(0, 50) + (firstTweet.text.length > 50 ? '...' : ''),
        date: firstTweet.created_at || '',
        tweetCount: tweetData.length,
        readingTime: `${Math.ceil(allTextLengths / 280)} min`,
        content: firstTweet.text
            .split('\n')
            .map(line => `<p>${line}</p>`)
            .join(''),
        images: firstTweet.entities.media?.map(media => media.url) || [],
        links: firstTweet.entities.urls?.map(url => url.expanded_url) || [],
        data: tweetData as Tweet[],
    };

    return thread;
};