// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package stack

import (
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
)

type stack[T any] struct {
	top    *stackNode[T]
	length int
}

type stackNode[T any] struct {
	value T
	prev  *stackNode[T]
}

// New returns a new, empty stack.
func New[T any]() *stack[T] {
	return &stack[T]{}
}

// Len returns the stack's depth.
func (s *stack[T]) Len() int {
	return s.length
}

// Peek returns the top item on the stack.
func (s *stack[T]) Peek() option.Option[T] {
	if s.length == 0 {
		return option.None[T]()
	}

	return option.Some(s.top.value)
}

// Pop returns the top item on the stack and removes it from the stack.
func (s *stack[T]) Pop() option.Option[T] {
	if s.length == 0 {
		return option.None[T]()
	}

	top := s.top
	s.top = top.prev
	s.length--

	return option.Some(top.value)
}

// Push puts the specified item on the top of the stack.
func (s *stack[T]) Push(value T) {
	node := &stackNode[T]{
		value: value,
		prev:  s.top,
	}
	s.top = node
	s.length++
}
