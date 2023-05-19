package ts

// A Module represents a group of related types from which a GQL schema can be constructed.
type Module struct {
	typePrefix string
	elements   []builderType
}

// NewModule creates a new Module
func NewModule() *Module {
	return &Module{}
}

// WithPrefix prefixes all types registered in this module with the given prefix.
// This is useful when using multiple modules as a form of namespacing
func (mt *Module) WithPrefix(prefix string) *Module {
	return &Module{
		typePrefix: prefix,
	}
}

func (mt *Module) addType(bt builderType) {
	mt.elements = append(mt.elements, bt)
}
