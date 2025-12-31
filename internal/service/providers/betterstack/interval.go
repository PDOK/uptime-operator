package betterstack

import "math"

func toSupportedInterval(intervalInMin int) int {
	// Better Stack only accepts a specific sets of intervals
	supportedIntervals := []int{30, 45, 60, 120, 180, 300, 600, 900, 1800}

	intervalInSec := intervalInMin * 60
	if intervalInSec <= 0 {
		return supportedIntervals[0] // use the smallest supported interval
	}
	if intervalInSec > supportedIntervals[len(supportedIntervals)-1] {
		return supportedIntervals[len(supportedIntervals)-1] // use the largest supported interval
	}

	nearestInterval := supportedIntervals[0]
	prevDiff := math.MaxInt

	// use nearest supported interval
	for _, si := range supportedIntervals {
		diff := int(math.Abs(float64(intervalInSec - si)))
		if diff < prevDiff {
			prevDiff = diff
			nearestInterval = si
		}
	}
	return nearestInterval
}
