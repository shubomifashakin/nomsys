package timeutil

import (
	"fmt"
	"time"
)

func ConvertMsToTime(ms int64) string {
	createdTime := time.UnixMilli(ms)

	d := time.Since(createdTime)

	h := int64(d.Hours())
	m := int64(d.Minutes()) % 60
	s := int64(d.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
