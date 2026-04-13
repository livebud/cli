package cli

import (
	"net/url"
	"strings"
	"time"
)

type Flag struct {
	name  string
	help  string
	short string
	env   *string
	value value
}

func (f *Flag) key() string {
	return "--" + f.name
}

// Short allows you to specify a short name for the flag.
func (f *Flag) Short(short byte) *Flag {
	f.short = string(short)
	return f
}

// Env allows you to use an environment variable to set the value of the flag.
func (f *Flag) Env(name string) *Flag {
	name = strings.TrimPrefix(name, "$")
	f.env = &name
	return f
}

func (f *Flag) Optional() *OptionalFlag {
	return &OptionalFlag{f}
}

func (f *Flag) Int(target *int) *Int {
	value := &Int{target, f.env, nil}
	f.value = &intValue{key: f.key(), inner: value}
	return value
}

func (f *Flag) Duration(target *time.Duration) *Duration {
	value := &Duration{target, f.env, nil}
	f.value = &durationValue{key: f.key(), inner: value}
	return value
}

func (f *Flag) Url(target *url.URL) *Url {
	value := &Url{target, f.env, nil}
	f.value = &urlValue{key: f.key(), inner: value}
	return value
}

func (f *Flag) String(target *string) *String {
	value := &String{target, f.env, nil}
	f.value = &stringValue{key: f.key(), inner: value}
	return value
}

func (f *Flag) Strings(target *[]string) *Strings {
	*target = []string{}
	value := &Strings{target, f.env, nil, false}
	f.value = &stringsValue{key: f.key(), inner: value}
	return value
}

func (f *Flag) Urls(target *[]*url.URL) *Urls {
	*target = []*url.URL{}
	value := &Urls{target, f.env, nil, false}
	f.value = &urlsValue{key: f.key(), inner: value}
	return value
}

func (f *Flag) Enums(target *[]string, possibilities ...string) *Enums {
	*target = []string{}
	value := &Enums{target: target, envvar: f.env, possibilities: possibilities}
	f.value = &enumsValue{key: f.key(), inner: value}
	return value
}

func (f *Flag) Durations(target *[]time.Duration) *Durations {
	*target = []time.Duration{}
	value := &Durations{target, f.env, nil, false}
	f.value = &durationsValue{key: f.key(), inner: value}
	return value
}

func (f *Flag) Enum(target *string, possibilities ...string) *Enum {
	value := &Enum{target, f.env, nil}
	f.value = &enumValue{key: f.key(), inner: value, possibilities: possibilities}
	return value
}

func (f *Flag) StringMap(target *map[string]string) *StringMap {
	*target = map[string]string{}
	value := &StringMap{target, f.env, nil, false}
	f.value = &stringMapValue{key: f.key(), inner: value}
	return value
}

func (f *Flag) Int64(target *int64) *Int64 {
	value := &Int64{target, f.env, nil}
	f.value = &int64Value{key: f.key(), inner: value}
	return value
}

func (f *Flag) Float32(target *float32) *Float32 {
	value := &Float32{target, f.env, nil}
	f.value = &float32Value{key: f.key(), inner: value}
	return value
}

func (f *Flag) Float64(target *float64) *Float64 {
	value := &Float64{target, f.env, nil}
	f.value = &float64Value{key: f.key(), inner: value}
	return value
}

func (f *Flag) Bool(target *bool) *Bool {
	value := &Bool{target, f.env, nil}
	f.value = &boolValue{key: f.key(), inner: value}
	return value
}

func (f *Flag) verify(name string) error {
	return f.value.verify()
}

type OptionalFlag struct {
	f *Flag
}

func (f *OptionalFlag) key() string {
	return "--" + f.f.name
}

func (f *OptionalFlag) String(target **string) *OptionalString {
	value := &OptionalString{target, f.f.env, nil}
	f.f.value = &optionalStringValue{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) Int(target **int) *OptionalInt {
	value := &OptionalInt{target, f.f.env, nil}
	f.f.value = &optionalIntValue{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) Duration(target **time.Duration) *OptionalDuration {
	value := &OptionalDuration{target, f.f.env, nil}
	f.f.value = &optionalDurationValue{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) Url(target **url.URL) *OptionalUrl {
	value := &OptionalUrl{target, f.f.env, nil}
	f.f.value = &optionalUrlValue{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) Int64(target **int64) *OptionalInt64 {
	value := &OptionalInt64{target, f.f.env, nil}
	f.f.value = &optionalInt64Value{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) Float32(target **float32) *OptionalFloat32 {
	value := &OptionalFloat32{target, f.f.env, nil}
	f.f.value = &optionalFloat32Value{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) Float64(target **float64) *OptionalFloat64 {
	value := &OptionalFloat64{target, f.f.env, nil}
	f.f.value = &optionalFloat64Value{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) Bool(target **bool) *OptionalBool {
	value := &OptionalBool{target, f.f.env, nil}
	f.f.value = &optionalBoolValue{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) Strings(target *[]string) *Strings {
	value := &Strings{target, f.f.env, nil, true}
	f.f.value = &stringsValue{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) Urls(target *[]*url.URL) *Urls {
	value := &Urls{target, f.f.env, nil, true}
	f.f.value = &urlsValue{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) Enums(target *[]string, possibilities ...string) *Enums {
	value := &Enums{target: target, envvar: f.f.env, possibilities: possibilities, optional: true}
	f.f.value = &enumsValue{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) Durations(target *[]time.Duration) *Durations {
	*target = []time.Duration{}
	value := &Durations{target, f.f.env, nil, true}
	f.f.value = &durationsValue{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) StringMap(target *map[string]string) *StringMap {
	value := &StringMap{target, f.f.env, nil, true}
	f.f.value = &stringMapValue{key: f.key(), inner: value}
	return value
}

func (f *OptionalFlag) Enum(target **string, possibilities ...string) *OptionalEnum {
	value := &OptionalEnum{target, f.f.env, nil}
	f.f.value = &optionalEnumValue{key: f.key(), inner: value, possibilities: possibilities}
	return value
}

func verifyFlags(flags []*Flag) error {
	for _, flag := range flags {
		if err := flag.verify(flag.name); err != nil {
			return err
		}
	}
	return nil
}
