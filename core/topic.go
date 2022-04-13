package core

import (
	"fmt"
	"net/http"
	"net/url"
)

type Topic struct {
	Url        string
	Cli        *http.Client
	TopicId    string
	Content    string
	ArticleUrl string
	Time       string
	Owner      string
}

func (t *Topic) Abstract() string {
	content, err := url.QueryUnescape(t.Content)
	if err != nil {
		return ""
	}

	// chinese format
	chn := []rune(content)
	size := 48
	if len(chn) <= size {
		size = len(chn)
	}
	return fmt.Sprintf("[%s]: %s", t.Owner, string(chn[:size]))
}

func (t *Topic) Href() string {
	return fmt.Sprintf("/topics/%s", t.TopicId)
}

func (t *Topic) CTime() string {
	return t.Time
}
