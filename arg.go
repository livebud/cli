package cli

type Arg struct {
	name  string
	help  string
	value value
}

func (a *Arg) key() string {
	return "<" + a.name + ">"
}

func (a *Arg) Optional() *OptionalArg {
	return &OptionalArg{a}
}

func (a *Arg) Int(target *int) *Int {
	value := &Int{target: target}
	a.value = &intValue{key: a.key(), inner: value}
	return value
}

func (a *Arg) String(target *string) *String {
	value := &String{target: target}
	a.value = &stringValue{key: a.key(), inner: value}
	return value
}

func (a *Arg) StringMap(target *map[string]string) *StringMap {
	*target = map[string]string{}
	value := &StringMap{target: target}
	a.value = &stringMapValue{key: a.key(), inner: value}
	return value
}

func (a *Arg) Enum(target *string, possibilities ...string) *Enum {
	value := &Enum{target: target}
	a.value = &enumValue{key: a.key(), inner: value, possibilities: possibilities}
	return value
}

// Custom allows you to define a custom parsing function
func (a *Arg) Custom(fn func(string) error) *Custom {
	value := &Custom{target: fn}
	a.value = &customValue{key: a.key(), inner: value}
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
	value := &OptionalString{target: target}
	a.a.value = &optionalStringValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArg) Int(target **int) *OptionalInt {
	value := &OptionalInt{target: target}
	a.a.value = &optionalIntValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArg) Bool(target **bool) *OptionalBool {
	value := &OptionalBool{target: target}
	a.a.value = &optionalBoolValue{key: a.key(), inner: value}
	return value
}

func (a *OptionalArg) Enum(target **string, possibilities ...string) *OptionalEnum {
	value := &OptionalEnum{target: target}
	a.a.value = &optionalEnumValue{key: a.key(), inner: value, possibilities: possibilities}
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
