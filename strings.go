package cli

import (
	"fmt"
	"strings"

	"github.com/kballard/go-shellquote"
)

type Strings struct {
	target   *[]string
	envvar   *string
	defval   *[]string // default value
	optional bool
}

func (v *Strings) Default(values ...string) {
	v.defval = &values
}

type stringsValue struct {
	key   string
	inner *Strings
	set   bool
}

var _ value = (*stringsValue)(nil)

func (v *stringsValue) optional() bool {
	return v.inner.optional
}

func (v *stringsValue) verify() error {
	if v.set {
		return nil
	} else if value, ok := lookupEnv(v.inner.envvar); ok {
		fields, err := shellquote.Split(value)
		if err != nil {
			return fmt.Errorf("%s: expected a list of strings but got %q", v.key, value)
		}
		for _, kv := range fields {
			if err := v.Set(kv); err != nil {
				return err
			}
		}
		return nil
	} else if v.hasDefault() {
		*v.inner.target = *v.inner.defval
		return nil
	} else if v.inner.optional {
		return nil
	}
	return &missingError{v.key, v.inner.envvar}
}

func (v *stringsValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *stringsValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return strings.Join(*v.inner.defval, ", "), true
}

func (v *stringsValue) Set(val string) error {
	*v.inner.target = append(*v.inner.target, val)
	v.set = true
	return nil
}

func (v *stringsValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strings.Join(*v.inner.target, ", ")
	} else if v.hasDefault() {
		return strings.Join(*v.inner.defval, ", ")
	}
	return ""
}
