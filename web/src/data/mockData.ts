import { UserProfile, Thread, TweetData, Tweet } from '@src/types';
import { convTweetData2Thread } from '../types/format';
import mockTwitterData from './mock/1.json';


export const userProfile: UserProfile = mockTwitterData?.length > 0 ? (mockTwitterData[0] as Tweet)?.author : {
    id: '1403881130802225152',
    name: '大宇',
    screen_name: 'BTCdayu',
    profile_image_url: 'https://pbs.twimg.com/profile_images/1862767252400967680/mjEMe7kp_normal.jpg',
    description: '📙纯粹二级、不接商务\n📍不滞于物，自此精修',
    followers_count: 263038,
    friends_count: 3663,
    rest_id: '',
    statuses_count: 0,
    created_at: '2021-06-13T01:05:28Z',
    verified: false,
    is_blue_verified: false
};

export const mockTweetData: TweetData = mockTwitterData as TweetData;
export const threads: Thread[] =
    [mockTwitterData, mockTwitterData]
        .map(tweetData => convTweetData2Thread(tweetData));
export const twitterDetailData: Thread = convTweetData2Thread(mockTwitterData);