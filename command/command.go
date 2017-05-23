package command

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
)

type Command struct {
	tokens []string
}

func NewCommand(str string) Command {
	return Command{
		regexp.MustCompile("\".*\"|\\S+").FindAllString(str, -1),
	}
}

func (command *Command) IsQuit() bool {
	return len(command.tokens) > 0 && strings.ToUpper(command.tokens[0]) == "QUIT"
}

func (command *Command) Serialize() bytes.Buffer {
	var buffer bytes.Buffer
	if len(command.tokens) > 0 {
		buffer.WriteString("*")
		buffer.WriteString(strconv.Itoa(len(command.tokens)))
		buffer.WriteString("\r\n")
		for _, tok := range command.tokens {
			tok = unquote(tok)
			buffer.WriteString("$")
			buffer.WriteString(strconv.Itoa(len(tok)))
			buffer.WriteString("\r\n")
			buffer.WriteString(tok)
			buffer.WriteString("\r\n")
		}
	}
	return buffer
}

func unquote(str string) string {
	if len(str) > 0 {
		if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
			return str[1 : len(str)-1]
		}
	}
	return str
}
