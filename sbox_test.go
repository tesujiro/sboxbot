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

const MAX_TEST_TIME = 60 * time.Second
const CHECK_TIMER = 10 * time.Second

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
		{
			command: fmt.Sprintf("echo hello 1! %v\n%v\n", now, tw.hashtag), expected: "hello 1!",
			replies: []status{status{command: fmt.Sprintf("echo hello 2! %v\n%v\n", now, tw.hashtag), expected: "hello 2!"}},
		},

		//{commands: fmt.Sprintf("sleep\n"), expected: fmt.Sprintf("hello world!\n")},
		//{commands: fmt.Sprintf("set\n")},
		//{commands: fmt.Sprintf("while : \ndo\n:\ndone\n"), expected: fmt.Sprintf("exit error: context deadline exceeded")},
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 600*time.Second)
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
	tick := time.NewTicker(time.Second * time.Duration(10)).C
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
						fmt.Printf("Found Quote for Id:%v FullText:%v\n", chk.tweet.Id, chk.tweet.FullText)
						r := regexp.MustCompile(chk.expectedFullText_regex)
						if !r.MatchString(tweet.FullText) {
							t.Errorf("tweet text not match:%v\n%v\n", chk.expectedFullText_regex, tweet.FullText)
						}

						//if c.expected != "" && actual != c.expected {
						//t.Errorf("got %v\nwant %v", actual, c.expected)
						//}
						//remove chk from checkQue
						checkQue = append(checkQue[:i], checkQue[i+1:]...)
						bot_tweet_list = append(bot_tweet_list, tweet.Id)
						break
					}
				}
				//if latestId < tweet.Id {
				//latestId = tweet.Id
				//}
			}
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
