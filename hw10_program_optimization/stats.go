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

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		user, err := getUser(line)
		if err != nil {
			return nil, fmt.Errorf("get users error: %w", err)
		}

		err = countDomains(user, domain, result)
		if err != nil {
			return nil, fmt.Errorf("get domain error: %w", err)
		}
	}

	if err = scanner.Err(); err != nil {
		return
	}

	return result, err
}

func getUser(line string) (result User, err error) {

	if err = result.UnmarshalJSON([]byte(line)); err != nil {
		return
	}

	return
}

func countDomains(user User, domain string, stat DomainStat) error {
	matched := strings.Contains(user.Email, "."+domain)

	if matched {
		key := strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])

		num := stat[key]
		num++
		stat[key] = num
	}

	return nil
}
