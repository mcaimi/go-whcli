package gui

import (
	"image"
	_ "image/jpeg"
	_ "image/png"

	"fyne.io/fyne/v2"

	"github.com/mcaimi/go-whcli/pkg/api"
)

const maxConcurrentThumbFetches = 6

// ThumbLoader fetches and decodes thumbnail images with bounded
// concurrency, backed by a shared ThumbCache.
type ThumbLoader struct {
	cache *ThumbCache
	sem   chan struct{}
}

// NewThumbLoader returns a loader backed by cache.
func NewThumbLoader(cache *ThumbCache) *ThumbLoader {
	return &ThumbLoader{
		cache: cache,
		sem:   make(chan struct{}, maxConcurrentThumbFetches),
	}
}

// Load resolves url to a decoded image and invokes onLoaded with the
// result, always via fyne.Do so it is safe to mutate Fyne widgets from
// onLoaded. A cache hit resolves without starting a goroutine.
func (l *ThumbLoader) Load(url string, onLoaded func(img image.Image, err error)) {
	if img, ok := l.cache.Get(url); ok {
		fyne.Do(func() { onLoaded(img, nil) })
		return
	}

	go func() {
		l.sem <- struct{}{}
		defer func() { <-l.sem }()

		resp, err := api.DoGet(url)
		if err != nil {
			fyne.Do(func() { onLoaded(nil, err) })
			return
		}
		defer resp.Body.Close()

		img, _, err := image.Decode(resp.Body)
		if err != nil {
			fyne.Do(func() { onLoaded(nil, err) })
			return
		}

		l.cache.Put(url, img)
		fyne.Do(func() { onLoaded(img, nil) })
	}()
}
