package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"

	"codeberg.org/mcaimi/go-whcli/pkg/common"
)

func buildSearchUrl(parms common.QueryParameters) (string, error) {
	endPoint := fmt.Sprintf("%s/%s/%s", common.API_ENDPOINT, common.API_VERSION, common.SEARCH_PATH)
	v := url.Values{}

	v.Add("purity", parms.Purity)
	if parms.AiArt {
		v.Add("ai_art_filter", "0")
	} else {
		v.Add("ai_art_filter", "1")
	}
	v.Add("atleast", parms.Resolution)
	v.Add("sorting", parms.Sorting)
	v.Add("categories", parms.Kind)
	if parms.Query != "" {
		v.Add("q", parms.Query)
	}
	if parms.Color != "" {
		v.Add("colors", parms.Color)
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

func Search(parms common.QueryParameters) common.SearchResponse {
	urlStr, err := buildSearchUrl(parms)
	if err != nil {
		log.Fatalf("%v", err)
	}

	resp, err := DoGet(urlStr)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read API response: %v", err)
	}

	var apiObject common.SearchResponse
	if err = json.Unmarshal(responseBytes, &apiObject); err != nil {
		log.Fatalf("failed to decode API response: %v", err)
	}

	return apiObject
}

func RandomWallpaper(parms common.QueryParameters) (string, error) {
	parms.Sorting = "random"
	searchResult := Search(parms)

	if len(searchResult.Data) > 0 {
		return searchResult.Data[0].ID, nil
	}
	return "", errors.New("no matching result")
}
