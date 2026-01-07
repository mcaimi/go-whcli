%{
package whcli

import (
        "os"
        "fmt"
        "unicode"
        "codeberg.org/mcaimi/go-whcli/pkg/api"
        "codeberg.org/mcaimi/go-whcli/pkg/common"
)
%}

%union{
Command string
Parameter string
Value string
}

%token <Command> SEARCHCMD INFOCMD TAGCMD PREVCMD NEXTCMD DOWNCMD ENVCMD QUITCMD
%token <Parameter> RESPARM PURITYPARM QUERYPARM KINDPARM COLORPARM ORDERPARM SORTPARM AIPARM PATHPARM
%token <Value> IDENTIFIER
%token EQUALS

%type <Parameter> variable
%start startcommand

%%
startcommand: commands
            | assignment

commands:
          QUITCMD {
              os.Exit(0);
          }
          | SEARCHCMD {
              GenericSearchResults = api.Search(CommandParameters);
              if len(GenericSearchResults.Data) == 0 {
                  fmt.Println("Search returned no items.")
              }
              CommandParameters.Seed = fmt.Sprintf("%s", GenericSearchResults.Meta.Seed);

              vP.SetData(GenericSearchResults);
              vP.Render();
          }
          | NEXTCMD {
              CommandParameters.Page += 1;
              GenericSearchResults = api.Search(CommandParameters);
              vP.SetData(GenericSearchResults)
              vP.Render()
          }
          | PREVCMD {
              CommandParameters.Page -= 1;
              if CommandParameters.Page < 1 { CommandParameters.Page = 1; }
              GenericSearchResults = api.Search(CommandParameters);
              vP.SetData(GenericSearchResults)
              vP.Render()
          }
          | INFOCMD IDENTIFIER {
              CommandParameters.Id = $2
              api.DisplayWallpaperInfo(api.GetById(CommandParameters));
          }
          | DOWNCMD IDENTIFIER {
              CommandParameters.Id = $2
              api.DownloadWallpaper(CommandParameters);
          }
          | TAGCMD IDENTIFIER {
              CommandParameters.Id = $2
              api.DisplayTagInfo(api.GetTagById(CommandParameters));
          }
          | ENVCMD variable {
              fmt.Printf("Variable { %s } is set to value [ %s ]\n", $2, yylex.(*lexer).options[$2]);
          }

assignment:
        variable EQUALS IDENTIFIER {
          yylex.(*lexer).options[$1] = $3;
          updateParameters($1, &CommandParameters, yylex.(*lexer));
        }

variable:
     RESPARM | PURITYPARM | KINDPARM | QUERYPARM | COLORPARM | ORDERPARM | SORTPARM | AIPARM | PATHPARM {
      $$ = $1
    }

%%

type lexer struct {
  input string
  options map[string]string
  parserError bool
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

func updateParameters(paramName string, paramStore *common.QueryParameters, lx *lexer) {
  var value string = lx.options[paramName];
  switch paramName {
  case "resolution":
      paramStore.Resolution = value;
      break;
  case "purity":
      paramStore.Purity = value;
      break;
  case "kind":
      paramStore.Kind = value;
      break;
  case "query":
      paramStore.Query = value;
      break;
  case "color":
      paramStore.Color = value;
      break;
  case "order":
      paramStore.Order = value;
      break;
  case "sorting":
      paramStore.Sorting = value;
      break;
  case "aifilter":
    if lx.options["aifilter"] == "true" {
      paramStore.AiArt = true;
    } else {
      paramStore.AiArt = false;
    }
    break;
  case "path":
      paramStore.Path = value;
      break;
  default:
    break;
  }
}

func nextToken(s string) (string, int) {
  var index int;
  var temp string = s[:];
  for index = 0; len(temp) > 0 && !(unicode.IsSpace(rune(temp[0]))); index++ { temp = temp[1:]; }

  // return next token
  return s[0:index], index;
}

func (l *lexer) Lex(lval *yySymType) int {
  // skip leading spaces
  for ; len(l.input) > 0 && unicode.IsSpace(rune(l.input[0])); l.input = l.input[1:] {}

  if isEndOfInput(l.input) { return 0; }

  // parse command
  tokenValue, tokenLen := nextToken(l.input);
  l.input = l.input[tokenLen:];
  switch(tokenValue) {
  case "search":
    lval.Command = tokenValue;
    return SEARCHCMD;
  case "info":
    lval.Command = tokenValue;
    return INFOCMD;
  case "tag":
    lval.Command = tokenValue;
    return TAGCMD;
  case "prev":
    lval.Command = tokenValue;
    return PREVCMD;
  case "next":
    lval.Command = tokenValue;
    return NEXTCMD;
  case "download":
    lval.Command = tokenValue;
    return DOWNCMD;
  case "env":
    lval.Command = tokenValue;
    return ENVCMD;
  case "quit":
    lval.Command = tokenValue;
    return QUITCMD;
  case "resolution":
    lval.Parameter = tokenValue;
    return RESPARM;
  case "purity":
    lval.Parameter = tokenValue;
    return PURITYPARM;
  case "kind":
    lval.Parameter = tokenValue;
    return KINDPARM;
  case "query":
    lval.Parameter = tokenValue;
    return QUERYPARM;
  case "color":
    lval.Parameter = tokenValue;
    return COLORPARM;
  case "order":
    lval.Parameter = tokenValue;
    return ORDERPARM;
  case "sorting":
    lval.Parameter = tokenValue;
    return SORTPARM;
  case "aifilter":
    lval.Parameter = tokenValue;
    return AIPARM;
  case "path":
    lval.Parameter = tokenValue;
    return PATHPARM;
  case "=":
    return EQUALS;
  default:
    lval.Value = tokenValue;
    return IDENTIFIER;
  }
}

