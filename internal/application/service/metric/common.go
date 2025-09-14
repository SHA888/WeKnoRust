package metric

import (
	"regexp"
	"strings"

	"github.com/Tencent/WeKnowRust/internal/types"
)

func sum(m map[string]int) int {
	s := 0
	for _, v := range m {
		s += v
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func splitSentences(text string) []string {
    // Compile regex to match Chinese period or English period
    re := regexp.MustCompile(`([。.])`)

    // Split text while keeping delimiters for positioning
    split := re.Split(text, -1)

	var sentences []string
	current := strings.Builder{}

    // Alternate between text segments and delimiters (odd indices are delimiters)
    for i, s := range split {
        if i%2 == 0 {
            current.WriteString(s)
        } else {
            // When encountering a delimiter, finalize the current sentence
            if current.Len() > 0 {
                sentence := strings.TrimSpace(current.String())
                if sentence != "" {
                    sentences = append(sentences, sentence)
                }
                current.Reset()
            }
        }
    }

    // Handle the last segment without a trailing delimiter
    if remaining := strings.TrimSpace(current.String()); remaining != "" {
        sentences = append(sentences, remaining)
    }

	return sentences
}

func splitIntoWords(sentences []string) []string {
    // Regex to match Chinese blocks, English blocks, and punctuation
    re := regexp.MustCompile(`([\p{Han}]+)|([a-zA-Z0-9_.,!?]+)|(\p{P})`)

	var tokens []string
	for _, text := range sentences {
		matches := re.FindAllStringSubmatch(text, -1)

		for _, groups := range matches {
			chineseBlock := groups[1]
			englishBlock := groups[2]
			punctuation := groups[3]

			            switch {
            case chineseBlock != "": // Handle Chinese segment
                words := types.Jieba.Cut(chineseBlock, true)
                tokens = append(tokens, words...)
            case englishBlock != "": // Handle English segment
                engTokens := strings.Fields(englishBlock)
                tokens = append(tokens, engTokens...)
            case punctuation != "": // Keep punctuation
                tokens = append(tokens, punctuation)
            }
        }
    }
    return tokens
}

func ToSet[T comparable](li []T) map[T]struct{} {
	res := make(map[T]struct{}, len(li))
	for _, v := range li {
		res[v] = struct{}{}
	}
	return res
}

func SliceMap[T any, Y any](li []T, fn func(T) Y) []Y {
	res := make([]Y, len(li))
	for i, v := range li {
		res[i] = fn(v)
	}
	return res
}

func Hit[T comparable](li []T, set map[T]struct{}) int {
	count := 0
	for _, v := range li {
		if _, exist := set[v]; exist {
			count++
		}
	}
	return count
}

func Fold[T any, Y any](slice []T, initial Y, f func(Y, T) Y) Y {
	accumulator := initial
	for _, item := range slice {
		accumulator = f(accumulator, item)
	}
	return accumulator
}
