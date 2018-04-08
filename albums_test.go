package main

import (
	"fmt"
	"testing"

	"github.com/zmb3/spotify"
)

func TestNewSideBar(t *testing.T) {
	client := NewDebugClient()
	sideBar, err := NewSideBar(client)
	if err != nil {
		t.Fatalf("Unexpected error occured: %s", err)
	}
	if len(sideBar.albumList.albumsDescriptions) != 135 {
		// Because DebugClient's implementation of CurrentUsersAlbumsOpt fetches 135 Spotify Albums
		t.Fatalf("Should fetch 135 album descripitons, fetched %d", len(sideBar.albumList.albumsDescriptions))
	}
}

type CallConfig struct {
	executionError bool
	returnValue    *spotify.SavedAlbumPage
}

type AlbumFetcherMock struct {
	call        int
	callConfigs []CallConfig
}

func (fake *AlbumFetcherMock) CurrentUsersAlbumsOpt(opt *spotify.Options) (*spotify.SavedAlbumPage, error) {
	if fake.callConfigs[fake.call].executionError == true {
		fake.call++
		return nil, fmt.Errorf("err")
	}
	returnValue := fake.callConfigs[fake.call].returnValue
	fake.call++
	return returnValue, nil
}

func TestFetchUserAlbumListFetchesNoPages(t *testing.T) {
	client := &DebugClient{}
	fetcherMock := &AlbumFetcherMock{}
	fetcherMock.callConfigs = []CallConfig{
		{
			executionError: false,
			returnValue:    &spotify.SavedAlbumPage{Albums: make([]spotify.SavedAlbum, 0)},
		},
	}
	client.UserAlbumFetcher = fetcherMock

	albumList := newEmptyAlbumList(client)
	albumList.fetchUserAlbums()

	if len(albumList.albumsDescriptions) != 0 {
		t.Fatalf("Expected albums descriptions to be empty, but it has length of %d", len(albumList.albumsDescriptions))
	}
}

// Wywala siee przy pobieraniu kolejnej strony

func TestFetchUserAlbumListFetchesSinglePage(t *testing.T) {
	client := &DebugClient{}
	fetcherMock := &AlbumFetcherMock{}

	saved := &spotify.SavedAlbumPage{Albums: constructNSpotifySavedAlbums(25)}
	saved.Total = 25 // Only one page

	fetcherMock.callConfigs = []CallConfig{
		{
			executionError: false,
			returnValue:    saved,
		},
	}
	client.UserAlbumFetcher = fetcherMock

	albumList := newEmptyAlbumList(client)
	err := albumList.fetchUserAlbums()
	if err != nil {
		t.Fatalf("Did not expect to fail, but it did")
	}
	if len(albumList.albumsDescriptions) != 25 {
		t.Fatalf("Expected albums descriptions to have 25 elements, but have %d elements", len(albumList.albumsDescriptions))
	}
	if fetcherMock.call != 1 {
		t.Fatalf("Expected CurrentUsersAlbumsOpt() to be called once, but it was called %d times", fetcherMock.call)
	}
}
func TestFetchUserAlbumListFetchesManyPages(t *testing.T) {
	defer func() { spotifyAPIPageOffset = 25 }() // Reset after test
	client := &DebugClient{}
	fetcherMock := &AlbumFetcherMock{}

	saved := &spotify.SavedAlbumPage{Albums: constructNSpotifySavedAlbums(25)}
	saved.Total = 50

	fetcherMock.callConfigs = []CallConfig{
		{
			executionError: false,
			returnValue:    saved,
		},
		{
			executionError: false,
			returnValue:    saved,
		},
	}
	client.UserAlbumFetcher = fetcherMock

	albumList := newEmptyAlbumList(client)
	err := albumList.fetchUserAlbums()
	if err != nil {
		t.Fatalf("Did not expect to fail, but it did")
	}
	if len(albumList.albumsDescriptions) != 50 {
		t.Fatalf("Expected albums descriptions to have 50 elements, but have %d elements", len(albumList.albumsDescriptions))
	}
	if fetcherMock.call != 2 {
		t.Fatalf("Expected CurrentUsersAlbumsOpt() to be called twice, but it was called %d times", fetcherMock.call)
	}
}
func TestFetchUserAlbumListFailsOnFirstCall(t *testing.T) {
	client := &DebugClient{}
	fetcherMock := &AlbumFetcherMock{}

	fetcherMock.callConfigs = []CallConfig{
		{
			executionError: true,
			returnValue:    nil,
		},
	}
	client.UserAlbumFetcher = fetcherMock

	albumList := newEmptyAlbumList(client)
	err := albumList.fetchUserAlbums()
	if err == nil {
		t.Fatalf("Expected to fail, but it didn't")
	}
	if fetcherMock.call != 1 {
		t.Fatalf("Expected CurrentUsersAlbumsOpt() to be called once, but it was called %d times", fetcherMock.call)
	}
}

func TestFetchUserAlbumListFailsWhenFetchingNotFirstPage(t *testing.T) {
	defer func() { spotifyAPIPageOffset = 25 }() // Reset after test
	t.Log(spotifyAPIPageOffset)
	client := &DebugClient{}
	fetcherMock := &AlbumFetcherMock{}

	saved := &spotify.SavedAlbumPage{Albums: constructNSpotifySavedAlbums(25)}
	saved.Total = 50

	fetcherMock.callConfigs = []CallConfig{
		{
			executionError: false,
			returnValue:    saved,
		},
		{
			executionError: true,
			returnValue:    nil,
		},
	}
	client.UserAlbumFetcher = fetcherMock

	albumList := newEmptyAlbumList(client)
	err := albumList.fetchUserAlbums()
	if err == nil {
		t.Fatalf("Expected to fail, but it didn't")
	}
	if fetcherMock.call != 2 {
		t.Fatalf("Expected CurrentUsersAlbumsOpt() to be called twice, but it was called %d times", fetcherMock.call)
	}
}

func TestTrimCommasIfTooLong(t *testing.T) {
	text := "Some text"
	cases := []struct {
		length         int
		expectedResult string
	}{
		{
			len(text),
			"Some text",
		},
		{
			len(text) - 1,
			"Some tex...",
		},
		{
			len(text) + 1,
			"Some text",
		},
	}
	for _, c := range cases {
		if result := trimWithCommasIfTooLong(text, c.length); result != c.expectedResult {
			t.Fatalf("Expected result to be %s, but it was %s", c.expectedResult, result)
		}
	}

}
