//TODO: move tests input to some binary files
package similarity

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/st-matskevich/item-based-recommendations/db"
)

var tolerance = .00001
var opt = cmp.Comparer(func(x, y float32) bool {
	diff := math.Abs(float64(x - y))
	return diff < tolerance
})

type FakePostTagLinkReader struct {
	rows []db.PostTagLink
	last int
}

func (fetcher *FakePostTagLinkReader) Next(data *db.PostTagLink) (bool, error) {
	result := false
	if fetcher.last < len(fetcher.rows) {
		*data = fetcher.rows[fetcher.last]
		fetcher.last++
		result = true
	}
	return result, nil
}

type FakeProfilesFetcher struct {
	userProfileReader, postsTagsReader FakePostTagLinkReader
	err                                error
}

func (client *FakeProfilesFetcher) GetUserProfile(id string) (db.PostTagLinkReader, error) {
	return &client.userProfileReader, client.err
}

func (client *FakeProfilesFetcher) GetPostsTags(id string) (db.PostTagLinkReader, error) {
	return &client.postsTagsReader, client.err
}

func TestNormalizeVector(t *testing.T) {
	tests := []struct {
		name string
		args map[int]float32
		want map[int]float32
	}{
		{
			name: "hand-made test",
			args: map[int]float32{1: 3, 2: 4, 3: 5},
			want: map[int]float32{1: 0.424264, 2: 0.565685, 3: 0.707106},
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
		args FakePostTagLinkReader
		want map[int]float32
		err  error
	}{
		{
			name: "hand-made test",
			args: FakePostTagLinkReader{
				rows: []db.PostTagLink{
					{1, 1}, {1, 2},
					{3, 1}, {3, 3},
					{5, 1}, {5, 4},
				},
				last: 0,
			},
			want: map[int]float32{1: 0.866025, 2: 0.288675, 3: 0.288675, 4: 0.288675},
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
		args FakePostTagLinkReader
		want map[int]map[int]float32
		err  error
	}{
		{
			name: "hand-made test",
			args: FakePostTagLinkReader{
				rows: []db.PostTagLink{
					{2, 1}, {2, 2},
					{4, 1}, {4, 5},
					{6, 2}, {6, 6},
					{7, 7}, {7, 8},
				},
				last: 0,
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
		fetcher FakeProfilesFetcher
		user    string
		top     int
		want    []PostSimilarity
		err     error
	}{
		{
			name: "hand-made test",
			fetcher: FakeProfilesFetcher{
				userProfileReader: FakePostTagLinkReader{
					rows: []db.PostTagLink{
						{1, 1}, {1, 2},
						{3, 1}, {3, 3},
						{5, 1}, {5, 4},
					},
					last: 0,
				},
				postsTagsReader: FakePostTagLinkReader{
					rows: []db.PostTagLink{
						{2, 1}, {2, 2},
						{4, 1}, {4, 5},
						{6, 2}, {6, 6},
						{7, 7}, {7, 8},
					},
					last: 0,
				},
			},
			user: "1",
			top:  3,
			want: []PostSimilarity{{2, 0.816496}, {4, 0.612372}, {6, 0.204124}},
			err:  nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//TODO: probably should change this errors logic
			test.fetcher.err = test.err

			result, err := GetSimilarPosts(&test.fetcher, test.user, test.top)

			if !cmp.Equal(err, test.err, opt) {
				t.Fatalf("GetSimilarPosts() error %v, wanted %v", err, test.err)
			}

			if !cmp.Equal(result, test.want, opt) {
				t.Fatalf("GetSimilarPosts() result %v, wanted %v", result, test.want)
			}
		})
	}
}
