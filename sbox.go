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
	for i, tweet := range t.search() {
		fmt.Printf("key:%d\tid:%d\tCreatedAt:%s\tUser.ScreenName:%s\n", i, tweet.Id, tweet.CreatedAt, tweet.User.ScreenName)
		//tweet = t.getTweet(tweet.Id)
		fmt.Println("=============================================")
		fmt.Println(tweet.Text)
		fmt.Println("=============================================")
		fmt.Println(tweet.Entities)
		fmt.Println("=============================================")
		fmt.Printf("%v\n", tweet)
		fmt.Println("=============================================")
		fmt.Printf("==>exec\n")
		cmd := strings.Replace(tweet.Text, t.hashtag, "", -1)

		ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		result, err := execOnContainer(ctxWithTimeout, cmd)
		if err != nil {
			result = fmt.Sprintf("%v\n%v\n", result, err)
		}
		t.quotedTweet(result, &tweet)

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
		return err
	}
mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			if err := quoteTweet(ctx, t); err != nil {
				return err
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
