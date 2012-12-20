// Copyright ©2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package kdtree

var (
	_ Interface  = nbPoints{}
	_ Comparable = nbPoint{}
)

// Randoms is the maximum number of random values to sample for calculation of median of
// random elements.
var nbRandoms = 100

// An nbPoint represents a point in a k-d space that satisfies the Comparable interface.
type nbPoint Point

func (p nbPoint) Clone() Comparable                   { return append(nbPoint(nil), p...) }
func (p nbPoint) Compare(c Comparable, d Dim) float64 { q := c.(nbPoint); return p[d] - q[d] }
func (p nbPoint) Dims() int                           { return len(p) }
func (p nbPoint) Distance(c Comparable) float64 {
	q := c.(nbPoint)
	var sum float64
	for dim, c := range p {
		d := c - q[dim]
		sum += d * d
	}
	return sum
}

// An nbPoints is a collection of point values that satisfies the Interface.
type nbPoints []nbPoint

func (p nbPoints) Index(i int) Comparable         { return p[i] }
func (p nbPoints) Len() int                       { return len(p) }
func (p nbPoints) Pivot(d Dim) int                { return nbPlane{nbPoints: p, Dim: d}.Pivot() }
func (p nbPoints) Slice(start, end int) Interface { return p[start:end] }

// An nbPlane is a wrapping type that allows a Points type be pivoted on a dimension.
type nbPlane struct {
	Dim
	nbPoints
}

func (p nbPlane) Less(i, j int) bool              { return p.nbPoints[i][p.Dim] < p.nbPoints[j][p.Dim] }
func (p nbPlane) Pivot() int                      { return Partition(p, MedianOfRandoms(p, nbRandoms)) }
func (p nbPlane) Slice(start, end int) SortSlicer { p.nbPoints = p.nbPoints[start:end]; return p }
func (p nbPlane) Swap(i, j int) {
	p.nbPoints[i], p.nbPoints[j] = p.nbPoints[j], p.nbPoints[i]
}
