package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/joeshaw/envdecode"
	"github.com/matryer/go-oauth/oauth"
)

var conn net.Conn

var (
	authClient    *oauth.Client
	creds         *oauth.Credentials
	authSetupOnce sync.Once
	httpClient    *http.Client
	ts            struct {
		ConsumerKey    string `env:SP_TWITTER_KEY, required`
		ConsumerSecret string `env:SP_TWITTER_SECRET, required`
		AccessToken    string `env:SP_TWITTER_ACCESSTOKEN, required`
		AccessSecret   string `env:SP_TWITTER_ACCESSSECRET, required`
		BearerToken    string `env:SP_TWITTER_BEARERTOKEN, required`
	}
	stream_path = "https://api.twitter.com/2/tweets/search/stream"
)

type tweet struct {
	Text string
}

type Rule struct {
	Value string `json:"value"`
	Tag   string `json:"tag,omitempty"`
}

func setupTwitterAuth() {
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

func makeRequest() (*http.Response, error) {
	query := url.Values{
		"tweet.fields": {"text"},
	}
	req, err := http.NewRequest("GET", "https://api.twitter.com/2/tweets/search/stream?"+query.Encode(), nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "applicaion/json")
	req.Header.Set("Authorization", "Bearer "+ts.BearerToken)
	return httpClient.Do(req)
}

func dial(netw, addr string) (net.Conn, error) {
	if conn != nil {
		conn.Close()
		conn = nil
	}
	netc, err := net.DialTimeout(netw, addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	conn = netc
	return netc, nil
}

var reader io.ReadCloser

func closeConn() {
	if conn != nil {
		conn.Close()
	}
	if reader != nil {
		reader.Close()
	}
}

func UpdateRule() {
	options, err := loadOptions()
	if err != nil {
		log.Fatalln(err)
	}
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
	options, err := loadOptions()
	if err != nil {
		log.Fatalln(err)
	}
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
func readFromTwitter(votes chan<- string) {
	options, err := loadOptions()
	if err != nil {
		log.Fatalln(err)
	}
	resp, err := makeRequest()
	if err != nil {
		log.Println("making request failed: ", err)
		return
	}
	reader := resp.Body
	decoder := json.NewDecoder(reader)
	for {
		var t tweet
		if err := decoder.Decode(&t); err != nil {
			break
		}
		for _, option := range options {
			if strings.Contains(strings.ToLower(t.Text), strings.ToLower(option)) {
				log.Print("vote: ", option)
				votes <- option
			}
		}
	}
}

func startTwitterStream(stopchan <-chan struct{}, votes chan<- string) <-chan struct{} {
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
				readFromTwitter(votes)
				log.Println(" (waiting)")
				time.Sleep(10 * time.Second)
			}
		}
	}()
	return stoppedchan
}
