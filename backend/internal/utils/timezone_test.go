package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var mockedTime = time.Date(2025, 12, 28, 11, 30, 0, 0, time.UTC)

func TestToJSTLocalTime(t *testing.T) {

	tests := []struct {
		name             string
		timeMock         time.Time
		expectedTimeZone string
	}{
		{
			name:             "[正常系] JSTに変換され意図した時刻で返ること",
			timeMock:         mockedTime,
			expectedTimeZone: "Asia/Tokyo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jstTime := ToJSTLocalTime(tt.timeMock)

			assert.Equal(t, tt.expectedTimeZone, jstTime.Location().String())
			assert.True(t, tt.timeMock.Equal(jstTime))
		})
	}
}
