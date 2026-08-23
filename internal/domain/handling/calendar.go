package handling

import "time"

type WorkdayCalendar struct {
	Location *time.Location
	Holidays map[string]bool
	Makeup   map[string]bool
}

func NewCalendar(location *time.Location, holidays, makeup []time.Time) WorkdayCalendar {
	if location == nil {
		location = time.UTC
	}
	c := WorkdayCalendar{Location: location, Holidays: map[string]bool{}, Makeup: map[string]bool{}}
	for _, d := range holidays {
		c.Holidays[c.key(d)] = true
	}
	for _, d := range makeup {
		c.Makeup[c.key(d)] = true
	}
	return c
}

func (c WorkdayCalendar) key(t time.Time) string { return t.In(c.Location).Format("2006-01-02") }

func (c WorkdayCalendar) IsWorkday(t time.Time) bool {
	key := c.key(t)
	if c.Makeup[key] {
		return true
	}
	if c.Holidays[key] {
		return false
	}
	w := t.In(c.Location).Weekday()
	return w != time.Saturday && w != time.Sunday
}

func (c WorkdayCalendar) AddWorkdays(start time.Time, days int) time.Time {
	current := start.In(c.Location)
	for remaining := days; remaining > 0; {
		current = current.AddDate(0, 0, 1)
		if c.IsWorkday(current) {
			remaining--
		}
	}
	return current.UTC()
}

func (c WorkdayCalendar) WorkdaysBetween(start, end time.Time) int {
	if !end.After(start) {
		return 0
	}
	count := 0
	for cursor := start.In(c.Location).AddDate(0, 0, 1); !cursor.After(end.In(c.Location)); cursor = cursor.AddDate(0, 0, 1) {
		if c.IsWorkday(cursor) {
			count++
		}
	}
	return count
}
