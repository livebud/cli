package cli

import (
	"fmt"
	"strings"
)

type StringMap struct {
	target *map[string]string
	defval *map[string]string // default value
}

func (v *StringMap) Default(value map[string]string) {
	v.defval = &value
}

func (v *StringMap) Optional() {
	v.defval = new(map[string]string)
}

type stringMapValue struct {
	key   string
	inner *StringMap
	set   bool
}

var _ value = (*stringMapValue)(nil)

func (v *stringMapValue) optional() bool {
	return false
}

func (v *stringMapValue) verify() error {
	if v.set {
		return nil
	} else if v.hasDefault() {
		*v.inner.target = *v.inner.defval
		return nil
	}
	return fmt.Errorf("missing %s", v.key)
}

func (v *stringMapValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *stringMapValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	str := new(strings.Builder)
	i := 0
	for k, v := range *v.inner.defval {
		if i > 0 {
			str.WriteString(" ")
		}
		str.WriteString(k + ":" + v)
		i++
	}
	return str.String(), true
}

func (v *stringMapValue) Set(val string) error {
	kv := strings.SplitN(val, ":", 2)
	if len(kv) != 2 {
		return fmt.Errorf("%s: invalid key:value pair for %q", v.key, val)
	}
	if *v.inner.target == nil {
		*v.inner.target = map[string]string{}
	}
	(*v.inner.target)[kv[0]] = kv[1]
	v.set = true
	return nil
}

func (v *stringMapValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return v.format(*v.inner.target)
	} else if v.hasDefault() {
		return v.format(*v.inner.defval)
	}
	return ""
}

// Format as a string
func (v *stringMapValue) format(kv map[string]string) (out string) {
	i := 0
	for k, v := range kv {
		if i > 0 {
			out += " "
		}
		out += k + ":" + v
		i++
	}
	return out
}
