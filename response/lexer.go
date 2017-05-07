package response

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// Errors.
type LexerError struct {
	index   int
	message string
}

func (e LexerError) Error() string {
	return fmt.Sprintf("(At [%d]): %s", e.index, e.message)
}

func eofError(index int) LexerError {
	return LexerError{index, "Unexpected end of input"}
}

func expectedError(index int, actual byte, expected byte, rest ...byte) LexerError {
	var buffer bytes.Buffer
	buffer.WriteString("Expected ")
	buffer.WriteByte('\'')
	buffer.WriteByte(expected)
	buffer.WriteByte('\'')
	for _, b := range rest {
		buffer.WriteString(", ")
		buffer.WriteByte('\'')
		buffer.WriteByte(b)
		buffer.WriteByte('\'')
	}
	buffer.WriteByte(';')
	buffer.WriteString(" found ")
	buffer.WriteByte('\'')
	buffer.WriteByte(actual)
	buffer.WriteByte('\'')

	return LexerError{index, fmt.Sprintf("%q", buffer.String())}
}

// Helpers.
type ByteReader struct {
	index  int
	buffer []byte
	reader io.Reader
}

func NewByteReader(reader io.Reader) ByteReader {
	return ByteReader{0, make([]byte, 1, 41), reader}
}

func (r *ByteReader) Read() (byte, error) {
	n, err := r.reader.Read(r.buffer)
	if n > 0 {
		r.index = r.index + n
	}
	return r.buffer[0], err
}

// Tokens.
type TokenType int

const (
	ERROR = iota

	INT
	FLOAT
	SIMPLE_STRING
	BULK_STRING

	LIST
	SET
)

type Token struct {
	tokenType TokenType
	value     []byte
}

func (t Token) String() string {
	return string(t.value)
}

func Lex(r io.Reader) ([]Token, error) {
	reader := NewByteReader(r)

	c, err := reader.Read()
	if err == nil {
		switch c {
		case '-':
			token, err := lexSimpleString(reader)
			return []Token{token}, err
		case ':':
			token, err := lexInt(reader)
			return []Token{token}, err
		case ';':
			token, err := lexFloat(reader)
			return []Token{token}, err
		case '+':
			token, err := lexSimpleString(reader)
			return []Token{token}, err
		case '$':
			token, err := lexBulkString(reader)
			return []Token{token}, err
		case '*':
			return lexArray(reader)
		default:
			return []Token{}, expectedError(reader.index, c, '-', ':', ';', '+', '$', '*', '&')
		}
	}

	return []Token{}, eofError(reader.index)
}

// Private lexing functions.

func lexToken(reader ByteReader) (Token, error) {
	c, err := reader.Read()
	if err == nil {
		switch c {
		case ':':
			return lexInt(reader)
		case ';':
			return lexFloat(reader)
		case '+':
			return lexSimpleString(reader)
		case '$':
			return lexBulkString(reader)
		default:
			return Token{}, expectedError(reader.index, c, ':', ';', '+', '$')
		}
	}
	return Token{}, eofError(reader.index)
}

func lexInt(reader ByteReader) (Token, error) {
	num, err := readInt(reader)
	if err == nil {
		return Token{INT, []byte(strconv.Itoa(num))}, nil
	}
	return Token{}, err
}

func lexFloat(reader ByteReader) (Token, error) {
	num, err := readFloat(reader)
	if err == nil {
		return Token{FLOAT, []byte(fmt.Sprintf("%f", num))}, nil
	}
	return Token{}, err
}

func lexSimpleString(reader ByteReader) (Token, error) {
	var value bytes.Buffer
	for {
		c, err := reader.Read()
		if err == nil {
			switch c {
			case '\r':
				if c, err := reader.Read(); err == nil && c == '\n' {
					return Token{SIMPLE_STRING, value.Bytes()}, nil
				} else {
					return Token{}, LexerError{reader.index, "Simple string must end with \\r\\n"}
				}
			case '\n':
				return Token{}, LexerError{reader.index, "Invalid line feed in simple string"}
			default:
				value.WriteByte(c)
			}
		} else {
			return Token{}, eofError(reader.index)
		}
	}
}

func lexBulkString(reader ByteReader) (Token, error) {
	len, err := readInt(reader)
	if err == nil {

		var buffer bytes.Buffer
		for i := 0; i < len; i++ {
			c, err := reader.Read()
			if err != nil {
				return Token{}, LexerError{reader.index, fmt.Sprintf("Invalid bulk string of length %d", len)}
			}
			buffer.WriteByte(c)
		}

		cr, _ := reader.Read()
		lf, _ := reader.Read()
		if cr != '\r' || lf != '\n' {
			return Token{}, LexerError{reader.index, "Bulk string must end with \\r\\n"}
		}

		return Token{BULK_STRING, buffer.Bytes()}, nil
	}
	return Token{}, err
}

func lexArray(reader ByteReader) ([]Token, error) {
	var tokens []Token
	len, err := readInt(reader)
	if err == nil {
		for i := 0; i < len; i++ {
			t, err := lexToken(reader)
			if err == nil {
				tokens = append(tokens, t)
			} else {
				return tokens, err
			}
		}
		return tokens, nil
	}
	return []Token{}, err
}

// Reading of data directly to types.

func readInt(reader ByteReader) (int, error) {
	str, err := lexSimpleString(reader)
	if err == nil {
		num, err := strconv.Atoi(string(str.value))
		if err == nil {
			return num, nil
		} else {
			return -1, LexerError{reader.index, "Invalid Integer"}
		}
	} else {
		return -1, err
	}
}

func readFloat(reader ByteReader) (float64, error) {
	str, err := lexSimpleString(reader)
	if err == nil {
		num, err := strconv.ParseFloat(string(str.value), 64)
		if err == nil {
			return num, nil
		} else {
			return -1, LexerError{reader.index, "Invalid Float"}
		}
	} else {
		return -1, err
	}
}
