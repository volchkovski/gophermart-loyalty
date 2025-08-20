package valid

import (
	"unicode"
)

func OrderNumber(number string) bool {
	sum := 0
	parity := len(number) % 2 // Это сделает проверку четности позиции правильно для строки

	for i, r := range number {
		if !unicode.IsDigit(r) {
			return false
		}

		digit := int(r - '0')
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}
