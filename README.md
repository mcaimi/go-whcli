### Wallhaven.cc CLI

A Go CLI tool to interact with the [Wallhaven.cc](https://wallhaven.cc) wallpaper API for searching and downloading wallpapers. Currently it supports:

- Wallpaper download
- Random wallpaper download
- Custom search with filters
- Interactive shell for pagination and live browsing
- Wallpaper and tag metadata queries

## Installation

```bash
# from source
go generate && go build

# or install directly
go install github.com/mcaimi/go-whcli@latest
```

## Architecture

The codebase is organized into three layers:

```
                  ┌─────────────────┐
                  │   main.go       │
                  │  (entry point)  │
                  └────────┬────────┘
                           │ calls
                           ▼
            ┌──────────────────────────────┐
            │      cmd/whcli/cli.go        │  ← Cobra CLI commands
            │  (root, info, download,      │
            │   search, random, tag)       │
            └───────────┬──────────────────┘
                        │ delegates to
            ┌───────────┴───────────┐
            │     cmd/whcli/        │
            │  grammar.y → grammar.go│  ← yacc-generated parser
            │  (lexer, yyParse)     │     for interactive shell
            └───────────┬───────────┘
                        │ uses
            ┌───────────┴───────────┬──────────────────────┐
            │   pkg/api/            │   pkg/common/         │   pkg/os/
            │  client.go  download   │   structs.go          │   viewport.go
            │  get.go   search.go    │   constants.go        │   (terminal I/O)
            └────────────────────────┴───────────────────────┴─────────┘
```

### Package responsibilities

| Package | Responsibility |
| --------- | --------------- |
| `main` | Single entry point, calls `whcli.Execute()` |
| `cmd/whcli` | Cobra CLI command definitions and yacc-based interactive shell lexer/parser |
| `pkg/api` | HTTP client, Wallhaven API request logic, download logic, response formatting |
| `pkg/common` | Shared data structs, API constants, query parameter definitions |
| `pkg/os` | Terminal/console I/O helpers: viewport creation, table rendering, prompt reading |

### Class Diagram

```
┌──────────────────────┐     1         * ┌──────────────────────┐
│     whcli.Execute()  │────────────────>│   cobra.Command      │
│  (main package)      │                 │   - infoCommand      │
└──────────────────────┘                 │   - downloadCommand  │
                                         │   - searchCommand    │
                                         │   - rndCommand       │
                                         │   - tagInfoCommand   │
                                         └──────────┬─────────┘
                                                    │ invokes
                                                    ▼
┌──────────────────────┐     *      1 ┌──────────────────────┐
│   lexer              │─────────────>│   Viewport           │
│  - input: string     │ lexer uses   │  - Items []SearchItem│
│  - options: map      │              │  - consoleInput      │
│  - line/pos: string  │              │  - Prompt            │
│  + Lex()             │              │  + Render()          │
│  + Error()           │              │  + ReadCommand()     │
└──────────────────────┘              └──────────────────────┘
                                               ▲
                                               │ uses
                                               │
┌──────────────────────┐     *      1         │
│  QueryParameters     │──────────────────────┘
│  - Resolution        │ sets data
│  - Purity, Kind,     │
│    Query, Color,     │
│    Sorting, Order,   │
│    Id, Path, Page,   │
│    Seed              │
└──────────────────────┘

┌──────────────────────┐
│  api package         │
│  ┌────────────────┐  │
│  │  DoGet/DoGet   │──┼──> HTTP requests to Wallhaven.cc API
│  │  Download      │──┼──> downloadHTTP
│  └────────────────┘  │
│  ┌────────────────┐  │
│  │  Search/Rnd()  │──┼──> buildSearchURL + doRequest
│  │  GetById()     │──┼──> fetch metadata by ID
│  │  DownloadWall()│──┼──> download file to disk
│  │  Display*()    │  │
│  └────────────────┘  │
└──────────────────────┘
         ▲
         │ uses
         │
┌──────────────────────┐
│  common structs      │
│  ┌────────────────┐  │
│  │SearchResponse  │  │    ┌──────────────────┐
│  │ + Data []Item  │  │    │SearchItem        │
│  │ + Meta         │  │    │ + ID, Path, URL  │
│  └────────────────┘  │    │ + Resolution     │
│  ┌────────────────┐  │    │ + Purity, Cat    │
│  │WallpaperInfo   │  │    └──────────────────┘
│  │Response        │  │    ┌──────────────────┐
│  │ + Data         │  │    │TagInfoResponse   │
│  └────────────────┘  │    │ + ID, Name, Cat  │
│  ┌────────────────┐  │    └──────────────────┘
│  │TagQueryResult  │  │
│  └────────────────┘  │
└──────────────────────┘

┌──────────────────────┐
│  api.Config          │
│  - httpClient        │  (global, 30s timeout)
│  - downloadHTTP      │  (global, 10m timeout)
│  - APIKey            │  (from WALLHAVEN_API_KEY env)
│  + SetApiKey()       │
│  + SetTimeout()      │
└──────────────────────┘
```

### Search flow

```
CLI (cobra)        api.Search()        buildSearchURL()        Wallhaven API
    │                   │                      │                      │
    │  search flags    │                      │                      │
    ├────────────────>│                      │                      │
    │                   ├────────────────────>│                      │
    │                   │  URL with query    │                      │
    │                   │                    │                      │
    │                   │    DoGet()         │                      │
    │                   ├───────────────────────────────────────────>│
    │                   │   HTTP GET          │                      │
    │                   │<───────────────────────────────────────────┤
    │                   │   JSON body        │                      │
    │                   │    unmarshal       │                      │
    │                   │    to SearchResp   │                      │
    │                   │                      │                      │
    │    formatted      │                      │                      │
    │<─────────────────┘                      │                      │
    │  table output                     (no interaction)              │
```

## Usage Examples

First, set your Wallhaven API key (optional but recommended):

```bash
export WALLHAVEN_API_KEY=your_api_key_here
```

### Search for wallpapers

Search with keyword filters, outputting a scrollable table:

```bash
whcli search -q "mountain sunset" -r "2560x1440" -c "red,orange"
```

Navigate through results inside the interactive shell:

```
WHCLI > next
WHCLI > resolution=3840x2160
WHCLI > search
WHCLI > prev
WHCLI > quit
```

### Download a specific wallpaper

```bash
whcli download -i abc123 -p ~/Pictures/wallpapers
```

### Download a random wallpaper

```bash
whcli random -q "night city" -k "010" -p "100" -r "1920x1080" -a
```

### View wallpaper info

```bash
whcli info -i abc123
```

### View tag info

```bash
whcli tag -i 12345
```

### Global flags

These apply to all subcommands:

```bash
whcli --timeout 1m --download-timeout 5m <command> [flags]
```

### Interactive shell commands

Inside any search, the following commands are available:

| Command | Description |
| --------- | ------------- |
| `search` | Re-run the current search with modified parameters |
| `next` | Go to the next page of results |
| `prev` | Go to the previous page (stays on page 1 minimum) |
| `info <id>` | Display metadata for a specific wallpaper by ID |
| `download <id>` | Download a specific wallpaper by ID |
| `tag <id>` | Display information about a specific tag |
| `env <param>` | Show the current value of a search parameter |
| `<param>=<value>` | Assign a new value to a search parameter (`<param> = <value>` also works) |
| `help` | Show shell usage: commands and valid parameter values |
| `quit` | Exit the shell |

Example parameter assignment and search re-execution:

```
WHCLI > purity=110
WHCLI > search
```

Parameter values are validated against the WallHaven API rules and invalid
values are rejected with a message (nothing is applied):

| Parameter | Valid values |
| --------- | ------------ |
| `resolution` | `WIDTHxHEIGHT`, e.g. `1920x1080` |
| `purity` | 3-character `0/1` string: `100` (sfw), `110` (+sketchy), `111` (+nsfw, needs API key) |
| `kind` | 3-character `0/1` string for general/anime/people, e.g. `101` |
| `query` | any non-empty text (spaces allowed), e.g. `+anime +summer` |
| `color` | 6-digit hex codes and/or palette names, e.g. `red,0066cc` |
| `sorting` | `date_added`, `relevance`, `random`, `views`, `favorites`, `toplist` |
| `order` | `desc`, `asc` |
| `path` | any non-empty path |

To clear `query`, `color`, `sorting` or `order`, assign an empty value: `query =`.
Ctrl-D (EOF) exits the shell.

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--timeout` | `30s` | API request timeout (Go duration format: `30s`, `5m`) |
| `--download-timeout` | `10m` | Download operation timeout (Go duration format) |

## API Key

Authenticate with Wallhaven.cc by setting the `WALLHAVEN_API_KEY` environment variable. You can also pass it at runtime via `--api-key` flag.

## TODO

- Better command line parser
- (Maybe) WAL-like colorscheme generator
