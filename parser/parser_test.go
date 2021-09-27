package parser

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/st-matskevich/item-based-recommendations/model"
)

var tolerance = .00001
var opt = cmp.Comparer(func(x, y float32) bool {
	diff := math.Abs(float64(x - y))
	return diff < tolerance
})

func TestParseUserProfile(t *testing.T) {
	////TODO: move tests input to some binary files
	tests := []struct {
		name string
		args []model.PostTagLink
		want map[int]float32
	}{
		{
			name: "normal test",
			args: []model.PostTagLink{
				{1, 1}, {1, 2},
				{3, 1}, {3, 3},
				{5, 1}, {5, 4},
			},
			want: map[int]float32{1: 0.866025, 2: 0.288675, 3: 0.288675, 4: 0.288675},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ParseUserProfile(test.args)

			if !cmp.Equal(result, test.want, opt) {
				t.Fatalf("ParseUserProfile() result %v, wanted %v", result, test.want)
			}
		})
	}
}

func TestParsePostsTags(t *testing.T) {
	////TODO: move tests input to some binary files
	tests := []struct {
		name string
		args []model.PostTagLink
		want map[int]map[int]float32
	}{
		{
			name: "normal test",
			args: []model.PostTagLink{
				{2, 1}, {2, 2},
				{4, 1}, {4, 5},
				{6, 2}, {6, 6},
				{7, 7}, {7, 8},
			},
			want: map[int]map[int]float32{
				2: {1: 0.707107, 2: 0.707107},
				4: {1: 0.707107, 5: 0.707107},
				6: {2: 0.707107, 6: 0.707107},
				7: {7: 0.707107, 8: 0.707107},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ParsePostsTags(test.args)

			if !cmp.Equal(result, test.want, opt) {
				t.Fatalf("ParsePostsTags() result %v, wanted %v", result, test.want)
			}
		})
	}
}
