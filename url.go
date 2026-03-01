package cli

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/kballard/go-shellquote"
)

type Url struct {
	target *url.URL
	envvar *string
	defval *url.URL
}

func (v *Url) Default(value url.URL) {
	v.defval = &value
}

type urlValue struct {
	key   string
	inner *Url
	set   bool
}

var _ value = (*urlValue)(nil)

func (v *urlValue) optional() bool {
	return false
}

func (v *urlValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *urlValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return v.inner.defval.String(), true
}

func (v *urlValue) verify() error {
	if v.set {
		return nil
	} else if value, ok := lookupEnv(v.inner.envvar); ok {
		return v.Set(value)
	} else if v.hasDefault() {
		*v.inner.target = *v.inner.defval
		return nil
	}
	return &missingInputError{v.key, v.inner.envvar}
}

func (v *urlValue) Set(val string) error {
	u, err := url.Parse(val)
	if err != nil {
		return fmt.Errorf("%s: expected a URL but got %q", v.key, val)
	}
	*v.inner.target = *u
	v.set = true
	return nil
}

func (v *urlValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return v.inner.target.String()
	} else if v.hasDefault() {
		return v.inner.defval.String()
	}
	return ""
}

type OptionalUrl struct {
	target **url.URL
	envvar *string
	defval *url.URL
}

func (v *OptionalUrl) Default(value url.URL) {
	v.defval = &value
}

type optionalUrlValue struct {
	key   string
	inner *OptionalUrl
	set   bool
}

var _ value = (*optionalUrlValue)(nil)

func (v *optionalUrlValue) optional() bool {
	return true
}

func (v *optionalUrlValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *optionalUrlValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	return v.inner.defval.String(), true
}

func (v *optionalUrlValue) verify() error {
	if v.set {
		return nil
	} else if value, ok := lookupEnv(v.inner.envvar); ok {
		return v.Set(value)
	} else if v.hasDefault() {
		*v.inner.target = v.inner.defval
		return nil
	}
	return nil
}

func (v *optionalUrlValue) Set(val string) error {
	u, err := url.Parse(val)
	if err != nil {
		return fmt.Errorf("%s: expected a URL but got %q", v.key, val)
	}
	*v.inner.target = u
	v.set = true
	return nil
}

func (v *optionalUrlValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return (*v.inner.target).String()
	} else if v.hasDefault() {
		return v.inner.defval.String()
	}
	return ""
}

type Urls struct {
	target   *[]*url.URL
	envvar   *string
	defval   *[]*url.URL
	optional bool
}

func (v *Urls) Default(values ...*url.URL) {
	v.defval = &values
}

type urlsValue struct {
	key   string
	inner *Urls
	set   bool
}

var _ value = (*urlsValue)(nil)

func (v *urlsValue) optional() bool {
	return v.inner.optional
}

func (v *urlsValue) verify() error {
	if v.set {
		return nil
	} else if value, ok := lookupEnv(v.inner.envvar); ok {
		fields, err := shellquote.Split(value)
		if err != nil {
			return fmt.Errorf("%s: expected a list of URLs but got %q", v.key, value)
		}
		for _, val := range fields {
			if err := v.Set(val); err != nil {
				return err
			}
		}
		return nil
	} else if v.hasDefault() {
		*v.inner.target = *v.inner.defval
		return nil
	} else if v.inner.optional {
		return nil
	}
	return &missingInputError{v.key, v.inner.envvar}
}

func (v *urlsValue) hasDefault() bool {
	return v.inner.defval != nil
}

func (v *urlsValue) Default() (string, bool) {
	if v.inner.defval == nil {
		return "", false
	}
	if len(*v.inner.defval) == 0 {
		return "[]", true
	}
	return strings.Join(stringifyURLSlice(*v.inner.defval), ", "), true
}

func (v *urlsValue) Set(val string) error {
	u, err := url.Parse(val)
	if err != nil {
		return fmt.Errorf("%s: expected a URL but got %q", v.key, val)
	}
	*v.inner.target = append(*v.inner.target, u)
	v.set = true
	return nil
}

func (v *urlsValue) String() string {
	if v.inner == nil {
		return ""
	} else if v.set {
		return strings.Join(stringifyURLSlice(*v.inner.target), ", ")
	} else if v.hasDefault() {
		return strings.Join(stringifyURLSlice(*v.inner.defval), ", ")
	}
	return ""
}

func stringifyURLSlice(values []*url.URL) []string {
	out := make([]string, len(values))
	for i, value := range values {
		if value == nil {
			continue
		}
		out[i] = value.String()
	}
	return out
}
