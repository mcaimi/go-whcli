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

// doRequest is a shared HTTP request helper: executes DoGet, reads body, unmarshals into result.
// All fetchers followed this identical pattern (same error handling, same steps).
func doRequest(url string, result interface{}) error {
	resp, err := DoGet(url)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read API response: %w", err)
	}

	if err = json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to decode API response: %w", err)
	}

	return nil
}

// doRequestOrFail mirrors doRequest but logs a fatal error on failure (matching existing CLI behavior).
func doRequestOrFail(url string, result interface{}) {
	if err := doRequest(url, result); err != nil {
		log.Fatalf("%v", err)
	}
}

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

// unmarshalSearchResponse properly decodes {data: [...], meta: {...}} from the API.
func unmarshalSearchResponse(body []byte) common.SearchResponse {
	var parsed struct {
		Data []common.SearchItem `json:"data,omitempty"`
		Meta struct {
			CurrentPage int `json:"current_page,omitempty"`
			LastPage    int `json:"last_page,omitempty"`
			PerPage     int `json:"per_page,omitempty"`
			Total       int `json:"total,omitempty"`
			Query       any `json:"query,omitempty"`
			Seed        any `json:"seed,omitempty"`
		} `json:"meta,omitempty"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		log.Fatalf("failed to decode API response: %v", err)
	}
	return common.SearchResponse{
		Data: parsed.Data,
		Meta: parsed.Meta,
	}
}

func Search(parms common.QueryParameters) common.SearchResponse {
	urlStr, err := buildSearchUrl(parms)
	if err != nil {
		log.Fatalf("%v", err)
	}

	resp, err := DoGet(urlStr)
	if err != nil {
		log.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read API response: %v", err)
	}

	return unmarshalSearchResponse(body)
}

func RandomWallpaper(parms common.QueryParameters) (string, error) {
	parms.Sorting = "random"
	searchResult := Search(parms)

	if len(searchResult.Data) > 0 {
		return searchResult.Data[0].ID, nil
	}
	return "", errors.New("no matching result")
}
