package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
)

// Переменная сборки, может быть задана через ldflags
var buildVersion = "dev"

// Feature flags - now controlled by command line flags
var (
	showIPAddress bool
	showCountry   bool
	showNameForm  bool
)

// CountryResponse represents the response from ipapi.co
type CountryResponse struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

// ClientInfo holds information about the client
type ClientInfo struct {
	IP      string
	Country string
	Name    string
	Surname string
}

func getClientIP(r *http.Request) string {
	// Check for forwarded headers first (for proxy/load balancer scenarios)
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Fall back to remote address
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}

func getCountryFromIP(ip string) string {
    ip = "212.5.142.217"

	if ip == "127.0.0.1" || ip == "::1" || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
		return "Local Network"
	}

	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s", ip))
	if err != nil {
		return "Unknown"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "Unknown"
	}

	var countryResp CountryResponse
	if err := json.Unmarshal(body, &countryResp); err != nil {
		return "Unknown"
	}

	if countryResp.Country == "" {
		return "Unknown"
	}

	return countryResp.Country
}

func handleForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		name := r.FormValue("name")
		surname := r.FormValue("surname")
		
		clientInfo := ClientInfo{
			Name:    name,
			Surname: surname,
		}

		if showIPAddress {
			clientInfo.IP = getClientIP(r)
		}

		if showCountry && clientInfo.IP != "" {
			clientInfo.Country = getCountryFromIP(clientInfo.IP)
		}

		renderResponse(w, clientInfo)
		return
	}

	// Show the form
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <title>Hello World - Enter Your Name</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .form-container { max-width: 400px; margin: 0 auto; }
        input[type="text"] { width: 100%; padding: 10px; margin: 10px 0; border: 1px solid #ddd; border-radius: 4px; }
        button { background: #007cba; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #005a87; }
    </style>
</head>
<body>
    <div class="form-container">
        <h2>Enter Your Name and Surname</h2>
        <form method="POST">
            <label for="name">Name:</label><br>
            <input type="text" id="name" name="name" required><br>
            <label for="surname">Surname:</label><br>
            <input type="text" id="surname" name="surname" required><br><br>
            <button type="submit">Submit</button>
        </form>
    </div>
</body>
</html>`

	t, _ := template.New("form").Parse(tmpl)
	t.Execute(w, nil)
}

func renderResponse(w http.ResponseWriter, info ClientInfo) {
	response := fmt.Sprintf("Build version: %s\n", buildVersion)
	
	if info.Name != "" && info.Surname != "" {
		response += fmt.Sprintf("Hello %s %s!\n", info.Name, info.Surname)
	}
	
	if showIPAddress && info.IP != "" {
		response += fmt.Sprintf("Your IP address: %s\n", info.IP)
	}
	
	if showCountry && info.Country != "" {
		response += fmt.Sprintf("Your country: %s\n", info.Country)
	}
	
	fmt.Fprintf(w, response)
}

func main() {
	// Define command line flags
	flag.BoolVar(&showIPAddress, "ip", false, "Show client IP address")
	flag.BoolVar(&showCountry, "country", false, "Show client country")
	flag.BoolVar(&showNameForm, "form", false, "Show name/surname form")
	
	// Add help flag
	help := flag.Bool("help", false, "Show this help message")
	
	// Parse command line flags
	flag.Parse()
	
	// Show help if requested
	if *help {
		fmt.Println("Go Hello World Server with Feature Flags")
		fmt.Println("Usage:")
		fmt.Println("  ./hello-world [flags]")
		fmt.Println()
		fmt.Println("Available flags:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  ./hello-world -ip                    # Show only IP address")
		fmt.Println("  ./hello-world -ip -country           # Show IP and country")
		fmt.Println("  ./hello-world -form                  # Show only the form")
		fmt.Println("  ./hello-world -ip -country -form     # Show all features")
		fmt.Println("  ./hello-world                        # Show basic message only")
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if showNameForm {
			handleForm(w, r)
		} else {
			// Original behavior with feature flags
			clientInfo := ClientInfo{}
			
			if showIPAddress {
				clientInfo.IP = getClientIP(r)
			}
			
			if showCountry && clientInfo.IP != "" {
				clientInfo.Country = getCountryFromIP(clientInfo.IP)
			}
			
			renderResponse(w, clientInfo)
		}
	})

	port := "8080"
	fmt.Printf("Starting server at http://localhost:%s\n", port)
	fmt.Printf("Feature flags: IP=%v, Country=%v, NameForm=%v\n", showIPAddress, showCountry, showNameForm)
	
	if !showIPAddress && !showCountry && !showNameForm {
		fmt.Println("Note: No features enabled. Use -help to see available options.")
	}
	
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
