package mat64

import (
	"github.com/gonum/blas"
	"github.com/gonum/blas/blas64"
)

var (
	symDense *SymDense

	_ Matrix          = symDense
	_ Symmetric       = symDense
	_ RawSymmmetricer = symDense
)

const (
	ErrUplo = "mat64: blas64.Symmetric not upper"
)

// SymDense is a symmetric matrix that uses Dense storage.
type SymDense struct {
	mat blas64.Symmetric
}

// Symmetric represents a symmetric matrix (where the element at {i, j} equals
// the element at {j, i}). Symmetric matrices are always square.
type Symmetric interface {
	Matrix
	// Symmetric returns the number of rows/columns in the matrix.
	Symmetric() int
}

// A RawSymmetricer can return a view of itself as a BLAS Symmetric matrix.
type RawSymmmetricer interface {
	RawSymmetric() blas64.Symmetric
}

// NewSymDense constructs an n x n symmetric matrix. If len(mat) == n * n,
// mat will be used to hold the underlying data, or if mat == nil, new data will be allocated.
// The underlying data representation is the same as a Dense matrix, except
// the values of the entries in the lower triangular portion are completely ignored.
func NewSymDense(n int, mat []float64) *SymDense {
	if n < 0 {
		panic("mat64: negative dimension")
	}
	if mat != nil && n*n != len(mat) {
		panic(ErrShape)
	}
	if mat == nil {
		mat = make([]float64, n*n)
	}
	return &SymDense{blas64.Symmetric{
		N:      n,
		Stride: n,
		Data:   mat,
		Uplo:   blas.Upper,
	}}
}

func (s *SymDense) Dims() (r, c int) {
	return s.mat.N, s.mat.N
}

func (s *SymDense) Symmetric() int {
	return s.mat.N
}

// RawSymmetric returns the matrix as a blas64.Symmetric. The returned
// value must be stored in upper triangular format.
func (s *SymDense) RawSymmetric() blas64.Symmetric {
	return s.mat
}

func (s *SymDense) IsZero() bool {
	return s.mat.N == 0
}

func (s *SymDense) CopySym(a Symmetric) int {
	n := a.Symmetric()
	n = min(n, s.mat.N)
	switch a := a.(type) {
	case RawSymmmetricer:
		amat := a.RawSymmetric()
		if amat.Uplo != blas.Upper {
			panic(ErrUplo)
		}
		for i := 0; i < n; i++ {
			copy(s.mat.Data[i*s.mat.Stride+i:i*s.mat.Stride+n], amat.Data[i*amat.Stride+i:i*amat.Stride+n])
		}
	default:
		for i := 0; i < n; i++ {
			stmp := s.mat.Data[i*s.mat.Stride : i*s.mat.Stride+n]
			for j := i; j < n; j++ {
				stmp[j] = a.At(i, j)
			}
		}
	}
	return n
}

// RankTwo performs a symmmetric rank-two update to the matrix a and stores
// the result in the receiver
//  m = a + alpha * (x * y' + y * x')
func (s *SymDense) RankTwo(a Symmetric, alpha float64, x, y []float64) {
	n := s.mat.N
	var w SymDense
	if s == a {
		w = *s
	}
	if w.IsZero() {
		w.mat = blas64.Symmetric{
			N:      n,
			Stride: n,
			Uplo:   blas.Upper,
			Data:   use(w.mat.Data, n*n),
		}
	} else if n != w.mat.N {
		panic(ErrShape)
	}
	if s != a {
		w.CopySym(a)
	}
	if len(x) != n {
		panic(ErrShape)
	}
	if len(y) != n {
		panic(ErrShape)
	}
	blas64.Syr2(alpha, blas64.Vector{Inc: 1, Data: x}, blas64.Vector{Inc: 1, Data: y}, w.mat)
	*s = w
	return
}
