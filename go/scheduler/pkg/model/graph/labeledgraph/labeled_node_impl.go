package labeledgraph

var (
	node *labeledNodeImpl

	_ LabeledNode = node
)

// labeledNodeImpl is the default implementation of the labeledgraph.LabeledNode interface
type labeledNodeImpl struct {
	id      int64
	label   string
	payload NodePayload
}

func newLabeledNodeImpl(id int64, label string) *labeledNodeImpl {
	return &labeledNodeImpl{
		id:    id,
		label: label,
	}
}

func (me *labeledNodeImpl) ID() int64 {
	return me.id
}

func (me *labeledNodeImpl) Label() string {
	return me.label
}

func (me *labeledNodeImpl) Payload() NodePayload {
	return me.payload
}

func (me *labeledNodeImpl) SetPayload(payload NodePayload) {
	me.payload = payload
}
