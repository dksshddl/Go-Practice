package test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
)

type Rule struct {
	Value string `json:"value"`
	Tag   string `json: "tag,omitempty"`
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

func TestTwitter(t *testing.T) {
	url_path := "https://api.twitter.com/2/tweets/search/stream/rules"
	method := "POST"
	token := os.Getenv("SP_TWITTER_BEARERTOKEN")
	u, err := url.Parse(url_path)
	if err != nil {
		t.Errorf("failed to parse URL, path : %s", url_path)
	}

	rule := Rule{Value: "cat has:images", Tag: "cats with images"}
	adder := map[string][]Rule{
		"add": {rule},
	}
	body, err := json.Marshal(adder)
	t.Log("body data : " + string(body) + "\n")
	t.Logf("body byte data : %v\n", body)

	if err != nil {
		t.Errorf("failed to Marshal input param")
	}
	req, err := http.NewRequest(method, u.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	t.Logf("%v\n", req)
	if err != nil {
		t.Errorf("failed to make new request,  (method, url) : (%s, %s)", method, u.String())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("failed to post request, %s", err)
	}
	t.Logf("%v\n", resp)
	if err != nil {
		t.Errorf("failed to post request, %s", err)
	}
	if resp.StatusCode/100 != 2 {
		t.Errorf("failed to request twitter api, error code: %v", resp.StatusCode)
	}

	query := url.Values{
		"tweet.fields": {"text"},
	}
	req, err = http.NewRequest("GET", "https://api.twitter.com/2/tweets/sample/stream?"+query.Encode(), nil)
	t.Logf("%v\n", req)
	if err != nil {
		t.Errorf("failed to make new request,  (method, url) : (GET, %s)", "https://api.twitter.com/2/tweets/sample/stream?"+query.Encode())
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("failed to get request, %s", err)
	}

	if resp.StatusCode/100 != 2 {
		t.Errorf("failed to request twitter api, error code: %v", resp.StatusCode)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	t.Logf("%v\n", string(respBody))
	if err != nil {
		t.Errorf("failed to read response body : %v", err)
	}

	rule = Rule{Value: "cat has:images", Tag: ""}
	remover := map[string]map[string][]string{
		"delete": {"values": {rule.Value}},
	}
	body, _ = json.Marshal(remover)

	req, _ = http.NewRequest(method, u.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	t.Logf("%v\n", req)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("failed to delete rule, %v", err)
	}
	if resp.StatusCode/100 != 2 {
		t.Errorf("failed to delete twitter api, error code: %v", resp.StatusCode)
	}
}

func TestStreamTwitter(t *testing.T) {
	query := url.Values{
		"tweet.fields": {"text"},
	}
	req, err := http.NewRequest("GET", "https://api.twitter.com/2/tweets/search/stream?"+query.Encode(), nil)
	bearer_token := os.Getenv("SP_TWITTER_BEARERTOKEN")
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "applicaion/json")
	req.Header.Set("Authorization", "Bearer "+bearer_token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	i := 0
	// bytesRead := 0
	reader := resp.Body
	decoder := json.NewDecoder(reader)

	for {
		var tw tweet
		if err := decoder.Decode(&tw); err != nil {
			t.Error(err)
			break
		}
		t.Log(tw)
		i++
		if i == 3 {
			break
		}
	}

	t.Fail()
}
