package ts

// An InputObjectType represents the GQL type of an input object created from Go structs
type InputObjectType[O any] struct {
	def string
}

// NewInputObjectType creates an InputObjectType and registers it with the given module
func NewInputObjectType[O any](mod *Module, def string) *InputObjectType[O] {
	it := &InputObjectType[O]{
		def: def,
	}
	mod.addType(&inputObjectTypeBuilder[O]{it: it})
	return it
}
