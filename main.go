package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/opnsense"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	apiUrl, err := getEnvOrError("OPNSENSE_API_URL")
	if err != nil {
		return err
	}
	apiKey, err := getEnvOrError("OPNSENSE_API_KEY")
	if err != nil {
		return err
	}
	apiSecret, err := getEnvOrError("OPNSENSE_API_SECRET")
	if err != nil {
		return err
	}
	allowInsecure, err := getEnvOrErrorBool("OPNSENSE_ALLOW_INSECURE")
	if err != nil {
		return err
	}
	loopInterval, err := getEnvOrErrorInt("LOOP_INTERVAL")
	if err != nil {
		return err
	}
	offlineTimeUntilReboot, err := getEnvOrErrorInt("OFFLINE_TIME_UNTIL_REBOOT")
	if err != nil {
		return err
	}

	// Create the api client
	apiClient := api.NewClient(api.Options{
		Uri:           apiUrl,
		APIKey:        apiKey,
		APISecret:     apiSecret,
		AllowInsecure: allowInsecure,
	})

	// Create the service client
	serviceClient := opnsense.NewClient(apiClient)

	// Main loop
	for {
		// Check if we are connected to the internet
		if !isConnectedToInternet() {
			fmt.Println("Offline, wait for reboot timer")
			time.Sleep(time.Duration(offlineTimeUntilReboot) * time.Second)
			// Check again
			if !isConnectedToInternet() {
				fmt.Println("Still offline, rebooting")
				// Reboot the system
				_, err := serviceClient.Core().SystemReboot(context.Background())
				if err != nil {
					fmt.Println("failed to reboot system:", err)
				} else {
					// Wait for the system to reboot
					fmt.Println("Rebooted, wait until system is started")
					time.Sleep(5 * time.Minute)
				}
			} else {
				fmt.Println("System got back online on its own")
			}
		} else {
			fmt.Println("System is online")
		}
		// Wait for the next loop
		time.Sleep(time.Duration(loopInterval) * time.Second)
	}
}

func getEnvOrError(key string) (string, error) {
	// Try with _FILE
	value, ok := os.LookupEnv(key + "_FILE")
	if ok {
		data, err := os.ReadFile(value)
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", value, err)
		}
		return string(data), nil
	}

	// Try the normal env var
	value, ok = os.LookupEnv(key)
	if !ok {
		// Still not found, exit with error
		return "", fmt.Errorf("%s(_FILE) environment variable not set", key)
	}
	return value, nil
}

func getEnvOrErrorBool(key string) (bool, error) {
	valueStr, err := getEnvOrError(key)
	if err != nil {
		return false, err
	}
	valueBool := strings.ToLower(valueStr) == "true" || valueStr == "1"
	return valueBool, nil
}

func getEnvOrErrorInt(key string) (int, error) {
	valueStr, err := getEnvOrError(key)
	if err != nil {
		return 0, err
	}
	valueInt, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s: %w", key, err)
	}
	return valueInt, nil
}

func isConnectedToInternet() bool {
	urlList := []string{
		"http://www.msftconnecttest.com/connecttest.txt",
		"http://connectivitycheck.gstatic.com/generate_204",
		"http://clients3.google.com/generate_204",
		"http://www.apple.com/library/test/success.html",
	}

	// Check each URL until one works
	for _, url := range urlList {
		_, err := http.Get(url)
		if err == nil {
			return true
		}
		time.Sleep(10 * time.Second)
	}
	// No url worked
	return false
}
