package api

import (
  "fmt"
  "io"
  "log"
  "net/http"
  "os"
  "path"

  "codeberg.org/mcaimi/go-whcli/pkg/common"
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
  defer func() {
    if stat, err := os.Stat(tempFile.Name()); err == nil && stat.Size() > 0 {
      os.Remove(tempFile.Name())
    }
  }();

  // download URL
  if downloadResponse, err = DoGetDownload(infoObject.Data.Path); err != nil {
    log.Fatalf("Download for URL %s -> %v", infoObject.Data.Path, err);
  }
  defer downloadResponse.Body.Close();

  // prepare destination path
  // make directory
  if err = os.MkdirAll(parms.Path, 0700); err != nil {
    log.Fatalf("%s", err);
  }
  destFile := path.Join(parms.Path, path.Base(infoObject.Data.Path));

  // start background goroutine
  syncChan := make(chan int);

  go func(c chan int) {
    downloadBuffer := make([]byte, common.CHUNK_SIZE)
    var bytesTransferred int;
    var written int
    for {
        if bytesTransferred, err = downloadResponse.Body.Read(downloadBuffer); err != nil {
          if err == io.EOF {
            break
          }
          c <- -1
          close(c)
          return
        }
        if bytesTransferred > 0 {
            written, err = tempFile.Write(downloadBuffer[:bytesTransferred])
            if err != nil {
              c <- -1
              close(c)
              return
            }
            c <- written;
        }
    }
    close(c)
  }(syncChan);

  var accumulator int
  var writeFailed bool
  for transferred := range syncChan {
    if transferred == -1 {
      writeFailed = true
      break
    }
    accumulator += transferred;
    var sizeStr string
    if downloadResponse.ContentLength > 0 {
      sizeStr = text.FgGreen.Sprintf("%d", downloadResponse.ContentLength)
    } else {
      sizeStr = "?"
    }
    fmt.Printf("Downloading file <%s> - [%s/%s] Bytes Received.\r", text.FgRed.Sprintf("%s", destFile), text.FgYellow.Sprintf("%d", accumulator), sizeStr)
  }
  fmt.Println()

  if writeFailed {
    log.Fatalf("Error writing file to disk")
  }

  // move file to destination
  if err = os.Rename(tempFile.Name(), destFile); err != nil {
    log.Fatalf("Failed to move file to %s: %s", destFile, err)
  }
}
