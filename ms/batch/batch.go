// Copyright 2014 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// Proides utility functions for interacting with batched sql statements
// in the same file or string.
package batch

import (
	"bytes"
	"errors"
	"strings"

	"bitbucket.org/kardianos/rdb"
)

func ExecuteBatchSql(cp *rdb.ConnPool, batchSql, separator string) error {
	ss := BatchSplitSql(batchSql, separator)
	cmd := &rdb.Command{
		Arity: rdb.Zero,
	}

	conn, err := cp.Connection()
	if err != nil {
		return err
	}
	defer conn.Close()

	for i := range ss {
		cmd.Sql = ss[i]

		_, err = conn.Query(cmd)
		if err != nil {
			if errList, is := err.(rdb.Errors); is {
				return SqlErrorWithContext(cmd.Sql, errList, 2)
			}
			return err
		}
	}
	return nil
}

func SqlErrorWithContext(sql string, msg rdb.Errors, contextLines int) error {
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

func BatchSplitCmd(cmd *rdb.Command, separator string) []*rdb.Command {
	sql := cmd.Sql
	localCmd := *cmd
	localCmd.Sql = ""

	ss := BatchSplitSql(sql, separator)
	ret := make([]*rdb.Command, len(ss))
	for i, item := range ss {
		itemCmd := localCmd
		itemCmd.Sql = item
		ret[i] = &itemCmd
	}

	return ret
}

func BatchSplitSql(sql, separator string) []string {
	if len(separator) == 0 || len(sql) < len(separator) {
		return []string{sql}
	}
	l := &lexer{
		Sql: sql,
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
	Sql   string
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
	return (l.At < len(l.Sql))
}
func (l *lexer) AddCurrent() bool {
	l.Add(l.Sql[l.Start:l.At])
	l.At += len(l.Sep)
	l.Start = l.At
	return (l.At < len(l.Sql))
}

type stateFn func(*lexer) stateFn

const (
	lineComment  = "--"
	leftComment  = "/*"
	rightComment = "*/"
)

func stateText(l *lexer) stateFn {
	for {
		ch := l.Sql[l.At]

		switch {
		case strings.HasPrefix(l.Sql[l.At:], lineComment):
			l.At += len(lineComment)
			return stateLineComment
		case strings.HasPrefix(l.Sql[l.At:], leftComment):
			l.At += len(leftComment)
			return stateMultiComment
		case ch == '\'':
			l.At += 1
			return stateString
		case ch == '\r', ch == '\n':
			l.At += 1
			return stateWhitespace
		default:
			if l.Next() == false {
				return nil
			}
		}
	}
}
func stateWhitespace(l *lexer) stateFn {
	if l.At >= len(l.Sql) {
		return nil
	}
	ch := l.Sql[l.At]

	switch {
	case ch == ' ', ch == '\t', ch == '\r', ch == '\n':
		l.At += 1
		return stateWhitespace
	case strings.HasPrefix(l.Sql[l.At:], l.Sep):
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
		ch := l.Sql[l.At]

		switch {
		case ch == '\r', ch == '\n':
			l.At += 1
			return stateWhitespace
		default:
			if l.Next() == false {
				return nil
			}
		}
	}
}
func stateMultiComment(l *lexer) stateFn {
	for {
		switch {
		case strings.HasPrefix(l.Sql[l.At:], rightComment):
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
		ch := l.Sql[l.At]
		chNext := rune(-1)
		if l.At+1 < len(l.Sql) {
			chNext = rune(l.Sql[l.At+1])
		}

		switch {
		case ch == '\'' && chNext == '\'':
			l.At += 2
		case ch == '\'' && chNext != '\'':
			l.At += 1
			return stateWhitespace
		default:
			if l.Next() == false {
				return nil
			}
		}
	}
}
