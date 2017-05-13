package command

import (
	"bytes"
	"regexp"
	"strconv"
)

func Serialize(str string) bytes.Buffer {
	tokens := regexp.MustCompile("\".*\"|\\S+").FindAllString(str, -1)
	var buffer bytes.Buffer
	if len(tokens) > 0 {
		buffer.WriteString("*")
		buffer.WriteString(strconv.Itoa(len(tokens)))
		buffer.WriteString("\r\n")
		for _, tok := range tokens {
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
