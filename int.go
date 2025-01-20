package cli

import (
	"fmt"
	"strconv"
)

type Int struct {
	target *int
	defval *int
}

func (v *Int) Default(value int) {
	v.defval = &value
}

type intValue struct {
	key   string
	inner *Int
	set   bool
}

var _ value = (*intValue)(nil)

func (v *intValue) optional() bool {
	return false
}

func (v *intValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *intValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return strconv.Itoa(*v.inner.defval), true
}

func (v *intValue) verify() error {
	if v.set {
		return nil
	} else if v.hasDefault() {
		*v.inner.target = *v.inner.defval
		return nil
	}
	return fmt.Errorf("missing %s", v.key)
}

func (v *intValue) Set(val string) error {
	n, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("%s: expected an integer but got %q", v.key, val)
	}
	*v.inner.target = n
	v.set = true
	return nil
}

func (v *intValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strconv.Itoa(*v.inner.target)
	} else if v.hasDefault() {
		return strconv.Itoa(*v.inner.defval)
	}
	return ""
}

type OptionalInt struct {
	target **int
	defval *int
}

func (v *OptionalInt) Default(value int) {
	v.defval = &value
}

type optionalIntValue struct {
	key   string
	inner *OptionalInt
	set   bool
}

var _ value = (*optionalIntValue)(nil)

func (v *optionalIntValue) optional() bool {
	return true
}

func (v *optionalIntValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *optionalIntValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return strconv.Itoa(*v.inner.defval), true
}

func (v *optionalIntValue) verify() error {
	if v.set {
		return nil
	} else if v.hasDefault() {
		*v.inner.target = v.inner.defval
		return nil
	}
	return nil
}

func (v *optionalIntValue) Set(val string) error {
	n, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("%s: expected an integer but got %q", v.key, val)
	}
	*v.inner.target = &n
	v.set = true
	return nil
}

func (v *optionalIntValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strconv.Itoa(**v.inner.target)
	} else if v.hasDefault() {
		return strconv.Itoa(*v.inner.defval)
	}
	return ""
}
