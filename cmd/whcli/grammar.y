%{
package whcli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/mcaimi/go-whcli/pkg/api"
	"github.com/mcaimi/go-whcli/pkg/common"
)
%}

%union{
Command string
Parameter string
Value string
}

%token <Command> SEARCHCMD INFOCMD TAGCMD PREVCMD NEXTCMD DOWNCMD ENVCMD QUITCMD
%token <Parameter> RESPARM PURITYPARM KINDPARM QUERYPARM COLORPARM ORDERPARM SORTPARM PATHPARM
%token <Value> IDENTIFIER
%token EQUALS
%token HELPCMD

%type <Parameter> variable
%type <Value> value
%start startcommand

%%
startcommand: commands
            | assignment

commands:
          QUITCMD {
              pendingAction = func() { os.Exit(0) };
          }
          | SEARCHCMD {
              pendingAction = doInteractiveSearch;
          }
          | NEXTCMD {
              pendingAction = doInteractiveNext;
          }
          | PREVCMD {
              pendingAction = doInteractivePrev;
          }
          | HELPCMD {
              pendingAction = printShellHelp;
          }
          | INFOCMD IDENTIFIER {
              id := $2;
              pendingAction = func() { doInteractiveWallpaperInfo(id) };
          }
          | DOWNCMD IDENTIFIER {
              id := $2;
              pendingAction = func() { doInteractiveDownload(id) };
          }
          | TAGCMD IDENTIFIER {
              id := $2;
              pendingAction = func() { doInteractiveTagInfo(id) };
          }
          | ENVCMD variable {
              name := $2;
              lx := yylex.(*lexer);
              pendingAction = func() { showVariable(lx, name) };
          }

assignment:
        variable EQUALS value {
          name, val := $1, $3
          lx := yylex.(*lexer)
          pendingAction = func() {
            valid, msg := validateValue(name, val);
            if !valid {
                fmt.Fprintf(os.Stderr, "whcli: %s\n", msg);
            } else {
                lx.options[name] = val;
                updateParameters(name, &CommandParameters, lx);
                fmt.Fprintf(os.Stdout, "whcli: %s is now [ %s ]\n", name, val);
            }
          };
        }
        | variable EQUALS {
          name := $1
          lx := yylex.(*lexer)
          pendingAction = func() {
            switch name {
            case "query", "color", "sorting", "order":
              lx.options[name] = ""
              updateParameters(name, &CommandParameters, lx)
              fmt.Fprintf(os.Stdout, "whcli: %s has been cleared\n", name)
            default:
              fmt.Fprintf(os.Stderr, "whcli: value for %s must not be empty.\n", name)
            }
          };
        }

variable:
     RESPARM | PURITYPARM | KINDPARM | QUERYPARM | COLORPARM | ORDERPARM | SORTPARM | PATHPARM {
      $$ = $1
    }

value:
     IDENTIFIER { $$ = $1 }
     | value IDENTIFIER { $$ = $1 + " " + $2 }

%%

// ---------------------------------------------------------------------------
// lexer
// ---------------------------------------------------------------------------

type lexer struct {
  input   string
  options map[string]string

  // context for user-friendly error messages
  line          string // the raw line being parsed
  pos           int    // byte offset where the next token starts in line
  eof           bool   // input exhausted; further Lex calls return 0
  lastToken     string // text of the token last returned by Lex
  lastTokenPos  int    // byte offset where lastToken started in line
  lastTokenEOF  bool   // true if the last Lex call returned EOF

  // queued tokens (used to expand the compact "var=value" form)
  queue []queuedToken
}

type queuedToken struct {
  tok   int
  value string
}

// Error is invoked by the generated parser when it cannot continue with the
// current token. It is called exactly once per rejected line, and the token
// the parser fails on is the last one returned by Lex (the parser reads
// one token ahead), so the lexer knows the exact context and can print a
// user-friendly, actionable message.
func (l *lexer) Error(e string) {
  offending, offPos := l.offendingToken()
  firstWord, _ := splitFirstWord(l.line)
  switch {
  case offPos == 0:
    // The first word of the line was rejected.
    if isKnownVariable(firstWord) {
      fmt.Fprintf(os.Stderr, "whcli: expected '=' after %s. Usage: %s = <value>. Type 'help' for usage.\n", firstWord, firstWord)
    } else if strings.Contains(l.line, "=") {
      fmt.Fprintf(os.Stderr, "whcli: unknown variable '%s'. Type 'help' for usage.\n", firstWord)
    } else if looksLikeWallpaperID(firstWord) {
      fmt.Fprintf(os.Stderr, "whcli: '%s' looks like a wallpaper ID. Try: info %s  or: download %s\n", firstWord, firstWord, firstWord)
    } else {
      fmt.Fprintf(os.Stderr, "whcli: unknown command '%s'. Type 'help' for usage.\n", firstWord)
    }
  case offending == "=":
    fmt.Fprintf(os.Stderr, "whcli: unexpected '='. Usage: <variable> = <value>. Type 'help' for usage.\n")
  case isKnownVariable(firstWord):
    // The line starts with a variable, so the parser was waiting for '='
    // or for a value (e.g. "query", "query =", "query foo bar").
    if strings.Contains(l.line, "=") {
      fmt.Fprintf(os.Stderr, "whcli: expected a value after '='. Usage: %s = <value>. Type 'help' for usage.\n", firstWord)
    } else if offending == "" {
      fmt.Fprintf(os.Stderr, "whcli: expected '=' after %s. Usage: %s = <value>. Type 'help' for usage.\n", firstWord, firstWord)
    } else {
      fmt.Fprintf(os.Stderr, "whcli: expected '=' after %s (got %s). Usage: %s = <value>. Type 'help' for usage.\n", firstWord, offending, firstWord)
    }
  case isCommandToken(firstWord):
    // The line starts with a known command, so the offending token was
    // found where an ID or nothing at all was expected
    // (e.g. "info", "info 123 456", "search extra").
    switch firstWord {
    case "info":
      fmt.Fprintf(os.Stderr, "whcli: 'info' takes exactly one wallpaper ID. Usage: info <ID>\n")
    case "download":
      fmt.Fprintf(os.Stderr, "whcli: 'download' takes exactly one wallpaper ID. Usage: download <ID>\n")
    case "tag":
      fmt.Fprintf(os.Stderr, "whcli: 'tag' takes exactly one tag ID. Usage: tag <ID>\n")
    case "env":
      fmt.Fprintf(os.Stderr, "whcli: 'env' takes exactly one variable. Usage: env <variable>\n")
    case "search":
      fmt.Fprintf(os.Stderr, "whcli: 'search' takes no arguments. Type 'help' for usage.\n")
    case "next":
      fmt.Fprintf(os.Stderr, "whcli: 'next' takes no arguments. Type 'help' for usage.\n")
    case "prev":
      fmt.Fprintf(os.Stderr, "whcli: 'prev' takes no arguments. Type 'help' for usage.\n")
    }
  case isCommandToken(offending):
    fmt.Fprintf(os.Stderr, "whcli: unexpected command %s. Type 'help' for usage.\n", offending)
  case isKnownVariable(offending):
    fmt.Fprintf(os.Stderr, "whcli: unexpected %s. Variables are set with '<variable> = <value>'. Type 'help' for usage.\n", offending)
  case offending != "":
    fmt.Fprintf(os.Stderr, "whcli: syntax error near '%s'. Type 'help' for usage.\n", offending)
  default:
    fmt.Fprintf(os.Stderr, "whcli: syntax error. Type 'help' for usage.\n")
  }
}

// offendingToken returns the token the parser failed on: the last token
// returned by Lex (the parser reads one token ahead), with its position.
// When the parser failed on end-of-input it returns ("", -1).
func (l *lexer) offendingToken() (string, int) {
  if l.lastTokenEOF {
    return "", -1
  }
  return l.lastToken, l.lastTokenPos
}

func (l *lexer) Lex(lval *yySymType) int {
  // emit queued tokens first (compact "var=value" expansion)
  if len(l.queue) > 0 {
    qt := l.queue[0]
    l.queue = l.queue[1:]
    if qt.tok == EQUALS {
      l.lastToken = "="
      l.lastTokenPos = l.pos
      l.lastTokenEOF = false
      return EQUALS
    }
    lval.Value = qt.value
    l.lastToken = qt.value
    l.lastTokenPos = l.pos
    l.lastTokenEOF = false
    return IDENTIFIER
  }

  if l.eof {
    l.lastTokenEOF = true
    return 0
  }
  // skip leading spaces
  for len(l.input) > 0 && unicode.IsSpace(rune(l.input[0])) {
    l.input = l.input[1:]
    l.pos++
  }
  if l.input == "" {
    l.eof = true
    l.lastTokenEOF = true
    return 0
  }

  // parse next token
  tokenValue, tokenLen := nextToken(l.input)
  l.pos += tokenLen
  l.input = l.input[tokenLen:]

  // remember the token for error reporting
  l.lastToken = tokenValue
  l.lastTokenPos = l.pos - tokenLen
  l.lastTokenEOF = false

  // accept the compact "var=value" form: expand it into VAR, EQUALS, VALUE
  if eq := strings.Index(tokenValue, "="); eq > 0 {
    if name := tokenValue[:eq]; isKnownVariable(name) {
      l.queue = []queuedToken{
        {tok: EQUALS},
        {tok: IDENTIFIER, value: tokenValue[eq+1:]},
      }
      lval.Parameter = name
      return variableToken(name)
    }
  }

  // accept "var =value" (no space after '='): split into EQUALS, VALUE
  if strings.HasPrefix(tokenValue, "=") && len(tokenValue) > 1 {
    l.queue = []queuedToken{{tok: IDENTIFIER, value: tokenValue[1:]}}
    return EQUALS
  }

  switch tokenValue {
  case "search":
    lval.Command = tokenValue
    return SEARCHCMD
  case "info":
    lval.Command = tokenValue
    return INFOCMD
  case "tag":
    lval.Command = tokenValue
    return TAGCMD
  case "prev":
    lval.Command = tokenValue
    return PREVCMD
  case "next":
    lval.Command = tokenValue
    return NEXTCMD
  case "download":
    lval.Command = tokenValue
    return DOWNCMD
  case "env":
    lval.Command = tokenValue
    return ENVCMD
  case "quit":
    lval.Command = tokenValue
    return QUITCMD
  case "help":
    lval.Command = tokenValue
    return HELPCMD
  case "resolution":
    lval.Parameter = tokenValue
    return RESPARM
  case "purity":
    lval.Parameter = tokenValue
    return PURITYPARM
  case "kind":
    lval.Parameter = tokenValue
    return KINDPARM
  case "query":
    lval.Parameter = tokenValue
    return QUERYPARM
  case "color":
    lval.Parameter = tokenValue
    return COLORPARM
  case "order":
    lval.Parameter = tokenValue
    return ORDERPARM
  case "sorting":
    lval.Parameter = tokenValue
    return SORTPARM
  case "path":
    lval.Parameter = tokenValue
    return PATHPARM
  case "=":
    return EQUALS
  default:
    lval.Value = tokenValue
    return IDENTIFIER
  }
}

// variableToken returns the goyacc token number for a known variable name.
func variableToken(name string) int {
  switch name {
  case "resolution":
    return RESPARM
  case "purity":
    return PURITYPARM
  case "kind":
    return KINDPARM
  case "query":
    return QUERYPARM
  case "color":
    return COLORPARM
  case "order":
    return ORDERPARM
  case "sorting":
    return SORTPARM
  case "path":
    return PATHPARM
  }
  return IDENTIFIER
}

// ---------------------------------------------------------------------------
// shell actions (shared with the REPL loop in cli.go)
// ---------------------------------------------------------------------------

func doInteractiveSearch() {
  resp, err := api.Search(CommandParameters)
  if err != nil {
    fmt.Fprintf(os.Stderr, "whcli: %s\n", apiErrorMessage("search", "", err))
    return
  }
  GenericSearchResults = resp
  if len(GenericSearchResults.Data) == 0 {
    fmt.Println("Search returned no items.")
  }
  CommandParameters.Seed = fmt.Sprintf("%s", GenericSearchResults.Meta.Seed)

  vP.SetData(GenericSearchResults)
  vP.Render()
}

func doInteractiveNext() {
  CommandParameters.Page += 1
  resp, err := api.Search(CommandParameters)
  if err != nil {
    fmt.Fprintf(os.Stderr, "whcli: %s\n", apiErrorMessage("next page request", "", err))
    return
  }
  GenericSearchResults = resp
  vP.SetData(GenericSearchResults)
  vP.Render()
}

func doInteractivePrev() {
  CommandParameters.Page -= 1
  if CommandParameters.Page < 1 {
    CommandParameters.Page = 1
  }
  resp, err := api.Search(CommandParameters)
  if err != nil {
    fmt.Fprintf(os.Stderr, "whcli: %s\n", apiErrorMessage("previous page request", "", err))
    return
  }
  GenericSearchResults = resp
  vP.SetData(GenericSearchResults)
  vP.Render()
}

func doInteractiveWallpaperInfo(id string) {
  CommandParameters.Id = id
  info, err := api.GetById(CommandParameters)
  if err != nil {
    fmt.Fprintf(os.Stderr, "whcli: %s\n", apiErrorMessage("wallpaper", id, err))
    return
  }
  api.DisplayWallpaperInfo(info)
}

func doInteractiveDownload(id string) {
  CommandParameters.Id = id
  if err := api.DownloadWallpaper(CommandParameters); err != nil {
    fmt.Fprintf(os.Stderr, "whcli: %s\n", apiErrorMessage("download of wallpaper", id, err))
  }
}

func doInteractiveTagInfo(id string) {
  CommandParameters.Id = id
  tag, err := api.GetTagById(CommandParameters)
  if err != nil {
    fmt.Fprintf(os.Stderr, "whcli: %s\n", apiErrorMessage("tag", id, err))
    return
  }
  api.DisplayTagInfo(tag)
}

// pendingAction holds the shell action recorded by the last parse.
// runShellLine executes it only after the whole line has been parsed,
// so a line with a trailing error (e.g. "search extra") never triggers
// its side effects.
var pendingAction func()

func showVariable(lx *lexer, name string) {
  if current, ok := lx.options[name]; ok {
      fmt.Fprintf(os.Stdout, "Variable %s is set to [ %s ]\n", name, current)
  } else {
      fmt.Fprintf(os.Stderr, "whcli: variable %s is not set. Type 'help' for usage.\n", name)
  }
}

// ---------------------------------------------------------------------------
// token helpers
// ---------------------------------------------------------------------------

func splitFirstWord(s string) (string, int) {
  index := 0
  for index < len(s) && !unicode.IsSpace(rune(s[index])) {
    index++
  }
  return s[0:index], index
}

func nextToken(s string) (string, int) {
  return splitFirstWord(s)
}

// looksLikeWallpaperID heuristically detects WallHaven wallpaper IDs
// (6 lowercase alphanumeric characters, e.g. "94x38z").
func looksLikeWallpaperID(token string) bool {
  if len(token) < 5 || len(token) > 7 {
    return false
  }
  for _, c := range token {
    if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
      return false
    }
  }
  return true
}

// isKnownVariable reports whether name is one of the settable variables.
func isKnownVariable(name string) bool {
  switch name {
  case "resolution", "purity", "kind", "query", "color", "order", "sorting", "path":
    return true
  }
  return false
}

// isCommandToken reports whether token is a shell command keyword.
func isCommandToken(token string) bool {
  switch token {
  case "search", "info", "tag", "prev", "next", "download", "env", "quit", "help":
    return true
  }
  return false
}

// ---------------------------------------------------------------------------
// value validation
// ---------------------------------------------------------------------------

var (
  validSortings = []string{"date_added", "relevance", "random", "views", "favorites", "toplist"}
  validOrders   = []string{"desc", "asc"}

  // official WallHaven palette (name -> hex)
  colorNames = map[string]string{
    "black":         "000000",
    "grey":          "cccccc",
    "white":         "ffffff",
    "navy":          "0066cc",
    "blue":          "0099cc",
    "cyan":          "66cccc",
    "teal":          "77cc33",
    "olive":         "669900",
    "green":         "336600",
    "lime":          "666600",
    "yellow":        "999900",
    "khaki":         "cccc33",
    "gold":          "ffff00",
    "yellow-orange": "ffcc33",
    "orange":        "ff9900",
    "brown-orange":  "ff6600",
    "red-brown":     "cc6633",
    "brown":         "996633",
    "maroon":        "663300",
    "red":           "cc0000",
    "red-pink":      "990000",
    "magenta":       "993399",
    "purple":        "663399",
    "violet":        "333399",
    "indigo":        "ea4c88",
    "silver":        "999999",
    "dark slate":    "424153",
  }
)

// validateValue checks a variable value against the WallHaven API rules.
// It returns (true, "") when the value is accepted, or (false, message)
// with a user-friendly explanation otherwise.
func validateValue(name, value string) (bool, string) {
  if value == "" {
    // empty values are only meaningful for query/color/sorting/order
    // (they reset the parameter); reject them everywhere else
    switch name {
    case "query", "color", "sorting", "order":
      return true, ""
    }
    return false, fmt.Sprintf("value for %s must not be empty.", name)
  }
  switch name {
  case "resolution":
    return validResolution(value)
  case "purity", "kind":
    return validBinaryTriplet(value, name)
  case "color":
    return validColor(value)
  case "sorting":
    return validOneOf(value, validSortings, name)
  case "order":
    return validOneOf(value, validOrders, name)
  case "query", "path":
    return true, ""
  default:
    return false, fmt.Sprintf("unknown variable %s. Type 'help' for usage.", name)
  }
}

func validOneOf(value string, allowed []string, name string) (bool, string) {
  for _, a := range allowed {
    if value == a {
      return true, ""
    }
  }
  return false, fmt.Sprintf("invalid value %s for %s. Valid values: %s.", value, name, strings.Join(allowed, ", "))
}

// validResolution accepts "WIDTHxHEIGHT" with positive integer dimensions.
func validResolution(value string) (bool, string) {
  v := strings.ToLower(strings.TrimSpace(value))
  i := strings.Index(v, "x")
  if i <= 0 || i == len(v)-1 {
    return false, fmt.Sprintf("invalid value %s for resolution. Expected WIDTHxHEIGHT, e.g. 1920x1080.", value)
  }
  w, errW := strconv.Atoi(v[:i])
  h, errH := strconv.Atoi(v[i+1:])
  if errW != nil || errH != nil || w <= 0 || h <= 0 {
    return false, fmt.Sprintf("invalid value %s for resolution. Expected WIDTHxHEIGHT, e.g. 1920x1080.", value)
  }
  return true, ""
}

// validBinaryTriplet accepts 3-character strings of 0/1 (e.g. purity 110, kind 101).
func validBinaryTriplet(value, name string) (bool, string) {
  msg := func() string {
    return fmt.Sprintf("invalid value %s for %s. Expected a 3-character string of 0 and 1, e.g. %s.", value, name, exampleFor(name))
  }
  if len(value) != 3 {
    return false, msg()
  }
  for _, c := range value {
    if c != '0' && c != '1' {
      return false, msg()
    }
  }
  return true, ""
}

func exampleFor(name string) string {
  if name == "purity" {
    return "100 (sfw only), 110 (sfw+sketchy), 111 (all)"
  }
  return "101 (general+anime), 010 (anime only), 111 (all)"
}

// validColor accepts a comma-separated list of 6-digit hex codes and/or
// color names from the official WallHaven palette.
func validColor(value string) (bool, string) {
  var problems []string
  for _, part := range strings.Split(value, ",") {
    c := strings.ToLower(strings.TrimSpace(part))
    if len(c) == 6 {
      ok := true
      for _, ch := range c {
        if !isHexDigit(ch) {
          ok = false
          break
        }
      }
      if ok {
        continue
      }
    }
    if _, found := colorNames[c]; found {
      continue
    }
    problems = append(problems, part)
  }
  if len(problems) > 0 {
    names := make([]string, 0, len(colorNames))
    for n := range colorNames {
      names = append(names, n)
    }
    sortStrings(names)
    return false, fmt.Sprintf("invalid color value(s): %s. Use 6-digit hex codes (e.g. 663399) and/or names: %s.",
      strings.Join(problems, ", "), strings.Join(names, ", "))
  }
  return true, ""
}

func isHexDigit(c rune) bool {
  return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')
}

func sortStrings(s []string) {
  // insertion sort: the lists are small
  for i := 1; i < len(s); i++ {
    j := i
    for j > 0 && s[j] < s[j-1] {
      s[j], s[j-1] = s[j-1], s[j]
      j--
    }
  }
}

// ---------------------------------------------------------------------------
// parameter bookkeeping
// ---------------------------------------------------------------------------

func updateParameters(paramName string, paramStore *common.QueryParameters, lx *lexer) {
  var value string = lx.options[paramName]
  switch paramName {
  case "resolution":
    paramStore.Resolution = value
  case "purity":
    paramStore.Purity = value
  case "kind":
    paramStore.Kind = value
  case "query":
    paramStore.Query = value
  case "color":
    paramStore.Color = value
  case "order":
    paramStore.Order = value
  case "sorting":
    paramStore.Sorting = value
  case "path":
    paramStore.Path = value
  }
}

// ---------------------------------------------------------------------------
// help
// ---------------------------------------------------------------------------

const shellHelpText = `Shell commands:
  search            re-run the search with the current parameters
  next              go to the next results page
  prev              go to the previous results page
  info <ID>         show metadata for a wallpaper (e.g. info 94x38z)
  download <ID>     download a wallpaper to the 'path' parameter directory
  tag <ID>          show information about a tag
  env <variable>    show the current value of a variable
  help              show this help
  quit              exit the shell

Settable variables (usage: <variable> = <value>  or  <variable>=<value>, e.g. purity = 110):
  resolution        minimum resolution, WIDTHxHEIGHT (e.g. 1920x1080)
  purity            0/1 triplet: 100 sfw, 110 +sketchy, 111 +nsfw (nsfw needs an API key)
  kind              0/1 triplet for general/anime/people (e.g. 101)
  query             search query, may contain spaces (e.g. anime +summer -people)
  color             color filter: 6-digit hex codes and/or names (e.g. red,0066cc)
  sorting           date_added | relevance | random | views | favorites | toplist
  order             desc | asc (only applies to date_added, views, favorites)
  path              directory where wallpapers are downloaded

Notes:
  - Parameter changes take effect on the next 'search', 'next' or 'prev'.
  - Invalid values are rejected with an explanation; nothing is applied.
  - To clear query, color, sorting or order: type '<variable> =' (empty value).
  - Type Ctrl-D to exit the shell.
`

func printShellHelp() {
  fmt.Fprint(os.Stdout, shellHelpText)
}
