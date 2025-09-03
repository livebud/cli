package cli

import (
	"fmt"
	"strconv"
)

type Bool struct {
	target *bool
	envvar *string
	defval *bool // default value
}

func (v *Bool) Default(value bool) {
	v.defval = &value
}

type boolValue struct {
	key   string
	inner *Bool
	set   bool
}

var _ value = (*boolValue)(nil)

func (v *boolValue) optional() bool {
	return false
}

func (v *boolValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *boolValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return strconv.FormatBool(*v.inner.defval), true
}

func (v *boolValue) verify() error {
	if v.set {
		return nil
	} else if value, ok := lookupEnv(v.inner.envvar); ok {
		return v.Set(value)
	} else if v.hasDefault() {
		*v.inner.target = *v.inner.defval
		return nil
	}
	return &missingInputError{v.key, v.inner.envvar}
}

func (v *boolValue) Set(val string) (err error) {
	*v.inner.target, err = strconv.ParseBool(val)
	if err != nil {
		return fmt.Errorf("%s: expected a boolean but got %q", v.key, val)
	}
	v.set = true
	return nil
}

func (v *boolValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strconv.FormatBool(*v.inner.target)
	} else if v.hasDefault() {
		return strconv.FormatBool(*v.inner.defval)
	}
	return "false"
}

// IsBoolFlag allows --flag to be an alias for --flag true
func (v *boolValue) IsBoolFlag() bool {
	return true
}

type OptionalBool struct {
	target **bool
	envvar *string
	defval *bool // default value
}

func (v *OptionalBool) Default(value bool) {
	v.defval = &value
}

type optionalBoolValue struct {
	key   string
	inner *OptionalBool
	set   bool
}

var _ value = (*optionalBoolValue)(nil)

// IsBoolFlag allows --flag to be an alias for --flag true
func (v *optionalBoolValue) IsBoolFlag() bool {
	return true
}

func (v *optionalBoolValue) optional() bool {
	return true
}

func (v *optionalBoolValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return strconv.FormatBool(*v.inner.defval), true
}

func (v *optionalBoolValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *optionalBoolValue) verify() error {
	if v.set {
		return nil
	} else if value, ok := lookupEnv(v.inner.envvar); ok {
		return v.Set(value)
	} else if v.hasDefault() {
		*v.inner.target = v.inner.defval
		return nil
	}
	return nil
}

func (v *optionalBoolValue) Set(val string) error {
	b, err := strconv.ParseBool(val)
	if err != nil {
		return fmt.Errorf("%s: expected a boolean but got %q", v.key, val)
	}
	*v.inner.target = &b
	v.set = true
	return nil
}

func (v *optionalBoolValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strconv.FormatBool(**v.inner.target)
	} else if v.hasDefault() {
		return strconv.FormatBool(*v.inner.defval)
	}
	return ""
}
