package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// findForms returns all <form> nodes from the given URL
func findForms(pageURL string) ([]*html.Node, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(pageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var forms []*html.Node
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "form" {
			forms = append(forms, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return forms, nil
}

// extractFormData parses a form node and returns its action, method, and all input fields
func extractFormData(form *html.Node) (action string, method string, fields map[string]string) {
	method = "GET"
	fields = make(map[string]string)

	for _, attr := range form.Attr {
		if attr.Key == "action" {
			action = attr.Val
		}
		if attr.Key == "method" {
			method = strings.ToUpper(attr.Val)
		}
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			var name, value string
			for _, attr := range n.Attr {
				if attr.Key == "name" {
					name = attr.Val
				}
				if attr.Key == "value" {
					value = attr.Val
				}
			}
			if name != "" {
				fields[name] = value
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(form)

	return
}

// submitForm injects the payload into all form fields and checks if the payload is reflected in the response
func submitForm(form *html.Node, baseURL string, payload string) (bool, error) {
	action, method, inputs := extractFormData(form)

	for key := range inputs {
		inputs[key] = payload
	}

	target := baseURL
	if action != "" {
		if strings.HasPrefix(action, "http") {
			target = action
		} else {
			base, err := url.Parse(baseURL)
			if err != nil {
				return false, err
			}
			target = base.ResolveReference(&url.URL{Path: action}).String()
		}
	}

	data := url.Values{}
	for k, v := range inputs {
		data.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	var resp *http.Response
	var err error

	if method == "POST" {
		resp, err = client.PostForm(target, data)
	} else {
		resp, err = client.Get(target + "?" + data.Encode())
	}
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if strings.Contains(string(bodyBytes), payload) {
		return true, nil
	}

	return false, nil
}

// loadPayloads reads the payload file line by line into a string slice
func loadPayloads(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var payloads []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			payloads = append(payloads, line)
		}
	}
	return payloads, scanner.Err()
}

func main() {
	urlFlag := flag.String("u", "", "Target URL")
	payloadFlag := flag.String("p", "", "Path to payload file")
	flag.Parse()

	if *urlFlag == "" {
		fmt.Println("Missing -u (URL) flag")
		flag.Usage()
		return
	}

	payloadFile := *payloadFlag
	if payloadFile == "" {
		payloadFile = "payload.txt"
		fmt.Println("No payload file specified, using default: payload.txt")
	}

	fmt.Println("Target:", *urlFlag)
	fmt.Println("Payload file:", payloadFile)

	payloads, err := loadPayloads(payloadFile)
	if err != nil {
		fmt.Println("Failed to read payloads:", err)
		return
	}

	forms, err := findForms(*urlFlag)
	if err != nil {
		fmt.Println("Error parsing forms:", err)
		return
	}

	if len(forms) == 0 {
		fmt.Println("No forms found")
		return
	}

	fmt.Printf("Found %d form(s)\n", len(forms))

	for i, form := range forms {
		fmt.Printf("Testing form #%d\n", i+1)
		for _, payload := range payloads {
			vulnerable, err := submitForm(form, *urlFlag, payload)
			if err != nil {
				fmt.Println("Submission error:", err)
				continue
			}
			if vulnerable {
				fmt.Printf("Possible XSS with payload: %s\n", payload)
			}
		}
	}
}
