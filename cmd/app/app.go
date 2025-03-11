package app

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
)

// HTML template for the main page
const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Service Status</title>
  <script src="https://unpkg.com/htmx.org@1.9.6"></script>
  <style>
    body {
      background-color: black;
      color: white;
    }
    .env-section {
      margin-bottom: 15px;
    }
    .env-header {
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .version-section {
      margin-bottom: 5px;
    }
    .service-buttons {
      display: flex;
      flex-wrap: wrap;
      gap: 5px;
    }
    a.status-up, a.status-down {
      text-decoration: none;
      padding: 5px 10px;
      margin: 2px;
      color: white;
      display: inline-block;
    }
    .status-up {
      background-color: green;
    }
    .status-down {
      background-color: red;
    }
	  footer {
      position: fixed;
      bottom: 0;
      left: 0;
      width: 100%;
      text-align: center;
      padding: 2px 0;
    }
  </style>
</head>
<body>
  <h1>Service Status Dashboard</h1>
  {{range .Environments}}
  <div class="env-section">
    <div class="env-header">
      <h2>Environment: {{.}}</h2>
      <button hx-get="/status/{{.}}" hx-target="#status-{{.}}">Refresh</button>
    </div>
    <div id="status-{{.}}" hx-get="/status/{{.}}" hx-trigger="load">
      <!-- Service status buttons will be loaded here -->
    </div>
  </div>
  {{end}}
  <footer>
    <p>Developed while Working from Home</p>
  </footer>
</body>
</html>`

// ServiceStatus holds the name and status of a service
type ServiceStatus struct {
	Name   string
	Status string // "up" or "down"
}

func constructURL(pattern, env, service, version string) string {
	// fmt.Println("pattern", pattern, env, service, version)
	// Step 1: Replace known placeholders with argument values
	url := strings.ReplaceAll(pattern, "{service}", service)
	url = strings.ReplaceAll(url, "{version}", version)
	url = strings.ReplaceAll(url, "{env}", env)

	// Step 2: Replace other placeholders with environment variable values
	for {
		start := strings.Index(url, "{")
		if start == -1 {
			break // No more placeholders
		}
		end := strings.Index(url[start:], "}")
		if end == -1 {
			break // Malformed pattern, stop processing
		}
		end += start
		placeholder := url[start+1 : end]
		// Skip if it's a placeholder we already handled
		if placeholder != "service" && placeholder != "version" && placeholder != "env" {
			envValue := os.Getenv(placeholder)
			if envValue == "" {
				// Optionally handle missing env vars (e.g., leave as-is or error)
				// fmt.Printf("Warning: Environment variable %s not set\n", placeholder)
			}
			url = url[:start] + envValue + url[end+1:]
		}
	}
	return url
}

// statusHandler fetches and returns service statuses for an environment
func statusHandler(w http.ResponseWriter, r *http.Request, env string) {
	services := strings.Split(os.Getenv("SERVICE_NAMES"), ",")
	versions := strings.Split(os.Getenv("VERSIONS"), ",")
	// baseURL := os.Getenv("BASE_URL")

	// Map to group statuses by version
	statusMap := make(map[string][]ServiceStatus)
	var wg sync.WaitGroup
	var mu sync.Mutex // For safe map access

	// Fetch statuses concurrently
	for _, version := range versions {
		for _, service := range services {
			wg.Add(1)
			go func(s, v string) {
				defer wg.Done()
				// url := fmt.Sprintf("http://%s%s/%s%s/health", "", baseURL, s, v)
				url := constructURL(os.Getenv("HEALTH_URL_PATTERN"), env, s, v)
				// url := fmt.Sprintf("https://%s%s/%s%s/health", env+".", baseURL, s, v)
				resp, err := http.Get(url)
				if err != nil {
					fmt.Println(url, "errored")
					statusMap[v] = append(statusMap[v], ServiceStatus{Name: s, Status: "down"})
					return
				}
				defer resp.Body.Close()
				// fmt.Println(env, url, resp.StatusCode)
				status := "down"
				if resp.StatusCode == http.StatusOK {
					status = "up"
				}
				mu.Lock()
				statusMap[v] = append(statusMap[v], ServiceStatus{Name: s, Status: status})
				mu.Unlock()
			}(service, version)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Sort services within each version by name
	for _, statuses := range statusMap {
		sort.Slice(statuses, func(i, j int) bool {
			return statuses[i].Name < statuses[j].Name
		})
	}

	// Generate HTML with each version on a separate line
	// logsBaseURL := strings.TrimSuffix(os.Getenv("LOGS_BASE_URL"), "/")
	// clusterName := os.Getenv("CLUSTER_NAME")
	// projectName := os.Getenv("PROJECT_NAME")
	var html strings.Builder
	for no, version := range versions {
		// html.WriteString(fmt.Sprintf(`<div class="version-section"><h3>Version: %s</h3><div style="display: flex; flex-wrap: wrap; gap: 10px;">`, version))
		html.WriteString(`<div>`)
		for _, status := range statusMap[version] {
			class := "status-down"
			if status.Status == "up" {
				class = "status-up"
			}
			// logsURL := fmt.Sprintf("%s/%s/%s/%s%s/logs?project=%s", logsBaseURL, clusterName, env, status.Name, version, projectName)
			logsURL := constructURL(os.Getenv("LOGS_URL_PATTERN"), env, status.Name, version)
			html.WriteString(fmt.Sprintf(`<a href="%s" class="%s">%s%s</a>`, logsURL, class, status.Name, version))
		}
		html.WriteString(`</div>`)
		if no < len(versions)-1 {
			html.WriteString(`<hr class="dashed">`)
		}

	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html.String()))
}

// mainHandler serves the initial dashboard page
func mainHandler(w http.ResponseWriter, r *http.Request) {
	environments := strings.Split(os.Getenv("ENVIRONMENTS"), ",")
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	data := struct {
		Environments []string
	}{environments}
	tmpl.Execute(w, data)
}

func Serve() {
	// check if required environment variables are set
	requiredEnvVars := []string{"ENVIRONMENTS", "SERVICE_NAMES", "VERSIONS", "BASE_URL", "HEALTH_URL_PATTERN", "LOGS_URL_PATTERN"}
	unsetList := []string{}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			unsetList = append(unsetList, envVar)
		}
	}
	if len(unsetList) > 0 {
		fmt.Printf("The following environment variables are not set: %s\n", strings.Join(unsetList, ", "))
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", mainHandler)
	mux.HandleFunc("/status/", func(w http.ResponseWriter, r *http.Request) {
		env := strings.TrimPrefix(r.URL.Path, "/status/")
		if env == "" {
			http.Error(w, "Environment not specified", http.StatusBadRequest)
			return
		}
		statusHandler(w, r, env)
	})
	port := os.Getenv("REST_PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, mux)
}
