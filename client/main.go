package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"howett.net/plist"
)

const bundleID = "com.nielshojen.macnamer"

func readPlistPref(path string, key string) string {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return ""
	}

	var config map[string]interface{}
	decoder := plist.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&config); err != nil {
		fmt.Println("Error decoding plist:", err)
		return ""
	}

	if val, ok := config[key]; ok {
		return fmt.Sprintf("%v", val)
	}

	// Optionally scan nested dictionaries
	for _, v := range config {
		if nested, ok := v.(map[string]interface{}); ok {
			if val, ok := nested[key]; ok {
				return fmt.Sprintf("%v", val)
			}
		}
	}

	return ""
}

func getPreference(domain, key string) string {
	paths := []string{
		"/Library/Managed Preferences/" + domain + ".plist",
		"/Library/Preferences/" + domain + ".plist",
	}

	if currentUser, err := user.Current(); err == nil {
		userPath := filepath.Join(currentUser.HomeDir, "Library/Preferences", domain+".plist")
		paths = append(paths, userPath)
	}

	for _, path := range paths {
		value := readPlistPref(path, key)
		if value != "" {
			return value
		}
	}

	return ""
}

func runCommand(command string, args ...string) string {
	out, err := exec.Command(command, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getSerialNumber() string {
	out := runCommand("ioreg", "-c", "IOPlatformExpertDevice")
	re := regexp.MustCompile(`"IOPlatformSerialNumber" = "([^"]+)"`)
	match := re.FindStringSubmatch(out)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func getIPAddress() string {
	out := runCommand("/sbin/ifconfig")
	re := regexp.MustCompile(`inet (\d+\.\d+\.\d+\.\d+)`)
	matches := re.FindAllStringSubmatch(out, -1)
	for _, m := range matches {
		if len(m) > 1 && m[1] != "127.0.0.1" {
			return m[1]
		}
	}
	return ""
}

func setName(nameType, value string) error {
	cmd := exec.Command("scutil", "--set", nameType, value)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set %s: %v\nOutput: %s", nameType, err, out)
	}
	return nil
}

func getCurrentName(nameType string) (string, error) {
	out, err := exec.Command("scutil", "--get", nameType).Output()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

func checkName(nameType, newValue string) error {
	currentValue, err := getCurrentName(nameType)
	if err != nil {
		return err
	}

	if currentValue == newValue {
		fmt.Printf("%s is already set to \"%s\", no change needed.\n", nameType, newValue)
		return nil
	}

	fmt.Printf("Updating %s from \"%s\" to \"%s\"\n", nameType, currentValue, newValue)
	return setName(nameType, newValue)
}

func setUnixHostname(newValue string) error {
	out, err := exec.Command("hostname").Output()
	if err != nil {
		return fmt.Errorf("failed to get CLI hostname: %v", err)
	}
	current := strings.TrimSpace(string(out))

	if current == newValue {
		fmt.Println("CLI hostname is already correct:", newValue)
		return nil
	}

	fmt.Printf("Updating CLI hostname from \"%s\" to \"%s\"\n", current, newValue)
	cmd := exec.Command("hostname", newValue)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set CLI hostname: %v\nOutput: %s", err, out)
	}

	return nil
}

func postToServer(serial, ip, key, urlStr string) (map[string]interface{}, error) {
	form := url.Values{}
	form.Set("serial", serial)
	form.Set("ip", ip)
	form.Set("key", key)

	req, err := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "macnamer/2.1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("JSON decode error: %v", err)
	}

	return result, nil
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("Only root can run this program.")
		os.Exit(1)
	}

	serial := getSerialNumber()
	ip := getIPAddress()

	serverURL := getPreference(bundleID, "ServerURL")
	key := getPreference(bundleID, "Key")
	url := serverURL + "/checkin/"

	data, err := postToServer(serial, ip, key, url)

	if err != nil {
		fmt.Println("Error posting to server:", err)
		return
	}

	fmt.Printf("Server responded with: %+v\n", data)

	name := strings.ToLower(data["name"].(string))
	prefix := strings.ToLower(data["prefix"].(string))
	divider := strings.ToLower(data["devider"].(string))
	length := int(data["length"].(float64))
	domain := strings.ToLower(data["domain"].(string))

	// Zero-pad the name string
	name = fmt.Sprintf("%0*s", length, name) // equivalent to Python's name.zfill(length)

	// Build newname just like in Python
	newname := prefix + divider + name

	fmt.Printf("Name should be: %+v\n", newname)
	fmt.Printf("Domain is: %+v\n", domain)

	checkName("LocalHostName", newname)
	checkName("ComputerName", newname)
	checkName("HostName", newname+"."+domain)
	setUnixHostname(newname + "." + domain)
}
