package xscraper

import (
	"context"
	"fmt"

	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/generated"
	"github.com/samber/lo"
)

// GetTweets returns the tweets with the given ID
func (x *XScraper) GetTweets(ctx context.Context, id string) ([]*Tweet, error) {
	p := generated.GetTweetDetailParams{}
	p.Variables.FocalTweetId = id
	p.Variables.Referrer = "home"
	p.Variables.RankingMode = "Relevance"
	p.Variables.IncludePromotedContent = true
	p.Variables.WithCommunity = true
	p.Variables.WithQuickPromoteEligibilityTweetFields = true
	p.Variables.WithBirdwatchNotes = true
	p.Variables.WithVoice = true

	p.Features.RwebVideoScreenEnabled = false
	p.Features.ProfileLabelImprovementsPcfLabelInPostEnabled = true
	p.Features.RwebTipjarConsumptionEnabled = true
	p.Features.VerifiedPhoneLabelEnabled = false
	p.Features.CreatorSubscriptionsTweetPreviewApiEnabled = true
	p.Features.ResponsiveWebGraphqlTimelineNavigationEnabled = true
	p.Features.ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled = false
	p.Features.PremiumContentApiReadEnabled = false
	p.Features.CommunitiesWebEnableTweetCommunityResultsFetch = true
	p.Features.C9sTweetAnatomyModeratorBadgeEnabled = true
	p.Features.ResponsiveWebGrokAnalyzeButtonFetchTrendsEnabled = false
	p.Features.ResponsiveWebGrokAnalyzePostFollowupsEnabled = true
	p.Features.ResponsiveWebJetfuelFrame = false
	p.Features.ResponsiveWebGrokShareAttachmentEnabled = true
	p.Features.ArticlesPreviewEnabled = true
	p.Features.ResponsiveWebEditTweetApiEnabled = true
	p.Features.GraphqlIsTranslatableRwebTweetIsTranslatableEnabled = true
	p.Features.ViewCountsEverywhereApiEnabled = true
	p.Features.LongformNotetweetsConsumptionEnabled = true
	p.Features.ResponsiveWebTwitterArticleTweetConsumptionEnabled = true
	p.Features.TweetAwardsWebTippingEnabled = false
	p.Features.ResponsiveWebGrokShowGrokTranslatedPost = false
	p.Features.ResponsiveWebGrokAnalysisButtonFromBackend = false
	p.Features.CreatorSubscriptionsQuoteTweetPreviewEnabled = false
	p.Features.FreedomOfSpeechNotReachFetchEnabled = true
	p.Features.StandardizedNudgesMisinfo = true
	p.Features.TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled = true
	p.Features.LongformNotetweetsRichTextReadEnabled = true
	p.Features.LongformNotetweetsInlineMediaEnabled = true
	p.Features.ResponsiveWebGrokImageAnnotationEnabled = true
	p.Features.ResponsiveWebEnhanceCardsEnabled = false

	p.FieldToggles.WithArticleRichContentState = true
	p.FieldToggles.WithArticlePlainText = false
	p.FieldToggles.WithGrokAnalyze = false
	p.FieldToggles.WithDisallowedReplyControls = false

	var resp generated.TweetDetailResponse
	err := x.GetGraphQL(ctx, "/i/api/graphql/xd_EMdYvB9hfZsZ6Idri0w/TweetDetail", &p, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get tweet detail: %w", err)
	}

	tweets, err := convertTimelineToTweets(resp.Data.ThreadedConversationWithInjectionsV2)
	if err != nil {
		return nil, fmt.Errorf("convert tweet: %w", err)
	}

	if len(tweets) == 0 {
		return nil, fmt.Errorf("no tweet found")
	}

	return tweets, nil
}

// SearchTweets searches for tweets matching the given query
func (x *XScraper) SearchTweets(ctx context.Context, query string, maxTweets int) ([]*Tweet, error) {
	p := generated.GetSearchTimelineParams{}

	if maxTweets == 0 {
		maxTweets = 20
	}
	if maxTweets > 50 {
		maxTweets = 50
	}
	p.Variables.Count = maxTweets
	p.Variables.RawQuery = query
	p.Variables.QuerySource = "typed_query"
	p.Variables.Product = "Top"

	p.Features.ArticlesPreviewEnabled = true
	p.Features.C9sTweetAnatomyModeratorBadgeEnabled = true
	p.Features.CommunitiesWebEnableTweetCommunityResultsFetch = true
	p.Features.CreatorSubscriptionsQuoteTweetPreviewEnabled = false
	p.Features.CreatorSubscriptionsTweetPreviewApiEnabled = true
	p.Features.FreedomOfSpeechNotReachFetchEnabled = true
	p.Features.GraphqlIsTranslatableRwebTweetIsTranslatableEnabled = true
	p.Features.LongformNotetweetsConsumptionEnabled = true
	p.Features.LongformNotetweetsInlineMediaEnabled = true
	p.Features.LongformNotetweetsRichTextReadEnabled = true
	p.Features.PremiumContentApiReadEnabled = false
	p.Features.ProfileLabelImprovementsPcfLabelInPostEnabled = true
	p.Features.ResponsiveWebEditTweetApiEnabled = true
	p.Features.ResponsiveWebEnhanceCardsEnabled = false
	p.Features.ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled = false
	p.Features.ResponsiveWebGraphqlTimelineNavigationEnabled = true
	p.Features.ResponsiveWebGrokAnalysisButtonFromBackend = false
	p.Features.ResponsiveWebGrokAnalyzeButtonFetchTrendsEnabled = false
	p.Features.ResponsiveWebGrokAnalyzePostFollowupsEnabled = true
	p.Features.ResponsiveWebGrokImageAnnotationEnabled = true
	p.Features.ResponsiveWebGrokShareAttachmentEnabled = true
	p.Features.ResponsiveWebGrokShowGrokTranslatedPost = false
	p.Features.ResponsiveWebJetfuelFrame = false
	p.Features.ResponsiveWebTwitterArticleTweetConsumptionEnabled = true
	p.Features.RwebTipjarConsumptionEnabled = true
	p.Features.StandardizedNudgesMisinfo = true
	p.Features.TweetAwardsWebTippingEnabled = false
	p.Features.TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled = true
	p.Features.VerifiedPhoneLabelEnabled = false
	p.Features.ViewCountsEverywhereApiEnabled = true

	var resp generated.SearchTimelineResponse
	err := x.GetGraphQL(ctx, "/i/api/graphql/VhUd6vHVmLBcw0uX-6jMLA/SearchTimeline", &p, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get tweet detail: %w", err)
	}

	tweets, err := convertTimelineToTweets(&resp.Data.SearchByRawQuery.SearchTimeline.Timeline)
	if err != nil {
		return nil, fmt.Errorf("convert tweets: %w", err)
	}

	return tweets, nil
}

// GetMentions returns the recent mentions of the user
func (x *XScraper) GetMentions(ctx context.Context) ([]*Tweet, error) {
	query := fmt.Sprintf("@%s", x.loginOpts.Username)
	maxTweets := 20
	// Fetch recent mentions
	tweets, err := x.SearchTweets(ctx, query, maxTweets)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentions: %w", err)
	}

	// Filter out tweets that are not mentions
	mentions := lo.Filter(tweets, func(tweet *Tweet, _ int) bool {
		return tweet.Author.ScreenName != x.loginOpts.Username
	})

	return mentions, nil
}
