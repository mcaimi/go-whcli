%{
package whcli

import (
        "os"
        "fmt"
        "unicode"
        "regexp"
        "github.com/mcaimi/go-whcli/pkg/api"
)
%}

%union{
Command string
Argument string
}

%token <Command> COMMAND
%token <Argument> ARGUMENT
%type <Argument> argument

%start expression

%%
expression:
          |
          COMMAND {
            switch $1 {
            case "next":
              CommandParameters.Page += 1;
              GenericSearchResults = api.Search(CommandParameters);
              vP.SetData(GenericSearchResults)
              vP.Render()
            case "prev":
              CommandParameters.Page -= 1;
              if CommandParameters.Page < 1 { CommandParameters.Page = 1; }
              GenericSearchResults = api.Search(CommandParameters);
              vP.SetData(GenericSearchResults)
              vP.Render()
            case "quit":
              os.Exit(0);
            default:
              break
            }
          }
          |
          COMMAND argument {
            switch $1 {
            case "info":
              CommandParameters.Id = $2
              api.DisplayWallpaperInfo(api.GetById(CommandParameters));
            case "get":
              CommandParameters.Id = $2
              CommandParameters.Path = "./"
              api.DownloadWallpaper(CommandParameters);
            case "tag":
              CommandParameters.Id = $2
              api.DisplayTagInfo(api.GetTagById(CommandParameters));
            default:
              break;
            }
          }

argument:
        ARGUMENT {
          $$ = $1
        }
%%

type lexer struct {
  input string
  command string
  argument string
  parserError bool
}

type token struct {
  matcher *regexp.Regexp
  tokenType int
}

var tokens = []token {
  token{ matcher: regexp.MustCompile("next|prev|info|get|tag|quit"), tokenType: COMMAND },
  token{ matcher: regexp.MustCompile("(?i)[a-z0-9]+"), tokenType: ARGUMENT },
}

func (l *lexer) Error(e string) {
  l.parserError = true;
  fmt.Println(e);
}

func isEndOfInput(s string) bool {
  if len(s) == 0 {
    return true;
  } else {
    return false;
  }
}

func (l *lexer) Lex(lval *yySymType) int {
  // skip spaces
  for ; len(l.input) > 0 && unicode.IsSpace(rune(l.input[0])); l.input = l.input[1:] {}

  if isEndOfInput(l.input) { return 0; }

  // parse command
  for _, tok := range tokens {
    strmatch := tok.matcher.FindString(l.input);
    if strmatch != "" {
      // pass value back to yacc
      if tok.tokenType == COMMAND {
        lval.Command = strmatch;
      } else {
        lval.Argument = strmatch;
      }
      l.input = l.input[len(strmatch):];
      return tok.tokenType;
    }
  }

  // no match
  return 0;
}

