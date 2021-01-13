package virtualtime

import "time"

func init() {
	virtualTime = defaultVirtualTime{}
}

type defaultVirtualTime struct{}

func (defaultVirtualTime) Parse(s string) (time.Time, error) {
	var format string
	if len(s) > 10 {
		format = "2006-1-2 15:04:05"
	} else {
		format = "2006-1-2"
	}
	return time.Parse(format, s)
}

func (defaultVirtualTime) Format(tm time.Time) string {
	return tm.Format("2006-01-02 15:04:05")
}

func (defaultVirtualTime) Now() time.Time {
	return time.Now()
}

func (defaultVirtualTime) Time(ts int64) time.Time {
	return time.Unix(ts, 0)
}

func (defaultVirtualTime) Location() *time.Location {
	return time.Local
}

func (defaultVirtualTime) FakeDuration() time.Duration {
	return 0
}
