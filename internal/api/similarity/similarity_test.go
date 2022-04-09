//TODO: move tests input to some binary files
package similarity

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
)

var tolerance = .00001
var opt = cmp.Comparer(func(x, y float32) bool {
	diff := math.Abs(float64(x - y))
	return diff < tolerance
})

type FakeResponseReader struct {
	rows []TaskTagLink
	last int
}

func (reader *FakeResponseReader) NextRow(dest ...interface{}) (bool, error) {
	result := false
	if reader.last < len(reader.rows) {
		*dest[0].(*utils.UID) = reader.rows[reader.last].TaskID
		*dest[1].(*utils.UID) = reader.rows[reader.last].TagID
		reader.last++
		result = true
	}
	return result, nil
}

func (reader *FakeResponseReader) GetRow(dest ...interface{}) error { return nil }
func (reader *FakeResponseReader) Close()                           {}

func TestNormalizeVector(t *testing.T) {
	tests := []struct {
		name string
		args map[utils.UID]float32
		want map[utils.UID]float32
	}{
		{
			name: "hand-made test",
			args: map[utils.UID]float32{1: 3, 2: 4, 3: 5},
			want: map[utils.UID]float32{1: 0.424264, 2: 0.565685, 3: 0.707106},
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
		want map[utils.UID]float32
		err  error
	}{
		{
			name: "hand-made test",
			args: FakeResponseReader{
				rows: []TaskTagLink{
					{1, 1}, {1, 2},
					{3, 1}, {3, 3},
					{5, 1}, {5, 4},
				},
				last: 0,
			},
			want: map[utils.UID]float32{1: 0.866025, 2: 0.288675, 3: 0.288675, 4: 0.288675},
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

func TestReadTasksTags(t *testing.T) {
	tests := []struct {
		name string
		args FakeResponseReader
		want map[utils.UID]map[utils.UID]float32
		err  error
	}{
		{
			name: "hand-made test",
			args: FakeResponseReader{
				rows: []TaskTagLink{
					{2, 1}, {2, 2},
					{4, 1}, {4, 5},
					{6, 2}, {6, 6},
					{7, 7}, {7, 8},
				},
				last: 0,
			},
			want: map[utils.UID]map[utils.UID]float32{
				2: {1: 0.707106, 2: 0.707106},
				4: {1: 0.894427, 5: 0.447213},
				6: {2: 0.894427, 6: 0.447213},
				7: {7: 0.707106, 8: 0.707106},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := readTasksTags(&test.args)

			if !cmp.Equal(err, test.err, opt) {
				t.Fatalf("readTasksTags() error %v, wanted %v", err, test.err)
			}

			if !cmp.Equal(result, test.want, opt) {
				t.Fatalf("readTasksTags() result %v, wanted %v", result, test.want)
			}
		})
	}
}

func TestGetSimilarTasks(t *testing.T) {
	tests := []struct {
		name      string
		readers   ProfilesReaders
		user      string
		threshold float32
		want      []TaskSimilarity
		err       error
	}{
		{
			name: "hand-made test",
			readers: ProfilesReaders{
				UserProfileReader: &FakeResponseReader{
					rows: []TaskTagLink{
						{1, 1}, {1, 2},
						{3, 1}, {3, 3},
						{5, 1}, {5, 4},
					},
					last: 0,
				},
				TasksTagsReader: &FakeResponseReader{
					rows: []TaskTagLink{
						{2, 1}, {2, 2},
						{4, 1}, {4, 5},
						{6, 2}, {6, 6},
						{7, 7}, {7, 8},
					},
					last: 0,
				},
			},
			threshold: 0.60,
			want:      []TaskSimilarity{{2, 0.816496}, {4, 0.774596}},
			err:       nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := getSimilarTasks(test.readers, test.threshold)

			if !cmp.Equal(err, test.err, opt) {
				t.Fatalf("GetSimilarTasks() error %v, wanted %v", err, test.err)
			}

			if !cmp.Equal(result, test.want, opt) {
				t.Fatalf("GetSimilarTasks() result %v, wanted %v", result, test.want)
			}
		})
	}
}
