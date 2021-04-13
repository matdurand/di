package di

import (
	"fmt"
	"reflect"
)

// schema is a dependency injection schema.
type schema interface {
	// find finds reflect.Type with matching Tags.
	find(t reflect.Type, tags Tags) (*node, error)
	// register cleanup
	cleanup(cleanup func())
}

// schema is a dependency injection schema.
type defaultSchema struct {
	nodes    map[reflect.Type][]*node
	cleanups []func()
}

func (s *defaultSchema) cleanup(cleanup func()) {
	s.cleanups = append(s.cleanups, cleanup)
}

// newDefaultSchema creates new dependency injection schema.
func newDefaultSchema() *defaultSchema {
	return &defaultSchema{
		nodes: map[reflect.Type][]*node{},
	}
}

// register registers reflect.Type provide function with optional Tags. Also, its registers
// type []<type> for group.
func (s *defaultSchema) register(n *node) {
	defer tracer.Trace("Register %s", n)
	if _, ok := s.nodes[n.rt]; !ok {
		s.nodes[n.rt] = []*node{n}
		return
	}
	s.nodes[n.rt] = append(s.nodes[n.rt], n)
}

// used depth-first topological sort algorithm
func (s *defaultSchema) prepare(n *node) error {
	var marks = map[*node]int{}
	if err := visit(s, n, marks); err != nil {
		return err
	}
	return nil
}

// find finds provideFunc by its reflect.Type and Tags.
func (s *defaultSchema) find(t reflect.Type, tags Tags) (*node, error) {
	nodes, ok := s.nodes[t]
	// type found
	if ok {
		matched := matchTags(nodes, tags)
		if len(matched) == 0 {
			return nil, fmt.Errorf("type %s%s %w", t, tags, ErrTypeNotExists)
		}
		if len(matched) > 1 {
			return nil, fmt.Errorf("multiple definitions of %s%s, maybe you need to use group type: []%s%s", t, tags, t, tags)
		}
		return matched[0], nil
	}
	// if not a group and not have di.Inject
	if t.Kind() != reflect.Slice && !canInject(t) {
		return nil, fmt.Errorf("type %s%s %w", t, tags, ErrTypeNotExists)
	}
	if canInject(t) {
		// constructor result with di.Inject - only addressable pointers
		// anonymous parameters with di.Inject - only struct
		//if t.Kind() == reflect.Ptr {
		//	return nil, fmt.Errorf("inject %s%s %w, use %s%s", t, tags, errFieldsNotSupported, t.Elem(), tags)
		//}
		node := &node{
			compiler: newTypeCompiler(t),
			rt:       t,
			rv:       new(reflect.Value),
		}
		// save node for future use
		s.nodes[t] = append(s.nodes[t], node)
		return node, nil
	}
	return s.group(t, tags)
}

func (s *defaultSchema) group(t reflect.Type, tags Tags) (*node, error) {
	group, ok := s.nodes[t.Elem()]
	if !ok {
		return nil, fmt.Errorf("type %s%s %w", t, tags, ErrTypeNotExists)
	}
	matched := matchTags(group, tags)
	if len(matched) == 0 {
		return nil, fmt.Errorf("type %s%s %w", t, tags, ErrTypeNotExists)
	}
	node := &node{
		compiler: newGroupCompiler(t, matched),
		rt:       t,
		tags:     tags,
		rv:       new(reflect.Value),
	}
	return node, nil
}
