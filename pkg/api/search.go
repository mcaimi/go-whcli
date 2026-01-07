package api

import (
  "encoding/json"
  "errors"
  "fmt"
  "io"
  "log"
  "net/http"
  "net/url"

  "codeberg.org/mcaimi/go-whcli/pkg/common"
)

func Search(parms common.QueryParameters) common.SearchResponse {
  var remoteResponse *http.Response
  var responseBytes []byte
  var err error
  var endPoint string

  endPoint = fmt.Sprintf("%s/%s/%s", common.API_ENDPOINT, common.API_VERSION, common.SEARCH_PATH)
  v := url.Values{}

  // set up query parameters
  v.Add("purity", parms.Purity)
  if parms.AiArt {
    v.Add("ai_art_filter", fmt.Sprintf("%d", 0));
  } else {
    v.Add("ai_art_filter", fmt.Sprintf("%d", 1));
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
    v.Add("seed", parms.Seed);
  }
  v.Add("page", fmt.Sprintf("%d", parms.Page));

  finalUrl, err := url.Parse(endPoint)
  if err != nil {
    log.Fatalf("%s -> %s\n", endPoint, err)
  }
  finalUrl.RawQuery = v.Encode()

  // get wallpaper information and download link (by wallpaper id)
  if remoteResponse, err = http.Get(finalUrl.String()); err != nil {
    log.Fatalf("%s -> %s\n", endPoint, err)
  }
  defer remoteResponse.Body.Close()

  if remoteResponse.StatusCode != 200 {
    log.Fatalf("Status Code -> %d\n", remoteResponse.StatusCode)
  }

  // read response bytes
  responseBytes, _ = io.ReadAll(remoteResponse.Body)
  // decode response into struct
  var apiObject common.SearchResponse
  if err = json.Unmarshal(responseBytes, &apiObject); err != nil {
    log.Fatalf("%s -> %s\n", "Error while decoding API response", err)
  }

  // return response
  return apiObject
}

func RandomWallpaper(parms common.QueryParameters) (string, error) {
  parms.Sorting = "random";
  searchResult := Search(parms);

  // return first result
  if len(searchResult.Data) > 0 {
    return searchResult.Data[0].ID, nil
  }
  return "", errors.New("No matching result.")
}
