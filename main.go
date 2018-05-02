package main

import (
	"context"
	"fmt"
	"time"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t := newTwitter()
	//tick := time.NewTicker(time.Second * time.Duration(60)).C
	tick := time.NewTicker(time.Second * time.Duration(10)).C
	hash := "#sboxbot"

mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			fmt.Printf("twitter.search %s\n", time.Now())
			for i, tweet := range t.search(hash) {
				createdAt, err := tweet.CreatedAtTime()
				if err != nil {
					panic(err)
				}
				if !createdAt.After(t.startedAt) {
					continue
				}
				fmt.Printf("key:%d\tid:%d\tCreatedAt:%s\tUser.ScreenName:%s\n", i, tweet.Id, tweet.CreatedAt, tweet.User.ScreenName)
				fmt.Println("=============================================")
				fmt.Println(tweet.Text)
				fmt.Println("=============================================")
				if tweet.RetweetCount == 0 {
					fmt.Printf("==>retweet\n")
					t.retweet(tweet.Id, true)
				}
				//result := execOnContainer(ctx, cmd)
				//fmt.Println(tweet.Text)
				//fmt.Println(tweet)
			}
		}
	}
}
