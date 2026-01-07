package api

import (
  "encoding/json"
  "fmt"
  "io"
  "log"
  "net/http"
  "os"

  "github.com/jedib0t/go-pretty/v6/table"
  "github.com/jedib0t/go-pretty/v6/text"
  "codeberg.org/mcaimi/go-whcli/pkg/common"
)

func GetById(parms common.QueryParameters) common.WallpaperInfoResponse {
  var remoteResponse *http.Response;
  var responseBytes []byte;
  var err error;
  var endPoint string;

  endPoint = fmt.Sprintf("%s/%s/%s/", common.API_ENDPOINT, common.API_VERSION, common.W_PATH);

  // get wallpaper information and download link (by wallpaper id)
  if remoteResponse, err = http.Get(endPoint + parms.Id); err != nil {
    log.Fatalf("%s -> %s\n", endPoint, err);
  }
  defer remoteResponse.Body.Close();

  if remoteResponse.StatusCode != 200 {
    log.Fatalf("Status Code -> %d\n", remoteResponse.StatusCode);
  }

  // read response bytes
  responseBytes, _ = io.ReadAll(remoteResponse.Body);
  // decode response into struct
  var apiObject common.WallpaperInfoResponse;
  if err = json.Unmarshal(responseBytes, &apiObject); err != nil {
    log.Fatalf("%s -> %s\n", "Error while decoding API response", err);
  }

  // return response
  return apiObject;
}

func GetTagById(parms common.QueryParameters) common.TagQueryResult {
  var remoteResponse *http.Response;
  var responseBytes []byte;
  var err error;
  var endPoint string;

  endPoint = fmt.Sprintf("%s/%s/%s/", common.API_ENDPOINT, common.API_VERSION, common.TAG_PATH);

  // get wallpaper information and download link (by wallpaper id)
  if remoteResponse, err = http.Get(endPoint + parms.Id); err != nil {
    log.Fatalf("%s -> %s\n", endPoint, err);
  }
  defer remoteResponse.Body.Close();

  if remoteResponse.StatusCode != 200 {
    log.Fatalf("Status Code -> %d\n", remoteResponse.StatusCode);
  }

  // read response bytes
  responseBytes, _ = io.ReadAll(remoteResponse.Body);
  // decode response into struct
  var apiObject common.TagQueryResult;
  if err = json.Unmarshal(responseBytes, &apiObject); err != nil {
    log.Fatalf("%s -> %s\n", "Error while decoding API response", err);
  }

  // return response
  return apiObject;
}

func DisplayWallpaperInfo(item common.WallpaperInfoResponse) {
  // display table
  tbl := table.NewWriter();
  tbl.SetOutputMirror(os.Stdout);
  tbl.AppendHeader(table.Row{"ID", "Wallpaper URL", "Resolution (Type)", "Uploaded By"});
  tbl.AppendRow(table.Row{text.FgGreen.Sprintf("%s", item.Data.ID), text.FgYellow.Sprintf("%s", item.Data.Path), text.FgYellow.Sprintf("%s (%s)", item.Data.Resolution, item.Data.Purity), text.FgBlue.Sprintf("%s", item.Data.Uploader.Username)});
  tbl.Render();

  fmt.Println(text.FgHiBlue.Sprintf("%s", "Tags associated with the Image"));
  tbl = table.NewWriter();
  tbl.SetOutputMirror(os.Stdout);
  tbl.AppendHeader(table.Row{"Tag", "ID", "Category", "Purity"});
  for i := range item.Data.Tags {
    field := item.Data.Tags[i];
    tbl.AppendRow(table.Row{text.FgGreen.Sprintf("%s", field.Name), text.FgRed.Sprintf("%d", field.ID), text.FgYellow.Sprintf("%s", field.Category), text.FgBlue.Sprintf("%s", field.Purity)});
  }
  tbl.Render();
}

func DisplayTagInfo(item common.TagQueryResult) {
  // display table
  tbl := table.NewWriter();
  tbl.SetOutputMirror(os.Stdout);
  tbl.AppendHeader(table.Row{"ID", "Name (alias)", "Category (ID)", "Purity"});
  tbl.AppendRow(table.Row{text.FgGreen.Sprintf("%d", item.Data.ID), text.FgYellow.Sprintf("%s (%s)", item.Data.Name, item.Data.Alias), text.FgYellow.Sprintf("%s (%d)", item.Data.Category, item.Data.CategoryID), text.FgBlue.Sprintf("%s", item.Data.Purity)});
  tbl.Render();
}
