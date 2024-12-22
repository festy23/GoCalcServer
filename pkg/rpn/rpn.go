package rpn

import (
	"errors"
	"strconv"
	"unicode"
)

// Calc - функция, которая вычисляет результат выражения expression, делает проверку его корректности
func Calc(expression string) (float64, error) {

	if !IsValidExpression(expression) {
		return 0.0, errors.New("invalid expression")
	}

	res, err := calculateExpression(expression)
	return res, err
}

// IsValidExpression Проверка корректности поданного на вход выражения
func IsValidExpression(expression string) bool {
	cntBraces := 0           // Проверка корректности скобок
	prevCharIsDigit := false // Последний символ — число
	inNumber := false        // Находимся ли внутри числа

	for i, c := range expression {
		if unicode.IsSpace(c) {
			return false // Пробелы не допускаются
		}

		if unicode.IsDigit(c) {
			prevCharIsDigit = true
			inNumber = true
		} else if c == '.' {
			if !inNumber {
				return false // Точка не может быть первым символом
			}
			if i == 0 || i == len(expression)-1 {
				return false // Точка не может быть первым или последним символом
			}
			// Проверка, чтобы предыдущий символ был цифрой, а следующий - тоже
			if !unicode.IsDigit(rune(expression[i-1])) || (i+1 < len(expression) && !unicode.IsDigit(rune(expression[i+1]))) {
				return false
			}
		} else if isValidMathOperation(c) {
			if !prevCharIsDigit { // Не может быть оператора после другого оператора
				if c != '-' || (i > 0 && expression[i-1] != '(' && !unicode.IsDigit(rune(expression[i-1]))) {
					// Разрешаем минус только в начале или после '(' или если предшествующий символ не является цифрой
					return false
				}
			}
			prevCharIsDigit = false
			inNumber = false
		} else if c == '(' {
			cntBraces++
			prevCharIsDigit = false
		} else if c == ')' {
			cntBraces--
			if cntBraces < 0 {
				return false // Некорректные скобки
			}
			prevCharIsDigit = true
		} else {
			return false // Недопустимый символ
		}
	}

	return cntBraces == 0 && prevCharIsDigit // Проверка на баланс скобок и последний символ
}

func isValidMathOperation(c rune) bool {
	return c == '+' || c == '-' || c == '*' || c == '/'
}

// Структура выражения, состоящая из чисел и операций
type stack struct {
	numbers []float64
	ops     []rune
}

func calculateExpression(expression string) (float64, error) {
	var s stack
	i := 0
	for i < len(expression) {
		if unicode.IsSpace(rune(expression[i])) {
			i++
			continue
		}

		if unicode.IsDigit(rune(expression[i])) || expression[i] == '.' {
			j := i
			for j < len(expression) && (unicode.IsDigit(rune(expression[j])) || expression[j] == '.') {
				j++
			}
			num, err := strconv.ParseFloat(expression[i:j], 64)
			if err != nil {
				return 0, err
			}
			s.numbers = append(s.numbers, num)
			i = j
		} else if isNegativeSign(expression, i) {
			num, err := parseNegativeNumber(expression, &i)
			if err != nil {
				return 0, err
			}
			s.numbers = append(s.numbers, num)
		} else if expression[i] == '(' {
			s.ops = append(s.ops, '(')
			i++
		} else if expression[i] == ')' {
			for len(s.ops) > 0 && s.ops[len(s.ops)-1] != '(' {
				if err := evaluate(&s); err != nil {
					return 0, err
				}
			}
			s.ops = s.ops[:len(s.ops)-1] // Remove '('
			i++
		} else {
			for len(s.ops) > 0 && priorityOfOperation(s.ops[len(s.ops)-1]) >= priorityOfOperation(rune(expression[i])) {
				if err := evaluate(&s); err != nil {
					return 0, err
				}
			}
			s.ops = append(s.ops, rune(expression[i]))
			i++
		}
	}

	for len(s.ops) > 0 {
		if err := evaluate(&s); err != nil {
			return 0, err
		}
	}

	if len(s.numbers) != 1 {
		return 0, errors.New("некорректное выражение")
	}
	return s.numbers[0], nil
}

// Evaluate выполняет последнюю операцию из стека
func evaluate(s *stack) error {
	// Проверяем, достаточно ли чисел и операторов для выполнения операции
	if len(s.numbers) < 2 || len(s.ops) == 0 {
		return errors.New("некорректное выражение")
	}
	a := s.numbers[len(s.numbers)-2]
	b := s.numbers[len(s.numbers)-1]
	op := s.ops[len(s.ops)-1]

	// Удаляем использованные операнды и оператор из стеков
	s.numbers = s.numbers[:len(s.numbers)-2]
	s.ops = s.ops[:len(s.ops)-1]

	var result float64
	switch op {
	case '+':
		result = a + b
	case '-':
		result = a - b
	case '*':
		result = a * b
	case '/':
		if b == 0 {
			return errors.New("деление на ноль") // Обработка деления на ноль
		}
		result = a / b
	}
	// Добавляем результат операции в стек чисел
	s.numbers = append(s.numbers, result)
	return nil
}

// isNegativeSign проверяет, является ли число отрицательным
func isNegativeSign(expression string, index int) bool {
	return expression[index] == '-' && (index == 0 || isValidMathOperation(rune(expression[index-1])) || expression[index-1] == '(')
}

// parseNegativeNumber парсит строку в отрицательное число
func parseNegativeNumber(expression string, index *int) (float64, error) {
	*index++
	j := *index
	for j < len(expression) && (unicode.IsDigit(rune(expression[j])) || expression[j] == '.') {
		j++
	}
	num, err := strconv.ParseFloat(expression[*index:j], 64)
	if err != nil {
		return 0, err
	}
	num = -num
	*index = j // Обновление индекса
	return num, nil
}

// Нахождение приоритета операции
func priorityOfOperation(operation rune) int {
	if operation == '+' || operation == '-' {
		return 1
	} else if operation == '*' || operation == '/' {
		return 2
	}
	return 0
}
