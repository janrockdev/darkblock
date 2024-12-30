package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/crypto/sha3"
)

func rnd(i int) int {
	return i + 1
}

func main() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var mu sync.Mutex // Mutex to control access and ensure sequential execution
	var i = 1

	for {
		<-ticker.C

		fmt.Printf("Ticker [%d] ticked at [%s]\n", i, time.Now())

		go func() {
			mu.Lock()         // Block until the previous execution completes
			defer mu.Unlock() // Ensure the lock is released after execution

			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			d := r.Intn(1000)
			time.Sleep(time.Duration(d+1000) * time.Millisecond) // Simulate some processing time
			fmt.Printf("Response [%d] (delay [%d]) executed at [%s]\n", rnd(i), d+1000, time.Now())

			hash := sha3.New512()

			// Data to hash
			data := []byte("Hello, this is an example using SHA-3 512!")
			// Write data to the hash object
			hash.Write(data)
			// Get the resulting hash as a byte slice
			hashBytes := hash.Sum(nil)
			// Print the hash in hexadecimal format
			fmt.Printf("SHA-3 512-bit hash: %x\n", hashBytes)
			i++
		}()
	}
}
