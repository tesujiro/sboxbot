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
			fmt.Printf("twitter.search now=%s\tlatestId=%v\n", time.Now(), t.savedata.LatestId)
			for i, tweet := range t.search() {
				fmt.Printf("key:%d\tid:%d\tCreatedAt:%s\tUser.ScreenName:%s\n", i, tweet.Id, tweet.CreatedAt, tweet.User.ScreenName)
				fmt.Println("=============================================")
				fmt.Println(tweet.Text)
				fmt.Println("=============================================")
				fmt.Printf("%v\n", tweet)
				fmt.Println("=============================================")
				fmt.Printf("==>exec\n")
				cmd := strings.Replace(tweet.Text, t.hashtag, "", -1)
				d := newDockerContainer()
				if err := d.run(ctx); err != nil {
					result := fmt.Sprintf("%v", err)
					t.quotedTweet(result, &tweet)
					panic(err)
				}
				if err := d.exec(cmd); err != nil {
					result, _ := d.exit()
					result += fmt.Sprintf("%v", err)
					t.quotedTweet(result, &tweet)
					panic(err)
				}
				result, err := d.exit()
				if err != nil {
					result += fmt.Sprintf("%v", err)
					t.quotedTweet(result, &tweet)
					panic(err)
				}
				t.quotedTweet(result, &tweet)
			}
		}
	}
}
