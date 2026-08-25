package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jossecurity/joss/pkg/i18n"
)

func handleBrevoConfig() {
	fmt.Println(i18n.Tr("brevoTitle"))

	envPath := GetEnvFile()
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		fmt.Println(i18n.Tr("brevoNoEnvError"))
		return
	}

	if hasArg("--enable") {
		key := getCLIOption("api-key")
		if key == "" {
			key = getCLIOption("key")
		}
		if key == "" {
			fmt.Println(i18n.Tr("brevoEmptyKeyError"))
			return
		}

		updateEnvFile(envPath, "BREVO_API", key)
		fmt.Println(i18n.Tr("brevoActivatedSuccess"))
		return
	}

	if hasArg("--disable") {
		removeEnvKey(envPath, "BREVO_API")
		fmt.Println(i18n.Tr("brevoDisabledSuccess"))
		return
	}

	fmt.Print("Deseas activar BREVO_API? (y/n): ")
	response := strings.ToLower(readLine())

	if response == "y" || response == "yes" || response == "s" || response == "si" {
		fmt.Print("Introduce tu Brevo API Key: ")
		key := readLine()

		if key == "" {
			fmt.Println(i18n.Tr("brevoEmptyKeyError"))
			return
		}

		updateEnvFile(envPath, "BREVO_API", key)
		fmt.Println(i18n.Tr("brevoActivatedSuccess"))
	} else {
		removeEnvKey(envPath, "BREVO_API")
		fmt.Println(i18n.Tr("brevoDisabledSuccess"))
	}
}
