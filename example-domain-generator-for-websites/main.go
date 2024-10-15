package main

import (
	"fmt"
	"github.com/bobesa/go-domain-util/domainutil"
	"github.com/fatih/color"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"os"
	"sort"

	"net"
	"strings"
	"unicode"
)

func main() {
	topleveldomains := []string{".com", ".com.br"}

	words := []string{
		"%¨&*$@#$%#$¨,", "luna", "cosmos", "digital", "digital", "galaxy", "cosmos", "pixel", "android", "golang", "go", "google.com", "google.com.br", "1234567890",
	}

	domains := generateDomains(words, topleveldomains)

	err := createFileDomains(domains)
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}
}

func generateDomains(words, topleveldomains []string) []string {

	words = sanitizeWords(words)
	domainsGenerated := generateDomain(words, topleveldomains)

	allDomainsGenerated := []string{}
	for _, domain := range domainsGenerated {

		validDomains := []string{}

		if !isDomainAcquired(domain) {
			validDomains = append(validDomains, domain)
			color.Green("%s disponível", domain)

		}

		allDomainsGenerated = append(allDomainsGenerated, validDomains...)
	}

	return allDomainsGenerated
}

func sanitizeWords(words []string) []string {
	words = removeDuplicates(words)
	var newWords []string
	for _, word := range words {
		newWord := sanitizeWord(word)
		if len(newWord) > 0 {
			newWords = append(newWords, newWord)
		}
	}

	return newWords
}

func generateDomain(words, topleveldomains []string) []string {
	domains := make([]string, 0)
	seenDomains := make(map[string]bool)

	for _, top := range topleveldomains {
		domains = append(domains, generateSingleWordDomains(words, top)...)
		domains = append(domains, generateTwoWordDomains(words, top, seenDomains)...)
	}

	sort.Strings(domains)
	return domains
}

func generateSingleWordDomains(words []string, topLevelDomain string) []string {
	var domains []string
	for _, word := range words {
		if len(word) >= 3 {
			domains = append(domains, fmt.Sprintf("%s%s", word, topLevelDomain))
		}
	}
	return domains
}

func generateTwoWordDomains(words []string, topLevelDomain string, seenDomains map[string]bool) []string {
	var domains []string
	for i := 0; i < len(words)-1; i++ {
		for j := i + 1; j < len(words); j++ {
			domain1 := fmt.Sprintf("%s%s%s", words[i], words[j], topLevelDomain)
			domain2 := fmt.Sprintf("%s%s%s", words[j], words[i], topLevelDomain)

			if !seenDomains[domain1] {
				domains = append(domains, domain1)
				seenDomains[domain1] = true
			}

			if !seenDomains[domain2] {
				domains = append(domains, domain2)
				seenDomains[domain2] = true
			}
		}
	}
	return domains
}

func isDomainAcquired(dominio string) bool {
	addrs, err := net.LookupHost(dominio)
	if err != nil {
		return false
	}

	if len(addrs) == 0 {
		return false
	}

	return true
}

func sanitizeWord(str string) string {
	validChars := "abcdefghijklmnopqrstuvwxyz0123456789-"

	str = strings.ToLower(str)

	if len(domainutil.DomainPrefix(str)) != 0 {
		str = domainutil.DomainPrefix(str)
	}

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, str)

	var resultadoFinal strings.Builder
	for _, char := range result {
		if strings.ContainsRune(validChars, unicode.ToLower(char)) {
			resultadoFinal.WriteRune(char)
		}
	}

	return resultadoFinal.String()
}

func createFileDomains(domains []string) error {
	file, err := os.Create("domains.txt")
	if err != nil {
		return fmt.Errorf("erro ao criar o file: %w", err)
	}
	defer file.Close()

	textInit := fmt.Sprintf("as suas sugestões geraram %d domínios válidos para aquicição", len(domains))

	_, err = fmt.Fprintln(file, textInit)
	if err != nil {
		return fmt.Errorf("erro ao escrever no file: %w", err)
	}

	for _, dominio := range domains {
		_, err = fmt.Fprintln(file, dominio)
		if err != nil {
			return fmt.Errorf("erro ao escrever no file: %w", err)
		}
	}

	return nil
}

func removeDuplicates(s []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for _, str := range s {
		if !encountered[str] {
			encountered[str] = true
			result = append(result, str)
		}
	}
	return result
}
