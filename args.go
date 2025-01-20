package cli

type Args struct {
	name  string
	help  string
	value value
}

func (a *Args) key() string {
	return "<" + a.name + "...>"
}

func (a *Args) verify() error {
	return a.value.verify()
}

func (a *Args) Optional() *OptionalArgs {
	return &OptionalArgs{a}
}

func (a *Args) Strings(target *[]string) *Strings {
	*target = []string{}
	value := &Strings{target: target}
	a.value = &stringsValue{key: a.key(), inner: value}
	return value
}

type OptionalArgs struct {
	a *Args
}

func (a *OptionalArgs) key() string {
	return "[<" + a.a.name + ">...]"
}

func (a *OptionalArgs) Strings(target *[]string) *Strings {
	*target = []string{}
	value := &Strings{target: target, optional: true}
	a.a.value = &stringsValue{key: a.key(), inner: value}
	return value
}
