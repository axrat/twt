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
	"time"
	"io"
	"encoding/csv"
)
type Client struct {
	TWITTER_CONSUMER_KEY        string
	TWITTER_CONSUMER_SECRET     string
	TWITTER_ACCESS_TOKEN        string
	TWITTER_ACCESS_TOKEN_SECRET string
}
var client Client
func setup(){
	err := godotenv.Overload()
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
		{
			Name:    "test",
			Usage:   "テスト",
			Action:  testAction,
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
func getApi()(*anaconda.TwitterApi){
	anaconda.SetConsumerKey(client.TWITTER_CONSUMER_KEY)
	anaconda.SetConsumerSecret(client.TWITTER_CONSUMER_SECRET)
	api := anaconda.NewTwitterApi(client.TWITTER_ACCESS_TOKEN, client.TWITTER_ACCESS_TOKEN_SECRET)
	api.SetLogger(anaconda.BasicLogger)
	return api
}
func testAction(c *cli.Context) {
	goz.Echo("TestAction")
	//FixOnoieTweet()
	//RemoveTweetFromCSV()
}
func getExpId()( []int64){
	return []int64{
		869525354259521536, //pin
		959537183773175808, //work
		}
}
func RemoveTweetFromCSV(){
	ReadTweetsCSV(0,getExpId())
}
func ReadTweetsCSV(ind int,skipids []int64){
	slice := csvToSlices()
	fmt.Println("columns:",len(slice[0]))
	rows:=len(slice)
	fmt.Println("rows:",rows)
	fmt.Println(slice[0])
	//[tweet_id in_reply_to_status_id in_reply_to_user_id timestamp source text retweeted_status_id retweeted_status_user_id retweeted_status_timestamp expanded_urls]
	for i := 1+ind; i < len(slice); i++ {
		//for api rate 15min / 900
		time.Sleep(1 * time.Second + 5 * time.Microsecond)

		s:=slice[i]
		tweet_id, _ := strconv.Atoi(s[0])
		timestamp:=s[3]
		skip:=false
		for _, id := range skipids {
			if id == int64(tweet_id){
				skip=true
			}
		}
		if !skip {
			tweet, err := getApi().GetTweet(int64(tweet_id),nil)
			if err == nil {
				if noMedia(tweet){
					TweetRemove(getApi(),int64(tweet_id))
				}
			}else{
				fmt.Println("ERROR")
			}
		}
		fmt.Println("No.", i, "/",rows,tweet_id ,timestamp,"SKIP",skip)
	}
}
func csvToSlices()([][]string){
	bulkCount := 100
	file, _ := os.Open("./tweets.csv")
	defer file.Close()
	reader := csv.NewReader(file)
	//header, _ := reader.Read()
	lines := make([][]string, 0, bulkCount)
	for {
		isLast := false
		for i := 0; i < bulkCount; i++ {
			line, err := reader.Read()
			if err == io.EOF {
				isLast = true
				break
			} else if err != nil {
				panic(err)
			}
			lines = append(lines, line)
		}
		if isLast {
			break
		}
	}
	return lines
}
func addQuery(base string,q string)(string){
	return base+" "+q
}
func createSinceUntilStr(year int,month int)(string){
	return fmt.Sprintf("since:%04d-%02d-%02d until:%04d-%02d-%02d",
		year,month,01,year,month,getLastDay(year,month))
}
func getLastDay(year int,month int)(lastDay int){
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	return time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, loc).
		AddDate(0, 0, -1).Day()
}
func TweetSearch(){
	api := getApi()
	var q string = "from:onoie3"
	q=addQuery(q,createSinceUntilStr(2018,1))
	//q=addQuery(q,"filter:images")
	searchResult, err := api.GetSearch(q, nil)
	if err != nil{
		panic(err)
	}
	for no, tweet := range searchResult.Statuses {
		fmt.Println(no,tweet.FullText)
	}
}
func tweetAction(c *cli.Context) {
	var isDebug = c.GlobalBool("debug")
	if isDebug {
		fmt.Println("TWITTER_CONSUMER_KEY:", client.TWITTER_CONSUMER_KEY)
		fmt.Println("TWITTER_CONSUMER_SECRET:", client.TWITTER_CONSUMER_SECRET)
		fmt.Println("TWITTER_ACCESS_TOKEN:", client.TWITTER_ACCESS_TOKEN)
		fmt.Println("TWITTER_ACCESS_TOKEN_SECRET:", client.TWITTER_ACCESS_TOKEN_SECRET)
	}else{
		if len(c.Args()) > 0 {
			var twtstr string
			for i := range c.Args() {
				twtstr+=" "+c.Args().Get(i)
			}
			fmt.Println(twtstr)
			Tweet(getApi(),twtstr)
		}else{
			fmt.Println("require tweet string")
		}
	}
}
func FixOnoieTweet(){
	v := createValues("onoie3")
	//addMaxId(v,"")
	lastid:= removeNotMediaTweet(getApi(),v,getExpId())
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
func removeNotMediaTweet(api *anaconda.TwitterApi,v url.Values, skipId []int64)(int64){
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
		skip:=false
		for _, id := range skipId {
			if id == tweet.Id{
				skip=true
			}
		}
		if skip {
			fmt.Println("=>SkipID")
		}else{
			if noMedia(tweet){
				fmt.Println(" = None")
				TweetRemove(api,tweet.Id)
			}else{
				fmt.Println(" =",tweet.Entities.Media[0].Id)
			}
		}
	}
	return last
}
func noMedia(tweet anaconda.Tweet)(bool){
	entities := tweet.Entities
	entirymedia := entities.Media
	if entirymedia == nil {
		return true
	}else {
		return false
	}
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

