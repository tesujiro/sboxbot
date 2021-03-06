package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"testing"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

const MAX_TEST_TIME = 600 * time.Second
const TEST_LOOP_TIMER = 10 * time.Second

type check struct {
	tweet                  anaconda.Tweet
	expectedFullText_regex string
}

type status struct {
	command  string
	expected string
	replies  []status
}

func TestRun(t *testing.T) {
	now := time.Now()
	tw := newTwitter("TEST_")
	latestId := int64(0)

	cases := []status{
		{
			command:  fmt.Sprintf("println(\"hello\")\n#%v\n%v\n", now, tw.hashtag),
			expected: "hello",
		},
		{
			command:  fmt.Sprintf("aaa=100\nprintln(aaa*123)\n#%v\n%v\n", now, tw.hashtag),
			expected: "12300",
		},
		{
			command:  fmt.Sprintf("%v\nXxxx\n#%v\n", tw.hashtag, now),
			expected: "ERROR",
		},
		{
			command: fmt.Sprintf("%v\nvar fmt=import(\"fmt\")\n#%v\n", tw.hashtag, now),
			replies: []status{status{command: fmt.Sprintf("%v\nfmt.Println(\"abc\")\n#%v\n", tw.hashtag, now), expected: "abc"}},
		},
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, MAX_TEST_TIME)
	defer cancel()

	checkQue := []check{}
	my_tweet_list := []int64{}
	bot_tweet_list := []int64{}

	var walk func([]status, int64, int64)
	walk = func(cases []status, replyUserId, replyStatusId int64) {
		for _, c := range cases {
			v := url.Values{}
			v.Add("in_reply_to_user_id", fmt.Sprintf("%d", replyUserId))
			v.Add("in_reply_to_user_id_str", fmt.Sprintf("%d", replyUserId))
			v.Add("in_reply_to_status_id", fmt.Sprintf("%d", replyStatusId))
			v.Add("in_reply_to_status_id_str", fmt.Sprintf("%d", replyStatusId))
			tweet, err := tw.post(c.command, v)
			if err != nil {
				panic(err)
			}
			checkQue = append(checkQue, check{tweet: tweet, expectedFullText_regex: c.expected})
			my_tweet_list = append(my_tweet_list, tweet.Id)
			fmt.Printf("add tweet(ID:%v) to checkQue\n", tweet.Id)
			//time.Sleep(TEST_LOOP_TIMER)
			walk(c.replies, tweet.User.Id, tweet.Id)
		}
	}
	// Post Test Tweets
	walk(cases, int64(0), int64(0))
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer func() {
		for _, id := range my_tweet_list {
			fmt.Printf("delete my tweet ID: %v\n", id)
			if _, err := tw.deleteTweet(id); err != nil {
				fmt.Printf("delete error : %v\n", err)
			}
		}
	}()

	// Check Tweets
	tick := time.NewTicker(TEST_LOOP_TIMER).C
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-signal_chan:
			break loop
		case <-tick:
			fmt.Printf("[test]twitter.search now=%s\tlatestId=%v\n", time.Now(), latestId)
			tweets, err := tw.getMentionsTimeline(latestId)
			if err != nil {
				panic(err)
			}
			for _, tweet := range tweets {
				fmt.Printf("Check Tweet Id:%v QuotedStatusID:%v\n", tweet.Id, tweet.QuotedStatusID)
				for i, chk := range checkQue {
					if tweet.QuotedStatusID == chk.tweet.Id {
						fmt.Printf("Found Quote for Id:%v FullText:[%v]\n", chk.tweet.Id, tweet.FullText)
						r := regexp.MustCompile(chk.expectedFullText_regex)
						if !r.MatchString(tweet.FullText) {
							t.Errorf("tweet text not match:%v\n%v\n", chk.expectedFullText_regex, tweet.FullText)
						}
						checkQue = append(checkQue[:i], checkQue[i+1:]...)
						bot_tweet_list = append(bot_tweet_list, tweet.Id)
						break
					}
				}
			}
			/*
				found := 0
				for i, chk := range checkQue {
					fmt.Printf("Check Tweet Id:%v QuotedStatusID:%v\n", chk.tweet.Id, chk.tweet.QuotedStatusID)
					quotes, err := tw.searchQuotedTweets(chk.tweet)
					if err != nil {
						panic(err)
					}
					for j, quote := range quotes {
						if quote.QuotedStatusID != chk.tweet.Id {
							fmt.Printf("==> other tweet Id:%v QuotedStatusID:%v Text:%v\n", quote.Id, quote.QuotedStatusID, quote.FullText)
							continue
						}
						r := regexp.MustCompile(chk.expectedFullText_regex)
						if r.MatchString(quote.FullText) {
							fmt.Printf("==> match index:%v\n", j)
							checkQue = append(checkQue[:(i-found)], checkQue[i-found+1:]...)
							found++
							bot_tweet_list = append(bot_tweet_list, quote.Id)
							break
						}
					}
				}
			*/
			if len(checkQue) == 0 {
				break loop
			}
		}
	}
	for _, chk := range checkQue {
		t.Errorf("Error Not Found Quote for Id:%v FullText:%v\n", chk.tweet.Id, chk.tweet.FullText)
	}
	bot_tw := newTwitter("")
	for _, id := range bot_tweet_list {
		fmt.Printf("delete bot tweet ID: %v\n", id)
		if _, err := bot_tw.deleteTweet(id); err != nil {
			fmt.Printf("delete error : %v\n", err)
		}
	}
	return
}
