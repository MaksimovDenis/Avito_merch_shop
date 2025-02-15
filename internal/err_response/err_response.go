package errresponse

import (
	"errors"
	"fmt"
	"strings"
)

func ErrResponse(err error, args ...any) error {
	var errorMessage string

	switch {
	case strings.Contains(err.Error(), "violates check constraint"):
		errorMessage = "недостаточно средств для покупки"
	case strings.Contains(err.Error(), "no rows in result set"):
		errorMessage = fmt.Sprintf("%v не найден", args)
	default:
		errorMessage = "ошибка при обновлении данных"
	}

	return errors.New(errorMessage)
}
