package pir 

// Performance logging code

import (
	"os"
	"fmt"
	"strconv"
	"encoding/csv"
)

// Parameters for the database polynomial, over F_2
// Since the polynomial is over F_2, it must be multilinear
type Params struct {
    M int          // num variables
    D int          // total degree
    N int          // total number of points (i.e., entries in the DB)
    Record_len int
}

type Metrics struct {
	p           *Params
	avg_time    float64
	stddev_time float64
	tput        float64
	download    int
}

type Perf struct {
	dbsz      int64
	recordlen int64

	encsz     int64
	theta     float64

	upload    int64
	download  int64

	time      float64
	stddev    float64

	tput      float64

    blowup    float64
    speedup   float64
	tputgains float64

	records   []string
}

type CSVLogger struct {
	file   *os.File
	writer *csv.Writer
}

func NewMetric(p *Params, times []float64, down int) *Metrics {
	m := &Metrics{
        p: p,
        avg_time: Avg(times),
        stddev_time: Stddev(times),
        download: down,
		tput: 0,
    }
	return m
}

func (m *Metrics) SetTput(v float64) {
	m.tput = v
}

func NewCSVLogger(path string) *CSVLogger {
	f, err := os.OpenFile(path, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		panic("Error")
	}

	w := csv.NewWriter(f)

	header := []string{
		"Num_records", 
		"Record_len (byte)",
		"DB size (byte)",
		"Encoding size (byte)",
		"M",
		"D",
		"Avg time/query (s)",
		"Std dev time/query (s)",
		"Avg time for linear scan (s)",
		"Std dev linear scan (s)",
		"Download (byte)",
		"Tput (q/s)",
		"Linear scan tput (q/s)",
		"Blowup (relative)",
		"Speedup (relative)",
		"Tput gains (relative)",
	}

	if err := w.Write(header); err != nil {
		fmt.Println(err)
		panic("Error")
	}

	w.Flush()
	return &CSVLogger{file: f, writer: w}
}

func (c *CSVLogger) Log(m_sublinear, m_xor *Metrics) {
	if (m_sublinear.p.N != m_xor.p.N) || (m_sublinear.p.Record_len != m_xor.p.Record_len) {
		fmt.Println(m_sublinear)
		fmt.Println(m_xor)
		panic("Parameters do not match!")
	}

	row := []string{
		strconv.Itoa(m_sublinear.p.N),
		strconv.Itoa(m_sublinear.p.Record_len),
		strconv.Itoa(m_sublinear.p.N * m_sublinear.p.Record_len),
		strconv.Itoa((1 << m_sublinear.p.M) * m_sublinear.p.Record_len),
		strconv.Itoa(m_sublinear.p.M),
		strconv.Itoa(m_sublinear.p.D),
		strconv.FormatFloat(m_sublinear.avg_time, 'f', 5, 64),
		strconv.FormatFloat(m_sublinear.stddev_time, 'f', 5, 64),
		strconv.FormatFloat(m_xor.avg_time, 'f', 5, 64),
		strconv.FormatFloat(m_xor.stddev_time, 'f', 5, 64),
		strconv.Itoa(m_sublinear.download),
		strconv.FormatFloat(m_sublinear.tput, 'f', 5, 64),
		strconv.FormatFloat(m_xor.tput, 'f', 5, 64),
		strconv.FormatFloat(float64((uint64(1) << m_sublinear.p.M)) / float64(m_sublinear.p.N),'f', 5, 64),
		strconv.FormatFloat(m_xor.avg_time/m_sublinear.avg_time, 'f', 5, 64),
		strconv.FormatFloat(m_sublinear.tput/m_xor.tput, 'f', 5, 64),
	}

	if err := c.writer.Write(row); err != nil {
		panic(err)
	}

	if m_sublinear.avg_time < m_xor.avg_time {
        Greenln("    time is better by a factor of: %f \n", m_xor.avg_time/m_sublinear.avg_time) // 31 GB/s
    } else {
        Redln("    time is worse by a factor of: %f \n", m_xor.avg_time/m_sublinear.avg_time) // 31 GB/s
    }

	if m_sublinear.tput > m_xor.tput {
        Greenln("    tput is better by a factor of: %f \n", m_sublinear.tput/m_xor.tput)
    } else {
        Redln("    tput is worse by a factor of: %f \n", m_sublinear.tput/m_xor.tput) 
    }

	c.writer.Flush()
}

func (c *CSVLogger) Close() {
	c.writer.Flush()

	if err := c.writer.Error(); err != nil {
		_ = c.file.Close()
		panic(err)
	}

	c.file.Close()
}

func ParseLog(file string) {
	f, err := os.Open(file)
	if err != nil {
		panic("Error opening file")
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		panic("Error reading CSV")
	}

	header := records[0]
	colIndex := make(map[string]int)
	for i, h := range header {
		colIndex[h] = i
	}

	frontier := make(map[int]Perf)
	for i, rec := range records {
		if i == 0 {
			continue
		}

		var p Perf
		var err1, err2, err3, err4, err5, err6, err7 error
		p.speedup, err1   = strconv.ParseFloat(rec[colIndex["Speedup (relative)"]], 64)
		p.encsz, err2     = strconv.ParseInt(rec[colIndex["Encoding size (byte)"]], 10, 64)
		p.dbsz, err3      = strconv.ParseInt(rec[colIndex["DB size (byte)"]], 10, 64)
		p.download, err4  = strconv.ParseInt(rec[colIndex["Download (byte)"]], 10, 64)
		p.time, err5      = strconv.ParseFloat(rec[colIndex["Avg time/query (s)"]], 64)
		p.stddev, err6    = strconv.ParseFloat(rec[colIndex["Std dev time/query (s)"]], 64)
		p.recordlen, err7 = strconv.ParseInt(rec[colIndex["Record_len (byte)"]], 10, 64)
		m, err8          := strconv.ParseInt(rec[colIndex["M"]], 10, 64)
		D, err9          := strconv.ParseInt(rec[colIndex["D"]], 10, 64) 

		if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil || err8 != nil || err9 != nil { 
			panic("Invalid data")
		}

		if p.speedup <= 1.0 {
			continue
		}

		p.blowup = float64(p.encsz)/float64(p.dbsz)
		p.theta = float64(D)/float64(m)
		p.records = rec
		p.upload = m

		subsumed := false

		for j, perf := range frontier {
			if perf.speedup > p.speedup && perf.blowup < p.blowup {
				subsumed = true
			}

			if perf.speedup < p.speedup && perf.blowup > p.blowup {
				delete(frontier, j)
			}
		}

		if !subsumed {
			frontier[i] = p
		}
	}
	
	for _, p := range frontier {
		fmt.Println("Record: ")
		fmt.Println(header)
		fmt.Println(p.records)
		Greenln("   SPEED-UP: %f\n ", p.speedup)
		Redln("  BLOW-UP:  %f\n ", p.blowup)
		fmt.Printf("     DB sz:    %f with record os length %d bytes\n", BytesToGB(int(p.dbsz)), p.recordlen)
		fmt.Printf("      Enc sz:    %f\n", BytesToGB(int(p.encsz)))
		fmt.Printf("      theta:    %f\n", p.theta)
		fmt.Printf("      upload:    %d bits\n", p.upload)
		fmt.Printf("      download:    %f KB = %f MB\n", BytesToKB(int(p.download)), BytesToMB(int(p.download)))
		fmt.Printf("      time:     %f +/- %f s\n\n", p.time, p.stddev)
	}
}
