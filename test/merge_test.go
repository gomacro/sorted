// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package heap

import (
	"github.com/gomacro/heap/int32/heap"
	"testing"
)

func Int64(a, b *int64) int {
	r := int(*a>>32) - int(*b>>32)
	if r != 0 {
		return r
	}
	return int(*a) - int(*b)
}

func Merge(compar func(*int64, *int64) int, dst []int64, srcs [][]int64) {

	// FOR PARALLEL
	//	cachei := len(dst)/2 - len(srcs)
	cachei := len(dst) - len(srcs)*2

	sli := dst[cachei:]
	sli = sli[:len(srcs)*2]

	// make the cache of starts and ends
	for i, s := range srcs {
		sli[i*2] = s[0] //starts
		sli[i*2+1] = s[len(s)-1]
	}

	// heap of starts
	heapslice := make([]int32, len(srcs))

	// ok now let's make the heap of starts
	hcompar := func(l, r *int32) int {
		//		fmt.Println("COMPARING: ",sli[*l],sli[*r])
		return compar(&sli[*l], &sli[*r])
	}

	for i := range srcs {
		heapslice[i] = int32(2 * i)
	}

	// build a heap
	heap.Heapify(hcompar, heapslice, heapslice)

	// todo: build the second heap from the reversed first heap

	// we also obtain the second smallest item
	heap.Another(hcompar, heapslice)

	dstl := dst
	dstl = dstl[:cachei]
	dst = dstl

	for {

		// check how many objects to copy
		tocopy := 0
		for tocopy = 0; tocopy < len(srcs[heapslice[0]/2]); tocopy++ {
			cmp := compar(&srcs[heapslice[0]/2][tocopy], &sli[heapslice[1]])

			if cmp > 0 {
				break
			}
		}

		// copy to the start
		copy(dstl, srcs[heapslice[0]/2][:tocopy])

		if tocopy >= len(dstl) {
			srcs[heapslice[0]/2] = srcs[heapslice[0]/2][tocopy:]
			break
		}

		dstl = dstl[tocopy:]

		// we must also shrink the slice of object 0

		if len(srcs[heapslice[0]/2]) == tocopy {

			srcs[heapslice[0]/2] = srcs[heapslice[0]/2][:0]

			heap.Remove(hcompar, &heapslice, 0)

			heap.Another(hcompar, heapslice)

		} else {

			// update cached start
			sli[heapslice[0]] = srcs[heapslice[0]/2][tocopy]

			srcs[heapslice[0]/2] = srcs[heapslice[0]/2][tocopy:]

			// bubble object 0

			heap.Fix(hcompar, heapslice, 0)
			heap.Another(hcompar, heapslice)

		}
	}

	for _ = range sli {
		min := 0

		for i := range heapslice {
			if compar(&srcs[heapslice[i]/2][0], &srcs[heapslice[min]/2][0]) < 0 {
				min = i
			}
		}

		dst = append(dst, srcs[heapslice[min]/2][0])

		srcs[heapslice[min]/2] = srcs[heapslice[min]/2][1:]

		if len(srcs[heapslice[min]/2]) == 0 {

			srcs[heapslice[min]/2] = srcs[heapslice[min]/2][:0]

			heap.Remove(hcompar, &heapslice, min)

		}

	}
}

func TestFooMerge0(t *testing.T) {
	var dst []int64
	var srcs [][]int64
	_ = dst

	dstlen := 0

	srcs = append(srcs, []int64{96, 99, 101, 111, 121, 122, 123, 124, 125, 126, 127})
	srcs = append(srcs, []int64{7, 9, 13, 45, 65, 78, 98, 99, 105, 116, 127, 135, 148})
	srcs = append(srcs, []int64{32, 38, 39, 40, 41, 44, 46, 49, 54, 129, 130, 131, 133})

	for _, s := range srcs {
		dstlen += len(s)
	}

	dst = make([]int64, dstlen)

	Merge(Int64, dst, srcs)

	for i := 1; i < len(dst); i++ {
		if dst[i-1] > dst[i] {
			t.Errorf("%v : Item %v > %v", i, dst[i-1], dst[i])
		}
	}
}
