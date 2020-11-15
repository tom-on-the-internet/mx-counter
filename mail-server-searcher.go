package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type kv struct {
	Key   string
	Value int
}

func main() {
	reader, err := getReader()
	if err != nil {
		log.Fatal(err)
	}

	emails, err := readEmails(reader)
	if err != nil {
		log.Fatal(err)
	}

	emails = unique(emails)
	emails = valid(emails)
	domains := uniqueDomains(emails)
	mailDomains := getMailDomains(domains)
	domainCounts := getDomainCounts(emails, mailDomains)
	orderedCounts := getOrderedCounts(domainCounts)

	render(orderedCounts)

	os.Exit(0)
}

func getReader() (*os.File, error) {
	var err error

	f := os.Stdin

	if len(os.Args) > 1 {
		f, err = os.Open(os.Args[1])
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}

func readEmails(f *os.File) ([]string, error) {
	defer f.Close()

	var emails []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		emails = append(emails, scanner.Text())
	}

	return emails, nil
}

func unique(emails []string) []string {
	keys := make(map[string]bool)
	uniqueEmails := []string{}

	for _, email := range emails {
		if _, value := keys[email]; !value {
			keys[email] = true

			uniqueEmails = append(uniqueEmails, email)
		}
	}

	return uniqueEmails
}

func valid(emails []string) []string {
	var validEmails []string

	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	for _, email := range emails {
		if len(email) > 2 && len(email) < 255 && emailRegex.MatchString(email) {
			validEmails = append(validEmails, email)
		}
	}

	return validEmails
}

func getDomain(email string) string {
	return strings.Split(email, "@")[1]
}

func uniqueDomains(emails []string) []string {
	domains := make([]string, len(emails))

	for i, email := range emails {
		domains[i] = getDomain(email)
	}

	return unique(domains)
}

func getMailDomains(domains []string) map[string]string {
	var wg sync.WaitGroup

	type result struct {
		Key   string
		Value string
	}

	c := make(chan result)
	m := make(map[string]string)

	for _, domain := range domains {
		wg.Add(1)

		go func(wg *sync.WaitGroup, domain string) {
			defer wg.Done()

			mxs, err := net.LookupMX(domain)
			if err != nil {
				return
			}

			minimumDomainParts := 3

			parts := strings.Split(mxs[0].Host, ".")
			if len(parts) < minimumDomainParts {
				return
			}

			d := parts[len(parts)-3] + "." + parts[len(parts)-2]
			c <- result{domain, d}
		}(&wg, domain)
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	for r := range c {
		m[r.Key] = r.Value
	}

	return m
}

func getDomainCounts(emails []string, mailDomains map[string]string) map[string]int {
	m := make(map[string]int)

	for _, email := range emails {
		domain := getDomain(email)
		mailDomain := mailDomains[domain]

		if len(mailDomain) > 0 {
			m[mailDomain]++
		}
	}

	return m
}

func getOrderedCounts(domainCounts map[string]int) []kv {
	orderedCounts := make([]kv, 0, len(domainCounts))

	for k, v := range domainCounts {
		orderedCounts = append(orderedCounts, kv{k, v})
	}

	sort.Slice(orderedCounts, func(i, j int) bool {
		return orderedCounts[i].Value > orderedCounts[j].Value
	})

	return orderedCounts
}

func render(orderedCounts []kv) {
	for _, count := range orderedCounts {
		fmt.Println(count.Key, count.Value)
	}
}
