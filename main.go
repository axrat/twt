package main

import (
	"github.com/TransAssist/goz"
	"github.com/ChimeraCoder/anaconda"
	"github.com/joho/godotenv"
)
import (
	"fmt"
	"log"
	"os"
	"net/url"
	"encoding/base64"
	"strconv"
	"io/ioutil"
	"github.com/urfave/cli"
)
type Client struct {
	TWITTER_CONSUMER_KEY        string
	TWITTER_CONSUMER_SECRET     string
	TWITTER_ACCESS_TOKEN        string
	TWITTER_ACCESS_TOKEN_SECRET string
}
var client Client
func setup(){
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	client = Client{
		os.Getenv("TWITTER_CONSUMER_KEY"),
		os.Getenv("TWITTER_CONSUMER_SECRET"),
		os.Getenv("TWITTER_ACCESS_TOKEN"),
		os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")}
}
func main() {
	app := cli.NewApp()
	app.Name = "twt"
	app.Usage = "TwitterApp"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "デバッグオプション",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "version",
			Aliases: []string{"-v"},
			Usage:   "バージョン",
			Action:  func(c *cli.Context) error {
				cli.ShowVersion(c)
				return nil
			},
		},
		{
			Name:    "tweet",
			Aliases: []string{"t"},
			Usage:   "引数でツイート",
			Action:  tweetAction,
		},
	}
	app.Before = func(c *cli.Context) error {
		setup()
		return nil
	}
	app.After = func(c *cli.Context) error {
		goz.Complete()
		return nil
	}
	app.Run(os.Args)
}
func tweetAction(c *cli.Context) {
	var isDebug = c.GlobalBool("debug")
	if isDebug {
		fmt.Println("TWITTER_CONSUMER_KEY:", client.TWITTER_CONSUMER_KEY)
		fmt.Println("TWITTER_CONSUMER_SECRET:", client.TWITTER_CONSUMER_SECRET)
		fmt.Println("TWITTER_ACCESS_TOKEN:", client.TWITTER_ACCESS_TOKEN)
		fmt.Println("TWITTER_ACCESS_TOKEN_SECRET:", client.TWITTER_ACCESS_TOKEN_SECRET)
	}
	if len(c.Args()) > 0 {
		anaconda.SetConsumerKey(client.TWITTER_CONSUMER_KEY)
		anaconda.SetConsumerSecret(client.TWITTER_CONSUMER_SECRET)
		api := anaconda.NewTwitterApi(client.TWITTER_ACCESS_TOKEN, client.TWITTER_ACCESS_TOKEN_SECRET)
		api.SetLogger(anaconda.BasicLogger)
		var twtstr string
		for i := range c.Args() {
			twtstr+=" "+c.Args().Get(i)
		}
		fmt.Println(twtstr)
		Tweet(api,twtstr)
	}else{
		fmt.Println("require tweet string")
	}
}
func fixOnoie3(api *anaconda.TwitterApi){
	v := createValues("onoie3")
	//addMaxId(v,"")
	var pin_tweet int64 = 869525354259521536
	lastid:=RemoveNotMediaTweet(api,v,pin_tweet)
	fmt.Println("LastTweetId:",lastid)
}
func createValues(screen_name string)(url.Values){
	v:=url.Values{}
	v.Set("screen_name", screen_name)
	return v
}
func addMaxId(v url.Values,maxid string)(url.Values){
	v.Set("max_id",maxid)
	return v
}
func RemoveNotMediaTweet(api *anaconda.TwitterApi,v url.Values, skipId int64)(int64){
	v.Set("include_rts", "true")
	v.Set("count", "200")
	tweets, err := api.GetUserTimeline(v)
	if err != nil {
		panic(err)
	}
	var last int64
	for index, tweet := range tweets {
		last=tweet.Id
		fmt.Print(tweet.Id," No.",index+1)
		if skipId != tweet.Id {
			entities := tweet.Entities
			entirymedia := entities.Media
			if entirymedia == nil {
				fmt.Println(" = None")
				TweetRemove(api,tweet.Id)
			}else {
				fmt.Println(" =",entities.Media[0].Id)
			}
		}
	}
	return last
}
func GetUserTimeline(api *anaconda.TwitterApi,screen_name string){
	v:=url.Values{}
	//v.Set("user_id", "261237740")
	v.Set("screen_name", screen_name,)
	v.Set("include_rts", "true")
	v.Set("count", "10")
	tweets, err := api.GetUserTimeline(v)
	if err != nil {
		panic(err)
	}
	for index, tweet := range tweets {
		fmt.Println(tweet.Id,"No.",index+1, tweet.FullText)
	}
}
func TweetRemove(api *anaconda.TwitterApi,id int64)(){
	tweet, err := api.DeleteTweet(id,false)
	if err != nil {
		panic(err)
	}
	fmt.Println(tweet.Text)
}
func GetUserID(api *anaconda.TwitterApi,username string)(int64){
	user, err := api.GetUsersShow(username,url.Values{})
	if err != nil {
		panic(err)
	}
	return user.Id
}
func Tweet(api *anaconda.TwitterApi,status string){
	tweet, err := api.PostTweet(status, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("tweet:",tweet)
}
func TweetWithBase64Image(api *anaconda.TwitterApi,base64String string,status string){
	media, _ := api.UploadMedia(base64String)
	v := url.Values{}
	v.Add("media_ids", media.MediaIDString)
	tweet, err := api.PostTweet(status, v)
	if err != nil {
		panic(err)
	}
	fmt.Println("tweet:",tweet)

}
func TweetWithLocalImage(api *anaconda.TwitterApi,filename string,status string){
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	mediaResponse, err := api.UploadMedia(base64.StdEncoding.EncodeToString(data))
	if err != nil {
		fmt.Println(err)
	}
	v := url.Values{}
	v.Set("media_ids", strconv.FormatInt(mediaResponse.MediaID, 10))
	result, err := api.PostTweet(status, v)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}
}

