package cli

import (
	"fmt"
	"strconv"
	"strings"
)

type Float32 struct {
	target *float32
	envvar *string
	defval *float32
}

func (v *Float32) Default(value float32) {
	v.defval = &value
}

type float32Value struct {
	key   string
	inner *Float32
	set   bool
}

var _ value = (*float32Value)(nil)

func (v *float32Value) optional() bool {
	return false
}

func (v *float32Value) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *float32Value) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return strconv.FormatFloat(float64(*v.inner.defval), 'f', -1, 32), true
}

func (v *float32Value) verify() error {
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

func (v *float32Value) Set(val string) error {
	n, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return fmt.Errorf("%s: expected a float32 but got %q", v.key, val)
	}
	*v.inner.target = float32(n)
	v.set = true
	return nil
}

func (v *float32Value) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strconv.FormatFloat(float64(*v.inner.target), 'f', -1, 32)
	} else if v.hasDefault() {
		return strconv.FormatFloat(float64(*v.inner.defval), 'f', -1, 32)
	}
	return ""
}

type OptionalFloat32 struct {
	target **float32
	envvar *string
	defval *float32
}

func (v *OptionalFloat32) Default(value float32) {
	v.defval = &value
}

type optionalFloat32Value struct {
	key   string
	inner *OptionalFloat32
	set   bool
}

var _ value = (*optionalFloat32Value)(nil)

func (v *optionalFloat32Value) optional() bool {
	return true
}

func (v *optionalFloat32Value) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *optionalFloat32Value) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return strconv.FormatFloat(float64(*v.inner.defval), 'f', -1, 32), true
}

func (v *optionalFloat32Value) verify() error {
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

func (v *optionalFloat32Value) Set(val string) error {
	n, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return fmt.Errorf("%s: expected a float32 but got %q", v.key, val)
	}
	f := float32(n)
	*v.inner.target = &f
	v.set = true
	return nil
}

func (v *optionalFloat32Value) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strconv.FormatFloat(float64(**v.inner.target), 'f', -1, 32)
	} else if v.hasDefault() {
		return strconv.FormatFloat(float64(*v.inner.defval), 'f', -1, 32)
	}
	return ""
}

type Float32s struct {
	target   *[]float32
	envvar   *string
	defval   *[]float32
	optional bool
}

func (v *Float32s) Default(values ...float32) {
	v.defval = &values
}

type float32sValue struct {
	key   string
	inner *Float32s
	set   bool
}

var _ value = (*float32sValue)(nil)

func (v *float32sValue) optional() bool {
	return v.inner.optional
}

func (v *float32sValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *float32sValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	if len(*v.inner.defval) == 0 {
		return "[]", true
	}
	strs := make([]string, len(*v.inner.defval))
	for i, f := range *v.inner.defval {
		strs[i] = strconv.FormatFloat(float64(f), 'f', -1, 32)
	}
	return strings.Join(strs, ", "), true
}

func (v *float32sValue) verify() error {
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

func (v *float32sValue) Set(val string) error {
	n, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return fmt.Errorf("%s: expected a float32 but got %q", v.key, val)
	}
	*v.inner.target = append(*v.inner.target, float32(n))
	v.set = true
	return nil
}

func (v *float32sValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		strs := make([]string, len(*v.inner.target))
		for i, f := range *v.inner.target {
			strs[i] = strconv.FormatFloat(float64(f), 'f', -1, 32)
		}
		return strings.Join(strs, ", ")
	} else if v.hasDefault() {
		strs := make([]string, len(*v.inner.defval))
		for i, f := range *v.inner.defval {
			strs[i] = strconv.FormatFloat(float64(f), 'f', -1, 32)
		}
		return strings.Join(strs, ", ")
	}
	return ""
}
