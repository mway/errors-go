package errgroup

// Options are used to configure a Group.
type Options struct {
	// FirstOnly controls whether only the first non-nil error encountered will
	// be returned, or if all errors will be appended in a chain and returned.
	FirstOnly bool
	// IgnoredErrors is used to filter out unhelpful or immaterial errors,
	// such as io.EOF.
	IgnoredErrors []error
	// Inline controls whether functions passed to Group.Add are handled
	// inline and serially in the calling goroutine, or if they will be
	// executed in parallel in a background goroutine. Note that if Inline
	// is true, Group.Add becomes a blocking call.
	Inline bool
}

// DefaultOptions returns a new Options with sane defaults. Using default
// Options verbatim is functionally equivalent to using a zero-value Group.
func DefaultOptions() Options {
	return Options{
		FirstOnly: false,
		Inline:    false,
	}
}

// With returns a new Options, using the current Options as a base and merging
// the given options down onto it.
func (o Options) With(opts ...Option) Options {
	for _, opt := range opts {
		opt.apply(&o)
	}
	return o
}

func (o Options) apply(opts *Options) {
	opts.FirstOnly = o.FirstOnly
	opts.Inline = o.Inline

	if o.IgnoredErrors != nil {
		opts.IgnoredErrors = append(opts.IgnoredErrors, o.IgnoredErrors...)
	}
}

// An Option configures a Group.
type Option interface {
	apply(*Options)
}

type optionFunc func(*Options)

func (f optionFunc) apply(o *Options) {
	f(o)
}

// WithFirstOnly returns an Option that configures a Group to return the first
// encountered error verbatim. Subsequently returned errors will be ignored.
func WithFirstOnly() Option {
	return optionFunc(func(o *Options) {
		o.FirstOnly = true
	})
}

// WithIgnoredErrors returns an Option that configures a Group to ignore errors
// that contain any of the given errors in their error chains.
func WithIgnoredErrors(errs ...error) Option {
	return optionFunc(func(o *Options) {
		tmp := make([]error, 0, len(o.IgnoredErrors)+len(errs))
		tmp = append(tmp, o.IgnoredErrors...)
		tmp = append(tmp, errs...)
		o.IgnoredErrors = tmp
	})
}

// WithInline returns an Option that configures a Group to execute all
// functions provided to Group.Add inline and serially within the calling
// goroutine. Note that this will make Group.Add a blocking call.
func WithInline() Option {
	return optionFunc(func(o *Options) {
		o.Inline = true
	})
}
