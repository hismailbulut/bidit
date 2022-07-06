package main

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

// LOGGING

// Override print func
func print(a ...any) {
	fmt.Println(a...)
}

func Prints(args ...any) string {
	message := ""
	for i, a := range args {
		message += fmt.Sprint(a)
		if i != len(args)-1 {
			message += " "
		}
	}
	return message
}

func PrintError(args ...any) error {
	return errors.New(Prints(args...))
}

func Assert(cond bool, msg ...any) {
	if !cond {
		panic(Prints("ASSERTION FAILED:", Prints(msg...)))
	}
}

// CONVERSION

func ConvertStringToFloat(v string) float64 {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return f
}

func ConvertStringToInt(v string) int {
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return int(i)
}

// CHECKING

func FloatValidator(s string) error {
	_, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return PrintError("Invalid float")
	}
	return nil
}

func IntValidator(s string) error {
	_, err := strconv.Atoi(s)
	if err != nil {
		return PrintError("Invalid integer")
	}
	return nil
}

func PasswordValidator(p string) error {
	pr := []rune(p)
	if len(pr) < 8 {
		return PrintError("Password must be at least 8 character length")
	}
	digits := 0
	puncts := 0
	letters := 0
	for _, r := range pr {
		switch {
		case unicode.IsSpace(r):
			return PrintError("Password can not contain space")
		case unicode.IsDigit(r):
			digits++
		case unicode.IsPunct(r):
			puncts++
		case unicode.IsLetter(r):
			letters++
		default:
			return PrintError("Password can only contain letters, digits and punctuation characters")
		}
	}
	if digits == 0 || puncts == 0 || letters == 0 {
		return PrintError("Password must contain at least one letter, one digit and one punctuation character")
	}
	return nil
}
