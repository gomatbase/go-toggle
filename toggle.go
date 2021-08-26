package toggle

import (
	"strconv"
	"sync"

	"github.com/gomatbase/go-env"
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
	mutex.Lock()
	active := getActiveToggleFromEnvironment(name)
	mutex.Unlock()

	if active < 0 || active >= len(options) {
		l.Error("Trying to run toggleable %s with incorrect toggle value : %s. Defaulting to 0", name, active)
		active = 0
	}
	return options[active]()
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

	active := -1
	// we first check if there was already a call to set the toggleable at a specific value
	if toggle, found := toggles[name]; found {
		if toggle < len(options) && toggle >= 0 {
			active = toggle
		} else {
			l.Debug("Invalid toggle value for toggleable", name, ":", toggle)
			delete(toggles, name)
		}
	}
	// even if the toggle was set manually before it was setup it was invalid
	if active == -1 {
		active = getActiveToggleFromEnvironment(name)
	}

	toggleable := toggleable{
		options: options,
		active:  active,
	}

	toggleables[name] = &toggleable
	return nil
}

func getActiveToggleFromEnvironment(name string) int {

	if toggle, found := toggles[name]; found {
		return toggle
	}

	variableName := "toggleable." + name
	if e := env.Var(variableName).
		From(env.CmlArgumentsSource().Name("T" + name)).
		Default("0").
		Add(); e != nil {
		l.Debugf("Unable to register toggleable variable %s : %v", name, e)
		return 0
	}

	if v, e := strconv.Atoi(env.Get(variableName).(string)); e != nil {
		l.Debug("Invalid active toggleable value for toggleable", name)
		return 0
	} else {
		toggles[name] = v
		return v
	}
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
	defer mutex.Unlock()
	toggles[name] = toggle
	toggleable, found := toggleables[name]
	if found {
		if toggle < len(toggleable.options) && toggle >= 0 {
			toggleable.active = toggle
		} else {
			return InvalidToggleError
		}
	}

	return nil
}
