package xscraper

import (
	"testing"

	"github.com/ipfs-force-community/threadmirror/pkg/xscraper/generated"
)

func TestMergeEntities(t *testing.T) {
	// Test case 1: Merge hashtags
	t.Run("MergeHashtags", func(t *testing.T) {
		original := generated.Entities{
			Hashtags: []generated.Hashtag{
				{"text": "golang"},
			},
		}
		updated := generated.Entities{
			Hashtags: []generated.Hashtag{
				{"text": "golang"},  // duplicate
				{"text": "twitter"}, // new
			},
		}

		merged := MergeEntities(original, updated)

		if len(merged.Hashtags) != 2 {
			t.Errorf("Expected 2 hashtags after merge, got %d", len(merged.Hashtags))
		}

		// Check that both hashtags are present
		foundGolang, foundTwitter := false, false
		for _, h := range merged.Hashtags {
			if h["text"] == "golang" {
				foundGolang = true
			}
			if h["text"] == "twitter" {
				foundTwitter = true
			}
		}
		if !foundGolang || !foundTwitter {
			t.Errorf("Expected both 'golang' and 'twitter' hashtags in merged result")
		}
	})

	// Test case 2: Merge URLs
	t.Run("MergeURLs", func(t *testing.T) {
		original := generated.Entities{
			Urls: []generated.Url{
				{Url: "https://example.com", DisplayUrl: "example.com"},
			},
		}
		updated := generated.Entities{
			Urls: []generated.Url{
				{Url: "https://golang.org", DisplayUrl: "golang.org"},
			},
		}

		merged := MergeEntities(original, updated)

		if len(merged.Urls) != 2 {
			t.Errorf("Expected 2 URLs after merge, got %d", len(merged.Urls))
		}

		// Check URLs are preserved
		urls := make(map[string]bool)
		for _, u := range merged.Urls {
			urls[u.Url] = true
		}
		if !urls["https://example.com"] || !urls["https://golang.org"] {
			t.Errorf("Expected both URLs to be present in merged result")
		}
	})

	// Test case 3: Merge media with nil handling
	t.Run("MergeMediaWithNil", func(t *testing.T) {
		originalMedia := []generated.Media{
			{IdStr: "123", MediaKey: "key123"},
		}

		original := generated.Entities{Media: &originalMedia}
		updated := generated.Entities{Media: nil} // nil media

		merged := MergeEntities(original, updated)

		if merged.Media == nil || len(*merged.Media) != 1 {
			t.Errorf("Expected 1 media item after merge with nil, got %v", merged.Media)
		}
	})

	// Test case 4: Merge with duplicate media
	t.Run("MergeDuplicateMedia", func(t *testing.T) {
		originalMedia := []generated.Media{
			{IdStr: "123", MediaKey: "key123", DisplayUrl: "pic.twitter.com/123"},
		}
		updatedMedia := []generated.Media{
			{IdStr: "123", MediaKey: "key123", DisplayUrl: "pic.twitter.com/123"}, // duplicate
			{IdStr: "456", MediaKey: "key456", DisplayUrl: "pic.twitter.com/456"}, // new
		}

		original := generated.Entities{Media: &originalMedia}
		updated := generated.Entities{Media: &updatedMedia}

		merged := MergeEntities(original, updated)

		if merged.Media == nil || len(*merged.Media) != 2 {
			t.Errorf("Expected 2 media items after merge, got %v", merged.Media)
		}

		// Check that media IDs are correct
		mediaIds := make(map[string]bool)
		for _, m := range *merged.Media {
			mediaIds[m.IdStr] = true
		}
		if !mediaIds["123"] || !mediaIds["456"] {
			t.Errorf("Expected media IDs 123 and 456 in merged result")
		}
	})

	// Test case 5: Merge timestamps
	t.Run("MergeTimestamps", func(t *testing.T) {
		originalTimestamps := []generated.Timestamp{
			{Seconds: 1234567890, Text: "timestamp1", Indices: []int{0, 10}},
		}
		updatedTimestamps := []generated.Timestamp{
			{Seconds: 1234567890, Text: "timestamp1", Indices: []int{0, 10}},  // duplicate
			{Seconds: 1234567891, Text: "timestamp2", Indices: []int{11, 21}}, // new
		}

		original := generated.Entities{Timestamps: &originalTimestamps}
		updated := generated.Entities{Timestamps: &updatedTimestamps}

		merged := MergeEntities(original, updated)

		if merged.Timestamps == nil || len(*merged.Timestamps) != 2 {
			t.Errorf("Expected 2 timestamps after merge, got %v", merged.Timestamps)
		}
	})
}

func TestEntityEqualityFunctions(t *testing.T) {
	// Test Media equality
	t.Run("MediaEquality", func(t *testing.T) {
		media1 := generated.Media{IdStr: "123", MediaKey: "key123"}
		media2 := generated.Media{IdStr: "123", MediaKey: "key123"}
		media3 := generated.Media{IdStr: "456", MediaKey: "key456"}

		if !isMediaEqual(media1, media2) {
			t.Error("Expected media1 and media2 to be equal")
		}
		if isMediaEqual(media1, media3) {
			t.Error("Expected media1 and media3 to be different")
		}
	})

	// Test Hashtag equality
	t.Run("HashtagEquality", func(t *testing.T) {
		hashtag1 := generated.Hashtag{"text": "golang", "indices": []int{0, 6}}
		hashtag2 := generated.Hashtag{"text": "golang", "indices": []int{10, 16}} // same text, different indices
		hashtag3 := generated.Hashtag{"text": "python", "indices": []int{0, 6}}   // different text

		if !isHashtagEqual(hashtag1, hashtag2) {
			t.Error("Expected hashtag1 and hashtag2 to be equal (same text)")
		}
		if isHashtagEqual(hashtag1, hashtag3) {
			t.Error("Expected hashtag1 and hashtag3 to be different (different text)")
		}
	})

	// Test Symbol equality
	t.Run("SymbolEquality", func(t *testing.T) {
		symbol1 := generated.Symbol{"text": "AAPL", "indices": []int{0, 4}}
		symbol2 := generated.Symbol{"text": "AAPL", "indices": []int{10, 14}} // same text, different indices
		symbol3 := generated.Symbol{"text": "GOOGL", "indices": []int{0, 5}}  // different text

		if !isSymbolEqual(symbol1, symbol2) {
			t.Error("Expected symbol1 and symbol2 to be equal (same text)")
		}
		if isSymbolEqual(symbol1, symbol3) {
			t.Error("Expected symbol1 and symbol3 to be different (different text)")
		}
	})

	// Test UserMention equality
	t.Run("UserMentionEquality", func(t *testing.T) {
		mention1 := generated.UserMention{"screen_name": "elonmusk", "indices": []int{0, 9}}
		mention2 := generated.UserMention{"screen_name": "elonmusk", "indices": []int{10, 19}} // same screen_name, different indices
		mention3 := generated.UserMention{"screen_name": "tim_cook", "indices": []int{0, 9}}   // different screen_name

		if !isUserMentionEqual(mention1, mention2) {
			t.Error("Expected mention1 and mention2 to be equal (same screen_name)")
		}
		if isUserMentionEqual(mention1, mention3) {
			t.Error("Expected mention1 and mention3 to be different (different screen_name)")
		}
	})

	// Test URL equality
	t.Run("URLEquality", func(t *testing.T) {
		url1 := generated.Url{Url: "https://example.com", DisplayUrl: "example.com"}
		url2 := generated.Url{Url: "https://example.com", DisplayUrl: "different.com"} // same URL, different display
		url3 := generated.Url{Url: "https://different.com", DisplayUrl: "different.com"}

		if !isUrlEqual(url1, url2) {
			t.Error("Expected url1 and url2 to be equal (same URL)")
		}
		if isUrlEqual(url1, url3) {
			t.Error("Expected url1 and url3 to be different (different URL)")
		}
	})

	// Test Timestamp equality
	t.Run("TimestampEquality", func(t *testing.T) {
		timestamp1 := generated.Timestamp{Seconds: 123, Text: "test", Indices: []int{0, 4}}
		timestamp2 := generated.Timestamp{Seconds: 123, Text: "test", Indices: []int{0, 4}}
		timestamp3 := generated.Timestamp{Seconds: 456, Text: "different", Indices: []int{0, 9}}

		if !isTimestampEqual(timestamp1, timestamp2) {
			t.Error("Expected timestamp1 and timestamp2 to be equal")
		}
		if isTimestampEqual(timestamp1, timestamp3) {
			t.Error("Expected timestamp1 and timestamp3 to be different")
		}
	})
}
