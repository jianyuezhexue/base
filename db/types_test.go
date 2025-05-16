// nolint
package db

import (
	"testing"
	"time"
)

func TestLocalTime_IsLessThanToday(t *testing.T) {
	now := time.Now().In(time.Local)
	tests := []struct {
		name string
		time LocalTime
		want bool
	}{
		{
			name: "Yesterday",
			time: LocalTime(now.AddDate(0, 0, -1)),
			want: true,
		},
		{
			name: "Today",
			time: LocalTime(now),
			want: true,
		},
		{
			name: "Tomorrow",
			time: LocalTime(now.AddDate(0, 0, 1)),
			want: false,
		},
		{
			name: "Zero time",
			time: LocalTime(time.Time{}),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.time.LteToday(); got != tt.want {
				t.Errorf("LteToday() = %v, want %v", got, tt.want)
			}
		})
	}
}
