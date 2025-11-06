package pir

import (
	"fmt"
)


// Returns (M, D) for representing a database of 'num_bits' bits as a polynomial
// Want D = theta * M, roughly. Also: D must be odd
func PickParams(num_records, record_len int, theta float64) *Params {
        if theta == 0 || theta > 0.5 {
                panic("PickParams: bad theta")
        }

        // Initial guess
		p := new(Params)
		p.Record_len = record_len
		p.N = num_records
		p.M = 10
        found_good_config := false

        // Iterative approximation for best M, D with binary search
        for {
                p.D = int(float64(p.M) * theta)
				if p.D % 2 == 0 {
					p.D += 1
				}

                N := Binomial(p.M, p.D)

                if N >= num_records {
						// now, need to decrease M and D
						found_good_config = true
						break
                } else {
                        // Make M a little bigger, see if works
                        p.M += 5
                }
        }

        if !found_good_config {
                panic("Never found good config")
        }

		for {
				newM := p.M - 1
				newD := int(float64(newM) * theta)
                if newD % 2 == 0 {
                        newD += 1
                }

				N := Binomial(newM, newD)

				if N >= num_records {
					p.M = newM
					p.D = newD
				} else {
					return p
				}
		}

		panic("Unreachable")
        return p
}

// EncodeIndex returns the i-th binary vector of length m with exactly D ones,
// in lexicographic order where 0 < 1. i must be < (m \choose D).
func EncodeIndex(i int, p *Params) []bool {
        total := Binomial(p.M, p.D)
        if i >= total {
                fmt.Printf("EncodeIndex: index %d out of range [0..%d)", i, total)
                panic("Failed")
        }

        res := make([]bool, p.M)
        remainOnes := p.D

        for pos := 0; pos < p.M; pos++ {
                if remainOnes == 0 {
                        break
                }

                // Count how many vectors start with a 0 here:
                // ((m-pos-1) \choose remainOnes).
                zerosBlock := Binomial(p.M-pos-1, remainOnes)

                if i < zerosBlock {
                        // our index is in the block of vectors beginning with 0
                        res[pos] = false
                } else {
                        // skip over the zeros-block and place a 1
                        i -= zerosBlock
                        res[pos] = true
                        remainOnes--
                }
        }

        return res
}

func EncodingToIndex(vec []bool) int {
		if len(vec) >= 63 {
			panic("M is >= 63. Not supported by this implementation!")
		}

        out := 0
        for i := 0; i < len(vec); i++ {
                if vec[i] {
                        out += (1 << i)
                }
        }

        return out
}
