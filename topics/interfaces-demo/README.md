Let's consider a small appplication where we have a football team with 11 player and each can kick the ball. KickBall calculates
shot based on player stat as below.

```go 
package main

import (
	"fmt"
	"math/rand"
)

type FootballPlayer struct {
	stamina int
	power   int
}

func (f FootballPlayer) KickBall() {
	shot := f.stamina + f.power
	fmt.Println("I'm kicking a ball", shot)
}

func main() {
	team := make([]FootballPlayer, 11)
	for i := 0; i < len(team); i++ {
		team[i] = FootballPlayer{
			stamina: rand.Intn(10),
			power:   rand.Intn(10),
		}
	}
	for i := 0; i < len(team); i++ {
		team[i].KickBall()
	}
}
```

There is nothing wrong with this code. But if we think about it, there are going to be different type of football players some with 
good defense, some left foot player etc. So you might endup with so many fields inside FootballPlayer struct. And while calculating
shot we are going to have lots of if else conditions. Testing it is going to be very very complicated.

In other languages we are going to have Base Football Player class and then other type of players inherit that stuff. But in Golang
we don't have any such thing. We can achieve similar thing in Golang using interfaces, so let's refactor this.

So only thing we are concerned about in this example is *KickBall* function.

```go 
package main

import (
	"fmt"
	"math/rand"
)

type Player interface {
	KickBall()
}
type FootballPlayer struct {
	stamina int
	power   int
}

func (f FootballPlayer) KickBall() {
	shot := f.stamina + f.power
	fmt.Println("I'm kicking a ball", shot)
}

type CR7 struct {
	stamina int
	power   int
    sui int
}

func (f CR7) KickBall() {
	shot := f.stamina + f.power * f.SUI
	fmt.Println("I'm kicking a ball", shot)
}


func main() {
	team := make([]Player, 11)
	for i := 0; i < len(team) - 1; i++ {
		team[i] = FootballPlayer{
			stamina: rand.Intn(10),
			power:   rand.Intn(10),
		}
	}
    team[len(team)] = CR7{
        stamina: 10,
        power: 10,
        sui: 8
    }
	for i := 0; i < len(team); i++ {
		team[i].KickBall()
	}
}
```

Now we are free to create any type of FootballPlayer with some additional qualities but as long as it has KickBall behaviour, it can
be put inside the team slice.
