import { ThreadDetail } from '@src/client/models';

// 模拟Thread数据
const originThreads: ThreadDetail[] = [
  {
    id: "thread_001",
    cid: "bafkreig4666666666666666666666666666666666666666666666666666666666666666",
    contentPreview: "最近研究了区块链的零知识证明技术，这里分享一些关于ZK-SNARKs和ZK-STARKs的见解...",
    createdAt: new Date("2023-08-15T08:42:35Z"),
    tweets: [
      {
        id: "1691375479704870912",
        restId: "1691375479704870912",
        text: "最近研究了区块链的零知识证明技术，这里分享一些关于ZK-SNARKs和ZK-STARKs的见解。\n\n(1/5) 零知识证明允许你在不泄露任何信息的情况下证明你知道某个秘密。这对于区块链隐私和扩展性至关重要。",
        createdAt: new Date("2023-08-15T08:42:35Z"),
        author: {
          id: "1403881130802225152",
          restId: "1403881130802225152",
          name: "大宇",
          screenName: "BTCdayu",
          profileImageUrl: "https://pbs.twimg.com/profile_images/1862767252400967680/mjEMe7kp_normal.jpg",
          description: "📙纯粹二级、不接商务\n📍不滞于物，自此精修",
          followersCount: 263038,
          friendsCount: 3663,
          statusesCount: 2175,
          createdAt: new Date("2021-06-13T01:05:28Z"),
          verified: false,
          isBlueVerified: false
        },
        entities: {
          hashtags: [],
          symbols: [],
          urls: [],
          userMentions: []
        },
        stats: {
          replyCount: 23,
          retweetCount: 142,
          favoriteCount: 561,
          quoteCount: 7,
          bookmarkCount: 89,
          viewCount: 120538,
          isNoteTweet: false,
          richtext: { richtextTags: [] }
        },
        isRetweet: false,
        isReply: false,
        isQuoteStatus: false,
        conversationId: "1691375479704870912",
        lang: "zh",
        source: "Twitter Web App",
        possiblySensitive: false,
        isTranslatable: true,
        hasBirdwatchNotes: false,
        isNoteTweet: false,
        richtext: { richtextTags: [] }
      },
      {
        id: "1691375485119811584",
        restId: "1691375485119811584",
        text: "(2/5) ZK-SNARKs (简洁的非交互式零知识证明)：\n- 优点：证明大小小，验证快\n- 缺点：需要可信设置\n- 应用：Zcash, Tornado Cash\n\n这种技术让交易完全私密，但初始设置阶段需要可信第三方。",
        createdAt: new Date("2023-08-15T08:42:37Z"),
        author: {
          id: "1403881130802225152",
          restId: "1403881130802225152",
          name: "大宇",
          screenName: "BTCdayu",
          profileImageUrl: "https://pbs.twimg.com/profile_images/1862767252400967680/mjEMe7kp_normal.jpg",
          description: "📙纯粹二级、不接商务\n📍不滞于物，自此精修",
          followersCount: 263038,
          friendsCount: 3663,
          statusesCount: 2175,
          createdAt: new Date("2021-06-13T01:05:28Z"),
          verified: false,
          isBlueVerified: false
        },
        entities: {
          hashtags: [],
          symbols: [],
          urls: [],
          userMentions: []
        },
        stats: {
          replyCount: 12,
          retweetCount: 98,
          favoriteCount: 426,
          quoteCount: 3,
          bookmarkCount: 67,
          viewCount: 89254,
          isNoteTweet: false,
          richtext: { richtextTags: [] }
        },
        isRetweet: false,
        isReply: true,
        isQuoteStatus: false,
        conversationId: "1691375479704870912",
        inReplyToStatusId: "1691375479704870912",
        inReplyToUserId: "1403881130802225152",
        lang: "zh",
        source: "Twitter Web App",
        possiblySensitive: false,
        isTranslatable: true,
        hasBirdwatchNotes: false,
        isNoteTweet: false,
        richtext: { richtextTags: [] }
      }
    ],
    numTweets: 2,
  },
  {
    id: "thread_002",
    cid: "bafkreig4666666666666666666666666666666666666666666666666666666666666666",
    contentPreview: "今日比特币分析：市场情绪指数显示我们可能正处在恐慌区域，这往往是长期投资者的良好买入时机...",
    createdAt: new Date("2023-09-22T15:28:10Z"),
    tweets: [
      {
        id: "1705231859417456993",
        restId: "1705231859417456993",
        text: "今日比特币分析：市场情绪指数显示我们可能正处在恐慌区域，这往往是长期投资者的良好买入时机。\n\n从技术面来看，BTC在19,800美元附近有强烈支撑，如果能守住这一水平，我们可能会看到反弹至22,000-23,000美元区间。",
        createdAt: new Date("2023-09-22T15:28:10Z"),
        author: {
          id: "1403881130802225152",
          restId: "1403881130802225152",
          name: "大宇",
          screenName: "BTCdayu",
          profileImageUrl: "https://pbs.twimg.com/profile_images/1862767252400967680/mjEMe7kp_normal.jpg",
          description: "📙纯粹二级、不接商务\n📍不滞于物，自此精修",
          followersCount: 263038,
          friendsCount: 3663,
          statusesCount: 2175,
          createdAt: new Date("2021-06-13T01:05:28Z"),
          verified: false,
          isBlueVerified: false
        },
        entities: {
          hashtags: [],
          symbols: [],
          urls: [],
          userMentions: []
        },
        stats: {
          replyCount: 45,
          retweetCount: 218,
          favoriteCount: 892,
          quoteCount: 15,
          bookmarkCount: 156,
          viewCount: 198742,
          isNoteTweet: false,
          richtext: { richtextTags: [] }
        },
        isRetweet: false,
        isReply: false,
        isQuoteStatus: false,
        conversationId: "1705231859417456993",
        lang: "zh",
        source: "Twitter for iPhone",
        possiblySensitive: false,
        isTranslatable: true,
        hasBirdwatchNotes: false,
        isNoteTweet: false,
        richtext: { richtextTags: [] }
      }
    ],
    numTweets: 1,
  },
  {
    id: "thread_003",
    cid: "bafkreig4666666666666666666666666666666666666666666666666666666666666666",
    contentPreview: "分享一个我正在使用的交易策略：结合RSI和MACD的背离信号，配合成交量确认，这在最近的市场环境中效果不错...",
    createdAt: new Date("2023-10-05T12:15:42Z"),
    tweets: [
      {
        id: "1709912485720211816",
        restId: "1709912485720211816",
        text: "分享一个我正在使用的交易策略：结合RSI和MACD的背离信号，配合成交量确认，这在最近的市场环境中效果不错。\n\n特别是在4小时图表上，当我们看到价格创新低但RSI没有跟随时，常常预示着反转即将到来。我会在推文中持续分享更多实例。",
        createdAt: new Date("2023-10-05T12:15:42Z"),
        author: {
          id: "1403881130802225152",
          restId: "1403881130802225152",
          name: "大宇",
          screenName: "BTCdayu",
          profileImageUrl: "https://pbs.twimg.com/profile_images/1862767252400967680/mjEMe7kp_normal.jpg",
          description: "📙纯粹二级、不接商务\n📍不滞于物，自此精修",
          followersCount: 263038,
          friendsCount: 3663,
          statusesCount: 2175,
          createdAt: new Date("2021-06-13T01:05:28Z"),
          verified: false,
          isBlueVerified: false
        },
        entities: {
          hashtags: [
            {
              "text": "交易策略",
              "indices": [5, 10]
            }
          ],
          symbols: [
            {
              "text": "BTC",
              "indices": [120, 123]
            }
          ],
          urls: [],
          userMentions: []
        },
        stats: {
          replyCount: 56,
          retweetCount: 325,
          favoriteCount: 1245,
          quoteCount: 28,
          bookmarkCount: 287,
          viewCount: 215683,
          isNoteTweet: false,
          richtext: { richtextTags: [] }
        },
        isRetweet: false,
        isReply: false,
        isQuoteStatus: false,
        conversationId: "1709912485720211816",
        lang: "zh",
        source: "Twitter Web App",
        possiblySensitive: false,
        isTranslatable: true,
        hasBirdwatchNotes: false,
        isNoteTweet: false,
        richtext: { richtextTags: [] }
      },
      {
        id: "1709912490698314957",
        restId: "1709912490698314957",
        text: "这里给大家展示一个近期的例子：\n\n在9月底，BTC价格跌至19,000以下，但4小时RSI创建了更高低点，形成明显背离。同时，MACD柱状图开始减缓并转向，成交量也逐步降低。\n\n这些信号组合在一起，给出了很强的买入提示。如果你按此操作，现在已经有20%的收益了。",
        createdAt: new Date("2023-10-05T12:15:43Z"),
        author: {
          id: "1403881130802225152",
          restId: "1403881130802225152",
          name: "大宇",
          screenName: "BTCdayu",
          profileImageUrl: "https://pbs.twimg.com/profile_images/1862767252400967680/mjEMe7kp_normal.jpg",
          description: "📙纯粹二级、不接商务\n📍不滞于物，自此精修",
          followersCount: 263038,
          friendsCount: 3663,
          statusesCount: 2175,
          createdAt: new Date("2021-06-13T01:05:28Z"),
          verified: false,
          isBlueVerified: false
        },
        entities: {
          hashtags: [],
          symbols: [
            {
              "text": "BTC",
              "indices": [16, 19]
            }
          ],
          urls: [],
          userMentions: []
        },
        stats: {
          replyCount: 42,
          retweetCount: 267,
          favoriteCount: 1098,
          quoteCount: 19,
          bookmarkCount: 246,
          viewCount: 184392,
          isNoteTweet: false,
          richtext: { richtextTags: [] }
        },
        isRetweet: false,
        isReply: true,
        isQuoteStatus: false,
        conversationId: "1709912485720211816",
        inReplyToStatusId: "1709912485720211816",
        inReplyToUserId: "1403881130802225152",
        lang: "zh",
        source: "Twitter Web App",
        possiblySensitive: false,
        isTranslatable: true,
        hasBirdwatchNotes: false,
        isNoteTweet: false,
        richtext: { richtextTags: [] }
      }
    ],
    numTweets: 2,
  },
];

const generateData = (count: number, startId: number = 1000): ThreadDetail[] => {
  const result: ThreadDetail[] = [];
  const templateThreads = originThreads;

  for (let i = 0; i < count; i++) {
    // 从模板中随机选择一个thread作为基础
    const templateIndex = i % templateThreads.length;
    const template = templateThreads[templateIndex];

    // 创建新thread，复制模板并修改部分属性
    const newThread: ThreadDetail = {
      ...template,
      id: `thread_${startId + i}`,
      createdAt: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000), // 随机设置在过去30天内的时间
    };

    // 为每个thread生成新的tweets，确保ID唯一
    if (newThread.tweets && newThread.tweets.length > 0) {
      newThread.tweets = newThread.tweets.map((tweet, idx) => {
        const tweetDate = new Date(newThread.createdAt.getTime());
        return {
          ...tweet,
          id: `tweet_${startId + i}_${idx}`,
          restId: `tweet_${startId + i}_${idx}`,
          createdAt: tweetDate,
          // 确保回复的时间稍晚于原thread
          ...(idx > 0 ? { createdAt: new Date(tweetDate.getTime() + idx * 60000) } : {})
        };
      });

      // 更新contentPreview以反映第一个tweet的内容
      if (newThread.tweets[0]) {
        const firstTweetText = newThread.tweets[0].text;
        newThread.contentPreview = firstTweetText.length > 100
          ? firstTweetText.substring(0, 97) + '...'
          : firstTweetText;
      }
    }

    result.push(newThread);
  }

  return result;
};

export const mockThreads: ThreadDetail[] = generateData(17);