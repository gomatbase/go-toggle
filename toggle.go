package toggle

import (
	"sync"

	"github.com/gomatbase/go-error"
	"github.com/gomatbase/go-log"
)

const (
	ToggleableExistsError   = err.Error("toggleable already exists")
	NotEnoughOptionsError   = err.Error("toggleables need at least 2 options to toggle")
	ToggleableNotFoundError = err.Error("toggleable does not exist")
	InvalidToggleError      = err.Error("toggle value is out of bounds for the toggleable")
)

var (
	toggleables = make(map[string]*toggleable)
	toggles     = make(map[string]int)
	mutex       = sync.Mutex{}
	l, _        = log.Get("TGL")
)

type toggleable struct {
	active  int
	options []func() error
}

func Run(name string, options ...func() error) error {
	if e := Add(name, options...); e != nil && e != ToggleableExistsError {
		return e
	}
	return Execute(name)
}

func Add(name string, options ...func() error) error {
	mutex.Lock()
	defer mutex.Unlock()

	if _, found := toggleables[name]; found {
		return ToggleableExistsError
	}

	if len(options) < 2 {
		return NotEnoughOptionsError
	}

	toggleable := toggleable{
		options: options,
	}
	if toggle, found := toggles[name]; found {
		if toggle < len(options) && toggle >= 0 {
			toggleable.active = toggle
		} else {
			l.Debug("Invalid toggle value for toggleable", name, ":", toggle)
		}
	}

	toggleables[name] = &toggleable
	return nil
}

func Execute(name string) error {
	mutex.Lock()
	toggleable, found := toggleables[name]
	mutex.Unlock()
	if !found {
		return ToggleableNotFoundError
	}

	return toggleable.options[toggleable.active]()
}

func Toggle(name string, toggle int) error {
	mutex.Lock()
	toggleable, found := toggleables[name]
	if !found {
		toggles[name] = toggle
		mutex.Unlock()
		return nil
	}
	mutex.Unlock()

	if toggle < len(toggleable.options) && toggle >= 0 {
		toggleable.active = toggle
	} else {
		return InvalidToggleError
	}

	return nil
}
