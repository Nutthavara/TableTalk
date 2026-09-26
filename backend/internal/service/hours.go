package service

import (
	"time"

	"tabletalk/backend/internal/model"
)

// withinOpeningHours reports whether [start, end) falls on a single local calendar day in loc and
// entirely inside the opening hours [open, close].
func withinOpeningHours(start, end time.Time, loc *time.Location, open, close model.TimeOfDay) bool {
	ls, le := start.In(loc), end.In(loc)
	sy, sm, sd := ls.Date()
	ey, em, ed := le.Date()
	if sy != ey || sm != em || sd != ed {
		return false
	}
	return secondOfDay(ls) >= int(open)*60 && secondOfDay(le) <= int(close)*60
}

func secondOfDay(t time.Time) int {
	return t.Hour()*3600 + t.Minute()*60 + t.Second()
}
