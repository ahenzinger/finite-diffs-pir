package pir

import (
	"os"
	"fmt"
	"time"
	"runtime"
	"runtime/pprof"
	"sync"
	"math"
	"unsafe"
)

func BytesToKB(b int) float64 {
        return float64(b) / 1024
}

func BytesToMB(b int) float64 {
        return float64(b) / 1024 / 1024
}

func BytesToGB(b int) float64 {
        return float64(b) / 1024 / 1024 / 1024
}

func BytesToTB(b int) float64 {
        return float64(b) / 1024 / 1024 / 1024 / 1024
}

func RandVec(length int) []bool {
        v := make([]bool, length)
	seed := RandomPRGKey()
	prg := NewBufPRG(NewPRG(seed))

        for i := 0; i < length; i++ {
		val := prg.Uint64()
                v[i] = ((val % 2) == 1)
        }

        return v
}

func RandByteVec(length int) []byte {
        v := make([]byte, length)
		maxByte := (1 << 8)

		workers := 192 * 5
		vals_per_worker := (length + workers - 1) / workers
		var wg sync.WaitGroup
	
		for w := 0; w < workers; w++ {
			wg.Add(1)

			go func(at int) {
				defer wg.Done()

				start := at * vals_per_worker
				l := vals_per_worker
				if start + l > length {
					l = length - start
				}

				if start >= length {
					return
				}

				seed := RandomPRGKey()
        			prg := NewBufPRG(NewPRG(seed))

				// Fill in 8‐byte chunks by reinterpreting the slice as []uint64
				filled := l / 8
				u64s := unsafe.Slice((*uint64)(unsafe.Pointer(&v[start])), filled)
				for i, _ := range u64s {
					u64s[i] = prg.Uint64()
				}
	
        			for i := filled * 8; i < l; i++ {
                			v[start + i] = byte(prg.Uint64() % uint64(maxByte))
        			}
			}(w)	
		}

		wg.Wait()

        return v
}

func VecAdd(v1, v2 []bool) []bool {
        if len(v1) != len(v2) {
                panic("VecAdd: Length mismatch")
        }

        u := make([]bool, len(v1))
        copy(u, v1)

		for i := 0; i < len(v1); i++ {
                if v2[i] {
                        u[i] = !u[i]
                }
        }

        return u
}

func VecByteAdd(v1, v2 []byte) []byte {
        if len(v1) != len(v2) {
                panic("VecByteAdd: Length mismatch")
        }

        u := make([]byte, len(v1))
        copy(u, v1)

        for i := 0; i < len(v1); i++ {
				u[i] ^= v2[i]
        }

        return u
}

func Binomial(n, k int) int {
        if k < 0 || k > n {
                return 0
        }

        if k > n-k {
                k = n - k
        }

        result := 1
        for i := 1; i <= k; i++ {
                result = result * int(n-k+i) / int(i)
        }

        return result
}

// initialize first combination [0,1,2,...,k-1]
func FirstCombination(k int) []int {
		idxs := make([]int, k)
		for i := 0; i < k; i++ {
				idxs[i] = i
        }
		return idxs
}

// nextCombination lexicographically advances the k-combination in-place.
// a must be sorted, of length k, with 0 <= a[0] < a[1] < … < a[k-1] < n.
// Returns false once it overflows past the last combination.
func NextCombination(a []int, n int) bool {
        k := len(a)
        for i := k - 1; i >= 0; i-- {
                if a[i] < n-k+i {
                        a[i]++

                        for j := i + 1; j < k; j++ {
                                a[j] = a[j-1] + 1
                        }

                        return true
                }
        }
        return false
}

func ProfileMemory(filename string) {
        f, err := os.Create(filename)
        if err != nil {
                fmt.Println(filename)
                panic("Could not create mem profile")
        }

        defer f.Close()

        runtime.GC()

        if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
                panic("Could not write to memory profile")
        }
}

func ProfileCPU(filename string) *os.File {
        f, err := os.Create(filename)
        if err != nil {
                fmt.Println(filename)
                panic("Could not create CPU profile")
        }

        if err := pprof.StartCPUProfile(f); err != nil {
                panic("Could not start CPU profile")
        }

        return f
}

func PrintMemUsage() {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)

        fmt.Printf("Alloc = %v GiB", m.Alloc / 1024 / 1024 / 1024)
        fmt.Printf("   TotalAlloc = %v GiB", m.TotalAlloc / 1024 / 1024 / 1024)
        fmt.Printf("   Sys = %v GiB", m.Sys / 1024 / 1024 / 1024)
        fmt.Printf("   NumGC = %v\n", m.NumGC)
}

func Avg(data []float64) float64 {
		sum := 0.0
		num := 0.0
		for _, elem := range data {
				sum += elem
				num += 1.0
		}
		return sum / num
}

func Stddev(data []float64) float64 {
		avg := Avg(data)
		sum := 0.0
		num := 0.0
		for _, elem := range data {
				sum += math.Pow(elem-avg, 2)
 				num += 1.0
		}
		variance := sum / num 
		return math.Sqrt(variance)
}

func PrintTime(start time.Time) time.Duration {
		elapsed := time.Since(start)
		fmt.Printf("\tElapsed: %s\n", elapsed)
		return elapsed
}

func Greenln(format string, a ...any) {
		fmt.Print("\033[32m")
		fmt.Printf(format, a...)
		fmt.Print("\033[0m")
}

func Redln(format string, a ...any) {
        fmt.Print("\033[31m")
        fmt.Printf(format, a...)
        fmt.Print("\033[0m")
}
