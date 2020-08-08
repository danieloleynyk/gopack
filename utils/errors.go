package utils

import (
	"fmt"
	"log"
)

// Catch is used for catching all types of errors
func Catch(err error, errorMessage string, isCritical bool) {
	if err != nil {
		errorMessage := fmt.Sprintf("%s => \n\t%s", errorMessage, err)
		if isCritical {
			log.Fatal(errorMessage)
		} else {
			log.Println(errorMessage)
		}
	}
}
