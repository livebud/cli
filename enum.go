package cli

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type Enum struct {
	target *string
	envvar *string
	defval *string // default value
}

func (v *Enum) Default(value string) {
	v.defval = &value
}

func (v *Enum) Env(name string) {
	v.envvar = &name
}

type enumValue struct {
	key           string
	inner         *Enum
	set           bool
	possibilities []string
}

var _ value = (*enumValue)(nil)

func (v *enumValue) optional() bool {
	return false
}

func (v *enumValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *enumValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return *v.inner.defval, true
}

func (v *enumValue) verify() error {
	if v.set {
		return nil
	} else if value, ok := lookupEnv(v.inner.envvar); ok {
		return v.Set(value)
	} else if v.hasDefault() {
		if err := verifyEnum(v.key, *v.inner.defval, v.possibilities...); err != nil {
			return err
		}
		*v.inner.target = *v.inner.defval
		return nil
	}
	return &missingInputError{v.key, v.inner.envvar}
}

func (v *enumValue) Set(val string) error {
	if err := verifyEnum(v.key, val, v.possibilities...); err != nil {
		return err
	}
	*v.inner.target = val
	v.set = true
	return nil
}

func (v *enumValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return *v.inner.target
	} else if v.hasDefault() {
		return *v.inner.defval
	}
	return ""
}

type OptionalEnum struct {
	target **string
	envvar *string
	defval *string // default value
}

func (v *OptionalEnum) Default(value string) {
	v.defval = &value
}

func (v *OptionalEnum) Env(name string) {
	v.envvar = &name
}

type optionalEnumValue struct {
	key           string
	inner         *OptionalEnum
	set           bool
	possibilities []string
}

var _ value = (*optionalEnumValue)(nil)

func (v *optionalEnumValue) optional() bool {
	return true
}

func (v *optionalEnumValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *optionalEnumValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return *v.inner.defval, true
}

func (v *optionalEnumValue) verify() error {
	if v.set {
		return nil
	} else if value, ok := lookupEnv(v.inner.envvar); ok {
		return v.Set(value)
	} else if v.hasDefault() {
		if err := verifyEnum(v.key, *v.inner.defval, v.possibilities...); err != nil {
			return err
		}
		*v.inner.target = v.inner.defval
		return nil
	}
	return nil
}

func (v *optionalEnumValue) Set(val string) error {
	if err := verifyEnum(v.key, val, v.possibilities...); err != nil {
		return err
	}
	*v.inner.target = &val
	v.set = true
	return nil
}

func (v *optionalEnumValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return **v.inner.target
	} else if v.hasDefault() {
		return *v.inner.defval
	}
	return ""
}

func verifyEnum(key, val string, possibilities ...string) error {
	if slices.Contains(possibilities, val) {
		return nil
	}
	s := new(strings.Builder)
	lp := len(possibilities)
	for i, p := range possibilities {
		if i == lp-1 {
			s.WriteString(" or ")
		} else if i > 0 {
			s.WriteString(", ")
		}
		s.WriteString(strconv.Quote(p))
	}
	return fmt.Errorf("%s %q must be either %s", key, val, s.String())
}
