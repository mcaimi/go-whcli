package api

import (
	"fmt"
	"os"

	"github.com/mcaimi/go-whcli/pkg/common"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func GetById(parms common.QueryParameters) common.WallpaperInfoResponse {
	var result common.WallpaperInfoResponse
	doRequestOrFail(fmt.Sprintf("%s/%s/%s/%s", common.API_ENDPOINT, common.API_VERSION, common.W_PATH, parms.Id), &result)
	return result
}

func GetTagById(parms common.QueryParameters) common.TagQueryResult {
	var result common.TagQueryResult
	doRequestOrFail(fmt.Sprintf("%s/%s/%s/%s", common.API_ENDPOINT, common.API_VERSION, common.TAG_PATH, parms.Id), &result)
	return result
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
