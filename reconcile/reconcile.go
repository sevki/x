// Copyright 2018 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reconcile // import "sevki.org/x/reconcile"

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
)

// StateWalkFunc walks a state. Walk should be hierarchical to ensure
// no cascading updates occur.
type StateWalkFunc func(key string, v interface{})

// State represents the interface which Reconciler Accepts
type State interface {
	Add(key string, v interface{})
	Update(key string, v interface{})
	Get(key string) interface{}
	Delete(key string)
	Walk(f StateWalkFunc)
}

type state int

func (i state) String() string {
	switch i {
	case 0:
		return "Add"
	case 1:
		return "Delete"
	case 2:
		return "Update"
	}
	return ""
}

const (
	new state = iota
	old
	dirty
)

// Reconcile takes two states and applies updates to them until they are the same
func Reconcile(current, desired State, verbose bool) {
	fix(current, diff(current, desired), verbose)
}

type update struct {
	key   string
	state state
	v     interface{}
	why   string
}

// Checksumed returns a unique Hash for an object for comparison
type Checksumed interface {
	Sum() []byte
}

var (
	ErrHashMismatch      = errors.New("checksum mismatch")
	ErrDeepEqualMismatch = errors.New("reflect.Deepequal mismatch")
)

// compare takes two interfaces and returns true if they are the same
func compare(a, b interface{}) error {
	hashableA, aok := a.(Checksumed)
	hashableB, bok := b.(Checksumed)
	if aok && bok {
		desiredSum := hashableA.Sum()
		currentSum := hashableB.Sum()
		if bytes.Compare(desiredSum, currentSum) != 0 {
			return fmt.Errorf("%v sum=%x != %v sum=%x", a, desiredSum[:5], b, currentSum[:5])
		}
	} else {
		if !reflect.DeepEqual(a, b) {
			return ErrDeepEqualMismatch
		}
	}
	return nil
}

func diff(current, desired State) []update {
	var updates []update
	desired.Walk(func(key string, v interface{}) {
		n := update{
			key: key,
			v:   v,
		}
		currentValue := current.Get(key)
		if currentValue == nil {
			n.state = new
			n.why = fmt.Sprintf("current state doesn't have %s doesn't exist", key)
			updates = append(updates, n)
			return
		}
		if err := compare(currentValue, v); err != nil {
			n.state = dirty
			n.why = err.Error()
			updates = append(updates, n)
			return
		}
	})
	current.Walk(func(key string, v interface{}) {
		if desired.Get(key) == nil {
			n := update{
				key: key,
				v:   nil,
			}
			n.state = old
			n.why = fmt.Sprintf("currentValue is with key %s marked for deletion", key)
			updates = append(updates, n)
		}
	})

	return updates
}

func fix(current State, updates []update, verbose bool) {
	for _, update := range updates {
		if verbose {
			log.Printf("key:%s state:%s\n\twhy:%s\n ", update.key, update.state, update.why)
		}
		switch update.state {
		case new:
			current.Add(update.key, update.v)
		case old:
			current.Delete(update.key)
		case dirty:
			current.Update(update.key, update.v)
		}
	}
}
