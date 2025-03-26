package main

/*
#include <stdlib.h>
#include <time.h>
*/
import "C"
import "fmt"

func main() {
	Seed(123)
	fmt.Println("Random: ", Random())
}

func Seed(i int) {
	C.srandom(C.uint(i))
}

func Random() int {
	return int(C.random())
}
