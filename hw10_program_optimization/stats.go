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

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	u, err := getUsers(r)
	if err != nil {
		return nil, fmt.Errorf("get users error: %w", err)
	}
	return countDomains(u, domain)
}

type users [100_000]User

func getUsers(r io.Reader) (result users, err error) {
	scanner := bufio.NewScanner(r)
	i := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var user User
		if err = user.UnmarshalJSON([]byte(line)); err != nil {
			return
		}
		result[i] = user
		i++
	}
	if err = scanner.Err(); err != nil {
		return
	}
	return
}

func countDomains(u users, domain string) (DomainStat, error) {
	result := make(DomainStat)

	for _, user := range u {
		matched := strings.Contains(user.Email, "."+domain)

		if matched {
			key := strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])

			num := result[key]
			num++
			result[key] = num
		}
	}
	return result, nil
}
