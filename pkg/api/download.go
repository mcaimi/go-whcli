package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"syscall"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/mcaimi/go-whcli/pkg/common"
)

// ProgressFunc is invoked as a wallpaper download progresses, with the
// destination path (known before the download starts), cumulative bytes
// transferred, and total content length (0 if unknown/not sent by the
// server). May be nil.
type ProgressFunc func(destFile string, transferred, total int64)

// StatusFunc is invoked for human-readable narration events that are not
// byte-level progress (e.g. the cross-device-link fallback notice). May be
// nil.
type StatusFunc func(message string)

// DownloadWallpaperWithProgress downloads the wallpaper identified by
// parms.Id into parms.Path, reporting progress/status via the supplied
// callbacks instead of printing to stdout, and returns the final
// destination path on success. It is the shared core behind both
// DownloadWallpaper (CLI) and the GUI's download flow.
func DownloadWallpaperWithProgress(parms common.QueryParameters, onProgress ProgressFunc, onStatus StatusFunc) (string, error) {
	var err error
	var downloadResponse *http.Response
	var infoObject common.WallpaperInfoResponse
	// query wallhaven for information about the requested wp ID
	infoObject, err = GetById(parms)
	if err != nil {
		return "", fmt.Errorf("failed to fetch wallpaper %s: %w", parms.Id, err)
	}
	if infoObject.Data.ID == "" {
		return "", fmt.Errorf("wallpaper %s not found", parms.Id)
	}

	// open a temp file
	var tempFile *os.File
	if tempFile, err = os.CreateTemp("", fmt.Sprintf("%s-", parms.Id)); err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer tempFile.Close()
	defer func() {
		if stat, err := os.Stat(tempFile.Name()); err == nil && stat.Size() > 0 {
			os.Remove(tempFile.Name())
		}
	}()

	// download URL
	if downloadResponse, err = DoGetDownload(infoObject.Data.Path); err != nil {
		return "", fmt.Errorf("download from %s failed: %w", infoObject.Data.Path, err)
	}
	defer downloadResponse.Body.Close()

	// prepare destination path
	// make directory
	if err = os.MkdirAll(parms.Path, 0o700); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", parms.Path, err)
	}
	destFile := path.Join(parms.Path, path.Base(infoObject.Data.Path))

	// start background goroutine
	syncChan := make(chan int)

	go func(c chan int) {
		downloadBuffer := make([]byte, common.CHUNK_SIZE)
		var bytesTransferred int
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
				c <- written
			}
		}
		close(c)
	}(syncChan)

	var accumulator int
	var writeFailed bool
	for transferred := range syncChan {
		if transferred == -1 {
			writeFailed = true
			break
		}
		accumulator += transferred
		if onProgress != nil {
			onProgress(destFile, int64(accumulator), downloadResponse.ContentLength)
		}
	}
	if onStatus != nil {
		onStatus("")
	}

	if writeFailed {
		return "", fmt.Errorf("error writing %s to disk", destFile)
	}

	// move file to destination
	if err = os.Rename(tempFile.Name(), destFile); err != nil {
		// try to overcome the "invalid Cross-Link Device" error (err code 18)
		lnkErr, isLinkErr := err.(*os.LinkError)
		var errID syscall.Errno
		if isLinkErr {
			if errno, isErrno := lnkErr.Err.(syscall.Errno); isErrno {
				errID = errno
			}
		}

		// is error code 18?
		if errID == 18 {
			if onStatus != nil {
				onStatus("Recover: Cannot move, Cross-Link error detected: trying to copy data instead...")
			}
			// copy instead of moving
			srcFile, err := os.Open(tempFile.Name())
			if err != nil {
				return "", fmt.Errorf("fatal error: %s", err)
			}
			defer srcFile.Close()
			dstFile, err := os.Create(destFile)
			if err != nil {
				return "", fmt.Errorf("fatal error: %s", err)
			}
			defer dstFile.Close()

			// copy data
			bytesCopied, err := io.Copy(dstFile, srcFile)
			if err != nil {
				return "", fmt.Errorf("fatal error: %s", err)
			}

			if onStatus != nil {
				onStatus(fmt.Sprintf("[%d] Bytes Copied, Syncing...", bytesCopied))
			}
			if err := dstFile.Sync(); err != nil {
				return "", fmt.Errorf("fatal error: %s", err)
			}

		} else {
			return "", fmt.Errorf("failed to move file to %s: %s", destFile, err)
		}
	}

	return destFile, nil
}

// DownloadWallpaper downloads the wallpaper identified by parms.Id into
// parms.Path, printing progress to stdout exactly as before. It returns a
// descriptive error instead of exiting, so that callers (the interactive
// shell) can stay alive.
func DownloadWallpaper(parms common.QueryParameters) error {
	destFile, err := DownloadWallpaperWithProgress(parms,
		func(destFile string, transferred, total int64) {
			var sizeStr string
			if total > 0 {
				sizeStr = text.FgGreen.Sprintf("%d", total)
			} else {
				sizeStr = "?"
			}
			fmt.Printf("Downloading file <%s> - [%s/%s] Bytes Received.\r", text.FgRed.Sprintf("%s", destFile), text.FgYellow.Sprintf("%d", transferred), sizeStr)
		},
		func(msg string) {
			fmt.Println(msg)
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("whcli: saved to %s\n", destFile)
	return nil
}
