package logger

import (
	"github.com/getsentry/raven-go"
)

type Reporter struct {
	sentry *raven.Client
}

func (r *Reporter) Report(err error) (string, error) {
	if r.sentry == nil {
		return "", nil
	}
	eventId := r.sentry.CaptureErrorAndWait(err, nil)
	return eventId, nil
}

func NewReporter(dsn *string) (Reporter, error) {
	if dsn == nil {
		return Reporter{
			sentry: nil,
		}, nil
	}
	sentry, e := raven.New(*dsn)
	if e != nil {
		return Reporter{sentry: nil}, e
	}
	return Reporter{
		sentry: sentry,
	}, nil

}
