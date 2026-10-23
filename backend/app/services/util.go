package services

import "time"

func ParseTime(value, layout string) (time.Time, error) {
	return time.Parse(layout, value)
}
