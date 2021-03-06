package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ChimeraCoder/anaconda"
)

const TWEET_MAX_CHARS = 280

type savedata struct {
	LatestId int64 `json:"latest_id"`
}

type Twitter struct {
	api *anaconda.TwitterApi
	//startedAt time.Time
	hashtag  string
	savefile string // json tmp file path
	savedata *savedata
}

func getTwitterApi(prefix string) *anaconda.TwitterApi {
	for _, v := range []string{"TWITTER_CONSUMER_KEY", "TWITTER_CONSUMER_SECRET", "TWITTER_ACCESS_TOKEN", "TWITTER_ACCESS_TOKEN_SECRET"} {
		env := prefix + v
		if os.Getenv(env) == "" {
			fmt.Printf("No Environment Variable (%v) set.\n", env)
			os.Exit(1)
		}
		fmt.Printf("%v=%v\n", env, os.Getenv(env))
	}
	anaconda.SetConsumerKey(os.Getenv(prefix + "TWITTER_CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv(prefix + "TWITTER_CONSUMER_SECRET"))
	api := anaconda.NewTwitterApi(os.Getenv(prefix+"TWITTER_ACCESS_TOKEN"), os.Getenv(prefix+"TWITTER_ACCESS_TOKEN_SECRET"))
	return api
}

func newTwitter(prefix string) *Twitter {
	v := "HASHTAG"
	hash := os.Getenv(v)
	if hash == "" {
		fmt.Printf("No Environment Variable (%v) set.\n", v)
		os.Exit(1)
	}
	fmt.Printf("%v=%v\n", v, os.Getenv(v))

	savefile := filepath.Join("./volume", hash+".json") //TODO:
	t := Twitter{
		api:      getTwitterApi(prefix),
		hashtag:  hash,
		savefile: savefile,
		//startedAt: time.Now(),
	}
	if err := t.readSavedata(); err != nil {
		fmt.Println("No Save File!")
	}
	return &t
}

func (t *Twitter) readSavedata() error {
	raw, _ := ioutil.ReadFile(t.savefile)

	var sd savedata
	if err := json.Unmarshal(raw, &sd); err != nil {
		t.savedata = &sd
		return err
	}
	t.savedata = &sd

	return nil
}

func (t *Twitter) writeSavedata() error {
	var js []byte
	js, err := json.Marshal(t.savedata)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(t.savefile, js, 0644); err != nil {
		return err
	}

	return nil
}

func (t *Twitter) search(latestId int64) ([]anaconda.Tweet, error) {
	v := url.Values{}
	if latestId != 0 {
		v.Set("since_id", fmt.Sprintf("%d", latestId+1))
	} else if t.savedata.LatestId != 0 {
		v.Set("since_id", fmt.Sprintf("%d", t.savedata.LatestId+1))
	}
	v.Set("include_entities", "true")
	v.Set("exclude", "retweets")
	//fmt.Printf("URL PARAM:%v\n", v)
	searchResult, err := t.api.GetSearch(t.hashtag, v)
	return searchResult.Statuses, err
}

func (t *Twitter) searchQuotedTweets(tweet anaconda.Tweet) ([]anaconda.Tweet, error) {
	v := url.Values{}
	//v.Set("since_id", fmt.Sprintf("%d", tweet.Id+1))
	//url := fmt.Sprintf("https://twitter.com/%s/status/%d", tweet.User.ScreenName, tweet.Id)
	//fmt.Printf("url=%v\n", url)
	//v.Set("q", url)
	v.Set("include_entities", "true")
	v.Set("result_type", "recent")
	//v.Set("exclude", "retweets")
	//fmt.Printf("URL PARAM:%v\n", v)
	searchResult, err := t.api.GetSearch(t.hashtag, v)
	//searchResult, err := t.api.GetSearch(url, v)
	return searchResult.Statuses, err
}

func (t *Twitter) getMentionsTimeline(latestId int64) ([]anaconda.Tweet, error) {
	v := url.Values{}
	if latestId != 0 {
		v.Set("since_id", fmt.Sprintf("%d", latestId+1))
	}
	v.Set("include_entities", "true")
	//return t.api.GetRetweetsOfMe(v)
	return t.api.GetMentionsTimeline(v)
}

func (t *Twitter) getTweet(id int64) (anaconda.Tweet, error) {
	v := url.Values{}
	//v.Set("include_entities", "true")
	return t.api.GetTweet(id, v)
}

func (t *Twitter) post(s string, v url.Values) (anaconda.Tweet, error) {
	if s == "" {
		s = "nil"
	}
	return t.api.PostTweet(s, v)
}

func (t *Twitter) retweet(id int64, trimUser bool) error {
	_, err := t.api.Retweet(id, trimUser)
	return err
}

func (t *Twitter) deleteTweet(id int64) (anaconda.Tweet, error) {
	return t.api.DeleteTweet(id, true)
}

func (t *Twitter) quotedTweet(result string, tweet *anaconda.Tweet) (anaconda.Tweet, error) {
	//status := fmt.Sprintf("@%s\n%s%s\nhttps://twitter.com/%s/status/%d", tweet.User.ScreenName, result, t.hashtag, tweet.User.ScreenName, tweet.Id)
	header := fmt.Sprintf("@%s\n", tweet.User.ScreenName)
	//footer := fmt.Sprintf("%s\nhttps://twitter.com/%s/status/%d", t.hashtag, tweet.User.ScreenName, tweet.Id)
	footer := fmt.Sprintf("https://twitter.com/%s/status/%d", tweet.User.ScreenName, tweet.Id)
	if len(header+result+footer) > TWEET_MAX_CHARS {
		result = result[0:TWEET_MAX_CHARS-1-len(header+footer)] + "\n"
	}
	status := header + result + footer
	fmt.Printf("len(status)=%d\n", len(status))
	v := url.Values{}
	v.Add("quoted_status_id", fmt.Sprintf("%d", tweet.Id))
	v.Add("quoted_status_id_str", tweet.IdStr)
	//v.Add("in_reply_to_user_id", fmt.Sprintf("%d", tweet.User.Id))
	//v.Add("in_reply_to_user_id_str", tweet.User.IdStr)
	//v.Add("in_reply_to_status_id", fmt.Sprintf("%d", tweet.Id))
	//v.Add("in_reply_to_status_id_str", tweet.User.IdStr)

	fmt.Println("=============================================")
	fmt.Printf("%s\n", status)
	fmt.Println("=============================================")

	return t.post(status, v)
}
