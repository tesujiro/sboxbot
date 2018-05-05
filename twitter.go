package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

const TWEET_MAX_CHARS = 180

type Twitter struct {
	api        *anaconda.TwitterApi
	startedAt  time.Time
	searchedId int64
}

func GetTwitterApi() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(os.Getenv("TWITTER_CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("TWITTER_CONSUMER_SECRET"))
	api := anaconda.NewTwitterApi(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"))
	return api
}

func newTwitter() *Twitter {
	return &Twitter{
		api:       GetTwitterApi(),
		startedAt: time.Now(),
	}
}

func (t *Twitter) search(hashtag string) []anaconda.Tweet {
	v := url.Values{}
	if t.searchedId != 0 {
		v.Add("since_id", fmt.Sprint(t.searchedId))
	}
	searchResult, err := t.api.GetSearch(hashtag, v)
	if err != nil {
		panic(err)
	}
	if len(searchResult.Statuses) > 0 {
		t.searchedId = searchResult.Metadata.MaxId
	}
	return searchResult.Statuses
}

func (t *Twitter) post(s string, v url.Values) {
	if s == "" {
		s = "nil"
	}
	s = s[0 : TWEET_MAX_CHARS-1]
	_, err := t.api.PostTweet(s, v)
	if err != nil {
		panic(err)
	}
}

func (t *Twitter) retweet(id int64, trimUser bool) {
	_, err := t.api.Retweet(id, trimUser)
	if err != nil {
		panic(err)
	}
}
