package cli

import (
	"fmt"
)

type Custom struct {
	target func(s string) error
	defval *string // default value
}

func (v *Custom) Default(value string) {
	v.defval = &value
}

func (v *Custom) Optional() {
	v.defval = new(string)
}

type customValue struct {
	key   string
	inner *Custom
	set   bool
}

var _ value = (*customValue)(nil)

func (v *customValue) optional() bool {
	return false
}

func (v *customValue) verify() error {
	if v.set {
		return nil
	} else if v.hasDefault() {
		return v.inner.target(*v.inner.defval)
	}
	return fmt.Errorf("missing %s", v.key)
}

func (v *customValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *customValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return *v.inner.defval, true
}

func (v *customValue) Set(val string) error {
	err := v.inner.target(val)
	if err != nil {
		return fmt.Errorf("%s: invalid value %q: %w", v.key, val, err)
	}
	v.set = true
	return nil
}

func (v *customValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return ""
	} else if v.hasDefault() {
		return *v.inner.defval
	}
	return ""
}
