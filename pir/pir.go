package pir

//#cgo CFLAGS: -O3 -march=native -mavx2
//#include "lookup_batch.h"
import "C"

import (
	"fmt"
	"time"
	"sort"
	"sync"
	"runtime"
	"slices"
	"unsafe"
	"runtime/debug"
	"math/rand"
)

// Returns x, query 1, query 2
func Query(i int, p *Params) (int, int, int) {
	x := EncodeIndex(i, p)
	r := RandVec(p.M)
	return EncodingToIndex(x), EncodingToIndex(r), EncodingToIndex(VecAdd(x, r))
}

func fetchCloud(p *Params) []int {
	var cloud []int

    for k := 0; k <= p.D/2; k++ {
        idxs := FirstCombination(k)
		for {
			point := 0
			for _, pos := range idxs {
				point |= (1 << pos)
			}
			cloud = append(cloud, point)

            if !NextCombination(idxs, p.M) {
                break
            }
        }
    }

	sort.Ints(cloud)

    return cloud
}

func Answer(enc *EncodedDatabase, cloud []int, query int) []byte {
	n := len(cloud)
	record_len := enc.Poly.Record_len
	ans := make([]byte, n * record_len)

	v1 := (*C.uint64_t)(unsafe.Pointer(&cloud[0]))
	v2 := (*C.uint8_t)(unsafe.Pointer(&enc.Data[0]))
	v3 := (*C.uint8_t)(unsafe.Pointer(&ans[0]))
	v4 := C.uint64_t(n)
	v6 := C.uint64_t(query)

	if record_len == 1 {
		C.lookupBatch_len1(v1, v2, v3, v4, v6)

	} else if record_len == 10 {
		C.lookupBatch_len10(v1, v2, v3, v4, v6)

	} else if record_len == 64 {
		C.lookupBatch_len64(v1, v2, v3, v4, v6)

	} else if record_len == 1024 {
		C.lookupBatch_len1024(v1, v2, v3, v4, v6)

	} else if record_len == 1024 * 100 {
		C.lookupBatch_len102400(v1, v2, v3, v4, v6)

	} else {
		v5 := C.uint64_t(record_len)
		C.lookupBatch(v1, v2, v3, v4, v5, v6)
	}

	return ans
}

func Recover(p *Params, cloud []int, state int, ans1 []byte, ans2 []byte) []byte {
	record_len := p.Record_len
	result := make([]byte, record_len)

	for at, point := range cloud {

		all_set := ((point & state) == point)

		if all_set {
			for j := 0; j < record_len; j++ {
				result[j] ^= ans1[at*record_len+j]
				result[j] ^= ans2[at*record_len+j]
			}
		}
    }

	return result
}

func RunPIRWithParams(p *Params) *Metrics {
	debug.SetGCPercent(-1)

	fmt.Printf("Executing DEPIR over database of %d %d-byte records: %f GB\n", p.N, p.Record_len, BytesToGB(p.N * p.Record_len))
	fmt.Printf("       picked parameters: m = %d, D = %d --> could handle up to %d records\n", p.M, p.D, Binomial(p.M, p.D))
	db := RandDatabase(p.N, p.Record_len)

	fmt.Println("    Setup...")
	start := time.Now()
	enc := EncodeDatabase(db, p)
	duration := time.Since(start)
    fmt.Printf("        took %v\n", duration)
    fmt.Printf("        original db size: %f GB --> encoded to %f GB\n", BytesToGB(p.N * p.Record_len), BytesToGB(enc.Bytelen()))
	runtime.GC()

	numQueries := 21
	fmt.Printf("    Building %d queries...\n", numQueries)
	start = time.Now()
	indices := make([]int, numQueries)
	clientState := make([]int, numQueries)
	query0 := make([]int, numQueries)
	query1 := make([]int, numQueries)
	for i := 0; i < numQueries; i++ {
		indices[i] = rand.Intn(p.N)
		clientState[i], query0[i], query1[i] = Query(indices[i], p)
	}
	duration = time.Since(start)
    fmt.Printf("       took %v\n", duration)
	runtime.GC()

	fmt.Printf("   Answering %d queries...\n", numQueries)
	cloud := fetchCloud(enc.Poly)
	ans0 := make([][]byte, numQueries)
	ans1 := make([][]byte, numQueries)
	times := make([]float64, 0)
	for i := 0; i < numQueries; i++ {
		start := time.Now()
		ans0[i] = Answer(enc, cloud, query0[i])
		duration = time.Since(start)
		times = append(times, duration.Seconds())

		start = time.Now()
		ans1[i] = Answer(enc, cloud, query1[i])
		duration = time.Since(start)
		//times = append(times, duration.Seconds())
	}

	times = times[1:]
    fmt.Printf("        answer size: %f KB = %f MB\n", BytesToKB(len(ans0[0])), BytesToMB(len(ans1[0])))
	fmt.Printf("        average answer time per query: %f s\n", Avg(times))
	fmt.Printf("        std dev of answer time per query: %f s\n", Stddev(times))

	fmt.Printf("    Reconstructing %d queries...\n", numQueries)
	start = time.Now()
	for i := 0; i < numQueries; i++ {
		recovered := Recover(p, cloud, clientState[i], ans0[i], ans1[i])

		if !slices.Equal(db.Read(indices[i]), recovered) {
			fmt.Printf("    Index %d: Get %b but should be %b\n", indices[i], recovered, db.Read(indices[i]))
			panic("Reconstruct failed!")
		}
	}
	duration = time.Since(start)
    fmt.Printf("        took %v\n", duration)
	runtime.GC()
	debug.SetGCPercent(100)

	return NewMetric(p, times, len(ans0[0]))
}

func RunPIR(N, record_len int, theta float64) *Metrics {
    p := PickParams(N, record_len, theta)
	return RunPIRWithParams(p)
}

func RunFakePIRWithParams(enc *EncodedDatabase, p *Params) *Metrics {
    debug.SetGCPercent(-1)

	fmt.Printf("Executing ~fake~ DEPIR over database of %d %d-byte-length records: %f GB\n", p.N, p.Record_len, BytesToGB(p.N * p.Record_len))
    fmt.Printf("       parameters: m = %d, D = %d --> could handle up to %d records\n", p.M, p.D, Binomial(p.M, p.D))

    fmt.Println("    Setup (~fake~)...")
    fmt.Printf("        original db size: %f GB --> encoded to %f GB\n", BytesToGB(p.N * p.Record_len), BytesToGB(enc.Bytelen()))
    runtime.GC()

    numQueries := 2 
    fmt.Printf("    Building %d queries...\n", numQueries)
    start := time.Now()
    query0 := make([]int, numQueries)
    for i := 0; i < numQueries; i++ {
        _, query0[i], _ = Query(0, p)
    }
    duration := time.Since(start)
    fmt.Printf("        took %v\n", duration)
    fmt.Printf("        query size: 64 bits\n")
    runtime.GC()

    fmt.Printf("    Answering %d queries...\n", numQueries)
	cloud := fetchCloud(enc.Poly)
    ans0 := make([][]byte, numQueries)
	times := make([]float64, 0)
    for i := 0; i < numQueries; i++ {
		start := time.Now()
        ans0[i] = Answer(enc, cloud, query0[i])
		duration := time.Since(start)
		times = append(times, duration.Seconds())
    }

	times = times[1:] // Drop first (to measure steady-state)
	fmt.Println(times)
    fmt.Printf("        answer size: %f KB = %f MB\n", BytesToKB(len(ans0[0])), BytesToMB(len(ans0[0])))
    fmt.Printf("        average answer time per query: %f s\n", Avg(times))
    fmt.Printf("        std dev of answer time per query: %f s\n", Stddev(times))
	fmt.Printf("        BIM storage would be: %f TB --> worse by %f x\n", BytesToTB(enc.Bytelen()) / float64(p.Record_len) * float64(len(ans0[0])), float64(len(ans0[0])) / float64(p.Record_len))

    debug.SetGCPercent(100)
	PrintMemUsage()

	return NewMetric(p, times, len(ans0[0]))
}

func RunFakePIR(N, record_len int, theta float64) *Metrics {
    p := PickParams(N, record_len, theta)
	enc := FakeEncodeDatabase(p)
	return RunFakePIRWithParams(enc, p)
}

func BenchFakeTput(enc *EncodedDatabase, p *Params) float64 {
	fmt.Printf("Tput for ~fake~ DEPIR over database of %d %d-byte-length records: %f GB\n", p.N, p.Record_len, BytesToGB(p.N * p.Record_len))
	cloud := fetchCloud(enc.Poly)

	// Serve the database for fixed time interval, see how many queries answered
	timeUp := false
	var timeMu sync.Mutex

	queriesAnswered := 0
	var answeredMu sync.Mutex

	n_threads := 192 * 10  
	PrintMemUsage()
	fmt.Println("    Tput experiment")

	start := time.Now()
	for i := 0; i < n_threads; i++ {
		go func(i int, cloud []int, enc *EncodedDatabase) {
			for {
				_, query, _ := Query(0, p)

				Answer(enc, cloud, query)

				answeredMu.Lock()
          		queriesAnswered += 1
          		answeredMu.Unlock()

          		timeMu.Lock()
          		if timeUp {
            		timeMu.Unlock()
            		return
          		}
          		timeMu.Unlock()

				runtime.GC()
			}
		}(i, cloud, enc)
	}

	var tput float64
	for j := 0; j < 20; j++ {
		time.Sleep(10 * time.Second) 

		answeredMu.Lock()
		elapsed := time.Since(start)
		tput = float64(queriesAnswered) / elapsed.Seconds()
		fmt.Printf("    Queries answered in %s: %d -- Tput: %f queries/s\n", elapsed, queriesAnswered, tput)
    	answeredMu.Unlock()
	}

	timeMu.Lock()
	timeUp = true
	timeMu.Unlock()

	time.Sleep(10 * time.Second)

	return tput
}
