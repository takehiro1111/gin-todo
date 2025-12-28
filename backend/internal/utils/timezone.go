package utils

import (
	"time"
)

func ToJSTLocalTime(timeNow time.Time) time.Time {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		loc = time.FixedZone("JST", 9*60*60) // 9 × 60分 × 60秒 = 32400秒(9h)
	}
	now := timeNow.In(loc)

	return now
}
