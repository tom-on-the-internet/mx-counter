package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var invalidEmailErr = errors.New("invalid email")

type kv struct {
	Key   string
	Value int
}

func main() {
	err := act()
	if err != nil {
		panic(err)
	}

	os.Exit(0)
}

func act() error {
	r, err := getReadCloser()
	if err != nil {
		return err
	}

	defer r.Close()

	emails, err := readEmails(r)
	if err != nil {
		return err
	}

	emails = unique(emails)
	emails = valid(emails)

	domains, err := uniqueDomains(emails)
	if err != nil {
		return err
	}

	mailDomains := getMailDomains(domains)

	domainCounts, err := getDomainCounts(emails, mailDomains)
	if err != nil {
		return err
	}

	orderedCounts := getOrderedCounts(domainCounts)

	output(orderedCounts)

	return nil
}

func getReadCloser() (io.ReadCloser, error) {
	var err error

	file := os.Stdin

	if len(os.Args) > 1 {
		file, err = os.Open(os.Args[1])
		if err != nil {
			return nil, err
		}
	}

	return file, nil
}

func readEmails(r io.Reader) ([]string, error) {
	var emails []string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		emails = append(emails, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return emails, err
	}

	return emails, nil
}

func unique(items []string) []string {
	keys := make(map[string]bool)
	unique := []string{}

	for _, email := range items {
		if _, ok := keys[email]; !ok {
			keys[email] = true

			unique = append(unique, email)
		}
	}

	return unique
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

func getDomain(email string) (string, error) {
	segments := strings.Split(email, "@")

	if len(segments) < 2 {
		return "", invalidEmailErr
	}

	return segments[1], nil
}

func uniqueDomains(emails []string) ([]string, error) {
	domains := make([]string, len(emails))

	for i, email := range emails {
		domain, err := getDomain(email)
		if err != nil {
			return domains, err
		}

		domains[i] = domain
	}

	return unique(domains), nil
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

			mailDomain := parts[len(parts)-3] + "." + parts[len(parts)-2]
			c <- result{domain, mailDomain}
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

func getDomainCounts(emails []string, mailDomains map[string]string) (map[string]int, error) {
	m := make(map[string]int)

	for _, email := range emails {
		domain, err := getDomain(email)
		if err != nil {
			return m, err
		}
		mailDomain := mailDomains[domain]

		if len(mailDomain) > 0 {
			m[mailDomain]++
		}
	}

	return m, nil
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

func output(orderedCounts []kv) {
	for _, count := range orderedCounts {
		fmt.Println(count.Key, count.Value)
	}
}
