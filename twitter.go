package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

type Twitter struct {
	api        *anaconda.TwitterApi
	createdAt  time.Time
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
		api: GetTwitterApi(),
		createdAttime.Now(),
	}
}

//const HASHTAG = "#sboxbot"
const HASHTAG = "#smallbox"

func (t *Twitter) search() {
	fmt.Printf("twitter.search %s\n", time.Now())
	v := url.Values{}
	if t.searchedId != 0 {
		v.Add("since_id", fmt.Sprint(t.searchedId))
	}
	searchResult, err := t.api.GetSearch(HASHTAG, v)
	if err != nil {
		panic(err)
	}
	for i, tweet := range searchResult.Statuses {
		fmt.Printf("key:%d\tid:%d\tCreatedAt:%s\tUser:%s\n", i, tweet.Id, tweet.CreatedAt, nil)
		//fmt.Println(tweet.Text)
		//fmt.Println(tweet)
		if t.searchedId < tweet.Id {
			t.searchedId = tweet.Id
		}
	}
}

func (t *Twitter) post(s string, v url.Values) {
	tweet, err := api.PostTweet(s, v)
	if err != nil {
		panic(err)
	}
}
