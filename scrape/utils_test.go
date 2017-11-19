package scrape

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestRandFloat(t *testing.T) {
	// This can be used to generate random floats in
	// other ranges, for example `0.5 <= f' < 1.5`.
	rand.Seed(time.Now().Unix())
	fmt.Println((rand.Float64()*1)+1)
	fmt.Println(rand.Float64()+0.5)
	 // This can be used to generate random floats in
    // other ranges, for example `5.0 <= f' < 10.0`.
	fmt.Println()
    fmt.Print((rand.Float64()*5)+5, ",")
    fmt.Print((rand.Float64() * 5) + 5)
    fmt.Println()
}

func TestRandF(t *testing.T) {
	s := 500* time.Millisecond
	inttt := int64(RandomF()*1000)
	//tt := time.Duration(inttt*s)
	fmt.Println(s, inttt)
}	

func TestRandInt(t *testing.T){
	//initial fetch delay
	s := 500* time.Millisecond
	//random ratio 
	rand := Random(500, 1500)
	m := s * time.Duration(rand)/1000
	fmt.Println(s, rand, m)
}

func TestGenerateCRC32(t *testing.T){
	fmt.Println(string(GenerateCRC32([]byte("test test test"))))
}