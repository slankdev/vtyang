package vtyang

import (
	"github.com/openconfig/goyang/pkg/yang"
)

func yangCompletionForConfigurationData(e *yang.Entry) *CompletionNode {
	node := &CompletionNode{}
	node.Name = e.Name

	if e.ReadOnly() {
		return nil
	}

	switch {
	case e.IsList():
		crNode := &CompletionNode{}
		crNode.Name = "<cr>"
		wildcardNode := &CompletionNode{}
		wildcardNode.Name = "NAME"
		wildcardNode.Childs = append(wildcardNode.Childs, crNode)
		for _, child := range e.Dir {
			if child.Name != e.Key {
				n := yangCompletionForConfigurationData(child)
				if n != nil {
					wildcardNode.Childs = append(wildcardNode.Childs, n)
				}
			}
		}
		node.Childs = append(node.Childs, wildcardNode)

	case e.IsLeaf():
		crNode := &CompletionNode{}
		crNode.Name = "<cr>"
		wildcardNode := &CompletionNode{}
		wildcardNode.Name = "NAME"
		wildcardNode.Childs = append(wildcardNode.Childs, crNode)
		node.Childs = append(node.Childs, wildcardNode)

	default:
		for _, child := range e.Dir {
			n := yangCompletionForConfigurationData(child)
			if n != nil {
				node.Childs = append(node.Childs, n)
			}
		}
	}

	return node
}

func IsListKey(e *yang.Entry) bool {
	return e.Parent != nil && e.Parent.IsList() && e.Parent.Key == e.Name
}

func yangCompletionForOperationalData(e *yang.Entry) *CompletionNode {
	node := &CompletionNode{}
	node.Name = e.Name

	switch {
	case e.IsList():
		crNode := &CompletionNode{}
		crNode.Name = "<cr>"
		wildcardNode := &CompletionNode{}
		wildcardNode.Name = "NAME"
		wildcardNode.Childs = append(wildcardNode.Childs, crNode)
		for _, child := range e.Dir {
			if child.Name != e.Key {
				n := yangCompletionForOperationalData(child)
				if n != nil {
					wildcardNode.Childs = append(wildcardNode.Childs, n)
				}
			}
		}
		if len(wildcardNode.Childs) > 1 {
			node.Childs = append(node.Childs, wildcardNode)
		}

	case e.IsLeaf():
		if !e.ReadOnly() && !IsListKey(e) {
			return nil
		}
		crNode := &CompletionNode{}
		crNode.Name = "<cr>"
		wildcardNode := &CompletionNode{}
		wildcardNode.Name = "NAME"
		wildcardNode.Childs = append(wildcardNode.Childs, crNode)
		node.Childs = append(node.Childs, wildcardNode)

	default:
		for _, child := range e.Dir {
			n := yangCompletionForOperationalData(child)
			if n != nil {
				node.Childs = append(node.Childs, n)
			}
		}
	}

	if len(node.Childs) == 0 {
		return nil
	}
	return node
}
