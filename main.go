package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Struct for handling API response
type Result struct {
	Results []struct {
		Name    string  `json:"name"`
		Address string  `json:"formatted_address"`
		Rating  float64 `json:"rating"`
	} `json:"results"`
}

func main() {
	// Define routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/search", searchHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/contact", contactHandler)


	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// Home page handler
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

// Search handler
func searchHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := "AIzaSyCxuNdUNd5Fgvz9oyafewxA1Gl2ia0r5tc" // Replace with your actual API key
	location := r.FormValue("location")
	ratingFilter := r.FormValue("rating")

	if location == "" {
		http.Error(w, "Location cannot be empty", http.StatusBadRequest)
		return
	}

	// Call Google Places API
	radius := 5000 // Search within 5km radius
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/place/textsearch/json?query=vegan+restaurants+in+%s&radius=%d&key=%s", location, radius, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Error fetching data", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the API response
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body)) // Debug: Print the raw API response in the terminal

	var result Result
	fmt.Println("Raw API Response:", string(body))
	err = json.Unmarshal(body, &result)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusInternalServerError)
		return
	}

	// Filter by rating if a filter is provided
	filteredResults := []struct {
		Name    string
		Address string
		Rating  float64
	}{}
	for _, r := range result.Results {
		if ratingFilter == "" || r.Rating >= parseRating(ratingFilter) {
			filteredResults = append(filteredResults, struct {
				Name    string
				Address string
				Rating  float64
			}{
				Name:    r.Name,
				Address: r.Address,
				Rating:  r.Rating,
			})
		}
	}

	// Pass filtered results to template
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, filteredResults)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        tmpl := template.Must(template.ParseFiles("templates/contact.html"))
        tmpl.Execute(w, nil)
    } else if r.Method == http.MethodPost {
        r.ParseForm()
        name := r.FormValue("name")
        email := r.FormValue("email")
        message := r.FormValue("message")

        fmt.Printf("New message from %s (%s): %s\n", name, email, message)

        http.Redirect(w, r, "/", http.StatusSeeOther) // Redirect to home after submission
    }
}


// Helper function to parse rating input
func parseRating(rating string) float64 {
	value, err := strconv.ParseFloat(rating, 64)
	if err != nil {
		return 0
	}
	return value
}
