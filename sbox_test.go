package main

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
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

func TestRun(t *testing.T) {
	now := time.Now()
	//go run(context.Background())
	tw := newTwitter("TEST_")
	//latestId := tw.savedata.LatestId
	latestId := int64(0)

	cases := []struct {
		commands string
		expected string
	}{
		{commands: fmt.Sprintf("echo hello world! %v\n%v\n", now, tw.hashtag), expected: "hello world!"},
		{commands: fmt.Sprintf("echo こんにちは、世界！%v\n%v\n", now, tw.hashtag), expected: "こんにちは、世界！"},
		{commands: fmt.Sprintf("echo no line break %v %v", now, tw.hashtag), expected: fmt.Sprintf("no line break")},
		{commands: fmt.Sprintf("echo with no command line %v\n \t\n%v\n", now, tw.hashtag), expected: fmt.Sprintf("with no command line")},
		{commands: fmt.Sprintf("echo hello long world! %v\nfor i in `seq 200`\ndo\n  echo i=$i\ndone\n%v\n", now, tw.hashtag), expected: fmt.Sprintf("i=20")},
		//{commands: fmt.Sprintf("sleep\n"), expected: fmt.Sprintf("hello world!\n")},
		//{commands: fmt.Sprintf("set\n")},
		//{commands: fmt.Sprintf("while : \ndo\n:\ndone\n"), expected: fmt.Sprintf("exit error: context deadline exceeded")},
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	checkQue := []check{}

	// Post Test Tweets
	for _, c := range cases {
		tweet, err := tw.post(c.commands, url.Values{})
		if err != nil {
			panic(err)
		}
		checkQue = append(checkQue, check{tweet: tweet, expectedFullText_regex: c.expected})
	}

	// Check Tweets
	tick := time.NewTicker(time.Second * time.Duration(10)).C
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-tick:
			fmt.Printf("[test]twitter.search now=%s\tlatestId=%v\n", time.Now(), tw.savedata.LatestId)
			tweets, err := tw.getMentionsTimeline(latestId)
			if err != nil {
				panic(err)
			}
			for _, tweet := range tweets {
				fmt.Printf("Check Tweet Id:%v QuotedStatusID:%v\n", tweet.Id, tweet.QuotedStatusID)
				for i, chk := range checkQue {
					if tweet.QuotedStatusID == chk.tweet.Id {
						fmt.Printf("Found Quote for :%v\n", chk.tweet)
						r := regexp.MustCompile(chk.expectedFullText_regex)
						if !r.MatchString(tweet.FullText) {
							t.Errorf("tweet text not match:%v\n%v\n", chk.expectedFullText_regex, tweet.FullText)
						}

						//if c.expected != "" && actual != c.expected {
						//t.Errorf("got %v\nwant %v", actual, c.expected)
						//}
						//remove chk from checkQue
						checkQue = append(checkQue[:i], checkQue[i+1:]...)
						break
					}
				}
				if latestId < tweet.Id {
					latestId = tweet.Id
				}
			}
			if len(checkQue) == 0 {
				break loop
			}
		}
	}
	for _, chk := range checkQue {
		t.Errorf("Error Not Found Quote for :%v\n", chk.tweet)
	}
	return
}
