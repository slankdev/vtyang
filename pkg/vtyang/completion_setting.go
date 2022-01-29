package vtyang

import (
	"fmt"
	"log"

	"github.com/openconfig/goyang/pkg/yang"
)

func do2(e *yang.Entry) CompletionNode {
	logstr := fmt.Sprintf("DO2 TRANSLATE FROM YANG %s", e.Name)
	switch {
	case e.IsList():
		logstr += fmt.Sprintf(" [%s]", e.Key)
	case e.IsLeaf():
		logstr += " (leaf)"
	}

	log.Println(logstr)

	node := CompletionNode{}
	node.Name = e.Name

	switch {

	case e.IsList():
		crNode := CompletionNode{}
		crNode.Name = "<cr>"
		wildcardNode := CompletionNode{}
		wildcardNode.Name = "NAME"
		wildcardNode.Childs = append(wildcardNode.Childs, crNode)
		for _, child := range e.Dir {
			if child.Name != e.Key {
				wildcardNode.Childs = append(wildcardNode.Childs, do2(child))
			}
		}
		node.Childs = append(node.Childs, wildcardNode)

	case e.IsLeaf():
		crNode := CompletionNode{}
		crNode.Name = "<cr>"
		wildcardNode := CompletionNode{}
		wildcardNode.Name = "NAME"
		wildcardNode.Childs = append(wildcardNode.Childs, crNode)
		node.Childs = append(node.Childs, wildcardNode)

	default:
		for _, child := range e.Dir {
			node.Childs = append(node.Childs, do2(child))
		}

	}

	return node
}

func setCompletionTreeForCommandShowOperationalData() {
	tree := GetCommandNode(CliModeView).tree
	root := DigNode(&tree.Root, []string{"show", "operational-data"})
	if root == nil {
		panic("OKASHII")
	}
	ents := dbm.DumpEntries()
	for _, e := range ents {
		log.Printf("hoge %s\n", e.Name)
		root.Childs = append(root.Childs, do2(e))
	}
}

func setCompletionTreeForCommandDelete() {
	tree := GetCommandNode(CliModeConfigure).tree
	root := DigNode(&tree.Root, []string{"delete"})
	if root == nil {
		panic("OKASHII")
	}
	ents := dbm.DumpEntries()
	for _, e := range ents {
		log.Printf("hoge %s\n", e.Name)
		root.Childs = append(root.Childs, do2(e))
	}
}

func setCompletionTreeForCommandSet() {
	tree := GetCommandNode(CliModeConfigure).tree
	root := DigNode(&tree.Root, []string{"set"})
	if root == nil {
		panic("OKASHII")
	}
	ents := dbm.DumpEntries()
	for _, e := range ents {
		log.Printf("hoge %s\n", e.Name)
		root.Childs = append(root.Childs, do2(e))
	}
}
