// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package depgraph

type stack struct {
	top    *stackNode
	length int
}

type stackNode struct {
	value interface{}
	prev  *stackNode
}

// newStack returns a new, empty stack.
func newStack() *stack {
	return &stack{}
}

// len returns the stack's depth.
func (s *stack) len() int {
	return s.length
}

// peek returns the top item on the stack.
func (s *stack) peek() interface{} {
	if s.length == 0 {
		return nil
	}

	return s.top.value
}

// pop returns the top item on the stack and removes it from the stack.
func (s *stack) pop() interface{} {
	if s.length == 0 {
		return nil
	}

	top := s.top
	s.top = top.prev
	s.length--

	return top.value
}

// push puts the specified item on the top of the stack.
func (s *stack) push(value interface{}) {
	node := &stackNode{
		value: value,
		prev:  s.top,
	}
	s.top = node
	s.length++
}
