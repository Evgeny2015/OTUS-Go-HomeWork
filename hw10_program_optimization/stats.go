package hw10programoptimization

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

//easyjson:json
type User struct {
	ID       int
	Name     string
	Username string
	Email    string
	Phone    string
	Password string
	Address  string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (result DomainStat, err error) {
	result = make(DomainStat)
	scanner := bufio.NewScanner(r)
	domain = "." + domain

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		email, err := GetEMail(line)
		if err != nil {
			return nil, fmt.Errorf("get users error: %w", err)
		}

		err = CountDomain(email, domain, result)
		if err != nil {
			return nil, fmt.Errorf("get domain error: %w", err)
		}
	}

	if err = scanner.Err(); err != nil {
		return
	}

	return result, err
}

func GetEMail(line string) (result string, err error) {

	index := strings.Index(line, "Email")
	if index < 0 {
		return "", fmt.Errorf("EMail not found")
	}

	rest := line[index+8:]
	index = strings.Index(rest, "\"")
	if index < 0 {
		return "", fmt.Errorf("JSON error")
	}
	result = rest[:index]

	return
}

func CountDomain(email string, domain string, stat DomainStat) error {
	matched := strings.Contains(email, domain)

	if matched {
		key := strings.ToLower(strings.SplitN(email, "@", 2)[1])

		num := stat[key]
		num++
		stat[key] = num
	}

	return nil
}
