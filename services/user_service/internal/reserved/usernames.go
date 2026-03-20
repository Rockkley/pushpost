package reserved

import (
	"golang.org/x/text/unicode/norm"
	"strings"
)

// usernames — зарезервированные имена пользователей.
// Защищают URL-пространство от захвата пользователями.
// Список обновляется при добавлении новых маршрутов в gateway.
//
// Хранится в user_service — это бизнес-логика продукта,
// а не инфраструктурный код.
var usernames = map[string]struct{}{
	// gateway
	"auth":    {},
	"users":   {},
	"friends": {},
	"api":     {},
	// common
	"admin":    {},
	"me":       {},
	"settings": {},
	"profile":  {},
	"login":    {},
	"logout":   {},
	"register": {},
	"static":   {},
	"favicon":  {},
	// product
	"support": {},
	"help":    {},
	"about":   {},
	"terms":   {},
	"privacy": {},
}

// IsReserved возвращает true если нормализованное имя зарезервировано.
func IsReserved(username string) bool {
	_, ok := usernames[normalize(username)]
	return ok
}

func normalize(username string) string {
	return strings.ToLower(strings.TrimSpace(norm.NFC.String(username)))
}
