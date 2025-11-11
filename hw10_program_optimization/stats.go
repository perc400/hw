package hw10programoptimization

import (
	"bufio"
	"io"
	"strings"

	easyjson "github.com/mailru/easyjson" //nolint:depguard
)

//easyjson:json
type User struct {
	ID       int    `json:"Id"`       //nolint:tagliatelle
	Name     string `json:"Name"`     //nolint:tagliatelle
	Username string `json:"Username"` //nolint:tagliatelle
	Email    string `json:"Email"`    //nolint:tagliatelle
	Phone    string `json:"Phone"`    //nolint:tagliatelle
	Password string `json:"Password"` //nolint:tagliatelle
	Address  string `json:"Address"`  //nolint:tagliatelle
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	return countDomains(r, domain)
}

func countDomains(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)

	domainSuffix := "." + domain

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var user User
		if err := easyjson.Unmarshal(scanner.Bytes(), &user); err != nil {
			return nil, err
		}
		domainHasSuffix := strings.HasSuffix(strings.ToLower(user.Email), domainSuffix)

		if domainHasSuffix {
			if atIndex := strings.IndexByte(user.Email, '@'); atIndex != -1 {
				domainPart := strings.ToLower(user.Email[atIndex+1:])
				num := result[domainPart]
				num++
				result[domainPart] = num
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
