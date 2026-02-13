package cli

import (
	"fmt"
	"strings"

	"github.com/kballard/go-shellquote"
)

type Enums struct {
	target        *[]string
	envvar        *string
	defval        *[]string
	possibilities []string
	optional      bool
}

func (v *Enums) Default(values ...string) {
	v.defval = &values
}

type enumsValue struct {
	key   string
	inner *Enums
	set   bool
}

var _ value = (*enumsValue)(nil)

func (v *enumsValue) optional() bool {
	return v.inner.optional
}

func (v *enumsValue) verify() error {
	if v.set {
		return nil
	} else if value, ok := lookupEnv(v.inner.envvar); ok {
		fields, err := shellquote.Split(value)
		if err != nil {
			return fmt.Errorf("%s: expected a list of strings but got %q", v.key, value)
		}
		for _, val := range fields {
			if err := v.Set(val); err != nil {
				return err
			}
		}
		return nil
	} else if v.hasDefault() {
		for _, val := range *v.inner.defval {
			if err := verifyEnum(v.key, val, v.inner.possibilities...); err != nil {
				return err
			}
		}
		*v.inner.target = *v.inner.defval
		return nil
	} else if v.inner.optional {
		return nil
	}
	return &missingInputError{v.key, v.inner.envvar}
}

func (v *enumsValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *enumsValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	def := strings.Join(*v.inner.defval, ", ")
	if def == "" {
		return "[]", true
	}
	return def, true
}

func (v *enumsValue) Set(val string) error {
	if err := verifyEnum(v.key, val, v.inner.possibilities...); err != nil {
		return err
	}
	*v.inner.target = append(*v.inner.target, val)
	v.set = true
	return nil
}

func (v *enumsValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strings.Join(*v.inner.target, ", ")
	} else if v.hasDefault() {
		return strings.Join(*v.inner.defval, ", ")
	}
	return ""
}
