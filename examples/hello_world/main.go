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

	node1 := spacetime.MustCreateNode("Node", "node1", map[string]interface{}{"description": "node1"}, 1.0)
	node2 := spacetime.MustCreateNode("Node", "node2", map[string]interface{}{"description": "node2"}, 1.0)
	spacetime.MustCreateLink(node1, "related", node2, map[string]interface{}{"description": "i'm a link!"}, 1.0)
}
