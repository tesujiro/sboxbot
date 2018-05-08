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

const TWEET_MAX_CHARS = 180

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
	hash := os.Getenv("HASHTAG")
	if hash == "" {
		fmt.Println("No Environment Variable (HASHTAG,TWITTER_CONSUMER_KEY,TWITTER_CONSUMER_SECRET,TWITTER_ACCESS_TOKEN,TWITTER_ACCESS_TOKEN_SECRET)")
		os.Exit(1)
	}

	savefile := filepath.Join("/tmp", hash+".json") //TODO:
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
		//return err
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

func (t *Twitter) search() []anaconda.Tweet {
	v := url.Values{}
	if t.savedata.LatestId != 0 {
		//fmt.Printf("since_id=%d\n", t.savedata.LatestId)
		v.Add("since_id", fmt.Sprintf("%d", t.savedata.LatestId+1))
	}
	searchResult, err := t.api.GetSearch(t.hashtag, v)
	if err != nil {
		panic(err)
	}
	if len(searchResult.Statuses) > 0 {
		t.savedata.LatestId = searchResult.Metadata.MaxId
		fmt.Printf("t.savedata.LatestId =%d\n", t.savedata.LatestId)
		if err := t.writeSavedata(); err != nil {
			panic(err)
		}
	}
	return searchResult.Statuses
}

func (t *Twitter) getTweet(id int64) anaconda.Tweet {
	v := url.Values{}
	searchResult, err := t.api.GetTweet(id, v)
	if err != nil {
		fmt.Println("GetTweet Error")
		panic(err)
	}
	return searchResult
}

func (t *Twitter) post(s string, v url.Values) {
	if s == "" {
		s = "nil"
	}
	if len(s) > TWEET_MAX_CHARS {
		s = s[0 : TWEET_MAX_CHARS-1]
	}
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

func (t *Twitter) quotedTweet(result string, tweet *anaconda.Tweet) {
	// func quotedTweet(tweet,result)
	status := fmt.Sprintf("@%s\n%s%s\nhttps://twitter.com/%s/status/%d", tweet.User.ScreenName, result, t.hashtag, tweet.User.ScreenName, tweet.Id)
	v := url.Values{}
	v.Add("quoted_status_id", fmt.Sprintf("%d", tweet.Id))
	v.Add("quoted_status_id_str", tweet.IdStr)

	/*
		jsonBytes, err := json.Marshal(tweet)
		if err != nil {
			fmt.Println("JSON Marshal error:", err)
			panic(err)
		}
		v.Add("quoted_status", fmt.Sprintf("%s", jsonBytes))
	*/

	v.Add("in_reply_to_status_id", fmt.Sprintf("%d", tweet.Id))
	v.Add("in_reply_to_status_id_str", tweet.IdStr)
	//v.Add("in_reply_to_user_id", fmt.Sprintf("%d", tweet.User.Id))
	fmt.Println("=============================================")
	fmt.Printf("%s\n", status)
	fmt.Println("=============================================")

	t.post(status, v)
}
