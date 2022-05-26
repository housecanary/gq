package query

import (
	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

type fieldSelectionIterator struct {
	current   *objectSelectorField
	selectors []*objectSelector
	fields    []*objectSelectorField
}

func newSelectionIterator(sel selector) *fieldSelectionIterator {
	for {
		switch t := sel.(type) {
		case listSelector:
			sel = t.ElementSelector
			continue
		case notNilSelector:
			sel = t.Delegate
			continue
		}
		break
	}

	switch t := sel.(type) {
	case *objectSelector:
		return &fieldSelectionIterator{nil, nil, t.Fields}
	case interfaceSelector:
		selectors := make([]*objectSelector, 0, len(t.Elements))
		for _, e := range t.Elements {
			selectors = append(selectors, e.(*objectSelector))
		}
		return &fieldSelectionIterator{nil, selectors, nil}
	case unionSelector:
		selectors := make([]*objectSelector, 0, len(t.Elements))
		for _, e := range t.Elements {
			selectors = append(selectors, e.(*objectSelector))
		}
		return &fieldSelectionIterator{nil, selectors, nil}
	}
	return &fieldSelectionIterator{nil, nil, nil}
}

func (i *fieldSelectionIterator) Next() bool {
	for {
		if len(i.fields) > 0 {
			i.current = i.fields[0]
			i.fields = i.fields[1:]
			return true
		}

		if len(i.selectors) > 0 {
			i.fields = i.selectors[0].Fields
			i.selectors = i.selectors[1:]
			continue
		}

		i.current = nil
		return false
	}
}

func (i *fieldSelectionIterator) Selection() *ast.Field {
	return i.current.AstField
}

func (i *fieldSelectionIterator) SchemaField() *schema.FieldDescriptor {
	return i.current.Field
}

func (i *fieldSelectionIterator) ChildFieldsIterator() schema.FieldSelectionIterator {
	return newSelectionIterator(i.current.Sel)
}
