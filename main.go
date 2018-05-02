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
			for i, tweet := range t.search(hash) {
				createdAt, err := tweet.CreatedAtTime()
				if err != nil {
					panic(err)
				}
				if !createdAt.After(t.startedAt) {
					continue
				}
				fmt.Printf("key:%d\tid:%d\tCreatedAt:%s\tSearchStatedAt:%s\n", i, tweet.Id, tweet.CreatedAt, t.startedAt)
				t.retweet(tweet.Id, true)
				//result := execOnContainer(ctx, cmd)
				//fmt.Println(tweet.Text)
				//fmt.Println(tweet)
			}
		}
	}
}
