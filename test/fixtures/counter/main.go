package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	stopAt := -1
	var err error
	if len(os.Args) > 1 {
		stopAt, err = strconv.Atoi(os.Args[1])
		if err != nil {
			panic(err)
		}
	}

	prefix := "Counter"
	if len(os.Args) > 2 {
		prefix = fmt.Sprintf("%v", os.Args[2])
	}

	fmt.Println("Starting count")
	counter := 0
	for {
		fmt.Printf("%v: %v\n", prefix, counter)
		counter++
		if stopAt >= 0 && counter >= stopAt {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

}
