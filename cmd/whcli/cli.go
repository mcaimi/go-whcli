package whcli

import (
  "fmt"
  "log"
  "os"

  "github.com/mcaimi/go-whcli/pkg/api"
  "github.com/mcaimi/go-whcli/pkg/os"
  "github.com/mcaimi/go-whcli/pkg/common"
  "github.com/jedib0t/go-pretty/v6/text"
  "github.com/spf13/cobra"
)

var (
  CommandParameters common.QueryParameters
  GenericSearchResults common.SearchResponse
  vP *consoleio.Viewport;

  rootCommand = &cobra.Command{
    Use:   "whcli",
    Short: "WallHaven.cc Text/Console Client",
    Long:  "Interact with WallHaven.cc APIs: Search and Download Images.",
  }

  infoCommand = &cobra.Command{
    Use:     "info",
    Short:   "Displays Info About a particular Wallpaper",
    Long:    "Queries Wallhaven.cc by Wallpaper ID and displays image information and associated Tags",
    Aliases: []string{"i", "info", "display"},
    Run: func(cmd *cobra.Command, args []string) {
      CommandParameters.Id, _ = cmd.Flags().GetString("id")

      api.DisplayWallpaperInfo(api.GetById(CommandParameters))

      os.Exit(0)
    },
  }

  tagInfoCommand = &cobra.Command{
    Use:     "tag",
    Short:   "Gets Info About a particular Tag",
    Long:    "Queries Wallhaven.cc by Tag ID and displays Tag information and associated properties",
    Aliases: []string{"t", "tag"},
    Run: func(cmd *cobra.Command, args []string) {
      CommandParameters.Id, _ = cmd.Flags().GetString("id")

      api.DisplayTagInfo(api.GetTagById(CommandParameters))

      os.Exit(0)
    },
  }

  downloadCommand = &cobra.Command{
    Use:     "download",
    Short:   "Downloads a specific wallpaper",
    Long:    "Downloads a wallpaper from Wallhaven.cc, identified by its ID. Optionally, set a target path for download.",
    Aliases: []string{"d", "dwl"},
    Run: func(cmd *cobra.Command, args []string) {
      CommandParameters.Id, _ = cmd.Flags().GetString("id")
      CommandParameters.Path, _ = cmd.Flags().GetString("path")

      api.DownloadWallpaper(CommandParameters);

      os.Exit(0)
    },
  }

  rndCommand = &cobra.Command{
    Use:     "random",
    Short:   "Download a random wallpaper",
    Long:    "Queries Wallhaven.cc and downloads the first result of a random wallpaper search. Search parameters can still be specified to narrow down the image properties.",
    Aliases: []string{"r", "rnd"},
    Run: func(cmd *cobra.Command, args []string) {
      var err error;

      CommandParameters.Resolution, _ = cmd.Flags().GetString("resolution")
      CommandParameters.Path, _ = cmd.Flags().GetString("path")
      CommandParameters.Purity, _ = cmd.Flags().GetString("purity")
      CommandParameters.Kind, _ = cmd.Flags().GetString("kind")
      CommandParameters.Query, _ = cmd.Flags().GetString("query")
      CommandParameters.Color, _ = cmd.Flags().GetString("color")
      CommandParameters.AiArt, _ = cmd.Flags().GetBool("aifilter");
      CommandParameters.Page = 1

      CommandParameters.Id, err = api.RandomWallpaper(CommandParameters);
      if err != nil {
        log.Fatalf(err.Error())
      } else {
        api.DisplayWallpaperInfo(api.GetById(CommandParameters))
        api.DownloadWallpaper(CommandParameters)
      }

      os.Exit(0)
    },
  }

  searchCommand = &cobra.Command{
    Use:     "search",
    Short:   "Generic search command",
    Long:    "Queries Wallhaven.cc APIs in the most generic way possible.",
    Aliases: []string{"r", "rnd"},
    Run: func(cmd *cobra.Command, args []string) {
      CommandParameters.Resolution, _ = cmd.Flags().GetString("resolution")
      CommandParameters.Purity, _ = cmd.Flags().GetString("purity")
      CommandParameters.Kind, _ = cmd.Flags().GetString("kind")
      CommandParameters.Query, _ = cmd.Flags().GetString("query")
      CommandParameters.Color, _ = cmd.Flags().GetString("color")
      CommandParameters.Order, _ = cmd.Flags().GetString("order")
      CommandParameters.Sorting, _ = cmd.Flags().GetString("sorting")
      CommandParameters.AiArt, _ = cmd.Flags().GetBool("aifilter");
      CommandParameters.Page = 1;

      // search and update params
      GenericSearchResults = api.Search(CommandParameters);
      if len(GenericSearchResults.Data) == 0 {
        fmt.Println("Search returned no items.")
      }
      CommandParameters.Seed = fmt.Sprintf("%s", GenericSearchResults.Meta.Seed);

      // display data
      vP = consoleio.NewViewPort();
      vP.SetData(GenericSearchResults);
      vP.Render();

      // start interactive loop
      var cliCmd string;
      vP.SetPrompt(text.FgGreen.Sprintf("%s", "WHCLI > "));
      for ;; {
        cliCmd = vP.ReadCommand();
        yyParse(&lexer{input: cliCmd});
      }
    },
  }
)

func init() {
  // info command params
  infoCommand.Flags().StringP("id", "i", "", "Wallpaper ID")

  // tag info command params
  tagInfoCommand.Flags().StringP("id", "i", "", "Wallpaper ID")

  // download command params
  downloadCommand.Flags().StringP("id", "i", "", "Wallpaper ID")
  downloadCommand.Flags().StringP("path", "p", "./", "Target directory where to download the wallpaper.")

  // random download command params
  rndCommand.Flags().StringP("path", "d", "./", "Target directory where to download the wallpaper.")
  rndCommand.Flags().StringP("resolution", "r", "1920x1080", "Wallpaper resolution")
  rndCommand.Flags().StringP("purity", "p", "100", "Wallpaper Purity Level (sfw,nsfw): String values made up of 3 integers: xxx where x can be 0 or 1 (sfw/sketchy/nsfw)")
  rndCommand.Flags().StringP("kind", "k", "111", "Wallpaper Kind: String value made up of 3 integers: xxx where x can be either 0 or 1 (anime/nature/people)")
  rndCommand.Flags().StringP("query", "q", "", "Specify wallpaprt search query, eg: +anime +summer")
  rndCommand.Flags().StringP("color", "c", "", "Search by wallpaper color")
  rndCommand.Flags().BoolP("aifilter", "a", true, "Filter out AI-generated content.")

  // search command params
  searchCommand.Flags().StringP("resolution", "r", "1920x1080", "Wallpaper resolution")
  searchCommand.Flags().StringP("purity", "p", "100", "Wallpaper Purity Level (sfw,nsfw): String values made up of 3 integers: xxx where x can be 0 or 1 (sfw/sketchy/nsfw)")
  searchCommand.Flags().StringP("kind", "k", "111", "Wallpaper Kind: String value made up of 3 integers: xxx where x can be either 0 or 1 (anime/nature/people)")
  searchCommand.Flags().StringP("query", "q", "", "Specify wallpaprt search query, eg: +anime +summer")
  searchCommand.Flags().StringP("color", "c", "", "Search by wallpaper color")
  searchCommand.Flags().StringP("sorting", "s", "", "Sorting Logic")
  searchCommand.Flags().StringP("order", "o", "", "Results ordering logic")
  searchCommand.Flags().BoolP("aifilter", "a", true, "Filter out AI-generated content.")

  rootCommand.AddCommand(infoCommand)
  rootCommand.AddCommand(tagInfoCommand)
  rootCommand.AddCommand(downloadCommand)
  rootCommand.AddCommand(rndCommand)
  rootCommand.AddCommand(searchCommand)
}

func Execute() {
  if err := rootCommand.Execute(); err != nil {
    fmt.Fprintf(os.Stderr, err.Error())
    os.Exit(1)
  }
}
