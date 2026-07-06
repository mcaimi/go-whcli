package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"codeberg.org/mcaimi/go-whcli/pkg/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func fetchById[T any](id string) T {
	resp, err := DoGet(fmt.Sprintf("%s/%s/%s/%s", common.API_ENDPOINT, common.API_VERSION, common.W_PATH, id))
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read API response: %v", err)
	}

	var result T
	if err = json.Unmarshal(body, &result); err != nil {
		log.Fatalf("failed to decode API response: %v", err)
	}

	return result
}

func fetchTagById[T any](id string) T {
	resp, err := DoGet(fmt.Sprintf("%s/%s/%s/%s", common.API_ENDPOINT, common.API_VERSION, common.TAG_PATH, id))
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read API response: %v", err)
	}

	var result T
	if err = json.Unmarshal(body, &result); err != nil {
		log.Fatalf("failed to decode API response: %v", err)
	}

	return result
}

func GetById(parms common.QueryParameters) common.WallpaperInfoResponse {
	return fetchById[common.WallpaperInfoResponse](parms.Id)
}

func GetTagById(parms common.QueryParameters) common.TagQueryResult {
	return fetchTagById[common.TagQueryResult](parms.Id)
}

func DisplayWallpaperInfo(item common.WallpaperInfoResponse) {
	tbl := table.NewWriter()
	tbl.SetOutputMirror(os.Stdout)
	tbl.AppendHeader(table.Row{"ID", "Wallpaper URL", "Resolution (Type)", "Uploaded By"})
	tbl.AppendRow(table.Row{text.FgGreen.Sprintf("%s", item.Data.ID), text.FgYellow.Sprintf("%s", item.Data.Path), text.FgYellow.Sprintf("%s (%s)", item.Data.Resolution, item.Data.Purity), text.FgBlue.Sprintf("%s", item.Data.Uploader.Username)})
	tbl.Render()

	fmt.Println(text.FgHiBlue.Sprintf("%s", "Tags associated with the Image"))
	tbl = table.NewWriter()
	tbl.SetOutputMirror(os.Stdout)
	tbl.AppendHeader(table.Row{"Tag", "ID", "Category", "Purity"})
	for i := range item.Data.Tags {
		field := item.Data.Tags[i]
		tbl.AppendRow(table.Row{text.FgGreen.Sprintf("%s", field.Name), text.FgRed.Sprintf("%d", field.ID), text.FgYellow.Sprintf("%s", field.Category), text.FgBlue.Sprintf("%s", field.Purity)})
	}
	tbl.Render()
}

func DisplayTagInfo(item common.TagQueryResult) {
	tbl := table.NewWriter()
	tbl.SetOutputMirror(os.Stdout)
	tbl.AppendHeader(table.Row{"ID", "Name (alias)", "Category (ID)", "Purity"})
	tbl.AppendRow(table.Row{text.FgGreen.Sprintf("%d", item.Data.ID), text.FgYellow.Sprintf("%s (%s)", item.Data.Name, item.Data.Alias), text.FgYellow.Sprintf("%s (%d)", item.Data.Category, item.Data.CategoryID), text.FgBlue.Sprintf("%s", item.Data.Purity)})
	tbl.Render()
}
