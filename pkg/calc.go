package pkg

import (
	"errors"
	"strings"
)

func GeneratePostfix(expression string) (result []string, isValid bool) {
	if len(expression) == 0 {
		return nil, true
	}
	tokens := tokenize(expression)
	postfix, err := translateToPostfix(tokens)
	if err != nil {
		return nil, false
	}
	return postfix, true
}

func tokenize(expr string) []string {
	var (
		tokens       []string
		currentToken strings.Builder
	)

	for _, char := range expr {
		switch char {
		case ' ':
			continue
		case '+', '-', '*', '/', '(', ')':
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}
			tokens = append(tokens, string(char))
		default:
			currentToken.WriteRune(char)
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}

func translateToPostfix(tokens []string) ([]string, error) {
	var (
		output              []string
		operators           = StackFabric[string]()
		operandCount        int
		operatorCount       int
		firstMustBeOperator bool // после любой ) должен идти только оператор. С помощью этого флага мы будем проверять
		// на наличие этого условия.
	)

	for _, token := range tokens {
		if IsNumber(token) {
			if firstMustBeOperator {
				return nil, InvalidExpression
			}
			output = append(output, token)
			operandCount++
		} else if token == "(" {
			if firstMustBeOperator {
				return nil, InvalidExpression
			}
			operators.Push(token)
		} else if token == ")" {
			for operators.Len() > 0 && operators.GetLast() != "(" {
				output = append(output, operators.Pop())
			}
			if operators.Len() == 0 {
				return nil, mismatchedParentheses
			}
			firstMustBeOperator = true
			operators.Pop()
		} else if IsOperator(token) {
			if firstMustBeOperator {
				firstMustBeOperator = false
			}
			for operators.Len() > 0 && getPriority(operators.GetLast()) >= getPriority(token) {
				output = append(output, operators.Pop())
			}
			operators.Push(token)
			operatorCount++
		} else {
			return nil, errors.New("invalid operator/operand")
		}
	}

	for operators.Len() > 0 {
		if operators.GetLast() == "(" {
			return nil, mismatchedParentheses
		}
		output = append(output, operators.Pop())
	}

	if operatorCount != operandCount-1 {
		return nil, InvalidExpression
	}

	return output, nil
}

func getPriority(op string) int {
	switch op {
	case "+", "-":
		return 1
	case "*", "/":
		return 2
	default:
		return 0
	}
}
