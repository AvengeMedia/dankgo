package num

import "fmt"

// Matrix3 defines a 3x3 matrix of float64 values.
type Matrix3 [9]float64

// NewMatrix3 creates a 3x3 Matrix3.
func NewMatrix3(x1, y1, z1, x2, y2, z2, x3, y3, z3 float64) Matrix3 {
	return Matrix3{
		x1, y1, z1,
		x2, y2, z2,
		x3, y3, z3,
	}
}

// Row returns the n-th row of the matrix.
func (m Matrix3) Row(n int) Vector3 {
	return Vector3{
		m[n*3],
		m[n*3+1],
		m[n*3+2],
	}
}

// Column returns the n-th column of the matrix.
func (m Matrix3) Column(n int) Vector3 {
	return Vector3{
		m[n],
		m[n+3],
		m[n+6],
	}
}

// Mul applies the matrix to a 3D vector and returns the resulting vector.
func (m Matrix3) Mul(v Vector3) Vector3 {
	return Vector3{
		m[0]*v[0] + m[1]*v[1] + m[2]*v[2],
		m[3]*v[0] + m[4]*v[1] + m[5]*v[2],
		m[6]*v[0] + m[7]*v[1] + m[8]*v[2],
	}
}

// Transpose transposes the Matrix3.
func (m Matrix3) Transpose() Matrix3 {
	return Matrix3{
		m[0], m[3], m[6],
		m[1], m[4], m[7],
		m[2], m[5], m[8],
	}
}

// Inverse inverses the M Matrix3. Returns false if m Matrix3 is not inversable.
func (m Matrix3) Inverse() (Matrix3, bool) {
	a, b, c := m[0], m[1], m[2]
	d, e, f := m[3], m[4], m[5]
	g, h, i := m[6], m[7], m[8]

	// Compute the determinant
	det := a*(e*i-f*h) - b*(d*i-f*g) + c*(d*h-e*g)
	if det == 0 {
		return Matrix3{}, false // Matrix is not invertible
	}
	invDet := 1.0 / det

	var inv Matrix3
	inv[0] = (e*i - f*h) * invDet
	inv[1] = -(b*i - c*h) * invDet
	inv[2] = (b*f - c*e) * invDet
	inv[3] = -(d*i - f*g) * invDet
	inv[4] = (a*i - c*g) * invDet
	inv[5] = -(a*f - c*d) * invDet
	inv[6] = (d*h - e*g) * invDet
	inv[7] = -(a*h - b*g) * invDet
	inv[8] = (a*e - b*d) * invDet

	return inv, true
}

// String implements Stringer method.
func (m Matrix3) String() string {
	return fmt.Sprintf(
		"[\n\t%.10f,%.10f,%.10f,\n\t%.10f,%.10f,%.10f,\n\t%.10f,%.10f,%.10f,\n]",
		m[0], m[1], m[2],
		m[3], m[4], m[5],
		m[6], m[7], m[8],
	)
}

// Vector3 defines a 3D vector.
type Vector3 [3]float64

// NewVector3 create new 3D vector: Vector3.
func NewVector3(x, y, z float64) Vector3 {
	return Vector3{x, y, z}
}

// VectorLike is any data that has 3 float64 values. Usually a color model.
type VectorLike interface {
	Values() (float64, float64, float64)
}

// set build constrains
var _ VectorLike = (*Vector3)(nil)

// NewVector create new 3D vector: Vector3.
func NewVector(v VectorLike) Vector3 {
	return NewVector3(v.Values())
}

// Map returns a new Vector3 where each component of v has been
// transformed by the given func f. Returns a new Vector3.
func (v Vector3) Map(f func(float64) float64) Vector3 {
	return Vector3{
		f(v[0]),
		f(v[1]),
		f(v[2]),
	}
}

// Scaled multiplies all values using a scalar float64 and retuns a new
// Vector3.
func (v Vector3) Scaled(s float64) Vector3 {
	return Vector3{
		v[0] * s,
		v[1] * s,
		v[2] * s,
	}
}

// Dot multiplies v Vector3 with given Vector3 and returns a float64.
func (v Vector3) Dot(vec Vector3) float64 {
	return v[0]*vec[0] + v[1]*vec[1] + v[2]*vec[2]
}

// MulElems multiplies v Vector3 with given Vector3 element-wise and returns a new Vector3.
func (v Vector3) MulElems(vec Vector3) Vector3 {
	return Vector3{
		v[0] * vec[0],
		v[1] * vec[1],
		v[2] * vec[2],
	}
}

// Add adds values of v Vector3 with given Vector3 and returns new Vector3.
func (v Vector3) Add(vec Vector3) Vector3 {
	return Vector3{
		v[0] + vec[0],
		v[1] + vec[1],
		v[2] + vec[2],
	}
}

// Values retuns all values of underlying Vector3.
func (v Vector3) Values() (float64, float64, float64) {
	return v[0], v[1], v[2]
}

func (v Vector3) String() string {
	return fmt.Sprintf("[%.10f, %.10f, %.10f]", v[0], v[1], v[2])
}
