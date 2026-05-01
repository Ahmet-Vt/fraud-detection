package distance

import (
	"testing"
	"time"
)

func TestDistanceRuleBursaToIstanbulNotReachableInTwoMinutes(t *testing.T) {
	bursa, ok := GetCityCoordinates("bursa")
	if !ok {
		t.Fatal("expected coordinates for bursa")
	}
	t.Logf("bursa coordinates: lat=%.6f lon=%.6f", bursa.Lat, bursa.Lon)

	istanbul, ok := GetCityCoordinates("istanbul")
	if !ok {
		t.Fatal("expected coordinates for istanbul")
	}
	t.Logf("istanbul coordinates: lat=%.6f lon=%.6f", istanbul.Lat, istanbul.Lon)

	kmDistance := HaversineDistance(bursa.Lat, bursa.Lon, istanbul.Lat, istanbul.Lon)
	t.Logf("calculated distance: %.2f km", kmDistance)
	if kmDistance < 70 || kmDistance > 200 {
		t.Fatalf("expected Bursa-Istanbul distance to be in a plausible range, got %.2f km", kmDistance)
	}

	elapsed := 2 * time.Minute
	t.Logf("elapsed between transactions: %.2f minutes", elapsed.Minutes())

	reachable := IsReachableWithPlane(kmDistance, elapsed)
	t.Logf("reachable by plane: %t", reachable)
	if reachable {
		t.Fatalf("expected Bursa-Istanbul trip to be unreachable in 2 minutes, distance was %.2f km", kmDistance)
	}
}

func TestDistanceRuleBursaToIstanbulReachableInTwentyMinutes(t *testing.T) {
	bursa, ok := GetCityCoordinates("bursa")
	if !ok {
		t.Fatal("expected coordinates for bursa")
	}
	t.Logf("bursa coordinates: lat=%.6f lon=%.6f", bursa.Lat, bursa.Lon)

	istanbul, ok := GetCityCoordinates("istanbul")
	if !ok {
		t.Fatal("expected coordinates for istanbul")
	}
	t.Logf("istanbul coordinates: lat=%.6f lon=%.6f", istanbul.Lat, istanbul.Lon)

	kmDistance := HaversineDistance(bursa.Lat, bursa.Lon, istanbul.Lat, istanbul.Lon)
	t.Logf("calculated distance: %.2f km", kmDistance)
	if kmDistance < 70 || kmDistance > 200 {
		t.Fatalf("expected Bursa-Istanbul distance to be in a plausible range, got %.2f km", kmDistance)
	}

	elapsed := 20 * time.Minute
	t.Logf("elapsed between transactions: %.2f minutes", elapsed.Minutes())

	reachable := IsReachableWithPlane(kmDistance, elapsed)
	t.Logf("reachable by plane: %t", reachable)
	if !reachable {
		t.Fatalf("expected Bursa-Istanbul trip to be reachable in 20 minutes, distance was %.2f km", kmDistance)
	}
}
