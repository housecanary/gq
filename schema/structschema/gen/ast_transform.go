package gen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ast/astutil"
)

type transformRule struct {
	matcher matcher
	action  func(c *astutil.Cursor)
}

type matcher func(n ast.Node) (bool, matcher)

func sequenceOf(head matcher, tail ...matcher) matcher {
	return func(n ast.Node) (bool, matcher) {
		if match, next := head(n); match {
			if next == nil {
				if len(tail) == 0 {
					return true, nil
				}
				next = tail[0]
				tail = tail[1:]
			}
			return true, sequenceOf(next, tail...)
		}

		return false, nil
	}
}

func oneOf(alts ...matcher) matcher {
	return func(n ast.Node) (bool, matcher) {
		for _, alt := range alts {
			if match, next := alt(n); match {
				return true, next
			}
		}
		return false, nil
	}
}

func matchAnyUntil(m matcher) matcher {
	return func(n ast.Node) (bool, matcher) {
		if match, next := m(n); match {
			return true, next
		}

		return true, matchAnyUntil(m)
	}
}

func match[T any](filter ...func(T) bool) matcher {
	return func(n ast.Node) (bool, matcher) {
		if t, ok := n.(T); ok {
			match := true
			for _, f := range filter {
				if !f(t) {
					match = false
					break
				}
			}
			if match {
				return true, nil
			}
		}

		return false, nil
	}
}

func matchPosition(pos token.Pos) matcher {
	return func(n ast.Node) (bool, matcher) {
		return n != nil && n.Pos() == pos, nil
	}
}

func matchField(ti *types.Info, pkg, name string, embedded bool) matcher {
	return match(func(n *ast.Field) bool {
		if embedded != (n.Names == nil) {
			return false
		}
		typ := ti.TypeOf(n.Type)
		fmt.Println(typ)
		if nt, ok := typ.(*types.Named); ok {
			return nt.Obj().Pkg().Path() == pkg && nt.Obj().Name() == name
		}
		return false
	})
}
func test(n ast.Node, m matcher) bool {
	stack := []matcher{m}
	matched := false

	astutil.Apply(n, func(c *astutil.Cursor) bool {
		if matched {
			return false
		}
		matches, next := stack[len(stack)-1](c.Node())
		if matches {
			stack = append(stack, next)
			if next == nil {
				matched = true
			}
			return true
		}
		return false
	}, func(c *astutil.Cursor) bool {
		stack = stack[:len(stack)-1]
		return !matched
	})
	return matched
}

func transform(n ast.Node, rules ...transformRule) {
	type matchState struct {
		currentMatcher matcher
		perform        bool
		action         func(c *astutil.Cursor)
	}

	var s0 []*matchState
	for _, r := range rules {
		s0 = append(s0, &matchState{
			currentMatcher: r.matcher,
			action:         r.action,
		})
	}

	stack := [][]*matchState{
		s0,
	}
	astutil.Apply(n, func(c *astutil.Cursor) bool {
		var nextStackEntry []*matchState

		needPost := false
		for _, ms := range stack[len(stack)-1] {
			if matches, next := ms.currentMatcher(c.Node()); matches {
				if next != nil {
					nextStackEntry = append(nextStackEntry, &matchState{
						currentMatcher: next,
						action:         ms.action,
					})
				} else {
					needPost = true
					ms.perform = true
				}
			}
		}

		if needPost || len(nextStackEntry) > 0 {
			stack = append(stack, nextStackEntry)
			return true
		}

		return false

	}, func(c *astutil.Cursor) bool {
		stack = stack[:len(stack)-1]
		for _, ms := range stack[len(stack)-1] {
			if ms.perform {
				ms.action(c)
			}
			ms.perform = false
		}
		return true
	})
}
