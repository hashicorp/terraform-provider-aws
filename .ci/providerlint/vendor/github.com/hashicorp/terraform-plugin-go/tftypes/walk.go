package tftypes

// these functions are based heavily on github.com/zclconf/go-cty
// used under the MIT License
//
// Copyright (c) 2017-2018 Martin Atkins
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import (
	"errors"
)

// stopWalkError is a well-known error for immediately stopping walk() without
// returning an actual error.
//
// The implementation of walk() will continue walking all attributes/elements
// within an object/collection since the boolean return value of the callback
// function is only intended to signal whether to stop descending into the same
// Value. Changing that behavior would be considered a breaking change.
//
// This could be considered for exporting to give external consumers better
// performance.
var stopWalkError = errors.New("walk stop requested")

// Walk traverses a Value, calling the passed function for every element and
// attribute in the Value. The AttributePath passed to the callback function
// will identify which attribute or element is currently being surfaced by the
// Walk, and the passed Value will be the element or attribute at that
// AttributePath. Returning true from the callback function will indicate that
// any attributes or elements of the surfaced Value should be walked, too;
// returning false short-circuits the walk at that element or attribute, and
// does not visit any of its descendants. The return value of the callback does
// not matter when the Value that has been surfaced has no elements or
// attributes. Walk uses a depth-first traversal.
func Walk(val Value, cb func(*AttributePath, Value) (bool, error)) error {
	_, err := walk(NewAttributePath(), val, cb)

	return err
}

// walk is the internal implementation of Walk(). It includes a bool return for
// whether callers should continue walking any remaining Value.
func walk(path *AttributePath, val Value, cb func(*AttributePath, Value) (bool, error)) (bool, error) {
	shouldContinue, err := cb(path, val)

	if errors.Is(err, stopWalkError) {
		return false, nil
	}

	if err != nil {
		return false, path.NewError(err)
	}

	if !shouldContinue {
		// The callback bool return is intended to signal that this Value should
		// no longer be descended. Changing this behavior is a breaking change.
		// A stopWalkError can be used to signal that all remaining Value can be
		// skipped.
		return true, nil
	}

	if val.IsNull() || !val.IsKnown() {
		return true, nil
	}

	switch val.Type().(type) {
	case List, Tuple:
		v, ok := val.value.([]Value)

		if !ok {
			return false, path.NewErrorf("cannot convert %T into []tftypes.Value", val.value)
		}

		for pos, el := range v {
			elementPath := path.WithElementKeyInt(pos)
			shouldContinue, err := walk(elementPath, el, cb)

			if err != nil {
				return false, elementPath.NewError(err)
			}

			if !shouldContinue {
				return false, nil
			}
		}
	case Map:
		v, ok := val.value.(map[string]Value)

		if !ok {
			return false, path.NewErrorf("cannot convert %T into map[string]tftypes.Value", val.value)
		}

		for k, el := range v {
			elementPath := path.WithElementKeyString(k)
			shouldContinue, err := walk(elementPath, el, cb)

			if err != nil {
				return false, elementPath.NewError(err)
			}

			if !shouldContinue {
				return false, nil
			}
		}
	case Object:
		v, ok := val.value.(map[string]Value)

		if !ok {
			return false, path.NewErrorf("cannot convert %T into map[string]tftypes.Value", val.value)
		}

		for k, el := range v {
			attributePath := path.WithAttributeName(k)
			shouldContinue, err := walk(attributePath, el, cb)

			if err != nil {
				return false, attributePath.NewError(err)
			}

			if !shouldContinue {
				return false, nil
			}
		}
	case Set:
		v, ok := val.value.([]Value)

		if !ok {
			return false, path.NewErrorf("cannot convert %T into []tftypes.Value", val.value)
		}

		for _, el := range v {
			elementPath := path.WithElementKeyValue(el)
			shouldContinue, err := walk(elementPath, el, cb)

			if err != nil {
				return false, elementPath.NewError(err)
			}

			if !shouldContinue {
				return false, nil
			}
		}
	}

	return true, nil
}

// Transform uses a callback to mutate a Value. Each element or attribute will
// be visited in turn, with the AttributePath and Value surfaced to the
// callback, as in Walk. Unlike in Walk, the callback returns a Value instead
// of a boolean; this is the Value that will be stored at that AttributePath.
// The callback must return the passed Value unmodified if it wishes to not
// mutate a Value. Elements and attributes of a Value will be passed to the
// callback prior to the Value they belong to being passed to the callback,
// which means a callback can overwrite its own modifications. Values passed to
// the callback will always reflect the results of earlier callback calls.
func Transform(val Value, cb func(*AttributePath, Value) (Value, error)) (Value, error) {
	return transform(NewAttributePath(), val, cb)
}

func transform(path *AttributePath, val Value, cb func(*AttributePath, Value) (Value, error)) (Value, error) {
	switch val.Type().(type) {
	case nil:
		return val, path.NewError(errors.New("invalid transform: value missing type"))
	}

	newVal, err := transformUnderlying(path, val, cb)

	if err != nil {
		return val, err
	}

	res, err := cb(path, newVal)

	if err != nil {
		return res, path.NewError(err)
	}

	newTy := newVal.Type()

	if newTy == nil {
		return val, path.NewError(errors.New("invalid transform: new value missing type"))
	}

	if !newTy.UsableAs(val.Type()) {
		return val, path.NewError(errors.New("invalid transform: value changed type"))
	}

	return res, err
}

// transformUnderlying returns the Value with any underlying attribute or
// element transformations completed.
func transformUnderlying(path *AttributePath, val Value, cb func(*AttributePath, Value) (Value, error)) (Value, error) {
	// If the Value is null or unknown, there is nothing to descend.
	if val.IsNull() || !val.IsKnown() {
		return val, nil
	}

	switch val.Type().(type) {
	case List, Tuple:
		elements, ok := val.value.([]Value)

		if !ok {
			return val, path.NewErrorf("cannot convert %T into []tftypes.Value", val.value)
		}

		if len(elements) == 0 {
			return val, nil
		}

		newElements := make([]Value, 0, len(elements))

		for index, element := range elements {
			elementPath := path.WithElementKeyInt(index)

			newElement, err := transform(elementPath, element, cb)

			if err != nil {
				return val, elementPath.NewError(err)
			}

			newElements = append(newElements, newElement)
		}

		newVal, err := newValue(val.Type(), newElements)

		if err != nil {
			return val, path.NewError(err)
		}

		return newVal, nil
	case Map:
		elements, ok := val.value.(map[string]Value)

		if !ok {
			return val, path.NewErrorf("cannot convert %T into map[string]tftypes.Value", val.value)
		}

		if len(elements) == 0 {
			return val, nil
		}

		newElements := make(map[string]Value, len(elements))

		for key, element := range elements {
			elementPath := path.WithElementKeyString(key)

			newElement, err := transform(elementPath, element, cb)

			if err != nil {
				return val, elementPath.NewError(err)
			}

			newElements[key] = newElement
		}

		newVal, err := newValue(val.Type(), newElements)

		if err != nil {
			return val, path.NewError(err)
		}

		return newVal, nil
	case Object:
		attributes, ok := val.value.(map[string]Value)

		if !ok {
			return val, path.NewErrorf("cannot convert %T into map[string]tftypes.Value", val.value)
		}

		if len(attributes) == 0 {
			return val, nil
		}

		newAttributes := make(map[string]Value, len(attributes))

		for name, attribute := range attributes {
			attributePath := path.WithAttributeName(name)

			newAttribute, err := transform(attributePath, attribute, cb)

			if err != nil {
				return val, attributePath.NewError(err)
			}

			newAttributes[name] = newAttribute
		}

		newVal, err := newValue(val.Type(), newAttributes)

		if err != nil {
			return val, path.NewError(err)
		}

		return newVal, nil
	case Set:
		elements, ok := val.value.([]Value)

		if !ok {
			return val, path.NewErrorf("cannot convert %T into []tftypes.Value", val.value)
		}

		if len(elements) == 0 {
			return val, nil
		}

		newElements := make([]Value, 0, len(elements))

		for _, element := range elements {
			elementPath := path.WithElementKeyValue(element)

			newElement, err := transform(elementPath, element, cb)

			if err != nil {
				return val, elementPath.NewError(err)
			}

			newElements = append(newElements, newElement)
		}

		newVal, err := newValue(val.Type(), newElements)

		if err != nil {
			return val, path.NewError(err)
		}

		return newVal, nil
	}

	return val, nil
}
