package gotree

import "errors"

//
const DefaultMaxLevel = 0 // no limit

var defaultFilter = nopFilter{}

//
type options struct {
	maxLevel   int
	noReport   bool
	filter     Filter
	noStd      bool
	noInternal bool
}

var defaultOptions = options{
	maxLevel: DefaultMaxLevel,
	noReport: false,
	filter:   defaultFilter,
}

type Option func(opts *options) error

func buildOpts(options ...Option) (options, error) {
	opts := defaultOptions
	for _, opt := range options {
		err := opt(&opts)
		if err != nil {
			return opts, err
		}
	}
	return opts, nil
}

// Set max level of tree
func WithMaxLevel(maxLevel int) Option {
	return func(opts *options) error {
		if maxLevel < 0 {
			return errors.New("max level must more than 0")
		}
		opts.maxLevel = maxLevel
		return nil
	}
}

// Turn off dep/direct/indirect count at end of tree listing
func WithNoReport(noReport bool) Option {
	return func(opts *options) error {
		opts.noReport = noReport
		return nil
	}
}

// Filter will filter out the matched dep
func WithFilter(filter Filter) Option {
	return func(opts *options) error {
		opts.filter = filter
		return nil
	}
}

// Filter out std lib
func WithNoStd(noStd bool) Option {
	return func(opts *options) error {
		opts.noStd = noStd
		return nil
	}
}

// Filter out internal pkg
func WithNoInternal(noInternal bool) Option {
	return func(opts *options) error {
		opts.noInternal = noInternal
		return nil
	}
}
