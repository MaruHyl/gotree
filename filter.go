package gotree

import "regexp"

// Filter will filter out the matched dep(return true)
type Filter interface {
	Filter(name string) bool
}

type nopFilter struct {
}

func (f nopFilter) Filter(name string) bool {
	return false
}

func NewNopFilter() Filter {
	return nopFilter{}
}

type regexpFilter struct {
	reg *regexp.Regexp
}

func NewRegexpFilter(pattern string) (Filter, error) {
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	f := &regexpFilter{reg}
	return f, nil
}

func (f regexpFilter) Filter(name string) bool {
	return f.reg.MatchString(name)
}

type reverseFilter struct {
	f Filter
}

func NewReverseFilter(f Filter) Filter {
	return reverseFilter{f}
}

func (f reverseFilter) Filter(name string) bool {
	return !f.f.Filter(name)
}
