package main

import (
	"fmt"
)

//const HASHTAG = "#sboxbot"
const HASHTAG = "#smallbox"

func search() {
	searchResult, err := api.GetSearch(HASHTAG, nil)
	if err != nil {
		panic(err)
	}
	for i, tweet := range searchResult.Statuses {
		fmt.Printf("key:%d\tid:%d\tCreatedAt:%s\tUser:%s\n", i, tweet.Id, tweet.CreatedAt, tweet.User)
		fmt.Println(tweet.Text)
		//fmt.Println(tweet)
	}
}

func main() {

	api := GetTwitterApi()
	tick := time.NewTicker(time.Second * time.Duration(60)).C

mainloop:
	for {
		select {
		case <-ctx.Done():
			break mainloop
		case <-tick:
			search()
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
