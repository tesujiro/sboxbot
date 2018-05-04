package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
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
				fmt.Printf("%s\n", tweet)
				fmt.Println("=============================================")
				if tweet.RetweetCount == 0 {
					fmt.Printf("==>exec\n")
					//t.retweet(tweet.Id, true)
					cmd := strings.Replace(tweet.Text, hash, "", -1)
					//cmd = strings.TrimRight(cmd, "\n")
					result := execOnContainer(ctx, cmd)
					v := url.Values{}
					v.Add("in_reply_to_status_id", fmt.Sprint(tweet.Id, true))
					t.post(result, v)
				}
				//fmt.Println(tweet.Text)
				//fmt.Println(tweet)
			}
		}
	}
}
