package main

import (
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"os"
	"sort"

	"math/rand"
	"net"
	"strings"
	"unicode"
)

func main() {
	topleveldomains := []string{".com", ".com.br"}

	words := []string{
		"luna", "galaxia", "cosmos", "digital", "galaxy", "cosmos", "pixel",
	}

	domains := generateDomains(words, topleveldomains)

	err := createFileDomains(domains)
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

}

func generateDomains(words, topleveldomains []string) []string {
	mapperDomain := make(map[string]bool)
	allDomainsGenerated := []string{}

	for i := 0; i < len(words)*len(words)*len(words); i++ {
		domainsGenerated := generateDomain(words, topleveldomains)
		above50 := false
		validDomains := []string{}
		numberValidDomain := 0
		for _, domain := range domainsGenerated {
			if _, exists := mapperDomain[domain]; !exists {
				mapperDomain[domain] = true
				if checkDomain(domain) {
					validDomains = append(validDomains, domain)
					numberValidDomain++
				}
				continue
			}
		}

		above50 = validPercentage(numberValidDomain, len(topleveldomains))
		displayDomains(validDomains, above50)
		allDomainsGenerated = append(allDomainsGenerated, validDomains...)
	}
	return allDomainsGenerated
}

func validPercentage(num1, num2 int) bool {
	return float64(num1) >= 0.70*float64(num2)
}

func displayDomains(domains []string, above50 bool) {

	for _, domain := range domains {
		if above50 {
			color.Green("%s disponível", domain)
		} else {
			fmt.Println(domain, " disponível")
		}
	}

}

func generateDomain(words, topleveldomains []string) []string {
	p1 := words[rand.Intn(len(words))]
	p2 := words[rand.Intn(len(words))]
	var domains []string

	for _, top := range topleveldomains {
		domain := fmt.Sprintf("%s%s%s", removeAccentsANDuplicates(p1), removeAccentsANDuplicates(p2), top)
		domains = append(domains, domain)
	}

	return domains
}

func checkDomain(dominio string) bool {
	addrs, err := net.LookupHost(dominio)
	if err != nil {
		return true
	}

	if len(addrs) == 0 {
		return true
	}

	return false
}

func removeAccentsANDuplicates(str string) string {

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, str)

	var resultadoFinal strings.Builder
	for i := 0; i < len(result); i++ {
		if i == 0 || result[i] != result[i-1] {
			resultadoFinal.WriteByte(result[i])
		}
	}

	return strings.ToLower(resultadoFinal.String())
}

func createFileDomains(domains []string) error {
	file, err := os.Create("domains.txt")
	if err != nil {
		return fmt.Errorf("erro ao criar o file: %w", err)
	}
	defer file.Close()

	orderStrings(domains)
	for _, dominio := range domains {
		_, err = fmt.Fprintln(file, dominio)
		if err != nil {
			return fmt.Errorf("erro ao escrever no file: %w", err)
		}
	}

	return nil
}

func orderStrings(s []string) {
	sort.Strings(s)
}
