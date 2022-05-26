package ts

// A ModuleType represents a group of related types from which a GQL schema can be constructed.
type ModuleType struct {
	typePrefix string
	elements   []builderType
}

// Module creates a new ModuleType
func Module() *ModuleType {
	return &ModuleType{}
}

// WithPrefix prefixes all types registered in this module with the given prefix.
// This is useful when using multiple modules as a form of namespacing
func (mt *ModuleType) WithPrefix(prefix string) *ModuleType {
	return &ModuleType{
		typePrefix: prefix,
	}
}

func (mt *ModuleType) addType(bt builderType) {
	mt.elements = append(mt.elements, bt)
}
