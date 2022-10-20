package tools

import (
	"fmt"
	"github.com/fatih/color"
)

func Yellow(str string, a ...interface{}) {
	if a == nil {
		fmt.Println(color.New(color.FgYellow).Sprint(str))
	} else {
		fmt.Println(color.New(color.FgYellow).Sprintf(str, a...))
	}
}

func Green(str string, a ...interface{}) {
	if a == nil {
		fmt.Println(color.GreenString(str))
	} else {
		fmt.Println(color.GreenString(str, a...))
	}
}

func Red(str string, a ...interface{}) {
	if a == nil {
		fmt.Println(color.RedString(str))
	} else {
		fmt.Println(color.RedString(str, a...))
	}
}

func White(str string, a ...interface{}) {
	if a == nil {
		fmt.Println(color.WhiteString(str))
	} else {
		fmt.Println(color.WhiteString(str, a...))
	}
}

func Scan(msg string) bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		Red(err.Error())
	}
	if response == "y" || response == "Y" {
		return true
	} else {
		if msg != "" {
			White(msg)
		}
	}

	return false
}
