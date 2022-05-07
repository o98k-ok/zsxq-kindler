package core

import (
	"encoding/json"
	"fmt"
	"github.com/o98k-ok/lazy/v2/assert"
	hp "github.com/o98k-ok/pcurl/http"
	"jaytaylor.com/html2text"
	"net/http"
	"net/url"
	"strconv"
)

type Group struct {
	groupId uint64
	scope   string
	url     string
	cli     *http.Client
	token   string
}

func NewGroup(groupId uint64, token string) *Group {
	return &Group{
		groupId: groupId,
		scope:   "by_owner",
		url:     fmt.Sprintf("https://api.zsxq.com/v2/groups/%d/topics", groupId),
		cli:     DefaultClient,
		token:   token,
	}
}

type Option struct {
	Count int
}

type TopicStruct struct {
	CreateTime string `json:"create_time"`
	TopicID    int64  `json:"topic_id"`
	Talk       struct {
		Owner struct {
			Name string `json:"name"`
		} `json:"Owner"`
		Text    string `json:"text"`
		Article struct {
			ArticleURL string `json:"article_url"`
		} `json:"article"`
	} `json:"talk"`
}

func (g *Group) constructCookie() string {
	return "zsxq_access_token=" + g.token + ";"
}

func (g *Group) ListTopics(option Option) ([]Topic, error) {
	res := make([]Topic, 0, option.Count)

	r := hp.NewRequest(g.cli, g.url).AddParam("scope", g.scope).AddParam("count", strconv.Itoa(option.Count)).AddHeader("cookie", g.constructCookie())
	r.AddHeader("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36")
	r.AddHeader("accept", " application/json, text/plain, */*")

	resp, err := r.Do()
	err = assert.IfNoError(err).ElIfEqual(http.StatusOK, resp.Code).Unwrap()
	if err != nil {
		return res, err
	}

	var body struct {
		RespData struct {
			Topics []TopicStruct `json:"topics"`
		} `json:"resp_data"`
		Code int `json:"code"`
	}

	err = json.NewDecoder(resp.Body).Decode(&body)
	err = assert.IfNoError(err).ElIfEqual(0, body.Code).Unwrap()
	if err != nil {
		return res, err
	}

	for _, topic := range body.RespData.Topics {
		res = append(res, Topic{
			TopicId:    strconv.FormatInt(topic.TopicID, 10),
			Content:    topic.Talk.Text,
			ArticleUrl: topic.Talk.Article.ArticleURL,
			Owner:      topic.Talk.Owner.Name,
			Time:       topic.CreateTime,
		})
	}

	return res, nil
}

func (g *Group) Fetch(topicId string) (string, error) {
	resp, err := hp.NewRequest(g.cli, "https://api.zsxq.com/v2/topics/"+topicId).AddHeader("cookie", g.constructCookie()).Do()
	err = assert.IfNoError(err).ElIfEqual(http.StatusOK, resp.Code).Unwrap()
	if err != nil {
		return "", err
	}

	var body struct {
		RespData struct {
			Topic TopicStruct `json:"topic"`
		} `json:"resp_data"`
		Code int `json:"code"`
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	err = assert.IfNoError(err).ElIfEqual(0, body.Code).Unwrap()
	if err != nil {
		return "", err
	}

	if len(body.RespData.Topic.Talk.Article.ArticleURL) == 0 {
		return url.QueryUnescape(body.RespData.Topic.Talk.Text)
	}

	resp, err = hp.NewRequest(g.cli, body.RespData.Topic.Talk.Article.ArticleURL).AddHeader("cookie", g.constructCookie()).Do()
	if err != nil {
		return "", err
	}
	return html2text.FromString(resp.String(), html2text.Options{PrettyTables: true})
}
