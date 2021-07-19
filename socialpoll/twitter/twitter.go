package twitter

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"socialpoll/utils"
	"strings"
	"sync"
	"time"

	"github.com/joeshaw/envdecode"
	"github.com/matryer/go-oauth/oauth"
)

var (
	options       []string
	authClient    *oauth.Client
	creds         *oauth.Credentials
	authSetupOnce sync.Once
	httpClient    *http.Client
	reader        io.ReadCloser
	conn          net.Conn
	ts            struct {
		ConsumerKey    string `env:SP_TWITTER_KEY, required`
		ConsumerSecret string `env:SP_TWITTER_SECRET, required`
		AccessToken    string `env:SP_TWITTER_ACCESSTOKEN, required`
		AccessSecret   string `env:SP_TWITTER_ACCESSSECRET, required`
		BearerToken    string `env:SP_TWITTER_BEARERTOKEN, required`
	}
	stream_path = "https://api.twitter.com/2/tweets/search/stream"
)

func init() {
	id := utils.Setting.Database[0].Id
	pw := utils.Setting.Database[0].Password
	path := utils.Setting.Database[0].Path

	dbpath := "mongodb+srv://" + id + ":" + pw + "@" + path
	client, ctx := utils.Dialdb(dbpath)
	log.Println("load options...")
	options = utils.LoadOptions(client, ctx)
	log.Println("finish load options: ", strings.Join(options, ", "))
}

type tweet struct {
	Data struct {
		Id   json.Number `json:"id,omitempty"`
		Text string      `json:"text,omitempty"`
	} `json:"data,omitempty"`
	MatchingRules []struct {
		Id  json.Number `json:"id,omitempty"`
		Tag string      `json:"tag,omitempty"`
	} `json:"matching_rules,omitempty"`
}

type Rule struct {
	Value string `json:"value"`
	Tag   string `json:"tag,omitempty"`
}

func SetupTwitterAuth() {
	if err := envdecode.Decode(&ts); err != nil {
		log.Fatalln(err)
	}
	creds = &oauth.Credentials{
		Token:  ts.AccessToken,
		Secret: ts.AccessSecret,
	}
	authClient = &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  ts.ConsumerKey,
			Secret: ts.ConsumerSecret,
		},
	}
}

func MakeRequest() (*http.Response, error) {
	query := url.Values{
		"tweet.fields": {"text"},
	}
	req, err := http.NewRequest("GET", "https://api.twitter.com/2/tweets/search/stream?"+query.Encode(), nil)
	bearer_token := os.Getenv("SP_TWITTER_BEARERTOKEN")

	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "applicaion/json")
	req.Header.Set("Authorization", "Bearer "+bearer_token)

	log.Println("request http client request, ", req.URL)
	return http.DefaultClient.Do(req)
}

func Dial(netw, addr string) (net.Conn, error) {
	if conn != nil {
		conn.Close()
		conn = nil
	}
	log.Println("Dial DB...")
	netc, err := net.DialTimeout(netw, addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	conn = netc
	return netc, nil
}

func CloseConn() {
	if conn != nil {
		conn.Close()
	}
	if reader != nil {
		reader.Close()
	}
}

func UpdateRule() {
	log.Println("delete rules... ")
	log.Println("  Options : " + strings.Join(options, ", "))
	rules := make([]Rule, 0)

	for _, option := range options {
		rules = append(rules, Rule{option, ""})
	}

	u, err := url.Parse(stream_path + "/rules")
	if err != nil {
		log.Fatalln(err)
	}

	adder := map[string][]Rule{
		"add": rules,
	}
	body, err := json.Marshal(adder)
	if err != nil {
		log.Fatalln(err)
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ts.BearerToken)
	if err != nil {
		log.Fatalln(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	} else if resp.StatusCode/100 != 2 {
		log.Fatalln("failed to update rule")
	}
	log.Println("sucecss to update rules " + strings.Join(options, ", "))
}

func DeleteRule() {
	log.Println("delete rules... ")
	log.Println("  Options : " + strings.Join(options, ", "))
	u, err := url.Parse(stream_path + "/rules")
	if err != nil {
		log.Fatalln(err)
	}
	deleter := map[string]map[string][]string{
		"delete": {"values": options},
	}
	body, err := json.Marshal(deleter)
	if err != nil {
		log.Fatalln(err)
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(body))
	if err != nil {
		log.Fatalln(err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	if resp.StatusCode/100 != 2 {
		log.Fatalln("failed to delete rules status code is " + string(resp.StatusCode))
	}
	log.Println("sucecss to delete rules " + strings.Join(options, ", "))
}

// votes라는 전송 전용 채널 사용
func ReadFromTwitter(votes chan<- string) {
	resp, err := MakeRequest()
	if err != nil {
		log.Println("making request failed: ", err)
		return
	}
	defer resp.Body.Close()
	reader = resp.Body
	decoder := json.NewDecoder(reader)
	log.Println("Request twitter data")
	for {
		var tw tweet
		if err := decoder.Decode(&tw); err != nil {
			break
		}
		for _, option := range options {
			if strings.Contains(strings.ToLower(tw.Data.Text), strings.ToLower(option)) {
				log.Print("vote: ", option)
				votes <- option
			}
		}
		time.Sleep(time.Second * 1)
	}
}

func StartTwitterStream(stopchan <-chan struct{}, votes chan<- string) <-chan struct{} {
	stoppedchan := make(chan struct{}, 1)
	go func() {
		defer func() {
			stoppedchan <- struct{}{}
		}()
		for {
			select {
			case <-stopchan:
				log.Println("stopping Twitter...")
				return
			default:
				log.Println("Querying Twitter...")
				ReadFromTwitter(votes)
				log.Println(" (waiting)")
				time.Sleep(10 * time.Second)
			}
		}
	}()
	return stoppedchan
}
