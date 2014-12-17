package parsers

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

/* all time is in UTC */
const TIME_FORMAT string = "2006.01.02-15:04:05"

func ParseTimeToEpoch(timestr string) (float64, error) {

	if strings.HasSuffix(timestr, "-ago") {
		relTimeUnit := timestr[:len(timestr)-4]

		unit := relTimeUnit[len(relTimeUnit)-1]

		val, err := strconv.ParseFloat(relTimeUnit[:len(relTimeUnit)-1], 64)
		if err != nil {
			return -1, err
		}

		currTime := float64(time.Now().UnixNano()) / 1000000000

		switch unit {
		case 's':
			return currTime - val, nil
		case 'm':
			return currTime - (val * 60), nil
		case 'h':
			return currTime - (val * 3600), nil
		case 'd':
			return currTime - (val * 86400), nil
		case 'w':
			return currTime - (val * 604800), nil
		default:
			return -1, fmt.Errorf("Invalid unit: %s", unit)
		}
	} else {
		t, err := time.Parse(TIME_FORMAT, timestr)
		if err != nil {
			val, err := strconv.ParseFloat(timestr, 64)
			if err != nil {
				return -1, err
			}
			return val, nil
		}
		return float64(t.UnixNano()) / 1000000000, nil
	}
}
