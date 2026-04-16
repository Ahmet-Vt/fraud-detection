package transaction

import (
	"math"
	"time"
)

func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0

	latDiffRadian := math.Abs(lat1-lat2) * math.Pi / 180.0
	lonDiffRadian := math.Abs(lon1-lon2) * math.Pi / 180.0

	haversineLatDiff := (1 - math.Cos(latDiffRadian)) / 2
	haversineLonDiff := (1 - math.Cos(lonDiffRadian)) / 2

	cosLatRadian1 := math.Cos(lat1 * math.Pi / 180.0)
	cosLatRadian2 := math.Cos(lat2 * math.Pi / 180.0)

	haversineTheta := haversineLatDiff + (cosLatRadian1 * cosLatRadian2 * haversineLonDiff)

	centralAngle := 2 * math.Atan2(math.Sqrt(haversineTheta), math.Sqrt(1 - haversineTheta))

	kmDistance := centralAngle * R

	return kmDistance
}

func IsReachableWithPlane(kmDistance float64, lastTransactionTime time.Time) bool {
	basePlaneSpeedKmAnHour := 800.0
	timeToReach := kmDistance / basePlaneSpeedKmAnHour

	timeSinceLastTransaction := time.Since(lastTransactionTime).Hours()

	if timeSinceLastTransaction >= timeToReach {
		return true
	}

	return false
}
