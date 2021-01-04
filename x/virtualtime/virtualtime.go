package virtualtime

import "time"

type VirtualTime interface {
	Now() time.Time
	Time(ts int64) time.Time
	Parse(string) (time.Time, error)
	Format(time.Time) string
	Location() *time.Location
	FakeDuration() time.Duration
}

var virtualTime VirtualTime

func Set(i VirtualTime) {
	virtualTime = i
}

func Parse(s string) (time.Time, error) {
	return virtualTime.Parse(s)
}

func Format(tm time.Time) string {
	return virtualTime.Format(tm)
}

func Now() time.Time {
	return virtualTime.Now()
}

func Time(ts int64) time.Time {
	return virtualTime.Time(ts)
}

func Location() *time.Location {
	return virtualTime.Location()
}

func FakeDuration() time.Duration {
	return virtualTime.FakeDuration()
}
