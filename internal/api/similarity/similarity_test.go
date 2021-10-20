//TODO: move tests input to some binary files
package similarity

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var tolerance = .00001
var opt = cmp.Comparer(func(x, y float32) bool {
	diff := math.Abs(float64(x - y))
	return diff < tolerance
})

type FakeResponseReader struct {
	rows []PostTagLink
	last int
}

func (reader *FakeResponseReader) Next(dest ...interface{}) (bool, error) {
	result := false
	if reader.last < len(reader.rows) {
		*dest[0].(*int64) = reader.rows[reader.last].PostID
		*dest[1].(*int64) = reader.rows[reader.last].TagID
		reader.last++
		result = true
	}
	return result, nil
}

func (reader *FakeResponseReader) Close() {}

func TestNormalizeVector(t *testing.T) {
	tests := []struct {
		name string
		args map[int64]float32
		want map[int64]float32
	}{
		{
			name: "hand-made test",
			args: map[int64]float32{1: 3, 2: 4, 3: 5},
			want: map[int64]float32{1: 0.424264, 2: 0.565685, 3: 0.707106},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.args
			normalizeVector(result)

			if !cmp.Equal(result, test.want, opt) {
				t.Fatalf("normalizeVector() result %v, wanted %v", result, test.want)
			}
		})
	}
}

func TestReadUserProfile(t *testing.T) {
	tests := []struct {
		name string
		args FakeResponseReader
		want map[int64]float32
		err  error
	}{
		{
			name: "hand-made test",
			args: FakeResponseReader{
				rows: []PostTagLink{
					{1, 1}, {1, 2},
					{3, 1}, {3, 3},
					{5, 1}, {5, 4},
				},
				last: 0,
			},
			want: map[int64]float32{1: 0.866025, 2: 0.288675, 3: 0.288675, 4: 0.288675},
			err:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := readUserProfile(&test.args)

			if !cmp.Equal(err, test.err, opt) {
				t.Fatalf("readUserProfile() error %v, wanted %v", err, test.err)
			}

			if !cmp.Equal(result, test.want, opt) {
				t.Fatalf("readUserProfile() result %v, wanted %v", result, test.want)
			}
		})
	}
}

func TestReadPostsTags(t *testing.T) {
	tests := []struct {
		name string
		args FakeResponseReader
		want map[int64]map[int64]float32
		err  error
	}{
		{
			name: "hand-made test",
			args: FakeResponseReader{
				rows: []PostTagLink{
					{2, 1}, {2, 2},
					{4, 1}, {4, 5},
					{6, 2}, {6, 6},
					{7, 7}, {7, 8},
				},
				last: 0,
			},
			want: map[int64]map[int64]float32{
				2: {1: 0.707107, 2: 0.707107},
				4: {1: 0.707107, 5: 0.707107},
				6: {2: 0.707107, 6: 0.707107},
				7: {7: 0.707107, 8: 0.707107},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := readPostsTags(&test.args)

			if !cmp.Equal(err, test.err, opt) {
				t.Fatalf("readPostsTags() error %v, wanted %v", err, test.err)
			}

			if !cmp.Equal(result, test.want, opt) {
				t.Fatalf("readPostsTags() result %v, wanted %v", result, test.want)
			}
		})
	}
}

func TestGetSimilarPosts(t *testing.T) {
	tests := []struct {
		name    string
		readers ProfilesReaders
		user    string
		top     int
		want    []PostSimilarity
		err     error
	}{
		{
			name: "hand-made test",
			readers: ProfilesReaders{
				UserProfileReader: &FakeResponseReader{
					rows: []PostTagLink{
						{1, 1}, {1, 2},
						{3, 1}, {3, 3},
						{5, 1}, {5, 4},
					},
					last: 0,
				},
				PostsTagsReader: &FakeResponseReader{
					rows: []PostTagLink{
						{2, 1}, {2, 2},
						{4, 1}, {4, 5},
						{6, 2}, {6, 6},
						{7, 7}, {7, 8},
					},
					last: 0,
				},
			},
			top:  3,
			want: []PostSimilarity{{2, 0.816496}, {4, 0.612372}, {6, 0.204124}},
			err:  nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := getSimilarPosts(test.readers, test.top)

			if !cmp.Equal(err, test.err, opt) {
				t.Fatalf("GetSimilarPosts() error %v, wanted %v", err, test.err)
			}

			if !cmp.Equal(result, test.want, opt) {
				t.Fatalf("GetSimilarPosts() result %v, wanted %v", result, test.want)
			}
		})
	}
}