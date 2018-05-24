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

const MAX_TEST_TIME = 1800 * time.Second
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
	//go run(context.Background())
	tw := newTwitter("TEST_")
	//latestId := tw.savedata.LatestId
	latestId := int64(0)

	cases := []status{
		{command: fmt.Sprintf("echo hello world! %v\n%v\n", now, tw.hashtag), expected: "hello world!"},
		{command: fmt.Sprintf("echo こんにちは、世界！%v\n%v\n", now, tw.hashtag), expected: "こんにちは、世界！"},
		//{command: fmt.Sprintf("echo no line break %v %v", now, tw.hashtag), expected: fmt.Sprintf("no line break")},
		//{command: fmt.Sprintf("echo with no command line %v\n \t\n%v\n", now, tw.hashtag), expected: fmt.Sprintf("with no command line")},
		//{command: fmt.Sprintf("echo hello long world! %v\nfor i in `seq 200`\ndo\n  echo i=$i\ndone\n%v\n", now, tw.hashtag), expected: fmt.Sprintf("i=20")},
		//{
		//command: fmt.Sprintf("echo hello 1! %v\n%v\n", now, tw.hashtag), expected: "hello 1!",
		//replies: []status{status{command: fmt.Sprintf("echo hello 2! %v\n%v\n", now, tw.hashtag), expected: "hello 2!"}},
		//},

		//{commands: fmt.Sprintf("sleep\n"), expected: fmt.Sprintf("hello world!\n")},
		//{commands: fmt.Sprintf("set\n")},
		//{commands: fmt.Sprintf("while : \ndo\n:\ndone\n"), expected: fmt.Sprintf("exit error: context deadline exceeded")},
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
			for i, chk := range checkQue {
				fmt.Printf("Check Tweet Id:%v QuotedStatusID:%v\n", chk.tweet.Id, chk.tweet.QuotedStatusID)
				tweets, err := tw.searchQuotedTweet(chk.tweet)
				if err != nil {
					panic(err)
				}
				for _, quote := range tweets {
					fmt.Printf("Found Quote for Id:%v FullText:%v\n", quote.Id, quote.FullText)
					r := regexp.MustCompile(chk.expectedFullText_regex)
					if !r.MatchString(quote.FullText) {
						t.Errorf("tweet text not match:%v\n%v\n", chk.expectedFullText_regex, quote.FullText)
						continue
					}
					checkQue = append(checkQue[:i], checkQue[i+1:]...)
					bot_tweet_list = append(bot_tweet_list, quote.Id)
					break
				}
				if len(checkQue) == 0 {
					break loop
				}
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
