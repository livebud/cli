package cli

import (
	"net/url"
	"time"
)

type Arg struct {
	name  string
	help  string
	value value
	env   *string
}

func (a *Arg) key() string {
	return "<" + a.name + ">"
}

// Env allows you to use an environment variable to set the value of the argument.
func (a *Arg) Env(name string) *Arg {
	a.env = &name
	return a
}

func (a *Arg) Optional() *OptionalArg {
	return &OptionalArg{a}
}

func (a *Arg) Int(target *int) *Int {
	value := &Int{target, a.env, nil}
	a.value = &intValue{key: a.key(), inner: value}
	return value
}

func (a *Arg) Duration(target *time.Duration) *Duration {
	value := &Duration{target, a.env, nil}
	a.value = &durationValue{key: a.key(), inner: value}
	return value
}

func (a *Arg) Url(target *url.URL) *Url {
	value := &Url{target, a.env, nil}
	a.value = &urlValue{key: a.key(), inner: value}
	return value
}

func (a *Arg) Int64(target *int64) *Int64 {
	value := &Int64{target, a.env, nil}
	a.value = &int64Value{key: a.key(), inner: value}
	return value
}

func (a *Arg) Float32(target *float32) *Float32 {
	value := &Float32{target, a.env, nil}
	a.value = &float32Value{key: a.key(), inner: value}
	return value
}

func (a *Arg) Float64(target *float64) *Float64 {
	value := &Float64{target, a.env, nil}
	a.value = &float64Value{key: a.key(), inner: value}
	return value
}

func (a *Arg) Bool(target *bool) *Bool {
	value := &Bool{target, a.env, nil}
	a.value = &boolValue{key: a.key(), inner: value}
	return value
}

func (a *Arg) String(target *string) *String {
	value := &String{target, a.env, nil}
	a.value = &stringValue{key: a.key(), inner: value}
	return value
}

func (a *Arg) Enum(target *string, possibilities ...string) *Enum {
	value := &Enum{target, a.env, nil}
	a.value = &enumValue{key: a.key(), inner: value, possibilities: possibilities}
	return value
}

// StringMap accepts a key-value pair in the form of "<key:value>".
func (a *Arg) StringMap(target *map[string]string) *StringMap {
	*target = map[string]string{}
	value := &StringMap{target, a.env, nil, false}
	a.value = &stringMapValue{key: "<key:value>", inner: value}
	return value
}

func (a *Arg) verify() error {
	return a.value.verify()
}

type OptionalArg struct {
	a *Arg
}

func (a *OptionalArg) key() string {
	return "<" + a.a.name + ">"
}

func (a *OptionalArg) String(target **string) *OptionalString {
	value := &OptionalString{target, a.a.env, nil}
	a.a.value = &optionalStringValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArg) Int(target **int) *OptionalInt {
	value := &OptionalInt{target, a.a.env, nil}
	a.a.value = &optionalIntValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArg) Duration(target **time.Duration) *OptionalDuration {
	value := &OptionalDuration{target, a.a.env, nil}
	a.a.value = &optionalDurationValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArg) Url(target **url.URL) *OptionalUrl {
	value := &OptionalUrl{target, a.a.env, nil}
	a.a.value = &optionalUrlValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArg) Int64(target **int64) *OptionalInt64 {
	value := &OptionalInt64{target, a.a.env, nil}
	a.a.value = &optionalInt64Value{key: a.key(), inner: value}
	return value
}

func (a *OptionalArg) Float32(target **float32) *OptionalFloat32 {
	value := &OptionalFloat32{target, a.a.env, nil}
	a.a.value = &optionalFloat32Value{key: a.key(), inner: value}
	return value
}

func (a *OptionalArg) Float64(target **float64) *OptionalFloat64 {
	value := &OptionalFloat64{target, a.a.env, nil}
	a.a.value = &optionalFloat64Value{key: a.key(), inner: value}
	return value
}

func (a *OptionalArg) Bool(target **bool) *OptionalBool {
	value := &OptionalBool{target, a.a.env, nil}
	a.a.value = &optionalBoolValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArg) Enum(target **string, possibilities ...string) *OptionalEnum {
	value := &OptionalEnum{target, a.a.env, nil}
	a.a.value = &optionalEnumValue{key: a.key(), inner: value, possibilities: possibilities}
	return value
}

func (a *OptionalArg) StringMap(target *map[string]string) *StringMap {
	*target = map[string]string{}
	value := &StringMap{target, a.a.env, nil, true}
	a.a.value = &stringMapValue{key: a.key(), inner: value}
	return value
}

func verifyArgs(args []*Arg) error {
	for _, arg := range args {
		if err := arg.verify(); err != nil {
			return err
		}
	}
	return nil
}
