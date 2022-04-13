package main

import (
	"fmt"
	"github.com/o98k-ok/zsxq-kindler/core"
	"os"
)

func main() {
	cookie := os.Getenv("cookie")
	group := core.NewGroup(48848444885258, cookie)
	topics, err := group.ListTopics(core.Option{1})
	fmt.Println(err)
	fmt.Println(topics)
}
