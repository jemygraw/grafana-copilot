package ernie

import (
	"bufio"
	"bytes"
	"strings"
)

const (
	TextCodeTagPrefix    = "```text"
	TextCodeTagDelimiter = "```"
)

const (
	JsonCodeTagPrefix    = "```json"
	JsonCodeTagDelimiter = "```"
)

var textStartPrefixList = []string{
	TextCodeTagPrefix,
	TextCodeTagDelimiter,
}

var jsonStartPrefixList = []string{
	JsonCodeTagPrefix,
	JsonCodeTagDelimiter,
}

func GetResponseTextContent(r string) (dst string) {
	var startPrefix string
	var startIndex = -1
	var endIndex = -1
	for _, prefix := range textStartPrefixList {
		startIndex = strings.Index(r, prefix)
		if startIndex != -1 {
			startPrefix = prefix
			break
		}
		// otherwise, fallback to next prefix check
	}
	if startIndex != -1 {
		// find the end index
		endIndex = strings.LastIndex(r, TextCodeTagDelimiter)
	} else {
		// non-standard text output
		dst = r
	}
	if (startIndex == -1 || endIndex == -1) || startIndex >= endIndex {
		dst = r
	} else {
		dst = r[startIndex+len(startPrefix) : endIndex]
	}
	dst = strings.TrimSpace(dst)
	return
}

func GetResponseJsonContent(r string) (dst string) {
	var startPrefix string
	var startIndex = -1
	var endIndex = -1
	for _, prefix := range jsonStartPrefixList {
		startIndex = strings.Index(r, prefix)
		if startIndex != -1 {
			startPrefix = prefix
			break
		}
		// otherwise, fallback to next prefix check
	}
	if startIndex != -1 {
		// find the end index
		endIndex = strings.LastIndex(r, JsonCodeTagDelimiter)
	} else {
		// non-standard json output
		return scanLinesToExtractJsonContent(r)
	}
	if (startIndex == -1 || endIndex == -1) || startIndex >= endIndex {
		dst = r
	} else {
		dst = r[startIndex+len(startPrefix) : endIndex]
	}
	dst = strings.TrimSpace(dst)
	return
}

func scanLinesToExtractJsonContent(r string) (dst string) {
	scanner := bufio.NewScanner(bytes.NewReader([]byte(r)))
	var jsonCloseTag string
	var jsonOpenIndex int
	for scanner.Scan() {
		line := scanner.Text()
		// json string startswith a `{` for object and `[` for array.
		if strings.HasPrefix(line, "{") {
			jsonCloseTag = "}"
			jsonOpenIndex = strings.Index(r, "{")
			break
		} else if strings.HasPrefix(line, "[") {
			jsonCloseTag = "]"
			jsonOpenIndex = strings.Index(r, "[")
			break
		}
	}
	if jsonOpenIndex == -1 {
		// failed to find json open tag
		dst = r
		return
	}
	// find the json close tag
	jsonCloseIndex := strings.LastIndex(r, jsonCloseTag)
	if jsonCloseIndex == -1 {
		// failed to find json close tag
		dst = r
		return
	}
	// check start and end index
	if jsonCloseIndex <= jsonOpenIndex {
		// failed to parse out json string
		dst = r
		return
	}

	// extract the json output
	dst = r[jsonOpenIndex : jsonCloseIndex+1]
	return
}
