package main

import (
	"fmt"

	"github.com/tristanls/sst"
)

func main() {
	fmt.Println("Hello World")

	config := &sst.Config{
		Name:            "my_test_graph_database",
		NodeCollections: []string{"Node"},
		Password:        "",
		URL:             "http://localhost:8529",
		Username:        "root",
	}
	spacetime, err := sst.NewSST(config)
	if err != nil {
		panic(err)
	}

	node1, err := spacetime.CreateNode("node1", "node1", "Node", 1.0)
	if err != nil {
		panic(err)
	}
	node2, err := spacetime.CreateNode("node2", "node2", "Node", 1.0)
	if err != nil {
		panic(err)
	}
	err = spacetime.CreateLink(node1, "related", node2, 1.0)
	if err != nil {
		panic(err)
	}
}
