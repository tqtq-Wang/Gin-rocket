package repository

import (
	"errors"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

var (
	ErrCacheMiss         = errors.New("cache miss")
	ErrUserAlreadyExists = errors.New("user already exists")
)

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
