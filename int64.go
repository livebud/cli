package cli

import (
	"fmt"
	"strconv"
	"strings"
)

type Int64 struct {
	target *int64
	envvar *string
	defval *int64
}

func (v *Int64) Default(value int64) {
	v.defval = &value
}

type int64Value struct {
	key   string
	inner *Int64
	set   bool
}

var _ value = (*int64Value)(nil)

func (v *int64Value) optional() bool {
	return false
}

func (v *int64Value) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *int64Value) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return strconv.FormatInt(*v.inner.defval, 10), true
}

func (v *int64Value) verify() error {
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

func (v *int64Value) Set(val string) error {
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf("%s: expected an int64 but got %q", v.key, val)
	}
	*v.inner.target = n
	v.set = true
	return nil
}

func (v *int64Value) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strconv.FormatInt(*v.inner.target, 10)
	} else if v.hasDefault() {
		return strconv.FormatInt(*v.inner.defval, 10)
	}
	return ""
}

type OptionalInt64 struct {
	target **int64
	envvar *string
	defval *int64
}

func (v *OptionalInt64) Default(value int64) {
	v.defval = &value
}

type optionalInt64Value struct {
	key   string
	inner *OptionalInt64
	set   bool
}

var _ value = (*optionalInt64Value)(nil)

func (v *optionalInt64Value) optional() bool {
	return true
}

func (v *optionalInt64Value) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *optionalInt64Value) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return strconv.FormatInt(*v.inner.defval, 10), true
}

func (v *optionalInt64Value) verify() error {
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

func (v *optionalInt64Value) Set(val string) error {
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf("%s: expected an int64 but got %q", v.key, val)
	}
	*v.inner.target = &n
	v.set = true
	return nil
}

func (v *optionalInt64Value) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strconv.FormatInt(**v.inner.target, 10)
	} else if v.hasDefault() {
		return strconv.FormatInt(*v.inner.defval, 10)
	}
	return ""
}

type Int64s struct {
	target   *[]int64
	envvar   *string
	defval   *[]int64
	optional bool
}

func (v *Int64s) Default(values ...int64) {
	v.defval = &values
}

type int64sValue struct {
	key   string
	inner *Int64s
	set   bool
}

var _ value = (*int64sValue)(nil)

func (v *int64sValue) optional() bool {
	return v.inner.optional
}

func (v *int64sValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *int64sValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	if len(*v.inner.defval) == 0 {
		return "[]", true
	}
	strs := make([]string, len(*v.inner.defval))
	for i, n := range *v.inner.defval {
		strs[i] = strconv.FormatInt(n, 10)
	}
	return strings.Join(strs, ", "), true
}

func (v *int64sValue) verify() error {
	if v.set {
		return nil
	} else if value, ok := lookupEnv(v.inner.envvar); ok {
		return v.Set(value)
	} else if v.hasDefault() {
		*v.inner.target = *v.inner.defval
		return nil
	} else if v.inner.optional {
		return nil
	}
	return &missingInputError{v.key, v.inner.envvar}
}

func (v *int64sValue) Set(val string) error {
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return fmt.Errorf("%s: expected an int64 but got %q", v.key, val)
	}
	*v.inner.target = append(*v.inner.target, n)
	v.set = true
	return nil
}

func (v *int64sValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		strs := make([]string, len(*v.inner.target))
		for i, n := range *v.inner.target {
			strs[i] = strconv.FormatInt(n, 10)
		}
		return strings.Join(strs, ", ")
	} else if v.hasDefault() {
		strs := make([]string, len(*v.inner.defval))
		for i, n := range *v.inner.defval {
			strs[i] = strconv.FormatInt(n, 10)
		}
		return strings.Join(strs, ", ")
	}
	return ""
}
