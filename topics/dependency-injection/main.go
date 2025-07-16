package main

import "fmt"

type RockClimber struct {
	rocksClimbed int
	kind         int
	sp           SafetyPlacer
}

// Injecting SafetyPlacer dependency
func newRockClimber(sp SafetyPlacer) *RockClimber {
	return &RockClimber{
		sp: sp,
	}
}

type SafetyPlacer interface {
	placeSafeties()
}

type IceSafetyPlacer struct {
	// db
	// data
	// api
}

func (sp *IceSafetyPlacer) placeSafeties() {
	fmt.Println("Placing my ICE safeties...")
}

type NopSafetyPlacer struct{}

func (sp *NopSafetyPlacer) placeSafeties() {
	fmt.Println("Placing no safeties...")
}

func (rc *RockClimber) climbRock() {
	rc.rocksClimbed++
	if rc.rocksClimbed == 10 {
		rc.sp.placeSafeties()
	}
}

func main() {
	rc := newRockClimber(&IceSafetyPlacer{})
	for i := 0; i < 11; i++ {
		rc.climbRock()
	}
	rc = newRockClimber(&NopSafetyPlacer{})
	for i := 0; i < 11; i++ {
		rc.climbRock()
	}
}
