package main

import (
	"context"
	"time"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//api := GetTwitterApi()
	t := newTwitter()
	tick := time.NewTicker(time.Second * time.Duration(60)).C

mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			t.search()
		}
	}
	/*
		text := "Hello world"
		tweet, err := api.PostTweet(text, nil)
		if err != nil {
			panic(err)
		}
	*/

	//fmt.Print(tweet.Text)
}
