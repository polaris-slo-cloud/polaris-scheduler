package labeledgraph

var (
	_simpleWeightWrapper *simpleWeightWrapper

	_ ComplexEdgeWeight = _simpleWeightWrapper
)

// Wrapper for using simple float values as ComplexEdgeWeight.
type simpleWeightWrapper struct {
	weight float64
}

func NewComplexEdgeWeightFromFloat(simpleWeight float64) ComplexEdgeWeight {
	return &simpleWeightWrapper{
		weight: simpleWeight,
	}
}

func (me *simpleWeightWrapper) SimpleWeight() float64 {
	return me.weight
}
