package cli

import (
	"fmt"
	"strings"
)

type Enum struct {
	target *string
	defval *string // default value
}

func (v *Enum) Default(value string) {
	v.defval = &value
}

type enumValue struct {
	inner         *Enum
	set           bool
	possibilities []string
}

func (v *enumValue) check(displayName, val string) error {
	for _, p := range v.possibilities {
		if p == val {
			return nil
		}
	}
	return fmt.Errorf("%s %q must be one of: %s", displayName, val, strings.Join(v.possibilities, ", "))
}

func (v *enumValue) verify(displayName string) error {
	if v.set {
		if err := v.check(displayName, *v.inner.target); err != nil {
			return err
		}
		return nil
	} else if v.inner.defval != nil {
		if err := v.check(displayName, *v.inner.defval); err != nil {
			return err
		}
		*v.inner.target = *v.inner.defval
		return nil
	}
	return fmt.Errorf("missing %s", displayName)
}

func (v *enumValue) Set(val string) error {
	*v.inner.target = val
	v.set = true
	return nil
}

func (v *enumValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return *v.inner.target
	} else if v.inner.defval != nil {
		return *v.inner.defval
	}
	return ""
}

type EnumValue struct {
	target **string
	defval *string // default value
}

func (v *EnumValue) Default(value string) {
	v.defval = &value
}

type optionalEnumValue struct {
	inner *EnumValue
	set   bool
}

var _ value = (*optionalEnumValue)(nil)

func (v *optionalEnumValue) verify(displayName string) error {
	if v.set {
		return nil
	} else if v.inner.defval != nil {
		*v.inner.target = v.inner.defval
		return nil
	}
	return nil
}

func (v *optionalEnumValue) Set(val string) error {
	*v.inner.target = &val
	v.set = true
	return nil
}

func (v *optionalEnumValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return **v.inner.target
	} else if v.inner.defval != nil {
		return *v.inner.defval
	}
	return ""
}
