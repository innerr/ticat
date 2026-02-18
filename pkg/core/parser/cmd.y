%{
package parser

import (
	"fmt"
	"strings"

	"github.com/innerr/ticat/pkg/core/model"
)

type yaccParseContext struct {
	cmdParser    *CmdParser
	envParser    *EnvParser
	cmds         *model.CmdTree
	envAbbrs     *model.EnvAbbrs
	currCmd      *model.CmdTree
	currEnvAbbrs *model.EnvAbbrs
	matchedPath  []string
	trivialLvl   int
	err          error
	isMinorErr   bool
	allowSub     bool
	rest         []string
	argIdx       int
}

%}

%union {
	str     string
	seg     parsedSeg
	segs    []parsedSeg
	env     model.ParsedEnv
	isEmpty bool
}

%token <str> WORD
%token <str> ENV
%token SEP

%type <segs> input segments
%type <seg> segment

%%

input:
	{
		$$ = nil
	}
	| segments
	{
		$$ = $1
		yylex.(*yyLex).SetResult($$)
	}

segments:
	segment
	{
		if $1.Type == parsedSegTypeSep && len($$) == 0 {
			$$ = nil
		} else if $1.Type == parsedSegTypeEnv {
			if env, ok := $1.Val.(model.ParsedEnv); !ok || len(env) == 0 {
				$$ = nil
			} else {
				$$ = []parsedSeg{$1}
			}
		} else {
			$$ = []parsedSeg{$1}
		}
	}
	| segments segment
	{
		if $2.Type == parsedSegTypeEnv {
			if env, ok := $2.Val.(model.ParsedEnv); !ok || len(env) == 0 {
				$$ = $1
			} else {
				$$ = append($1, $2)
			}
		} else if $2.Type == parsedSegTypeSep {
			if len($1) > 0 && $1[len($1)-1].Type == parsedSegTypeSep {
				$$ = $1
			} else {
				$$ = append($1, $2)
			}
		} else {
			$$ = append($1, $2)
		}
	}

segment:
	WORD
	{
		$$ = yylex.(*yyLex).ProcessWord($1)
	}
	| ENV
	{
		$$ = yylex.(*yyLex).ProcessEnv($1)
		// After processing env, allow sub command
		yylex.(*yyLex).ctx.allowSub = true
	}
	| SEP
	{
		$$ = parsedSeg{Type: parsedSegTypeSep, Val: nil}
		// After separator, allow sub command
		yylex.(*yyLex).ctx.allowSub = true
	}

%%

type yyLex struct {
	tokens     []yyToken
	pos        int
	result     []parsedSeg
	ctx        *yaccParseContext
	hasError   bool
}

type yyToken struct {
	typ int
	str string
}

func newYYLex(tokens []yyToken, ctx *yaccParseContext) *yyLex {
	return &yyLex{
		tokens: tokens,
		pos:    0,
		ctx:    ctx,
	}
}

func (l *yyLex) Lex(lval *yySymType) int {
	if l.pos >= len(l.tokens) || l.hasError {
		return 0
	}
	tok := l.tokens[l.pos]
	l.pos++
	lval.str = tok.str
	return tok.typ
}

func (l *yyLex) Error(s string) {
	l.hasError = true
}

func (l *yyLex) SetResult(segs []parsedSeg) {
	l.result = segs
}

func (l *yyLex) ProcessWord(word string) parsedSeg {
	ctx := l.ctx
	
	if ctx.allowSub && ctx.currCmd != nil {
		sub := ctx.currCmd.GetSub(word)
		if sub != nil {
			ctx.currCmd = sub
			if ctx.currEnvAbbrs != nil {
				ctx.currEnvAbbrs = ctx.currEnvAbbrs.GetSub(word)
			}
			ctx.matchedPath = append(ctx.matchedPath, word)
			ctx.allowSub = false
			return parsedSeg{Type: parsedSegTypeCmd, Val: model.MatchedCmd{Name: word, Cmd: sub}}
		}
	}
	
	if ctx.allowSub {
		errStr := "unknow input '" + word + "' ..., should be sub cmd"
		ctx.err = fmt.Errorf("[CmdParser.parse] %s: %s", l.displayPath(), errStr)
		ctx.isMinorErr = false
		l.hasError = true
		return parsedSeg{Type: parsedSegTypeCmd, Val: model.MatchedCmd{Name: word}}
	}
	
	env := l.tryParseAsArg(word)
	if env != nil {
		return parsedSeg{Type: parsedSegTypeEnv, Val: env}
	}
	
	return parsedSeg{Type: parsedSegTypeCmd, Val: model.MatchedCmd{Name: word}}
}

func (l *yyLex) tryParseAsArg(word string) model.ParsedEnv {
	ctx := l.ctx
	if ctx.currCmd == nil || ctx.currCmd.Cmd() == nil {
		return nil
	}
	
	args := ctx.currCmd.Args()
	if args.IsEmpty() {
		return nil
	}
	
	env := model.ParsedEnv{}
	
	realName := args.Realname(word)
	if len(realName) > 0 {
		if l.pos < len(l.tokens) {
			nextTok := l.tokens[l.pos]
			if nextTok.typ == WORD {
				l.pos++
				env[realName] = model.NewParsedEnvArgv(word, nextTok.str)
				return env
			}
		}
	}
	
	names := args.Names()
	if len(names) > ctx.argIdx {
		name := names[ctx.argIdx]
		env[name] = model.NewParsedEnvArgv(name, word)
		ctx.argIdx++
		return env
	}
	
	return nil
}

func (l *yyLex) ProcessEnv(envStr string) parsedSeg {
	ctx := l.ctx
	env := parseEnvString(envStr, ctx.cmdParser.envParser)
	
	// Debug: print env parsing result
	// fmt.Printf("ProcessEnv: %q -> env=%v\n", envStr, env)
	
	if len(ctx.matchedPath) > 0 && env != nil {
		for k, v := range env {
			env[k] = model.ParsedEnvVal{
				Val:            v.Val,
				IsArg:          v.IsArg,
				IsSysArg:       v.IsSysArg,
				MatchedPath:    append(ctx.matchedPath, v.MatchedPath...),
				MatchedPathStr: strings.Join(append(ctx.matchedPath, v.MatchedPath...), ctx.cmdParser.cmdSep),
			}
		}
	}
	
	return parsedSeg{Type: parsedSegTypeEnv, Val: env}
}

func (l *yyLex) displayPath() string {
	ctx := l.ctx
	if len(ctx.matchedPath) == 0 {
		return ctx.cmdParser.cmdRootNodeName
	}
	return strings.Join(ctx.matchedPath, ctx.cmdParser.cmdSep)
}

func parseEnvString(s string, envParser *EnvParser) model.ParsedEnv {
	if !strings.HasPrefix(s, envParser.brackets.Left) || !strings.HasSuffix(s, envParser.brackets.Right) {
		return nil
	}
	
	content := s[len(envParser.brackets.Left) : len(s)-len(envParser.brackets.Right)]
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		return nil
	}
	
	env := make(model.ParsedEnv)
	parts := splitBySep(content, envParser.kvSep)
	
	i := 0
	for i < len(parts) {
		part := strings.TrimSpace(parts[i])
		if i+2 < len(parts) && strings.TrimSpace(parts[i+1]) == envParser.kvSep {
			key := part
			// Collect all parts until the next key (which doesn't start with '=')
			valParts := []string{strings.TrimSpace(parts[i+2])}
			i += 3
			// Continue collecting value parts if they contain '=' (meaning '=' in value)
			for i < len(parts) {
				nextPart := strings.TrimSpace(parts[i])
				if nextPart == envParser.kvSep && i+1 < len(parts) {
					// This '=' is part of the value, add it
					valParts[len(valParts)-1] += envParser.kvSep + strings.TrimSpace(parts[i+1])
					i += 2
				} else {
					break
				}
			}
			val := strings.TrimSpace(strings.Join(valParts, ""))
			env[key] = model.NewParsedEnvVal(key, val)
		} else if i+1 < len(parts) && strings.Contains(part, envParser.kvSep) {
			// Handle inline key=value (split on first '=' only)
			kv := strings.SplitN(part, envParser.kvSep, 2)
			if len(kv) == 2 {
				env[strings.TrimSpace(kv[0])] = model.NewParsedEnvVal(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]))
			}
			i++
		} else {
			i++
		}
	}
	
	return env
}

func splitBySep(s string, sep string) []string {
	var parts []string
	current := ""
	i := 0
	for i < len(s) {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			if len(current) > 0 {
				parts = append(parts, current)
			}
			parts = append(parts, sep)
			current = ""
			i += len(sep)
		} else {
			current += string(s[i])
			i++
		}
	}
	if len(current) > 0 {
		parts = append(parts, current)
	}
	return parts
}
