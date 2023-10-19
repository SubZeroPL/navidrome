package persistence

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/beego/beego/v2/client/orm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/navidrome/navidrome/conf"
	"github.com/navidrome/navidrome/db"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/model/request"
	"github.com/navidrome/navidrome/tests"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPersistence(t *testing.T) {
	tests.Init(t, true)

	//os.Remove("./test-123.db")
	//conf.Server.DbPath = "./test-123.db"
	conf.Server.DbPath = "file::memory:?cache=shared"
	_ = orm.RegisterDataBase("default", db.Driver, conf.Server.DbPath)
	db.Init()
	log.SetLevel(log.LevelError)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Persistence Suite")
}

var (
	genreElectronic = model.Genre{ID: "gn-1", Name: "Electronic"}
	genreRock       = model.Genre{ID: "gn-2", Name: "Rock"}
	testGenres      = model.Genres{genreElectronic, genreRock}
)

var (
	artistKraftwerk = model.Artist{ID: "2", Name: "Kraftwerk", AlbumCount: 1, FullText: " kraftwerk"}
	artistBeatles   = model.Artist{ID: "3", Name: "The Beatles", AlbumCount: 2, FullText: " beatles the"}
	testArtists     = model.Artists{
		artistKraftwerk,
		artistBeatles,
	}
)

var (
	albumSgtPeppers    = model.Album{ID: "101", Name: "Sgt Peppers", Artist: "The Beatles", OrderAlbumName: "sgt peppers", AlbumArtistID: "3", Genre: "Rock", Genres: model.Genres{genreRock}, EmbedArtPath: P("/beatles/1/sgt/a day.mp3"), SongCount: 1, MaxYear: 1967, FullText: " beatles peppers sgt the"}
	albumAbbeyRoad     = model.Album{ID: "102", Name: "Abbey Road", Artist: "The Beatles", OrderAlbumName: "abbey road", AlbumArtistID: "3", Genre: "Rock", Genres: model.Genres{genreRock}, EmbedArtPath: P("/beatles/1/come together.mp3"), SongCount: 1, MaxYear: 1969, FullText: " abbey beatles road the"}
	albumRadioactivity = model.Album{ID: "103", Name: "Radioactivity", Artist: "Kraftwerk", OrderAlbumName: "radioactivity", AlbumArtistID: "2", Genre: "Electronic", Genres: model.Genres{genreElectronic, genreRock}, EmbedArtPath: P("/kraft/radio/radio.mp3"), SongCount: 2, FullText: " kraftwerk radioactivity"}
	testAlbums         = model.Albums{
		albumSgtPeppers,
		albumAbbeyRoad,
		albumRadioactivity,
	}
)

var (
	songDayInALife    = model.MediaFile{ID: "1001", Title: "A Day In A Life", ArtistID: "3", Artist: "The Beatles", AlbumID: "101", Album: "Sgt Peppers", Genre: "Rock", Genres: model.Genres{genreRock}, Path: P("/beatles/1/sgt/a day.mp3"), FullText: " a beatles day in life peppers sgt the"}
	songComeTogether  = model.MediaFile{ID: "1002", Title: "Come Together", ArtistID: "3", Artist: "The Beatles", AlbumID: "102", Album: "Abbey Road", Genre: "Rock", Genres: model.Genres{genreRock}, Path: P("/beatles/1/come together.mp3"), FullText: " abbey beatles come road the together"}
	songRadioactivity = model.MediaFile{ID: "1003", Title: "Radioactivity", ArtistID: "2", Artist: "Kraftwerk", AlbumID: "103", Album: "Radioactivity", Genre: "Electronic", Genres: model.Genres{genreElectronic}, Path: P("/kraft/radio/radio.mp3"), FullText: " kraftwerk radioactivity"}
	songAntenna       = model.MediaFile{ID: "1004", Title: "Antenna", ArtistID: "2", Artist: "Kraftwerk", AlbumID: "103", Genre: "Electronic", Genres: model.Genres{genreElectronic, genreRock}, Path: P("/kraft/radio/antenna.mp3"), FullText: " antenna kraftwerk"}
	testSongs         = model.MediaFiles{
		songDayInALife,
		songComeTogether,
		songRadioactivity,
		songAntenna,
	}
)

var (
	m3uLinks = model.RadioLinks{
		model.RadioLink{Name: "1", Url: "https://example.com:6000/stream"},
		model.RadioLink{Name: "2", Url: "https://example.com:6000/stream/20"},
		model.RadioLink{Name: "3", Url: "https://example.com:6000/stream/40"},
		model.RadioLink{Name: "4", Url: "https://example.com:6000/stream/100"},
	}
	m3uExtendedLinks = model.RadioLinks{
		model.RadioLink{Name: "Sample Radio 10", Url: "https://example.com:8000/stream"},
		model.RadioLink{Name: "Sample Radio 100", Url: "https://example.com:8000/stream/20"},
		model.RadioLink{Name: "3", Url: "https://example.com:8000/stream/40"},
		model.RadioLink{Name: "4", Url: "https://example.com:8000/stream/100"},
	}
	plsLinks = model.RadioLinks{
		model.RadioLink{Name: "Sample Radio 1", Url: "https://example.com:1000/stream"},
		model.RadioLink{Name: "2", Url: "https://example.com:1000/stream/4"},
		model.RadioLink{Name: "Sample Radio 2", Url: "https://example.com:1000/stream/2"},
		model.RadioLink{Name: "6", Url: "http://example.com:1000/stream/1000"},
	}
)

var (
	radioWithoutHomePage = model.Radio{ID: "1235", StreamUrl: "https://example.com:8000/1/stream.mp3", HomePageUrl: "", Name: "No Homepage", IsPlaylist: false}
	radioWithHomePage    = model.Radio{ID: "5010", StreamUrl: "https://example.com/stream.mp3", Name: "Example Radio", HomePageUrl: "https://example.com", IsPlaylist: false}
	radioWithM3u         = model.Radio{ID: "1234", StreamUrl: "https://example.com/stream.m3u", Name: "M3U Playlist", IsPlaylist: true, Links: m3uLinks}
	radioWithExtendedM3u = model.Radio{ID: "2345", StreamUrl: "https://example.com/stream-extended.m3u", Name: "M3U Extended Playlist", IsPlaylist: true, Links: m3uExtendedLinks}
	radioWithPls         = model.Radio{ID: "3456", StreamUrl: "https://example.com/stream.pls", Name: "PLS Playlist", IsPlaylist: true, Links: plsLinks}

	testRadios = model.Radios{
		radioWithoutHomePage,
		radioWithHomePage,
		radioWithM3u,
		radioWithExtendedM3u,
		radioWithPls,
	}
)

var (
	plsBest       model.Playlist
	plsCool       model.Playlist
	testPlaylists []*model.Playlist
)

func P(path string) string {
	return filepath.FromSlash(path)
}

func SetupRadio(repo *radioRepository, client *tests.FakeHttpClient) {
	repo.client = client
	client.Res = http.Response{
		Body:       io.NopCloser(bytes.NewBufferString("")),
		StatusCode: 200,
	}

	err := repo.Put(&radioWithoutHomePage)
	if err != nil {
		panic(err)
	}

	err = repo.Put(&radioWithHomePage)
	if err != nil {
		panic(err)
	}

	f, _ := os.Open("tests/fixtures/radios/radio.m3u")
	header := http.Header{}
	header.Set("Content-Type", "application/mpegurl")
	client.Res = http.Response{
		Body: f, StatusCode: 200, Header: header,
	}
	err = repo.Put(&radioWithM3u)
	if err != nil {
		panic(err)
	}

	f, _ = os.Open("tests/fixtures/radios/radio-extended.m3u")
	client.Res = http.Response{
		Body: f, StatusCode: 200, Header: header,
	}
	err = repo.Put(&radioWithExtendedM3u)
	if err != nil {
		panic(err)
	}

	f, _ = os.Open("tests/fixtures/radios/radio.pls")
	header.Set("Content-Type", "audio/x-scpls")
	client.Res = http.Response{
		Body: f, StatusCode: 200, Header: header,
	}
	err = repo.Put(&radioWithPls)
	if err != nil {
		panic(err)
	}
}

// Initialize test DB
// TODO Load this data setup from file(s)
var _ = BeforeSuite(func() {
	o := orm.NewOrm()
	ctx := log.NewContext(context.TODO())
	user := model.User{ID: "userid", UserName: "userid", IsAdmin: true}
	ctx = request.WithUser(ctx, user)

	ur := NewUserRepository(ctx, o)
	err := ur.Put(&user)
	if err != nil {
		panic(err)
	}

	gr := NewGenreRepository(ctx, o)
	for i := range testGenres {
		g := testGenres[i]
		err := gr.Put(&g)
		if err != nil {
			panic(err)
		}
	}

	mr := NewMediaFileRepository(ctx, o)
	for i := range testSongs {
		s := testSongs[i]
		err := mr.Put(&s)
		if err != nil {
			panic(err)
		}
	}

	alr := NewAlbumRepository(ctx, o).(*albumRepository)
	for i := range testAlbums {
		a := testAlbums[i]
		err := alr.Put(&a)
		if err != nil {
			panic(err)
		}
	}

	arr := NewArtistRepository(ctx, o)
	for i := range testArtists {
		a := testArtists[i]
		err := arr.Put(&a)
		if err != nil {
			panic(err)
		}
	}

	rar := NewRadioRepository(ctx, o)

	testClient := &tests.FakeHttpClient{}

	SetupRadio(rar.(*radioRepository), testClient)

	plsBest = model.Playlist{
		Name:      "Best",
		Comment:   "No Comments",
		OwnerID:   "userid",
		OwnerName: "userid",
		Public:    true,
		SongCount: 2,
	}
	plsBest.AddTracks([]string{"1001", "1003"})
	plsCool = model.Playlist{Name: "Cool", OwnerID: "userid", OwnerName: "userid"}
	plsCool.AddTracks([]string{"1004"})
	testPlaylists = []*model.Playlist{&plsBest, &plsCool}

	pr := NewPlaylistRepository(ctx, o)
	for i := range testPlaylists {
		err := pr.Put(testPlaylists[i])
		if err != nil {
			panic(err)
		}
	}

	// Prepare annotations
	if err := arr.SetStar(true, artistBeatles.ID); err != nil {
		panic(err)
	}
	ar, _ := arr.Get(artistBeatles.ID)
	artistBeatles.Starred = true
	artistBeatles.StarredAt = ar.StarredAt
	testArtists[1] = artistBeatles

	if err := alr.SetStar(true, albumRadioactivity.ID); err != nil {
		panic(err)
	}
	al, _ := alr.Get(albumRadioactivity.ID)
	albumRadioactivity.Starred = true
	albumRadioactivity.StarredAt = al.StarredAt
	testAlbums[2] = albumRadioactivity

	if err := mr.SetStar(true, songComeTogether.ID); err != nil {
		panic(err)
	}
	mf, _ := mr.Get(songComeTogether.ID)
	songComeTogether.Starred = true
	songComeTogether.StarredAt = mf.StarredAt
	testSongs[1] = songComeTogether
})
