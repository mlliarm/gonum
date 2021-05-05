// Copyright ©2021 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gonum

import (
	"math"

	"gonum.org/v1/gonum/internal/asm/f64"
	"gonum.org/v1/gonum/lapack"
)

// Dlangb returns the given norm of an n×n band matrix with kl sub-diagonals and
// ku super-diagonals.
//
// When norm is lapack.MaxColumnSum, the length of work must be at least n.
func (impl Implementation) Dlangb(norm lapack.MatrixNorm, n, kl, ku int, ab []float64, ldab int, work []float64) float64 {
	ncol := kl + 1 + ku
	switch {
	case norm != lapack.MaxAbs && norm != lapack.MaxRowSum && norm != lapack.MaxColumnSum && norm != lapack.Frobenius:
		panic(badNorm)
	case n < 0:
		panic(nLT0)
	case kl < 0:
		panic(klLT0)
	case ku < 0:
		panic(kuLT0)
	case ldab < ncol:
		panic(badLdA)
	}

	// Quick return if possible.
	if n == 0 {
		return 0
	}

	switch {
	case len(ab) < (n-1)*ldab+ncol:
		panic(shortAB)
	case len(work) < n && norm == lapack.MaxColumnSum:
		panic(shortWork)
	}

	var value float64
	switch norm {
	case lapack.MaxAbs:
		for i := 0; i < n; i++ {
			l := max(0, kl-i)
			u := min(n+kl-i, ncol)
			for _, aij := range ab[i*ldab+l : i*ldab+u] {
				aij = math.Abs(aij)
				if aij > value || math.IsNaN(aij) {
					value = aij
				}
			}
		}
	case lapack.MaxRowSum:
		for i := 0; i < n; i++ {
			l := max(0, kl-i)
			u := min(n+kl-i, ncol)
			sum := f64.L1Norm(ab[i*ldab+l : i*ldab+u])
			if sum > value || math.IsNaN(sum) {
				value = sum
			}
		}
	case lapack.MaxColumnSum:
		work = work[:n]
		for j := range work {
			work[j] = 0
		}
		for i := 0; i < n; i++ {
			l := max(0, kl-i)
			u := min(n+kl-i, ncol)
			for jb, aij := range ab[i*ldab+l : i*ldab+u] {
				j := l + jb - kl + i
				work[j] += math.Abs(aij)
			}
		}
		for _, sumj := range work {
			if sumj > value || math.IsNaN(sumj) {
				value = sumj
			}
		}
	case lapack.Frobenius:
		scale := 0.0
		ssq := 1.0
		for i := 0; i < n; i++ {
			l := max(0, kl-i)
			u := min(n+kl-i, ncol)
			ilen := u - l
			rowscale, rowssq := impl.Dlassq(ilen, ab[i*ldab+l:], 1, 0, 1)
			scale, ssq = impl.Dcombssq(scale, ssq, rowscale, rowssq)
		}
		value = scale * math.Sqrt(ssq)
	}
	return value
}
