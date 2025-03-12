package pkg

import "errors"

var NotImplementedError = errors.New("not implemented")

var (
	mismatchedParentheses = errors.New("mismatched parentheses")
	InvalidExpression     = errors.New("invalid expression")
)
