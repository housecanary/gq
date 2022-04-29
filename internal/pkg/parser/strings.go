package parser

import (
	"bufio"
	"strconv"
	"strings"
)

func parseString(image string) string {
	if strings.HasPrefix(image, `"""`) {
		return parseBlockString(image)
	}

	return parseSimpleString(image)
}

func parseBlockString(image string) string {
	s := image[3 : len(image)-3]
	s = strings.ReplaceAll(s, `\"""`, `"""`)
	return trimBlockString(s)
}

func trimBlockString(rawValue string) string {
	// Let {lines} be the result of splitting {rawValue} by {LineTerminator}.
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(rawValue))
	for sc.Scan() {
		line := sc.Text()
		lines = append(lines, line)
	}

	// Let {commonIndent} be {null}.
	commonIndent := -1

	// For each {line} in {lines}:
	//   - If {line} is the first item in {lines}, continue to the next line.
	//   - Let {length} be the number of characters in {line}.
	//   - Let {indent} be the number of leading consecutive {WhiteSpace} characters in {line}.
	//     - If {indent} is less than {length}:
	//       - If {commonIndent} is {null} or {indent} is less than {commonIndent}:
	//         - Let {commonIndent} be {indent}.
	for _, line := range lines[1:] {
		for i := 0; i < len(line); i++ {
			c := line[i]
			if c == ' ' || c == '\t' {
				continue
			}
			if commonIndent == -1 || i < commonIndent {
				commonIndent = i
				break
			}
		}
	}

	// If {commonIndent} is not {null}:
	if commonIndent != -1 {
		// For each {line} in {lines}:
		//   - If {line} is the first item in {lines}, continue to the next line.
		//   - Remove {commonIndent} characters from the beginning of {line}.
		for i, line := range lines[1:] {
			indent := commonIndent
			if indent > len(line) {
				lines[i+1] = ""
				continue
			}
			lines[i+1] = line[commonIndent:]
		}
	}

	// While the first item {line} in {lines} contains only {WhiteSpace}: Remove the first item from {lines}.
	for len(lines) > 0 {
		line := lines[0]
		hasNonSpace := false
		for i := 0; i < len(line); i++ {
			if line[i] != ' ' && line[i] != '\t' {
				hasNonSpace = true
				break
			}
		}

		if hasNonSpace {
			break
		}

		lines = lines[1:]
	}

	// While the last item {line} in {lines} contains only {WhiteSpace}: Remove the last item from {lines}.
	for len(lines) > 0 {
		line := lines[len(lines)-1]
		hasNonSpace := false
		for i := 0; i < len(line); i++ {
			if line[i] != ' ' && line[i] != '\t' {
				hasNonSpace = true
				break
			}
		}

		if hasNonSpace {
			break
		}

		lines = lines[0 : len(lines)-1]
	}

	// Let {formatted} be the empty character sequence.
	formatted := strings.Builder{}

	// For each {line} in {lines}
	for i, line := range lines {

		// If {line} is the first item in {lines}:
		if i == 0 {
			// Append {formatted} with {line}.
			formatted.WriteString(line)
		} else { // Otherwise
			// Append {formatted} with a line feed character (U+000A).
			formatted.WriteString("\n")
			// Append {formatted} with {line}.
			formatted.WriteString(line)
		}
	}

	return formatted.String()
}

func parseSimpleString(image string) string {
	sb := strings.Builder{}
	sb.Grow(len(image) - 2)

	for i := 1; i < len(image)-1; i++ {
		c := image[i]
		if c != '\\' {
			sb.WriteByte(c)
			continue
		}

		escapeCode := image[i+1]
		i += 1
		switch escapeCode {
		case '"', '/', '\\':
			sb.WriteByte(escapeCode)
		case 'b':
			sb.WriteByte('\b')
		case 'f':
			sb.WriteByte('\f')
		case 'n':
			sb.WriteByte('\n')
		case 'r':
			sb.WriteByte('\r')
		case 't':
			sb.WriteByte('\t')
		case 'u':
			charCode, _ := strconv.ParseUint(image[i+1:i+5], 16, 32)
			sb.WriteRune(rune(charCode))
			i += 4
		}
	}
	return sb.String()
}
