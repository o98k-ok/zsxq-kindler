package main

import (
	"fmt"
	"github.com/o98k-ok/zsxq-kindler/core"
)

func main() {
	topic := core.Topic{}
	abstract := topic.Abstract()
	fmt.Println(abstract)
}
