package main

import (
	"fmt"
	"slices"
	"strings"
)


func ValidateChirp(body string) (string, error) {
	if len(body) > 140  {
		return "", fmt.Errorf("Chirp is too long")
	}

	cleanedBody := CleanChirp(body)
	return cleanedBody, nil
}

func CleanChirp(text string) string {
	blackList := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(text, " ")
	for i, word := range words {
		if slices.Contains(blackList, strings.ToLower(word)) {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

