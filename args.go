package cli

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
	*target = []string{}
	value := &Strings{target, a.env, nil, false}
	a.value = &stringsValue{key: a.key(), inner: value}
	return value
}

func (a *Args) StringMap(target *map[string]string) *StringMap {
	*target = map[string]string{}
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

func (a *OptionalArgs) StringMap(target *map[string]string) *StringMap {
	value := &StringMap{target, a.a.env, nil, true}
	a.a.value = &stringMapValue{key: a.key(), inner: value}
	return value
}
