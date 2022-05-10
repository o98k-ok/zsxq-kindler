package core

import (
	"encoding/json"
	"fmt"
	"github.com/matryer/try"
	"github.com/o98k-ok/lazy/v2/assert"
	hp "github.com/o98k-ok/pcurl/http"
	"jaytaylor.com/html2text"
	"net/http"
	"net/url"
	"strconv"
)

type Group struct {
	groupId   uint64
	domainUrl string
	cli       *http.Client
	token     string
}

func NewGroup(groupId uint64, token string) *Group {
	return &Group{
		groupId:   groupId,
		domainUrl: "https://api.zsxq.com/v2",
		cli:       DefaultClient,
		token:     token,
	}
}

type Option struct {
	Count  int
	Preset bool
	Tag    string
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

func (g *Group) ListMenus() ([]Menu, error) {
	res := make([]Menu, 0, 5)

	u := fmt.Sprintf("%s/groups/%d/menus", g.domainUrl, g.groupId)
	r := hp.NewRequest(g.cli, u).AddHeader("cookie", g.constructCookie())

	var body struct {
		RespData struct {
			Menus []struct {
				MenuID     int64  `json:"menu_id"`
				Title      string `json:"title"`
				Preset     bool   `json:"preset,omitempty"`
				PresetType string `json:"preset_type,omitempty"`
				Hashtag    struct {
					HashtagID int64 `json:"hashtag_id"`
				} `json:"hashtag,omitempty"`
			} `json:"menus"`
		} `json:"resp_data"`
		Code int `json:"code"`
	}

	e := try.Do(func(attempt int) (bool, error) {
		resp, err := r.Do()
		err = assert.NoError(err).AndEqual(http.StatusOK, resp.Code).Unwrap()
		if err != nil {
			return false, err
		}

		err = json.NewDecoder(resp.Body).Decode(&body)
		err = assert.NoError(err).AndEqual(0, body.Code).Unwrap()
		if err == nil {
			return false, nil
		}
		body.Code = 0
		return attempt < 3, err
	})
	if e != nil {
		return res, e
	}

	for _, menu := range body.RespData.Menus {
		if menu.Preset {
			res = append(res, &PresetMenu{
				gid:   g.groupId,
				name:  menu.Title,
				pType: menu.PresetType,
			})
		} else {
			res = append(res, &CustomMenu{
				gid:  g.groupId,
				name: menu.Title,
				hId:  menu.Hashtag.HashtagID,
			})
		}
	}
	return res, nil
}

func (g *Group) ListTopics(option Option) ([]Topic, error) {
	res := make([]Topic, 0, option.Count)
	var r *hp.Request
	if option.Preset {
		// https://api.zsxq.com/v2/groups/48848444885258/topics?scope=all&count=20s
		u := fmt.Sprintf("%s/groups/%d/topics", g.domainUrl, g.groupId)
		r = hp.NewRequest(g.cli, u).AddParam("scope", option.Tag).AddParam("count", strconv.Itoa(option.Count))
	} else {
		// https://api.zsxq.com/v2/hashtags/51122412445154/topics?count=20
		u := fmt.Sprintf("%s/hashtags/%s/topics", g.domainUrl, option.Tag)
		r = hp.NewRequest(g.cli, u).AddParam("count", strconv.Itoa(option.Count))
	}
	r.AddHeader("cookie", g.constructCookie())
	r.AddHeader("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36")
	r.AddHeader("accept", " application/json, text/plain, */*")

	var body struct {
		RespData struct {
			Topics []TopicStruct `json:"topics"`
		} `json:"resp_data"`
		Code int `json:"code"`
	}

	// retry 3 times
	// when http code OK && json code != 0
	e := try.Do(func(attempt int) (bool, error) {
		resp, err := r.Do()
		err = assert.NoError(err).AndEqual(http.StatusOK, resp.Code).Unwrap()
		if err != nil {
			return false, err
		}

		err = json.NewDecoder(resp.Body).Decode(&body)
		err = assert.NoError(err).AndEqual(0, body.Code).Unwrap()
		if err == nil {
			return false, nil
		}
		body.Code = 0
		return attempt < 3, err
	})
	if e != nil {
		return res, e
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
	var body struct {
		RespData struct {
			Topic TopicStruct `json:"topic"`
		} `json:"resp_data"`
		Code int `json:"code"`
	}

	// retry 3 times
	// when http code OK && json code != 0
	e := try.Do(func(attempt int) (bool, error) {
		resp, err := hp.NewRequest(g.cli, "https://api.zsxq.com/v2/topics/"+topicId).AddHeader("cookie", g.constructCookie()).Do()
		err = assert.NoError(err).AndEqual(http.StatusOK, resp.Code).Unwrap()
		if err != nil {
			return false, err
		}

		err = json.NewDecoder(resp.Body).Decode(&body)
		err = assert.NoError(err).AndEqual(0, body.Code).Unwrap()
		if err == nil {
			return false, nil
		}
		body.Code = 0
		return attempt < 3, err
	})
	if e != nil {
		return "", e
	}

	if len(body.RespData.Topic.Talk.Article.ArticleURL) == 0 {
		return url.QueryUnescape(body.RespData.Topic.Talk.Text)
	}
	resp, err := hp.NewRequest(g.cli, body.RespData.Topic.Talk.Article.ArticleURL).AddHeader("cookie", g.constructCookie()).Do()
	if err != nil {
		return "", err
	}
	return html2text.FromString(resp.String(), html2text.Options{PrettyTables: true})
}
