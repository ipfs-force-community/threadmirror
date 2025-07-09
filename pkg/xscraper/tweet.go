package xscraper

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/generated"
	"github.com/samber/lo"
)

// GetTweets returns the tweets with the given ID and completeness information
func (x *XScraper) GetTweets(ctx context.Context, id string) (*TweetsResult, error) {
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

	// Handle GraphQL-level errors
	if resp.Errors != nil && len(*resp.Errors) > 0 {
		msgs := lo.Map(*resp.Errors, func(e generated.ErrorResponse, _ int) string { return e.Message })
		return nil, fmt.Errorf("tweet detail: %s", strings.Join(msgs, "; "))
	}

	tweetsResult, err := convertTimelineToTweets(resp.Data.ThreadedConversationWithInjectionsV2)
	if err != nil {
		return nil, fmt.Errorf("convert tweet: %w", err)
	}

	if len(tweetsResult.Tweets) == 0 {
		return nil, fmt.Errorf("no tweet found")
	}

	return tweetsResult, nil
}

// GetTweetDetail 调用 TweetDetail GraphQL接口，返回目标推文及其线程（可能包含多条推文）。
// 参数仅需 tweet `id`，其余字段按 OpenAPI 默认/示例填充。
func (x *XScraper) GetTweetDetail(ctx context.Context, id string) ([]*Tweet, error) {
	p := generated.GetTweetDetailParams{}

	// Variables
	p.Variables.FocalTweetId = id
	p.Variables.Referrer = "home"
	p.Variables.RankingMode = "Relevance"
	p.Variables.IncludePromotedContent = true
	p.Variables.WithCommunity = true
	p.Variables.WithQuickPromoteEligibilityTweetFields = true
	p.Variables.WithBirdwatchNotes = true
	p.Variables.WithVoice = true
	p.Variables.WithRuxInjections = false

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
	p.Features.RwebVideoScreenEnabled = false
	p.Features.StandardizedNudgesMisinfo = true
	p.Features.TweetAwardsWebTippingEnabled = false
	p.Features.TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled = true
	p.Features.VerifiedPhoneLabelEnabled = false
	p.Features.ViewCountsEverywhereApiEnabled = true

	// Field toggles
	p.FieldToggles.WithArticleRichContentState = true
	p.FieldToggles.WithArticlePlainText = false
	p.FieldToggles.WithGrokAnalyze = false
	p.FieldToggles.WithDisallowedReplyControls = false

	var resp generated.TweetDetailResponse
	var berr *BadRequestError
	err := x.GetGraphQL(ctx, "/i/api/graphql/xd_EMdYvB9hfZsZ6Idri0w/TweetDetail", &p, &resp)
	if err != nil {
		if errors.As(err, &berr) && berr.StatusCode == http.StatusNotFound {
			return []*Tweet{}, nil
		}
		return nil, fmt.Errorf("failed to get tweet detail: %w", err)
	}

	// Handle GraphQL-level errors
	if resp.Errors != nil && len(*resp.Errors) > 0 {
		msgs := lo.Map(*resp.Errors, func(e generated.ErrorResponse, _ int) string { return e.Message })
		return nil, fmt.Errorf("tweet detail: %s", strings.Join(msgs, "; "))
	}

	tweetsResult, err := convertTimelineToTweets(resp.Data.ThreadedConversationWithInjectionsV2)
	if err != nil {
		return nil, fmt.Errorf("convert tweet: %w", err)
	}
	return tweetsResult.Tweets, nil
}

// GetTweetDetail returns the tweet with the given ID
func (x *XScraper) GetTweetResultByRestId(ctx context.Context, id string) (*Tweet, error) {
	p := generated.GetTweetResultByRestIdParams{}
	p.Variables.TweetId = id
	p.Variables.IncludePromotedContent = false
	p.Variables.WithCommunity = false
	p.Variables.WithVoice = false

	// Feature toggles (aligned with other calls)
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
	p.Features.ResponsiveWebEditTweetApiEnabled = true
	p.Features.ResponsiveWebEnhanceCardsEnabled = false
	p.Features.ResponsiveWebGraphqlExcludeDirectiveEnabled = true
	p.Features.ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled = false
	p.Features.ResponsiveWebGraphqlTimelineNavigationEnabled = true
	p.Features.ResponsiveWebTwitterArticleTweetConsumptionEnabled = true
	p.Features.RwebTipjarConsumptionEnabled = true
	p.Features.RwebVideoTimestampsEnabled = true
	p.Features.StandardizedNudgesMisinfo = true
	p.Features.TweetAwardsWebTippingEnabled = false
	p.Features.TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled = true
	p.Features.TweetWithVisibilityResultsPreferGqlMediaInterstitialEnabled = true
	p.Features.TweetypieUnmentionOptimizationEnabled = true
	p.Features.VerifiedPhoneLabelEnabled = false
	p.Features.ViewCountsEverywhereApiEnabled = true

	// Field toggles
	p.FieldToggles.WithArticlePlainText = false
	p.FieldToggles.WithArticleRichContentState = true

	var resp generated.TweetResultByRestIdResponse
	var berr *BadRequestError
	err := x.GetGraphQL(ctx, "/i/api/graphql/7xflPyRiUxGVbJd4uWmbfg/TweetResultByRestId", &p, &resp)
	if err != nil {
		if errors.As(err, &berr) && berr.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("tweet not found")
		}
		return nil, fmt.Errorf("failed to get tweet detail: %w", err)
	}

	// Handle GraphQL-level errors
	if resp.Errors != nil && len(*resp.Errors) > 0 {
		msgs := lo.Map(*resp.Errors, func(e generated.ErrorResponse, _ int) string { return e.Message })
		return nil, fmt.Errorf("tweet detail: %s", strings.Join(msgs, "; "))
	}

	if resp.Data.TweetResult == nil || resp.Data.TweetResult.Result == nil {
		return nil, fmt.Errorf("no tweet found")
	}

	genTweet, err := resp.Data.TweetResult.Result.AsTweet()
	if err != nil {
		return nil, fmt.Errorf("result is not a tweet: %w", err)
	}

	tweet, err := convertGeneratedTweetToTweet(&genTweet)
	if err != nil {
		return nil, fmt.Errorf("convert tweet: %w", err)
	}

	return tweet, nil
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
	var berr *BadRequestError
	err := x.GetGraphQL(ctx, "/i/api/graphql/VhUd6vHVmLBcw0uX-6jMLA/SearchTimeline", &p, &resp)
	if err != nil {
		if errors.As(err, &berr) && (berr.StatusCode == http.StatusNotFound) {
			return []*Tweet{}, nil
		}
		return nil, fmt.Errorf("failed to get tweet detail: %w", err)
	}

	// Handle GraphQL-level errors
	if resp.Errors != nil && len(*resp.Errors) > 0 {
		msgs := lo.Map(*resp.Errors, func(e generated.ErrorResponse, _ int) string { return e.Message })
		return nil, fmt.Errorf("search tweets: %s", strings.Join(msgs, "; "))
	}

	tweetsResult, err := convertTimelineToTweets(&resp.Data.SearchByRawQuery.SearchTimeline.Timeline)
	if err != nil {
		return nil, fmt.Errorf("convert tweets: %w", err)
	}

	return tweetsResult.Tweets, nil
}

// GetMentions returns the recent mentions of the user
func (x *XScraper) GetMentions(ctx context.Context, filter func(*Tweet) bool) ([]*Tweet, error) {
	return x.GetMentionsByScreenName(ctx, x.LoginOpts.Username, filter)
}

// GetMentionsByScreenName returns the recent mentions of the user
func (x *XScraper) GetMentionsByScreenName(ctx context.Context, screenName string, filter func(*Tweet) bool) ([]*Tweet, error) {
	query := fmt.Sprintf("(@%s) filter:replies", screenName)
	maxTweets := 20
	// Fetch recent mentions
	tweets, err := x.SearchTweets(ctx, query, maxTweets)
	if err != nil {
		return nil, fmt.Errorf("failed to get mentions: %w", err)
	}

	return lo.Filter(tweets, func(tweet *Tweet, _ int) bool {
		if tweet.RestID == "" {
			return false
		}

		if tweet.Author.ScreenName == screenName {
			return false
		}

		return filter(tweet)
	}), nil
}

type NewTweet struct {
	Text                    string
	MediaIDs                []string
	TaggedUsers             [][]string
	InReplyToTweetId        *string
	AttachmentUrl           *string
	ConversationControlMode *string
}

func (x *XScraper) CreateTweet(ctx context.Context, newTweet NewTweet) (*Tweet, error) {
	r := generated.PostCreateTweetJSONBody{}

	r.QueryId = "IID9x6WsdMnTlXnzXGq8ng"
	r.Features.ArticlesPreviewEnabled = true
	r.Features.C9sTweetAnatomyModeratorBadgeEnabled = true
	r.Features.CommunitiesWebEnableTweetCommunityResultsFetch = true
	r.Features.CreatorSubscriptionsQuoteTweetPreviewEnabled = false
	r.Features.FreedomOfSpeechNotReachFetchEnabled = true
	r.Features.GraphqlIsTranslatableRwebTweetIsTranslatableEnabled = true
	r.Features.LongformNotetweetsConsumptionEnabled = true
	r.Features.LongformNotetweetsInlineMediaEnabled = true
	r.Features.LongformNotetweetsRichTextReadEnabled = true
	r.Features.PremiumContentApiReadEnabled = false
	r.Features.ProfileLabelImprovementsPcfLabelInPostEnabled = true
	r.Features.ResponsiveWebEditTweetApiEnabled = true
	r.Features.ResponsiveWebEnhanceCardsEnabled = false
	r.Features.ResponsiveWebGraphqlSkipUserProfileImageExtensionsEnabled = false
	r.Features.ResponsiveWebGraphqlTimelineNavigationEnabled = true
	r.Features.ResponsiveWebGrokAnalysisButtonFromBackend = false
	r.Features.ResponsiveWebGrokAnalyzeButtonFetchTrendsEnabled = false
	r.Features.ResponsiveWebGrokAnalyzePostFollowupsEnabled = true
	r.Features.ResponsiveWebGrokImageAnnotationEnabled = true
	r.Features.ResponsiveWebGrokShareAttachmentEnabled = true
	r.Features.ResponsiveWebGrokShowGrokTranslatedPost = false
	r.Features.ResponsiveWebJetfuelFrame = false
	r.Features.ResponsiveWebTwitterArticleTweetConsumptionEnabled = true
	r.Features.RwebTipjarConsumptionEnabled = true
	r.Features.StandardizedNudgesMisinfo = true
	r.Features.TweetAwardsWebTippingEnabled = false
	r.Features.TweetWithVisibilityResultsPreferGqlLimitedActionsPolicyEnabled = true
	r.Features.VerifiedPhoneLabelEnabled = false
	r.Features.ViewCountsEverywhereApiEnabled = true

	r.Variables.AttachmentUrl = newTweet.AttachmentUrl

	r.Variables.TweetText = newTweet.Text

	if len(newTweet.MediaIDs) > 0 {
		mediaEntities := lo.Map(newTweet.MediaIDs, func(mediaID string, index int) struct {
			MediaId     string   `json:"media_id"`
			TaggedUsers []string `json:"tagged_users"`
		} {
			taggedUsers := []string{}
			if index < len(newTweet.TaggedUsers) {
				taggedUsers = newTweet.TaggedUsers[index]
			}
			return struct {
				MediaId     string   `json:"media_id"`
				TaggedUsers []string `json:"tagged_users"`
			}{
				MediaId:     mediaID,
				TaggedUsers: taggedUsers,
			}
		})
		r.Variables.Media.MediaEntities = &mediaEntities
	}

	if newTweet.InReplyToTweetId != nil {
		r.Variables.Reply = &struct {
			ExcludeReplyUserIds []map[string]any `json:"exclude_reply_user_ids"`
			InReplyToTweetId    string           `json:"in_reply_to_tweet_id"`
		}{
			ExcludeReplyUserIds: []map[string]any{},
			InReplyToTweetId:    *newTweet.InReplyToTweetId,
		}
	}

	r.Variables.DisallowedReplyOptions = nil

	if newTweet.ConversationControlMode != nil {
		r.Variables.ConversationControl = &struct {
			Mode string "json:\"mode\""
		}{
			Mode: *newTweet.ConversationControlMode,
		}
	}

	var resp generated.CreateTweetResponse
	err := x.PostGraphQL(ctx, "/i/api/graphql/IID9x6WsdMnTlXnzXGq8ng/CreateTweet", &r, &resp)
	if err != nil {
		return nil, fmt.Errorf("create tweet: %w", err)
	}

	// Validate response structure
	if resp.Errors != nil {
		if len(*resp.Errors) == 0 {
			return nil, fmt.Errorf("create tweet: unknown error")
		}
		msgs := lo.Map(*resp.Errors, func(e generated.ErrorResponse, _ int) string { return e.Message })
		return nil, fmt.Errorf("create tweet: %s", strings.Join(msgs, "; "))
	}

	// Convert generated tweet to simplified Tweet struct
	tweet, err := convertGeneratedTweetToTweet(&resp.Data.CreateTweet.TweetResults.Result)
	if err != nil {
		return nil, fmt.Errorf("convert tweet: %w", err)
	}

	return tweet, nil
}
