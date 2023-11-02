package searcher

type searchTree struct {
	firstNodes []*node
	idx        int
	stat       []int
}

func newSearchTree() *searchTree {
	return &searchTree{
		firstNodes: nil,
	}
}

func (t *searchTree) empty() {
	t.firstNodes = nil
	t.idx = 0
}

func (t *searchTree) setFirstNodes(nodes []*node) {
	t.firstNodes = nodes
}
