package srv

import (
	"encoding/json"
	"fmt"
	"github.com/o98k-ok/zsxq-kindler/core"
	"html/template"
	"net/http"
	"os"
	"strconv"
)

type Planet struct {
	*http.ServeMux
	group *core.Group
	conf  map[string]string
}

func (p *Planet) ListTopics(writer http.ResponseWriter, request *http.Request) {
	tmpl := template.Must(template.ParseFiles("template/index.html"))
	option := core.Option{}

	var err error
	option.Count, err = strconv.Atoi(p.conf["count"])
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "got error %s", err.Error())
		return
	}

	topics, err := p.group.ListTopics(option)
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "got error %s", err.Error())
		return
	}

	type htmlVar struct {
		Abstract string
		Href     string
		Time     string
	}

	var ss []htmlVar
	for _, t := range topics {
		ss = append(ss, struct {
			Abstract string
			Href     string
			Time     string
		}{Abstract: t.Abstract(), Href: t.Href(), Time: t.CTime()})
	}

	tmpl.Execute(writer, struct {
		Topics []htmlVar
	}{ss})
}

func (p *Planet) TopicDetail(writer http.ResponseWriter, request *http.Request) {
	res, err := p.group.Fetch(request.URL.Path[len("/topics/"):])
	if err != nil {
		fmt.Println(err)
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(writer, "got error %s", err.Error())
		return
	}
	fmt.Fprint(writer, res)
}

func NewPlanet() *Planet {
	xq := &Planet{}
	xq.ServeMux = http.NewServeMux()

	xq.ServeMux.HandleFunc("/", xq.ListTopics)
	xq.ServeMux.HandleFunc("/topics/", xq.TopicDetail)

	f, err := os.Open("./conf/config.json")
	if err != nil {
		fmt.Println(err)
		return nil
	}

	defer f.Close()
	err = json.NewDecoder(f).Decode(&(xq.conf))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	i, err := strconv.ParseUint(xq.conf["groupId"], 10, 64)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	xq.group = core.NewGroup(i, xq.conf["token"])
	return xq
}
