package cli

import (
	"net/url"
	"time"
)

type Args struct {
	name  string
	help  string
	value value
	env   *string
}

func (a *Args) key() string {
	return "<" + a.name + "...>"
}

func (a *Args) verify() error {
	return a.value.verify()
}

// Env allows you to use an environment variable to set the value of the argument.
func (a *Args) Env(name string) *Args {
	a.env = &name
	return a
}

func (a *Args) Optional() *OptionalArgs {
	return &OptionalArgs{a}
}

func (a *Args) Strings(target *[]string) *Strings {
	if target != nil {
		*target = []string{}
	}
	value := &Strings{target, a.env, nil, false}
	a.value = &stringsValue{key: a.key(), inner: value}
	return value
}

func (a *Args) Urls(target *[]*url.URL) *Urls {
	if target != nil {
		*target = []*url.URL{}
	}
	value := &Urls{target, a.env, nil, false}
	a.value = &urlsValue{key: a.key(), inner: value}
	return value
}

func (a *Args) Enums(target *[]string, possibilities ...string) *Enums {
	if target != nil {
		*target = []string{}
	}
	value := &Enums{target: target, envvar: a.env, possibilities: possibilities}
	a.value = &enumsValue{key: a.key(), inner: value}
	return value
}

func (a *Args) Durations(target *[]time.Duration) *Durations {
	if target != nil {
		*target = []time.Duration{}
	}
	value := &Durations{target, a.env, nil, false}
	a.value = &durationsValue{key: a.key(), inner: value}
	return value
}

func (a *Args) Int64s(target *[]int64) *Int64s {
	if target != nil {
		*target = []int64{}
	}
	value := &Int64s{target, a.env, nil, false}
	a.value = &int64sValue{key: a.key(), inner: value}
	return value
}

func (a *Args) Float32s(target *[]float32) *Float32s {
	if target != nil {
		*target = []float32{}
	}
	value := &Float32s{target, a.env, nil, false}
	a.value = &float32sValue{key: a.key(), inner: value}
	return value
}

func (a *Args) Float64s(target *[]float64) *Float64s {
	if target != nil {
		*target = []float64{}
	}
	value := &Float64s{target, a.env, nil, false}
	a.value = &float64sValue{key: a.key(), inner: value}
	return value
}

func (a *Args) StringMap(target *map[string]string) *StringMap {
	if target != nil {
		*target = map[string]string{}
	}
	value := &StringMap{target, a.env, nil, false}
	a.value = &stringMapValue{key: "<key:value...>", inner: value}
	return value
}

type OptionalArgs struct {
	a *Args
}

func (a *OptionalArgs) key() string {
	return "[<" + a.a.name + ">...]"
}

func (a *OptionalArgs) Strings(target *[]string) *Strings {
	value := &Strings{target, a.a.env, nil, true}
	a.a.value = &stringsValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArgs) Urls(target *[]*url.URL) *Urls {
	value := &Urls{target, a.a.env, nil, true}
	a.a.value = &urlsValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArgs) Enums(target *[]string, possibilities ...string) *Enums {
	value := &Enums{target: target, envvar: a.a.env, possibilities: possibilities, optional: true}
	a.a.value = &enumsValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArgs) Durations(target *[]time.Duration) *Durations {
	if target != nil {
		*target = []time.Duration{}
	}
	value := &Durations{target, a.a.env, nil, true}
	a.a.value = &durationsValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArgs) Int64s(target *[]int64) *Int64s {
	value := &Int64s{target, a.a.env, nil, true}
	a.a.value = &int64sValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArgs) Float32s(target *[]float32) *Float32s {
	value := &Float32s{target, a.a.env, nil, true}
	a.a.value = &float32sValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArgs) Float64s(target *[]float64) *Float64s {
	value := &Float64s{target, a.a.env, nil, true}
	a.a.value = &float64sValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArgs) StringMap(target *map[string]string) *StringMap {
	value := &StringMap{target, a.a.env, nil, true}
	a.a.value = &stringMapValue{key: a.key(), inner: value}
	return value
}
