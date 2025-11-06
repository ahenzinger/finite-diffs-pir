package pir

import (
	"fmt"
	"slices"
	"testing"
	"runtime/pprof"
)

func testEncoding(t *testing.T, N, record_len int, theta float64) {
    db := RandDatabase(N, record_len)
    p := PickParams(N, record_len, theta)
    enc := EncodeDatabase(db, p)

    for i := 0; i < N; i++ {
        vec := EncodeIndex(i, p)
        index := EncodingToIndex(vec)
        record := enc.Read(index)

        if !slices.Equal(record, db.Read(i)) {
            fmt.Printf("Position %d: read %b, should be %b\n", i, record, db.Read(i))
            t.Fail()
            panic("Encoding failed")
        }
    }
}

func TestEncoding1(t *testing.T) {
    fmt.Println("TestEncodingSmall")
    testEncoding(t, 95, 1, 0.5)
}

func TestEncoding10(t *testing.T) {
    fmt.Println("TestEncodingSmall")
    testEncoding(t, 96, 10, 0.5)
}

func TestEncoding100(t *testing.T) {
    fmt.Println("TestEncodingSmall")
    testEncoding(t, 97, 100, 0.5)
}

func TestEncoding1024(t *testing.T) {
    fmt.Println("TestEncodingSmall")
    testEncoding(t, 98, 1024, 0.5)
}

func TestEncoding10240(t *testing.T) {
    fmt.Println("TestEncodingSmall")
    testEncoding(t, 99, 10240, 0.5)
}

func TestEncoding102400(t *testing.T) {
    fmt.Println("TestEncodingSmall")
    testEncoding(t, 10, 102400, 0.5)
}

func TestEncodingSmall(t *testing.T) {
	fmt.Println("TestEncodingSmall")
	testEncoding(t, 10, 1, 0.5)
}

//func TestEncodingMed(t *testing.T) {
//        fmt.Println("TestEncodingMed")
//        testEncoding(t, 1000 * 8, 1, 0.5) 
//}

func testPIR(t *testing.T, N, record_len int, theta float64) {
	RunPIR(N, record_len, theta)
}

func TestPIRSmall1(t *testing.T) {
        fmt.Println("TestPIRSmall1")
        testPIR(t, 100, 1, 0.5)
}

func TestPIRSmall10(t *testing.T) {
        fmt.Println("TestPIRSmall10")
        testPIR(t, 100, 10, 0.5)
}

func TestPIRMed1(t *testing.T) {
        fmt.Println("TestPIRMed1")
        testPIR(t, 1024 * 8, 1, 0.5)
}

func TestPIRMed10(t *testing.T) {
        fmt.Println("TestPIRMed10")
        testPIR(t, 1024, 10, 0.5)
}

func TestPIRMed100(t *testing.T) {
        fmt.Println("TestPIRMed100")
        testPIR(t, 1024, 100, 0.5)
}

func TestPIRMed1024(t *testing.T) {
        fmt.Println("TestPIRMed1024")
        testPIR(t, 1024, 1024, 0.5)
}

func TestPIRMed10240(t *testing.T) {
        fmt.Println("TestPIRMed10240")
        testPIR(t, 1024 * 7, 10240, 0.5)
}

//func TestPIRMed102400(t *testing.T) {
//       fmt.Println("TestPIRMed102400")
//       testPIR(t, 1024 * 3, 102400, 0.5)
//}
/*
func TestPIRBig1(t *testing.T) {
    fmt.Println("TestPIRBig1")
    testPIR(t, 100 * 1024 * 8, 1, 0.5)
}

func TestPIRBig10(t *testing.T) {
    fmt.Println("TestPIRBig10")
    testPIR(t, 10 * 1024 * 8, 10, 0.5)
}*/

func testFakePIR(t *testing.T, N, record_len int, theta float64) {
    RunFakePIR(N, record_len, theta)
}

func TestFakePIRBig1(t *testing.T) {
    f := ProfileCPU("cpu_test.prof")
    defer f.Close()
    defer pprof.StopCPUProfile()

    fmt.Println("TestFakePIR1")
    testFakePIR(t, 128 * 1024 * 1024, 1, 0.5)
}

func TestFakePIRBig10(t *testing.T) {
    f := ProfileCPU("cpu_test.prof")
	defer f.Close()
	defer pprof.StopCPUProfile()

    fmt.Println("TestFakePIR10")
    testFakePIR(t, 1024 * 1024, 2000, 0.5)
}

// shrink N to make the GB of the database the largest integer <= N
func shrinkDBToIntegerSz(p *Params) {
    dbsz := int(BytesToGB(p.N * p.Record_len))
    p.N = (dbsz * 1024 * 1024 * 1024) / p.Record_len
}

func BenchmarkLatencyPIR(b *testing.B) {
    fmt.Println("Benchmark DEPIR, latency")
	p := new(Params)
    p.M, p.D = 40, 19
    p.N = Binomial(p.M, p.D)
	p.Record_len = 1
	shrinkDBToIntegerSz(p)

	enc := FakeEncodeDatabase(p)
	RunFakePIRWithParams(enc, p)
}

func BenchmarkTputPIR(b *testing.B) {
	fmt.Println("Benchmark DEPIR, tput")

	p := new(Params) 
    p.M, p.D = 35, 17
    p.N = Binomial(p.M, p.D)
    p.Record_len = 32
	shrinkDBToIntegerSz(p)

	enc := FakeEncodeDatabase(p)
	BenchFakeTput(enc, p)
}
