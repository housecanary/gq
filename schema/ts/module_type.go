package ts

import "context"

type ModuleType[P any] struct {
	typePrefix string
	elements   []builderType
}

func Module[P any]() *ModuleType[P] {
	return &ModuleType[P]{}
}

func (mt *ModuleType[P]) Provider(ctx context.Context, provider P) context.Context {
	return context.WithValue(ctx, mt, provider)
}

func (mt *ModuleType[P]) GetProvider(ctx context.Context) P {
	return ctx.Value(mt).(P)
}

func (mt *ModuleType[P]) WithPrefix(prefix string) *ModuleType[P] {
	return &ModuleType[P]{
		typePrefix: prefix,
	}
}

func (mt *ModuleType[P]) addType(bt builderType) {
	mt.elements = append(mt.elements, bt)
}

func (mt *ModuleType[P]) namePrefix() string {
	return mt.typePrefix
}

func (mt *ModuleType[P]) types() []builderType {
	return mt.elements
}
