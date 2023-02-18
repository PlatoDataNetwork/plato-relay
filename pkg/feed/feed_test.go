package feed

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

const samplePubKey = "1870bcd5f6081ef7ea4b17204ffa4e92de51670142be0c8140e0635b355ca85f"
const sampleUrlForPublicKey = "https://nitter.moomoo.me/Bitcoin/rss"
const samplePrivateKeyForPubKey = "27660ab89e69f59bb8d9f0bd60da4a8515cdd3e2ca4f91d72a242b086d6aaaa7"
const testSecret = "test"

const sampleInvalidUrl = "https:// nostr.example/"
const sampleInvalidUrlContentType = "https://accounts.google.com/.well-known/openid-configuration"
const sampleRedirectingUrl = "https://httpstat.us/301"
const sampleValidDirectFeedUrl = "https://mastodon.social/@Gargron.rss"
const sampleValidIndirectFeedUrl = "https://www.rssboard.org/"
const sampleValidIndirectFeedUrlExpected = "http://feeds.rssboard.org/rssboard"
const sampleValidWithoutFeedUrl = "https://go.dev/"
const sampleValidWithRelativeFeedUrl = "https://golangweekly.com/"
const sampleValidWithRelativeFeedUrlExpected = "https://golangweekly.com/rss"

var actualTime = time.Now()
var sampleNitterFeed = gofeed.Feed{
	Title:           "Coldplay / @coldplay",
	Description:     "Twitter feed for: @coldplay. Generated by nitter.moomoo.me",
	Link:            "http://nitter.moomoo.me/coldplay",
	FeedLink:        "https://nitter.moomoo.me/coldplay/rss",
	Links:           []string{"http://nitter.moomoo.me/coldplay"},
	PublishedParsed: &actualTime,
	Language:        "en-us",
	Image: &gofeed.Image{
		URL:   "http://nitter.moomoo.me/pic/pbs.twimg.com%2Fprofile_images%2F1417506973877211138%2FYIm7dOQH_400x400.jpg",
		Title: "Coldplay / @coldplay",
	},
}

var sampleStackerNewsFeed = gofeed.Feed{
	Title:           "Stacker News",
	Description:     "Like Hacker News, but we pay you Bitcoin.",
	Link:            "https://stacker.news",
	FeedLink:        "https://stacker.news/rss",
	Links:           []string{"https://blog.cryptographyengineering.com/2014/11/zero-knowledge-proofs-illustrated-primer.html"},
	PublishedParsed: &actualTime,
	Language:        "en",
}

var sampleNitterFeedRTItem = gofeed.Item{
	Title:           "RT by @coldplay: TOMORROW",
	Description:     "Sample description",
	Content:         "Sample content",
	Link:            "http://nitter.moomoo.me/coldplay/status/1622148481740685312#m",
	UpdatedParsed:   &actualTime,
	PublishedParsed: &actualTime,
	GUID:            "http://nitter.moomoo.me/coldplay/status/1622148481740685312#m",
	DublinCoreExt: &ext.DublinCoreExtension{
		Creator: []string{"@nbcsnl"},
	},
}

var sampleNitterFeedResponseItem = gofeed.Item{
	Title:           "R to @coldplay: Sample",
	Description:     "Sample description",
	Content:         "Sample content",
	Link:            "http://nitter.moomoo.me/elonmusk/status/1621544996167122944#m",
	UpdatedParsed:   &actualTime,
	PublishedParsed: &actualTime,
	GUID:            "http://nitter.moomoo.me/elonmusk/status/1621544996167122944#m",
	DublinCoreExt: &ext.DublinCoreExtension{
		Creator: []string{"@elonmusk"},
	},
}

var sampleDefaultFeedItem = gofeed.Item{
	Title:           "Golang Weekly",
	Description:     "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus nec condimentum orci. Vestibulum at nunc porta, placerat ex sit amet, consectetur augue. Donec cursus ipsum sed venenatis maximus. Nunc tincidunt dui nec congue lacinia. In mollis magna eu nisi viverra luctus. Ut ultrices eros gravida, lacinia nibh vitae, tristique massa. Sed eu scelerisque erat. Sed eget tortor et turpis feugiat interdum. Nulla sit amet nibh vel massa bibendum congue. Quisque sed tempor velit. Interdum et malesuada fames ac ante ipsum primis in faucibus. Curabitur suscipit mollis fringilla. Integer quis sodales tortor, at hendrerit lacus. Cras posuere maximus nisi. Mauris eget.",
	Content:         "Sample content",
	Link:            "https://golangweekly.com/issues/446",
	UpdatedParsed:   &actualTime,
	PublishedParsed: &actualTime,
	GUID:            "https://golangweekly.com/issues/446",
}

var sampleDefaultFeedItemWithComments = gofeed.Item{
	Title:           "Golang Weekly",
	Description:     "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Phasellus nec condimentum orci. Vestibulum at nunc porta, placerat ex sit amet, consectetur augue. Donec cursus ipsum sed venenatis maximus. Nunc tincidunt dui nec congue lacinia. In mollis magna eu nisi viverra luctus. Ut ultrices eros gravida, lacinia nibh vitae, tristique massa. Sed eu scelerisque erat. Sed eget tortor et turpis feugiat interdum. Nulla sit amet nibh vel massa bibendum congue. Quisque sed tempor velit. Interdum et malesuada fames ac ante ipsum primis in faucibus. Curabitur suscipit mollis fringilla. Integer quis sodales tortor, at hendrerit lacus. Cras posuere maximus nisi. Mauris eget.",
	Content:         "Sample content",
	Link:            "https://golangweekly.com/issues/446",
	UpdatedParsed:   &actualTime,
	PublishedParsed: &actualTime,
	GUID:            "https://golangweekly.com/issues/446",
	Custom: map[string]string{
		"comments": "https://golangweekly.com/issues/446",
	},
}

var sampleDefaultFeedItemExpectedContent = fmt.Sprintf("**%s**\n\n%s", sampleDefaultFeedItem.Title, sampleDefaultFeedItem.Description)
var sampleDefaultFeedItemExpectedContentSubstring = sampleDefaultFeedItemExpectedContent[0:249]

var sampleStackerNewsFeedItem = gofeed.Item{
	Title:           "Zero Knowledge Proofs: An illustrated primer",
	Description:     "<a href=\"https://stacker.news/items/131533\">Comments</a>",
	Content:         "Sample content",
	Link:            "https://blog.cryptographyengineering.com/2014/11/zero-knowledge-proofs-illustrated-primer.html",
	UpdatedParsed:   &actualTime,
	PublishedParsed: &actualTime,
	GUID:            "https://stacker.news/items/131533",
	Custom: map[string]string{
		"comments": "https://stacker.news/items/131533",
	},
}

var sampleDefaultFeed = gofeed.Feed{
	Title:           "Golang Weekly",
	Description:     "A weekly newsletter about the Go programming language",
	Link:            "https://golangweekly.com/rss",
	FeedLink:        "https://golangweekly.com/rss",
	Links:           []string{"https://golangweekly.com/issues/446"},
	PublishedParsed: &actualTime,
	Language:        "en-us",
	Image:           nil,
}

func TestGetFeedURLWithInvalidURLReturnsEmptyString(t *testing.T) {
	feed := GetFeedURL(sampleInvalidUrl)
	assert.Empty(t, feed)
}

func TestGetFeedURLWithInvalidContentTypeReturnsEmptyString(t *testing.T) {
	feed := GetFeedURL(sampleInvalidUrlContentType)
	assert.Empty(t, feed)
}

func TestGetFeedURLWithRedirectingURLReturnsEmptyString(t *testing.T) {
	feed := GetFeedURL(sampleRedirectingUrl)
	assert.Empty(t, feed)
}

func TestGetFeedURLWithValidUrlOfValidTypesReturnsSameUrl(t *testing.T) {
	feed := GetFeedURL(sampleValidDirectFeedUrl)
	assert.Equal(t, sampleValidDirectFeedUrl, feed)
}

func TestGetFeedURLWithValidUrlOfHtmlTypeWithFeedReturnsFoundFeed(t *testing.T) {
	feed := GetFeedURL(sampleValidIndirectFeedUrl)
	assert.Equal(t, sampleValidIndirectFeedUrlExpected, feed)
}

func TestGetFeedURLWithValidUrlOfHtmlTypeWithRelativeFeedReturnsFoundFeed(t *testing.T) {
	feed := GetFeedURL(sampleValidWithRelativeFeedUrl)
	assert.Equal(t, sampleValidWithRelativeFeedUrlExpected, feed)
}

func TestGetFeedURLWithValidUrlOfHtmlTypeWithoutFeedReturnsEmpty(t *testing.T) {
	feed := GetFeedURL(sampleValidWithoutFeedUrl)
	assert.Empty(t, feed)
}

func TestParseFeedWithValidUrlReturnsParsedFeed(t *testing.T) {
	feed, err := ParseFeed(sampleValidWithRelativeFeedUrlExpected)
	assert.NotNil(t, feed)
	assert.NoError(t, err)
}

func TestParseFeedWithValidUrlWithoutFeedReturnsError(t *testing.T) {
	feed, err := ParseFeed(sampleValidWithoutFeedUrl)
	assert.Nil(t, feed)
	assert.Error(t, err)
}

func TestParseFeedWithCachedUrlReturnsCachedParsedFeed(t *testing.T) {
	_, _ = ParseFeed(sampleValidWithRelativeFeedUrlExpected)
	feed, err := ParseFeed(sampleValidWithRelativeFeedUrlExpected)
	assert.NotNil(t, feed)
	assert.NoError(t, err)
}

func TestEntryFeedToSetMetadata(t *testing.T) {
	testCases := []struct {
		pubKey                   string
		feed                     *gofeed.Feed
		originalUrl              string
		enableAutoRegistration   bool
		defaultProfilePictureUrl string
	}{
		{
			pubKey:                   samplePubKey,
			feed:                     &sampleNitterFeed,
			originalUrl:              sampleNitterFeed.FeedLink,
			enableAutoRegistration:   true,
			defaultProfilePictureUrl: "https://image.example",
		},
		{
			pubKey:                   samplePubKey,
			feed:                     &sampleDefaultFeed,
			originalUrl:              sampleDefaultFeed.FeedLink,
			enableAutoRegistration:   true,
			defaultProfilePictureUrl: "https://image.example",
		},
	}
	for _, tc := range testCases {
		metadata := EntryFeedToSetMetadata(tc.pubKey, tc.feed, tc.originalUrl, tc.enableAutoRegistration, tc.defaultProfilePictureUrl)
		assert.NotEmpty(t, metadata)
		assert.Equal(t, samplePubKey, metadata.PubKey)
		assert.Equal(t, 0, metadata.Kind)
		assert.Empty(t, metadata.Sig)
	}
}

func TestPrivateKeyFromFeed(t *testing.T) {
	sk := PrivateKeyFromFeed(sampleUrlForPublicKey, testSecret)
	assert.Equal(t, samplePrivateKeyForPubKey, sk)
}

func TestItemToTextNote(t *testing.T) {
	testCases := []struct {
		pubKey           string
		item             *gofeed.Item
		feed             *gofeed.Feed
		defaultCreatedAt time.Time
		originalUrl      string
		expectedContent  string
	}{
		{
			pubKey:           samplePubKey,
			item:             &sampleNitterFeedRTItem,
			feed:             &sampleNitterFeed,
			defaultCreatedAt: actualTime,
			originalUrl:      sampleNitterFeed.FeedLink,
			expectedContent:  fmt.Sprintf("**RT %s:**\n\n%s\n\n%s", sampleNitterFeedRTItem.DublinCoreExt.Creator[0], sampleNitterFeedRTItem.Description, strings.ReplaceAll(sampleNitterFeedRTItem.Link, "http://", "https://")),
		},
		{
			pubKey:           samplePubKey,
			item:             &sampleNitterFeedResponseItem,
			feed:             &sampleNitterFeed,
			defaultCreatedAt: actualTime,
			originalUrl:      sampleNitterFeed.FeedLink,
			expectedContent:  fmt.Sprintf("**Response to %s:**\n\n%s\n\n%s", "@coldplay", sampleNitterFeedResponseItem.Description, strings.ReplaceAll(sampleNitterFeedResponseItem.Link, "http://", "https://")),
		},
		{
			pubKey:           samplePubKey,
			item:             &sampleDefaultFeedItem,
			feed:             &sampleDefaultFeed,
			defaultCreatedAt: actualTime,
			originalUrl:      sampleDefaultFeed.FeedLink,
			expectedContent:  sampleDefaultFeedItemExpectedContentSubstring + "…" + "\n\n" + sampleDefaultFeedItem.Link,
		},
		{
			pubKey:           samplePubKey,
			item:             &sampleDefaultFeedItemWithComments,
			feed:             &sampleDefaultFeed,
			defaultCreatedAt: actualTime,
			originalUrl:      sampleDefaultFeed.FeedLink,
			expectedContent:  sampleDefaultFeedItemExpectedContentSubstring + "…\nComments: " + sampleDefaultFeedItemWithComments.Custom["comments"] + "\n\n" + sampleDefaultFeedItem.Link,
		},
		{
			pubKey:           samplePubKey,
			item:             &sampleStackerNewsFeedItem,
			feed:             &sampleStackerNewsFeed,
			defaultCreatedAt: actualTime,
			originalUrl:      sampleStackerNewsFeed.FeedLink,
			expectedContent:  fmt.Sprintf("**%s**\n\nComments: %s\n\n%s", sampleStackerNewsFeedItem.Title, sampleStackerNewsFeedItem.GUID, sampleStackerNewsFeedItem.Link),
		},
	}
	for _, tc := range testCases {
		event := ItemToTextNote(tc.pubKey, tc.item, tc.feed, tc.defaultCreatedAt, tc.originalUrl)
		assert.NotEmpty(t, event)
		assert.Equal(t, tc.pubKey, event.PubKey)
		assert.Equal(t, tc.defaultCreatedAt, event.CreatedAt)
		assert.Equal(t, 1, event.Kind)
		assert.Equal(t, tc.expectedContent, event.Content)
		assert.Empty(t, event.Sig)
		assert.Empty(t, event.Tags)
	}
}

func TestDeleteExistingInvalidFeed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when closing a stub database connection", err)
		}
	}(db)

	mock.ExpectExec("DELETE FROM feeds").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectClose()
	DeleteInvalidFeed(sampleUrlForPublicKey, db)
}

func TestDeleteNonExistingInvalidFeed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when closing a stub database connection", err)
		}
	}(db)

	mock.ExpectExec("DELETE FROM feeds").WillReturnError(errors.New(""))
	mock.ExpectClose()
	DeleteInvalidFeed(sampleUrlForPublicKey, db)
}
