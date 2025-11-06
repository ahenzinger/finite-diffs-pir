package pir

import (
	"golang.org/x/sys/unix"
)

// Lookup table that encodes the database (i.e., holds the truth table of the database polynomial)
type EncodedDatabase struct {
    Poly       *Params // parameters of the database polynomial
    Data       []byte  // truth table... could also pack into uint64
}

func buildLookupTable(terms [][]bool, coeffs []byte, vars int) []byte {
	if vars == 0 {
		var out byte

		for i, t := range terms {
			all_false := true
			for _, val := range t {
				if val {
					all_false = false
					break
				}
			}

			if all_false {
				out ^= coeffs[i]
			}
		}

		return []byte{out}
	}

	// split terms by whether bit (vars-1) is set
	var t0, t1 [][]bool
	var coeffs0, coeffs1 []byte
	for i, t := range terms {
        if !t[vars-1] {
			t0 = append(t0, t)
			coeffs0 = append(coeffs0, coeffs[i])
        } else {
			t[vars-1] = false
			t1 = append(t1, t) 
			coeffs1 = append(coeffs1, coeffs[i])
        }
	}

	// recurse on the lower m-1 bits
	D0 := buildLookupTable(t0, coeffs0, vars-1)
	D1 := buildLookupTable(t1, coeffs1, vars-1)

    D := make([]byte, (1 << vars))
    half := 1 << (vars-1)

	copy(D, D0) // assignments with top bit = 0

	// stitch back together
	for s := 0; s < half; s++ {
        D[s + half] = (D0[s] ^ D1[s]) // with top bit = 1 we include both groups
    }

	return D
}

func interleave(parts [][]byte) []byte {
	var total int
 	for _, p := range parts {
		total += len(p)
	}

	out := make([]byte, 0, total)

 	for i := 0; i < len(parts[0]); i++ {
        for _, p := range parts {
			if i < len(p) {
				out = append(out, p[i])
			}
		}
	}

	return out
}

func EncodeDatabase(db *Database, p *Params) *EncodedDatabase {
    enc := new(EncodedDatabase)
    enc.Poly = p

	// Encode database, byte by byte (for each record)
	data := make([][]byte, 0)

	for i := 0; i < db.Record_len; i++ {

		// Precompute encoding for every term
		var terms [][]bool
		var coeffs []byte
		for l := 0; l < db.Num_records; l++ {
			terms = append(terms, EncodeIndex(l, p))
			coeffs = append(coeffs, db.Read(l)[i])
		}

		data = append(data, buildLookupTable(terms, coeffs, p.M))
	}

	// Interleave encoded databases (to amortize RAM lookups in online phase)
	enc.Data = interleave(data)

	_ = unix.Madvise(enc.Data, unix.MADV_HUGEPAGE | unix.MADV_RANDOM)

	return enc
}

func FakeEncodeDatabase(p *Params) *EncodedDatabase {
    enc := new(EncodedDatabase)
    enc.Poly = p
    enc.Data = RandByteVec((1 << p.M) * p.Record_len)

	_ = unix.Madvise(enc.Data, unix.MADV_HUGEPAGE | unix.MADV_RANDOM)

    return enc
}

func (enc *EncodedDatabase) Read(index int) []byte {
	at := index * enc.Poly.Record_len

	if at >= len(enc.Data) {
		panic("Read on encoded database is out of bounds")
	}

	return enc.Data[at : at + enc.Poly.Record_len]
}

func (enc *EncodedDatabase) Bytelen() int {
	return len(enc.Data)
}
