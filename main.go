package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t := newTwitter()
	//tick := time.NewTicker(time.Second * time.Duration(60)).C
	tick := time.NewTicker(time.Second * time.Duration(10)).C

mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			fmt.Printf("twitter.search %s\n", time.Now())
			for i, tweet := range t.search() {
				fmt.Printf("key:%d\tid:%d\tCreatedAt:%s\tUser.ScreenName:%s\n", i, tweet.Id, tweet.CreatedAt, tweet.User.ScreenName)
				fmt.Println("=============================================")
				fmt.Println(tweet.Text)
				fmt.Println("=============================================")
				//fmt.Printf("%s\n", tweet)
				//fmt.Println("=============================================")
				if tweet.RetweetCount == 0 {
					fmt.Printf("==>exec\n")
					cmd := strings.Replace(tweet.Text, t.hashtag, "", -1)
					result := execOnContainer(ctx, cmd)
					t.quotedTweet(result, &tweet)
				}
			}
		}
	}
}
