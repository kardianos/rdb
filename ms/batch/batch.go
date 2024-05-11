// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// Package batch functions for interacting with batched sql statements
// in the same file or string.
package batch

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"unicode"

	"github.com/kardianos/rdb"
)

// ExecuteSQL runs the batchSql on the connection pool on a single
// connection after separating out each commend, joined with separator.
func ExecuteSQL(ctx context.Context, cp *rdb.ConnPool, batchSQL, separator string) error {
	ss := SplitSQL(batchSQL, separator)
	cmd := &rdb.Command{
		Arity: rdb.Zero,
	}

	conn, err := cp.Connection()
	if err != nil {
		return err
	}
	defer conn.Close()

	for i := range ss {
		cmd.SQL = ss[i]

		_, err = conn.Query(ctx, cmd)
		if err != nil {
			if errList, is := err.(rdb.Errors); is {
				return SQLErrorWithContext(cmd.SQL, errList, 2)
			}
			return err
		}
	}
	return nil
}

// SQLErrorWithContext highlights errors in the SQL script displaying
// the number lines of contextLines for each error.
func SQLErrorWithContext(sql string, msg rdb.Errors, contextLines int) error {
	if contextLines < 0 {
		contextLines = 0
	}

	localMsg := &bytes.Buffer{}
	sqlLines := strings.Split(sql, "\n")
	for _, msg := range msg {
		localMsg.WriteString(msg.String())
		localMsg.WriteString("; context:\n")
		lineIndex := int(msg.LineNumber) - 1
		min := lineIndex - contextLines
		max := lineIndex + contextLines
		if min < 0 {
			min = 0
		}
		if max >= len(sqlLines) {
			max = len(sqlLines) - 1
		}
		for ci := min; ci <= max; ci++ {
			if ci == lineIndex {
				localMsg.WriteString("-->")
			} else {
				localMsg.WriteString("   ")
			}
			localMsg.WriteString(sqlLines[ci])
			localMsg.WriteRune('\n')
		}
	}

	return errors.New(localMsg.String())
}

// SplitCmd takes a single command and uses separator to split them
// into mutliple commands.
func SplitCmd(cmd *rdb.Command, separator string) []*rdb.Command {
	sql := cmd.SQL
	localCmd := *cmd
	localCmd.SQL = ""

	ss := SplitSQL(sql, separator)
	ret := make([]*rdb.Command, len(ss))
	for i, item := range ss {
		itemCmd := localCmd
		itemCmd.SQL = item
		ret[i] = &itemCmd
	}

	return ret
}

// SplitSQL takes SQL text and splits it with separator.
func SplitSQL(sql, separator string) []string {
	if len(separator) == 0 || len(sql) < len(separator) {
		return []string{sql}
	}
	l := &lexer{
		SQL: sql,
		Sep: separator,
		At:  0,
	}
	state := stateWhitespace
	for state != nil {
		state = state(l)
	}
	l.AddCurrent()
	return l.Batch
}

type lexer struct {
	SQL   string
	Sep   string
	At    int
	Start int

	Batch []string
}

func (l *lexer) Add(b string) {
	if len(b) == 0 {
		return
	}
	l.Batch = append(l.Batch, b)
}

func (l *lexer) Next() bool {
	l.At++
	return (l.At < len(l.SQL))
}
func (l *lexer) AddCurrent() bool {
	l.Add(l.SQL[l.Start:l.At])
	l.At += len(l.Sep)
	l.Start = l.At
	return (l.At < len(l.SQL))
}

type stateFn func(*lexer) stateFn

const (
	lineComment  = "--"
	leftComment  = "/*"
	rightComment = "*/"
)

func stateText(l *lexer) stateFn {
	for {
		ch := l.SQL[l.At]

		switch {
		case strings.HasPrefix(l.SQL[l.At:], lineComment):
			l.At += len(lineComment)
			return stateLineComment
		case strings.HasPrefix(l.SQL[l.At:], leftComment):
			l.At += len(leftComment)
			return stateMultiComment
		case ch == '\'':
			l.At++
			return stateString
		case ch == '\r', ch == '\n':
			l.At++
			return stateWhitespace
		default:
			if !l.Next() {
				return nil
			}
		}
	}
}
func stateWhitespace(l *lexer) stateFn {
	if l.At >= len(l.SQL) {
		return nil
	}
	ch := l.SQL[l.At]

	switch {
	case ch == ' ', ch == '\t', ch == '\r', ch == '\n':
		l.At++
		return stateWhitespace
	case hasPrefixFold(l.SQL[l.At:], l.Sep):
		next := l.SQL[l.At+len(l.Sep):]
		space := len(next) == 0 || unicode.IsSpace(rune(next[0]))
		if !space {
			return stateText
		}
		if l.AddCurrent() {
			return stateWhitespace
		}
		return nil
	default:
		return stateText
	}
}
func stateLineComment(l *lexer) stateFn {
	for {
		ch := l.SQL[l.At]

		switch {
		case ch == '\r', ch == '\n':
			l.At++
			return stateWhitespace
		default:
			if !l.Next() {
				return nil
			}
		}
	}
}
func stateMultiComment(l *lexer) stateFn {
	for {
		switch {
		case strings.HasPrefix(l.SQL[l.At:], rightComment):
			l.At += len(leftComment)
			return stateWhitespace
		default:
			if l.Next() == false {
				return nil
			}
		}
	}
}
func stateString(l *lexer) stateFn {
	for {
		ch := l.SQL[l.At]
		chNext := rune(-1)
		if l.At+1 < len(l.SQL) {
			chNext = rune(l.SQL[l.At+1])
		}

		switch {
		case ch == '\'' && chNext == '\'':
			l.At += 2
		case ch == '\'' && chNext != '\'':
			l.At++
			return stateWhitespace
		default:
			if !l.Next() {
				return nil
			}
		}
	}
}

func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[0:len(prefix)], prefix)
}
