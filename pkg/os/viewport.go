package consoleio

import (
  "os"
  "fmt"
  "bufio"
  "strings"

  "golang.org/x/term"

  "github.com/jedib0t/go-pretty/v6/table"
  "github.com/jedib0t/go-pretty/v6/text"
  "github.com/mcaimi/go-whcli/pkg/common"
)

// console viewport structure
type Viewport struct {
  w int
  h int
  CurrentPage int
  LastPage int
  PerPage int
  Total int
  Items []common.SearchItem
  consoleOutput int
  consoleInput *bufio.Reader
  Prompt string
}

// argument must be a file descriptor (e.g Stdin/Stdout)
func NewViewPort() *Viewport {
  // viewport
  var vp Viewport;
  var err error;

  vp.consoleOutput = int(os.Stdout.Fd());
  if ! term.IsTerminal(vp.consoleOutput) { return nil; }

  // get terminal size
  if vp.w, vp.h, err = term.GetSize(vp.consoleOutput); err != nil {
    return nil;
  }

  // finalize
  vp.consoleInput = bufio.NewReader(os.Stdin);

  // return
  return &vp;
}

func (v *Viewport) SetData(apiData common.SearchResponse) {
  // fill Search Response Data
  if apiData.Data != nil {
    v.Items = apiData.Data;
  }
  // Fill Metadata
  v.CurrentPage = apiData.Meta.CurrentPage;
  v.PerPage = apiData.Meta.PerPage;
  v.LastPage = apiData.Meta.LastPage;
  v.Total = apiData.Meta.Total;
}

func (v *Viewport) Render() {
  // display table
  tbl := table.NewWriter();
  tbl.SetOutputMirror(os.Stdout);
  tbl.AppendHeader(table.Row{"ID", "Wallpaper URL", "Resolution (Type)", "Category"});

  // add entries to the table
  for _, item := range v.Items {
    tbl.AppendRow(table.Row{text.FgGreen.Sprintf("%s", item.ID), text.FgYellow.Sprintf("%s", item.Path), text.FgRed.Sprintf("%s (%s)", item.Resolution, item.Purity), text.FgBlue.Sprintf("%s", item.Category)});
  }

  // summarize info
  fmt.Printf("Current Page: %d, Last Page: %d, Items Per Page: %d, Total Items: %d\n", v.CurrentPage, v.LastPage, v.PerPage, v.Total);

  // render table
  tbl.Render();
}

func (v *Viewport) SetPrompt(p string) {
  v.Prompt = p;
}

func (v *Viewport) ReadCommand() string {
  // display prompt
  fmt.Printf("%s ", v.Prompt);

  // Read line from terminal
  if line, err := v.consoleInput.ReadString('\n'); err != nil {
    fmt.Println("Error reading from Standard Input");
    return "";
  } else {
    // return line
    return strings.TrimSuffix(line, "\n");
  }
}
