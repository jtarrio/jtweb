package extensions

import "github.com/yuin/goldmark/ast"

func isBlankNode(n ast.Node) bool {
	if n.Kind() != ast.KindText {
		return false
	}
	t := n.(*ast.Text)
	return t.Segment.Start == t.Segment.Stop
}

func getNonBlankNodes(nodes []ast.Node) []ast.Node {
	out := make([]ast.Node, 0, len(nodes))
	for _, n := range nodes {
		if !isBlankNode(n) {
			out = append(out, n)
		}
	}
	return out
}

func previousNonBlankSibling(n ast.Node) ast.Node {
	s := n.PreviousSibling()
	for s != nil && isBlankNode(s) {
		s = s.PreviousSibling()
	}
	return s
}

func nextNonBlankSibling(n ast.Node) ast.Node {
	s := n.NextSibling()
	for s != nil && isBlankNode(s) {
		s = s.NextSibling()
	}
	return s
}
