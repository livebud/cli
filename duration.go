package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/duration"
)

type Duration struct {
	target *time.Duration
	envvar *string
	defval *time.Duration
}

func (v *Duration) Default(value time.Duration) {
	v.defval = &value
}

type durationValue struct {
	key   string
	inner *Duration
	set   bool
}

var _ value = (*durationValue)(nil)

func (v *durationValue) optional() bool {
	return false
}

func (v *durationValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *durationValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return v.inner.defval.String(), true
}

func (v *durationValue) verify() error {
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

func (v *durationValue) Set(val string) error {
	d, err := duration.Parse(val)
	if err != nil {
		return fmt.Errorf("%s: expected a duration but got %q", v.key, val)
	}
	*v.inner.target = d
	v.set = true
	return nil
}

func (v *durationValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return v.inner.target.String()
	} else if v.hasDefault() {
		return v.inner.defval.String()
	}
	return ""
}

type OptionalDuration struct {
	target **time.Duration
	envvar *string
	defval *time.Duration
}

func (v *OptionalDuration) Default(value time.Duration) {
	v.defval = &value
}

type optionalDurationValue struct {
	key   string
	inner *OptionalDuration
	set   bool
}

var _ value = (*optionalDurationValue)(nil)

func (v *optionalDurationValue) optional() bool {
	return true
}

func (v *optionalDurationValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *optionalDurationValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return v.inner.defval.String(), true
}

func (v *optionalDurationValue) verify() error {
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

func (v *optionalDurationValue) Set(val string) error {
	d, err := time.ParseDuration(val)
	if err != nil {
		return fmt.Errorf("%s: expected a duration but got %q", v.key, val)
	}
	*v.inner.target = &d
	v.set = true
	return nil
}

func (v *optionalDurationValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return (*v.inner.target).String()
	} else if v.hasDefault() {
		return v.inner.defval.String()
	}
	return ""
}

type Durations struct {
	target   *[]time.Duration
	envvar   *string
	defval   *[]time.Duration
	optional bool
}

func (v *Durations) Default(values ...time.Duration) {
	v.defval = &values
}

type durationsValue struct {
	key   string
	inner *Durations
	set   bool
}

var _ value = (*durationsValue)(nil)

func (v *durationsValue) optional() bool {
	return v.inner.optional
}

func (v *durationsValue) verify() error {
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

func (v *durationsValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *durationsValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	if len(*v.inner.defval) == 0 {
		return "[]", true
	}
	strs := make([]string, len(*v.inner.defval))
	for i, d := range *v.inner.defval {
		strs[i] = d.String()
	}
	return strings.Join(strs, ", "), true
}

func (v *durationsValue) Set(val string) error {
	d, err := time.ParseDuration(val)
	if err != nil {
		return fmt.Errorf("%s: expected a duration but got %q", v.key, val)
	}
	*v.inner.target = append(*v.inner.target, d)
	v.set = true
	return nil
}

func (v *durationsValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		strs := make([]string, len(*v.inner.target))
		for i, d := range *v.inner.target {
			strs[i] = d.String()
		}
		return strings.Join(strs, ", ")
	} else if v.hasDefault() {
		strs := make([]string, len(*v.inner.defval))
		for i, d := range *v.inner.defval {
			strs[i] = d.String()
		}
		return strings.Join(strs, ", ")
	}
	return ""
}
