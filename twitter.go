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

func GetTwitterApi() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(os.Getenv("TWITTER_CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("TWITTER_CONSUMER_SECRET"))
	api := anaconda.NewTwitterApi(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"))
	return api
}

func newTwitter() *Twitter {
	for _, v := range []string{"HASHTAG", "TWITTER_CONSUMER_KEY", "TWITTER_CONSUMER_SECRET", "TWITTER_ACCESS_TOKEN", "TWITTER_ACCESS_TOKEN_SECRET"} {
		fmt.Printf("%v=%v\n", v, os.Getenv(v))
	}
	hash := os.Getenv("HASHTAG")
	if hash == "" {
		fmt.Println("No Environment Variable (HASHTAG,TWITTER_CONSUMER_KEY,TWITTER_CONSUMER_SECRET,TWITTER_ACCESS_TOKEN,TWITTER_ACCESS_TOKEN_SECRET)")
		os.Exit(1)
	}

	savefile := filepath.Join("./volume", hash+".json") //TODO:
	t := Twitter{
		api:      GetTwitterApi(),
		hashtag:  hash,
		savefile: savefile,
		//startedAt: time.Now(),
	}
	if err := t.readSavedata(); err != nil {
		panic(err)
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

func (t *Twitter) search() ([]anaconda.Tweet, error) {
	v := url.Values{}
	if t.savedata.LatestId != 0 {
		v.Set("since_id", fmt.Sprintf("%d", t.savedata.LatestId+1))
	}
	v.Set("include_entities", "true")
	v.Set("exclude", "retweets")
	//fmt.Printf("URL PARAM:%v\n", v)
	searchResult, err := t.api.GetSearch(t.hashtag, v)
	return searchResult.Statuses, err
}

func (t *Twitter) getTweet(id int64) (anaconda.Tweet, error) {
	v := url.Values{}
	//v.Set("include_entities", "true")
	return t.api.GetTweet(id, v)
}

func (t *Twitter) post(s string, v url.Values) error {
	if s == "" {
		s = "nil"
	}
	_, err := t.api.PostTweet(s, v)
	return err
}

func (t *Twitter) retweet(id int64, trimUser bool) error {
	_, err := t.api.Retweet(id, trimUser)
	return err
}

func (t *Twitter) quotedTweet(result string, tweet *anaconda.Tweet) error {
	//status := fmt.Sprintf("@%s\n%s%s\nhttps://twitter.com/%s/status/%d", tweet.User.ScreenName, result, t.hashtag, tweet.User.ScreenName, tweet.Id)
	header := fmt.Sprintf("@%s\n", tweet.User.ScreenName)
	footer := fmt.Sprintf("%s\nhttps://twitter.com/%s/status/%d", t.hashtag, tweet.User.ScreenName, tweet.Id)
	if len(header+result+footer) > TWEET_MAX_CHARS {
		result = result[0:TWEET_MAX_CHARS-1-len(header+footer)] + "\n"
	}
	status := header + result + footer
	fmt.Printf("len(status)=%d\n", len(status))
	v := url.Values{}
	//v.Add("quoted_status_id", fmt.Sprintf("%d", tweet.Id))
	//v.Add("quoted_status_id_str", tweet.IdStr)
	v.Add("in_reply_to_user_id", fmt.Sprintf("%d", tweet.User.Id))
	v.Add("in_reply_to_user_id_str", tweet.User.IdStr)
	v.Add("in_reply_to_status_id", fmt.Sprintf("%d", tweet.Id))
	v.Add("in_reply_to_status_id_str", tweet.User.IdStr)

	fmt.Println("=============================================")
	fmt.Printf("%s\n", status)
	fmt.Println("=============================================")

	return t.post(status, v)
}
