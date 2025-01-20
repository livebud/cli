package cli

import "fmt"

type String struct {
	target *string
	defval *string // default value
}

func (v *String) Default(value string) {
	v.defval = &value
}

type stringValue struct {
	key   string
	inner *String
	set   bool
}

func (v *stringValue) optional() bool {
	return false
}

var _ value = (*stringValue)(nil)

func (v *stringValue) verify() error {
	if v.set {
		return nil
	} else if v.hasDefault() {
		*v.inner.target = *v.inner.defval
		return nil
	}
	return fmt.Errorf("missing %s", v.key)
}

func (v *stringValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *stringValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return *v.inner.defval, true
}

func (v *stringValue) Set(val string) error {
	*v.inner.target = val
	v.set = true
	return nil
}

func (v *stringValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return *v.inner.target
	} else if v.hasDefault() {
		return *v.inner.defval
	}
	return ""
}

type OptionalString struct {
	target **string
	defval *string // default value
}

func (v *OptionalString) Default(value string) {
	v.defval = &value
}

type optionalStringValue struct {
	key   string
	inner *OptionalString
	set   bool
}

var _ value = (*optionalStringValue)(nil)

func (v *optionalStringValue) optional() bool {
	return true
}

func (v *optionalStringValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *optionalStringValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return *v.inner.defval, true
}

func (v *optionalStringValue) verify() error {
	if v.set {
		return nil
	} else if v.hasDefault() {
		*v.inner.target = v.inner.defval
		return nil
	}
	return nil
}

func (v *optionalStringValue) Set(val string) error {
	*v.inner.target = &val
	v.set = true
	return nil
}

func (v *optionalStringValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return **v.inner.target
	} else if v.hasDefault() {
		return *v.inner.defval
	}
	return ""
}
