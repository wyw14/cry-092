package clock

import "time"

type UTCSource func() time.Time

func (source UTCSource) Now() time.Time {
	return source().UTC()
}

func Live() UTCSource {
	return func() time.Time { return time.Now() }
}

func At(value time.Time) UTCSource {
	frozen := value.UTC()
	return func() time.Time { return frozen }
}
