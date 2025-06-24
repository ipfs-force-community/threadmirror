// Define the structure for the 'author' object
interface Author {
    id: string;
    rest_id: string;
    name: string;
    screen_name: string;
    profile_image_url: string;
    description: string;
    followers_count: number;
    friends_count: number;
    statuses_count: number;
    created_at: string;
    verified: boolean;
    is_blue_verified: boolean;
}

// Define the structure for the 'media' object within 'entities'
interface Media {
    id: string;
    media_key: string;
    type: string;
    url: string;
    display_url: string;
    expanded_url: string;
    width: number;
    height: number;
}

// Define the structure for the 'entities' object
interface Entities {
    hashtags: null | any[]; // Assuming it can be null or an array of something, adjust if specific type is known
    symbols: null | any[];
    urls: null | any[];
    user_mentions: null | any[];
    media?: Media[]; // Optional, as it's not present in all objects
}

// Define the structure for the 'stats' object
interface Stats {
    reply_count: number;
    retweet_count: number;
    favorite_count: number;
    quote_count: number;
    bookmark_count: number;
    view_count: number;
}

// Define the structure for the 'edit_control' object
interface EditControl {
    edit_tweet_ids: string[];
    editable_until_msecs: string;
    edits_remaining: string;
    is_edit_eligible: boolean;
}

// Define the structure for the 'views' object
interface Views {
    count: string;
    state: string;
}

// Define the main structure for a single tweet/thread item
export interface Tweet {
    id: string;
    rest_id: string;
    text: string;
    created_at: string;
    author: Author;
    entities: Entities;
    stats: Stats;
    is_retweet: boolean;
    is_reply: boolean;
    is_quote_status: boolean;
    conversation_id: string;
    in_reply_to_status_id?: string; // Optional
    in_reply_to_user_id?: string;   // Optional
    has_birdwatch_notes: boolean;
    lang: string;
    source: string;
    possibly_sensitive: boolean;
    is_translatable: boolean;
    edit_control: EditControl;
    views: Views;
}

// Define the type for the array of tweets
export type TweetData = Tweet[];
export type UserProfile = Author;
export interface Thread {
    id: string;
    title: string;

    date: string;
    tweetCount: number;
    readingTime: string;
    content: string;
    images?: string[];
    links?: string[];

    data: TweetData;
}


export interface UserListResponse {
    auth: Author;
    last_text: string;
}

export interface UserTwitterResponse {
    auth: Author;
    msg: TweetData[];
    total?: number;
}

export type TwitterDetailResponse = TweetData;

