package source

import (
	"regexp"
	"strconv"
)

// Error represents an errror in the srouce code
type Error struct {
	File         string
	Line, Column int
	Message      string
}

/*
Parse a message that is almost the standard in error messages that are outputed by
most modern compilers and tools that work with source code

format is either

xxx.yyy:01:01: some message
{filename}.{fileext}:{line}:{column}: {message}

or
xxx.yyy:01: some message
{filename}.{fileext}:{line}: {message}

*/
var validMessage = regexp.MustCompile(`([[:alnum:]]+.[[:alnum:]]+):([0-9]+):([0-9]+)?:? (.*)`)

// ParseSourceErrors takes the log of a process and
// returns it's sourcecode errors
func ParseSourceErrors(message string) []Error {
	var errors []Error
	messages := validMessage.FindAllStringSubmatch(message, -1)
	for _, message := range messages {
		e := Error{
			File: message[1],
		}
		if line, err := strconv.Atoi(message[2]); err == nil {
			e.Line = line
		}
		if column, err := strconv.Atoi(message[3]); err == nil {
			e.Column = column
		} else {
			e.Column = -1
			e.Message = message[3]
		}
		if len(message) > 3 {
			e.Message = message[4]
		}
		errors = append(errors, e)
	}
	return errors
}