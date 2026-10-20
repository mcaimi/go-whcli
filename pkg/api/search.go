package api

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/mcaimi/go-whcli/pkg/common"
)

// buildSearchUrl assembles the WallHaven search URL from query parameters.
func buildSearchUrl(parms common.QueryParameters) (string, error) {
	endPoint := fmt.Sprintf("%s/%s/%s", common.API_ENDPOINT, common.API_VERSION, common.SEARCH_PATH)
	v := url.Values{}

	if parms.Purity != "" {
		v.Add("purity", parms.Purity)
	}
	if parms.AiArt {
		v.Add("ai_art_filter", "0")
	} else {
		v.Add("ai_art_filter", "1")
	}
	v.Add("atleast", parms.Resolution)
	if parms.Sorting != "" {
		v.Add("sorting", parms.Sorting)
	}
	v.Add("categories", parms.Kind)
	if parms.Query != "" {
		v.Add("q", parms.Query)
	}
	if parms.Color != "" {
		v.Add("colors", parms.Color)
	}
	if parms.Order != "" {
		v.Add("order", parms.Order)
	}
	if parms.Sorting == "random" && parms.Seed != "" {
		v.Add("seed", parms.Seed)
	}
	v.Add("page", fmt.Sprintf("%d", parms.Page))

	finalUrl, err := url.Parse(endPoint)
	if err != nil {
		return "", fmt.Errorf("failed to parse endpoint %s: %w", endPoint, err)
	}
	finalUrl.RawQuery = v.Encode()
	return finalUrl.String(), nil
}

// Search queries the WallHaven search endpoint and decodes the response.
func Search(parms common.QueryParameters) (common.SearchResponse, error) {
	var resp common.SearchResponse

	urlStr, err := buildSearchUrl(parms)
	if err != nil {
		return common.SearchResponse{}, err
	}
	if err := doRequest(urlStr, &resp); err != nil {
		return common.SearchResponse{}, err
	}
	return resp, nil
}

// RandomWallpaper returns the ID of the first result of a random search.
func RandomWallpaper(parms common.QueryParameters) (string, error) {
	parms.Sorting = "random"
	result, err := Search(parms)
	if err != nil {
		return "", err
	}

	if len(result.Data) > 0 {
		return result.Data[0].ID, nil
	}
	return "", errors.New("no matching result")
}
