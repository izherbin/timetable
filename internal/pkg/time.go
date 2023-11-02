package pkg

import (
	"fmt"
	"regexp"
	"time"
)

var re = regexp.MustCompile(`^([0-9]|0[0-9]|1[0-9]|2[0-3]):([0-9]|[0-5][0-9])$`)

var Now time.Time

func init() {
	Now = time.Now()
}

func ParseHM(tStr string) (time.Time, error) {
	var h, m int
	if tStr == "" {
		return time.Time{}, nil
	}
	_, err := fmt.Sscanf(tStr, "%d:%d", &h, &m)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(Now.Year(), Now.Month(), Now.Day(), h, m, 0, Now.Nanosecond(), Now.Location()), nil
}

func ValidateTime(tStr string) bool {
	return re.MatchString(tStr)
}
