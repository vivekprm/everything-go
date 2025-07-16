Let's look at very basic rock climbing application.

```go 
package main

import "fmt"

type RockClimber struct {
	rocksClimbed int
}

func (rc *RockClimber) climbRock() {
	rc.rocksClimbed++
	if rc.rocksClimbed == 10 {
		rc.placeSafeties()
	}
}

func (rc *RockClimber) placeSafeties() {
	fmt.Println("Placing my safeties")
}

func main() {
  rc := &RockClimber{}
  for i := 0; i < 11; i++ {
    rc.climbRock()
  }
}
```

Here rock can be of different type e.g. sandy, concrete or ice, so climb rock depends on the type of rock being climbed. So we can 
modify our struct and have a ```kind``` field as below.

```go 
type RockClimber struct {
	rocksClimbed int
	kind         int
}

func (rc *RockClimber) placeSafeties() {
	switch rc.kind {
	case 1:
		// ICE
	case 2:
		// Sand
	case 3:
		// Concrete
	}
	fmt.Println("Placing my safeties")
}

```

But placing safety can be very very heavy function, may be placing safeties is going to go to our database or third party API.
May be only for ICE rocks we are going to database may be because we want to check what kind of temperature ICE is going to have
or something and may be based on that we connect to some third party API and pick right tool for that type of ice.

That means placing safety can have big effects on rock climber, which basically means that placing safeties is a hard dependency.

So we are going to extract placing dependency out of the domain of RockClimber.

```go 
type SafetyPlacer struct {
	kind int
}

func (sp *SafetyPlacer) placeSafeties() {
	switch sp.kind {
	case 1:
	// ICE
	case 2:
	// Sand
	case 3:
		// Concrete
	}
	fmt.Println("Placing my safeties...")
}
```

Now we are going to say rock climber will depend on safety placer.

```go 
type RockClimber struct {
	rocksClimbed int
	kind         int
	sp           SafetyPlacer
}

func (rc *RockClimber) climbRock() {
	rc.rocksClimbed++
	if rc.rocksClimbed == 10 {
		rc.sp.placeSafeties()
	}
}
```

It's already a better design because the RockClimber domain doesn't have logic of placing safeties because it's in the SafetyPlacer.
But actually you are moving the problem somewhere else but not solving it, as RockClimber still depends on SafetyPlacer.

How can we solve it?
So we are going to say SafetyPlacer is not going to be a struct but an interface.

```go 
type SafetyPlacer interface {
  placeSafeties()
}
```

So now RockClimber still depends on SafetyPlacer but it doesn't depend on the implementation of the SafetyPlacer.
If your program depends on implementation you have hard dependency. If your program depdends on a behaviour it's ok.

In case we are depending on the implementation if we need to test, eventhough we don't want to test RockClimber on placing
the safeties we still need to inject, we still need to construct RockClimber with the SafetyPlacer implementation which could have
a database setup, it can have api keys etc for API call. So we have to provide all that eventhough we don't want to test it.

Since now we are depending upon the behaviour, we are free to do anything.

Now we are going to add specific implementations:

```go 
type IceSafetyPlacer struct {
	// db
	// data
	// api
}

func (sp *IceSafetyPlacer) placeSafeties() {
	fmt.Println("Placing my ICE safeties...")
}
```

Now let's add a constructor for RockClimber.

```go 
// Injecting SafetyPlacer dependency
func newRockClimber(sp SafetyPlacer) *RockClimber {
	return &RockClimber{
		sp: sp,
	}
}

func main() {
	rc := newRockClimber(&IceSafetyPlacer{})
	for i := 0; i < 11; i++ {
		rc.climbRock()
	}
}
```

