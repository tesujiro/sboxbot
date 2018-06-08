package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

const CHECK_TIMER = 30 * time.Second
const CONTAINER_TIMER = 10 * time.Second
const TWEET_TIMER = 10 * time.Second

type sbox struct {
	twitter *Twitter
}

func new() *sbox {
	return &sbox{twitter: newTwitter("")}
}

func (s *sbox) execOnContainer(ctx context.Context, commands []string) (string, error) {
	d, err := newDockerContainer(ctx, "centos", []string{"/bin/bash"})
	if err != nil {
		return "", err
	}
	defer d.remove()

	if err := d.run(ctx); err != nil {
		return "", err
	}
	for _, cmd := range commands {
		if err := d.exec(cmd); err != nil {
			result, _ := d.exit()
			//result += fmt.Sprintf("%v", err)
			return result, err
		}
	}
	return d.exit()
}

func (s *sbox) getCommand(text string) (string, error) {
	addLineBreak := func(s string) string {
		if len(s) > 0 && s[len(s)-1] != '\n' {
			return s + "\n"
		} else {
			return s
		}
	}
	text = strings.Replace(text, s.twitter.hashtag, "", -1)
	text = addLineBreak(text)
	if strings.Replace(text, " \t\n", "", -1) == "" {
		return "", nil
	}
	return text, nil
}

func (s *sbox) upTree(tweet anaconda.Tweet) ([]string, error) {
	if tweet.QuotedStatusID != 0 {
		fmt.Printf("skip because of QuotedStatusID:%v\n", tweet.QuotedStatusID)
		return []string{}, nil
	}

	// extract command from TweetText
	command, err := s.getCommand(tweet.FullText)
	if err != nil {
		return []string{command}, err
	}

	if tweet.InReplyToStatusID == 0 {
		// Top tweet
		return []string{command}, nil
	} else {
		// Intermediate tweet
		reply, err := s.twitter.getTweet(tweet.InReplyToStatusID)
		if err != nil {
			return []string{}, err
		}
		commands, err := s.upTree(reply)
		return append(commands, command), err
	}
}

func (s *sbox) run(ctx context.Context) error {
	fmt.Printf("twitter.search now=%s\tlatestId=%v\n", time.Now(), s.twitter.savedata.LatestId)
	tweets, err := s.twitter.search(0)
	if err != nil {
		fmt.Printf("search error:%v\n", err)
	}
	for i, tweet := range tweets {
		fmt.Printf("key:%d\tid:%d\tCreatedAt:%s\tUser.ScreenName:%s\n", i, tweet.Id, tweet.CreatedAt, tweet.User.ScreenName)
		//Save LatestId
		if s.twitter.savedata.LatestId < tweet.Id {
			s.twitter.savedata.LatestId = tweet.Id
			fmt.Printf("s.twitter.savedata.LatestId =%d\n", s.twitter.savedata.LatestId)
			if err := s.twitter.writeSavedata(); err != nil {
				return err
			}
		}
		//tweet = s.twitter.getTweet(tweet.Id)
		fmt.Println("=============================================")
		fmt.Println(tweet.FullText)
		//fmt.Println("=============================================")
		//fmt.Println(tweet.Entities)
		fmt.Println("=============================================")
		fmt.Printf("%+v\n", tweet)
		fmt.Println("=============================================")
		fmt.Printf("InReplyToUserID:%v\n", tweet.InReplyToUserID)
		fmt.Printf("InReplyToStatusID:%v\n", tweet.InReplyToStatusID)
		fmt.Printf("QuotedStatusID:%v\n", tweet.QuotedStatusID)
		fmt.Println("=============================================")

		commands, err := s.upTree(tweet)
		if err != nil {
			return err
		}

		fmt.Printf("==>execute command\n")
		ctxWithTimeout, cancel := context.WithTimeout(ctx, CONTAINER_TIMER)
		defer cancel()
		result, err := s.execOnContainer(ctxWithTimeout, commands)
		if err != nil {
			result = fmt.Sprintf("%v\n%v\n", result, err)
		}
		if _, err := s.twitter.quotedTweet(result, &tweet); err != nil {
			return err
		}
		time.Sleep(TWEET_TIMER)

	}
	return nil
}

func run(ctx context.Context) error {

	s := new()
	tick := time.NewTicker(CHECK_TIMER).C

	if err := s.run(ctx); err != nil {
		fmt.Printf("quoteTweet error:%v\n", err)
	}
mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			if err := s.run(ctx); err != nil {
				fmt.Printf("quoteTweet error:%v\n", err)
			}
		}
	}
	return nil
}

func _main() int {
	if envvar := os.Getenv("GOMAXPROCS"); envvar == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := run(ctx); err != nil {
		return 1
	}

	return 0
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "Error:\n%s", err)
			os.Exit(1)
		}
	}()
	os.Exit(_main())
}
