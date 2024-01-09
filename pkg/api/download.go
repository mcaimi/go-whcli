package api

import (
  "fmt"
  "io"
  "log"
  "net/http"
  "os"
  "strings"

  "github.com/mcaimi/go-whcli/pkg/common"
  "github.com/jedib0t/go-pretty/v6/text"
)

func DownloadWallpaper(parms common.QueryParameters) {
  var err error;
  var downloadResponse *http.Response;
  var infoObject common.WallpaperInfoResponse;
  // query wallhaven for information about the requested wp ID
  infoObject = GetById(parms);

  // open a temp file
  var tempFile *os.File;
  if tempFile, err = os.CreateTemp("", fmt.Sprintf("%s-", parms.Id)); err != nil {
    log.Fatalf("%s", err);
  }
  defer tempFile.Close();

  // download URL
  if downloadResponse, err = http.Get(infoObject.Data.Path); err != nil {
    log.Fatalf("Download for URL %s -> %s", infoObject.Data.Path, err);
  }
  defer downloadResponse.Body.Close();

  // prepare destination destinationPath
  pathSpec := strings.Split(infoObject.Data.Path, "/");
  // make directory
  if err = os.MkdirAll(parms.Path, 0700); err != nil {
    log.Fatalf("%s", err);
  }
  destFile := fmt.Sprintf("%s/%s", parms.Path, pathSpec[len(pathSpec)-1]);

  // start background goroutine
  syncChan := make(chan int);

  go func(c chan int) {
    // download file from the net
    downloadBuffer := make([]byte, common.CHUNK_SIZE)
    var bytesTransferred int;
    for {
        if bytesTransferred, err = downloadResponse.Body.Read(downloadBuffer); err != nil {
          if err == io.EOF {
            close(syncChan);
            break
          }
          log.Fatalf("Error while downloading: %v", err)
        }
        if bytesTransferred > 0 {
            tempFile.Write(downloadBuffer[:bytesTransferred])
            syncChan <- bytesTransferred;
        }
    }
  }(syncChan);

  var accumulator int;
  for transferred := range syncChan {
    accumulator = accumulator + transferred;
    fmt.Printf(fmt.Sprintf("Downloading file <%s> - [%s/%s] Bytes Received.\r", text.FgRed.Sprintf("%s", destFile), text.FgYellow.Sprintf("%d", accumulator), text.FgGreen.Sprintf("%d", downloadResponse.ContentLength)));
  }
  fmt.Println();

  // move file to destination
  os.Rename(tempFile.Name(), destFile);
}
