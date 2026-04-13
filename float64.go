package cli

import (
	"fmt"
	"strconv"
	"strings"
)

type Float64 struct {
	target *float64
	envvar *string
	defval *float64
}

func (v *Float64) Default(value float64) {
	v.defval = &value
}

type float64Value struct {
	key   string
	inner *Float64
	set   bool
}

var _ value = (*float64Value)(nil)

func (v *float64Value) optional() bool {
	return false
}

func (v *float64Value) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *float64Value) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return strconv.FormatFloat(*v.inner.defval, 'f', -1, 64), true
}

func (v *float64Value) verify() error {
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

func (v *float64Value) Set(val string) error {
	n, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return fmt.Errorf("%s: expected a float64 but got %q", v.key, val)
	}
	*v.inner.target = n
	v.set = true
	return nil
}

func (v *float64Value) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strconv.FormatFloat(*v.inner.target, 'f', -1, 64)
	} else if v.hasDefault() {
		return strconv.FormatFloat(*v.inner.defval, 'f', -1, 64)
	}
	return ""
}

type OptionalFloat64 struct {
	target **float64
	envvar *string
	defval *float64
}

func (v *OptionalFloat64) Default(value float64) {
	v.defval = &value
}

type optionalFloat64Value struct {
	key   string
	inner *OptionalFloat64
	set   bool
}

var _ value = (*optionalFloat64Value)(nil)

func (v *optionalFloat64Value) optional() bool {
	return true
}

func (v *optionalFloat64Value) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *optionalFloat64Value) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return strconv.FormatFloat(*v.inner.defval, 'f', -1, 64), true
}

func (v *optionalFloat64Value) verify() error {
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

func (v *optionalFloat64Value) Set(val string) error {
	n, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return fmt.Errorf("%s: expected a float64 but got %q", v.key, val)
	}
	*v.inner.target = &n
	v.set = true
	return nil
}

func (v *optionalFloat64Value) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strconv.FormatFloat(**v.inner.target, 'f', -1, 64)
	} else if v.hasDefault() {
		return strconv.FormatFloat(*v.inner.defval, 'f', -1, 64)
	}
	return ""
}

type Float64s struct {
	target   *[]float64
	envvar   *string
	defval   *[]float64
	optional bool
}

func (v *Float64s) Default(values ...float64) {
	v.defval = &values
}

type float64sValue struct {
	key   string
	inner *Float64s
	set   bool
}

var _ value = (*float64sValue)(nil)

func (v *float64sValue) optional() bool {
	return v.inner.optional
}

func (v *float64sValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *float64sValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	if len(*v.inner.defval) == 0 {
		return "[]", true
	}
	strs := make([]string, len(*v.inner.defval))
	for i, f := range *v.inner.defval {
		strs[i] = strconv.FormatFloat(f, 'f', -1, 64)
	}
	return strings.Join(strs, ", "), true
}

func (v *float64sValue) verify() error {
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

func (v *float64sValue) Set(val string) error {
	n, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return fmt.Errorf("%s: expected a float64 but got %q", v.key, val)
	}
	*v.inner.target = append(*v.inner.target, n)
	v.set = true
	return nil
}

func (v *float64sValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		strs := make([]string, len(*v.inner.target))
		for i, f := range *v.inner.target {
			strs[i] = strconv.FormatFloat(f, 'f', -1, 64)
		}
		return strings.Join(strs, ", ")
	} else if v.hasDefault() {
		strs := make([]string, len(*v.inner.defval))
		for i, f := range *v.inner.defval {
			strs[i] = strconv.FormatFloat(f, 'f', -1, 64)
		}
		return strings.Join(strs, ", ")
	}
	return ""
}
