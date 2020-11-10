package main

import (
	"flag"
	"fmt"
	"math/big"
	"time"
)

const DEF_WRITE_SPEED = 100
const MAX_WRITE_SPEED = 3000

func genFib(ch chan *big.Int) {
	a := big.NewInt(0)
	b := big.NewInt(1)
	c := big.NewInt(0).Add(b, a)

	ch <- a
	ch <- b
	ch <- c

	for {
		a = b
		b = c
		c = big.NewInt(0).Add(b, a)
		ch <- c
	}
}

/*

if you prefer to structure generator and throttling logic separately,
however it is more complex from execution standpoint since it
requires two threads

func throttleFib(outCh chan *big.Int){
    numPerSecond := 1
    var nextNumber *big.Int
    startTime := time.Now()
    total := 0
    numbersCh := make(chan *big.Int, numPerSecond)
    go genFib(numbersCh)
	for {
        if total < numPerSecond {
            nextNumber = <-numbersCh
            outCh <- nextNumber
            total += 1
        } else if time.Since(startTime) > time.Second {
            startTime = time.Now()
            total = 0
        }
    }
}*/

func main() {
	var numFlag = flag.Int("generation_speed", DEF_WRITE_SPEED, "throttle output in numbers per second")
	var debugFlag = flag.Bool("debug", false, "enable debug mode")
	flag.Parse()

	numPerSecond := *numFlag
	if numPerSecond > MAX_WRITE_SPEED {
		numPerSecond = MAX_WRITE_SPEED
	}

	debug := *debugFlag

	numbersCh := make(chan *big.Int, numPerSecond)
	var nextNumber *big.Int
	go genFib(numbersCh)

	startTime := time.Now()
	total := 0
	for {
		if total < numPerSecond {
			nextNumber = <-numbersCh
			fmt.Println(nextNumber)
			if debug {
				fmt.Println("Current time: ", time.Now())
			}
			total += 1
		} else if time.Since(startTime) > time.Second {
			startTime = time.Now()
			total = 0
		}
	}
}
