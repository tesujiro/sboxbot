package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

func execOnContainer(ctx context.Context, cmd string) (string, error) {
	d := newDockerContainer()
	if err := d.run(ctx); err != nil {
		return "", err
	}
	if err := d.exec(cmd); err != nil {
		result, _ := d.exit()
		//result += fmt.Sprintf("%v", err)
		return result, err
	}
	return d.exit()
}

func quoteTweet(ctx context.Context, t *Twitter) error {
	fmt.Printf("twitter.search now=%s\tlatestId=%v\n", time.Now(), t.savedata.LatestId)
	tweets, err := t.search()
	if err != nil {
		fmt.Printf("search error:%v\n", err)
	}
	for i, tweet := range tweets {
		fmt.Printf("key:%d\tid:%d\tCreatedAt:%s\tUser.ScreenName:%s\n", i, tweet.Id, tweet.CreatedAt, tweet.User.ScreenName)
		//tweet = t.getTweet(tweet.Id)
		fmt.Println("=============================================")
		fmt.Println(tweet.FullText)
		//fmt.Println("=============================================")
		//fmt.Println(tweet.Entities)
		fmt.Println("=============================================")
		fmt.Printf("%v\n", tweet)
		fmt.Println("=============================================")
		fmt.Printf("==>exec\n")
		addLineBreak := func(s string) string {
			if len(s) > 0 && s[len(s)-1] != '\n' {
				return s + "\n"
			} else {
				return s
			}
		}
		cmd := strings.Replace(tweet.FullText, t.hashtag, "", -1)
		cmd = addLineBreak(cmd)
		if strings.Replace(cmd, " \t\n", "", -1) != "" {

			ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			result, err := execOnContainer(ctxWithTimeout, cmd)
			if err != nil {
				result = fmt.Sprintf("%v\n%v\n", result, err)
			}
			if err := t.quotedTweet(result, &tweet); err != nil {
				return err
			}
		}

		//Save LatestId
		if t.savedata.LatestId < tweet.Id {
			t.savedata.LatestId = tweet.Id
			fmt.Printf("t.savedata.LatestId =%d\n", t.savedata.LatestId)
			if err := t.writeSavedata(); err != nil {
				return err
			}
		}
	}
	return nil
}

func run(ctx context.Context) error {

	t := newTwitter()
	//tick := time.NewTicker(time.Second * time.Duration(60)).C
	tick := time.NewTicker(time.Second * time.Duration(10)).C

	if err := quoteTweet(ctx, t); err != nil {
		fmt.Printf("quoteTweet error:%v\n", err)
	}
mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			if err := quoteTweet(ctx, t); err != nil {
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
