package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var googleOauthConfig *oauth2.Config
var oauthStateString = "random_string"
var store = sessions.NewCookieStore([]byte("your-secret-key"))

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	googleOauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/calendar.readonly"},
		Endpoint:     google.Endpoint,
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get user token from session (ensure it's a *oauth2.Token)
	userTokenInterface := getUserTokenFromSession(r)
	if userTokenInterface == nil {
		http.Error(w, "User token not found", http.StatusUnauthorized)
		return
	}

	// Type assertion to convert to *oauth2.Token
	userToken, ok := userTokenInterface.(*oauth2.Token)
	if !ok {
		http.Error(w, "Failed to assert token", http.StatusInternalServerError)
		return
	}

	// Fetch Google Calendar events
	events, err := getGoogleCalendarEvents(userToken)
	if err != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}

	// Pass events to dashboard.html
	data := struct {
		Events []GoogleEvent
	}{
		Events: events,
	}

	// Render dashboard.html with events
	tmpl := template.Must(template.New("dashboard").ParseFiles("dashboard.html"))
	tmpl.Execute(w, data)
}

func getUserTokenFromSession(r *http.Request) any {
	session, err := store.Get(r, "session-name")
	if err != nil {
		return nil
	}
	token := session.Values["user_token"]
	return token
}

func storeTokenInSession(w http.ResponseWriter, token *oauth2.Token) {
	// Make sure to use the correct session handling code
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	session.Values["token"] = token
	session.Save(w, r) // Pass both ResponseWriter and Request
}

// getGoogleCalendarEvents fetches the Google Calendar events using the user's token
func getGoogleCalendarEvents(token *oauth2.Token) ([]GoogleEvent, error) {
	// Use the token to make a request to Google Calendar API
	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/calendar/v3/calendars/primary/events")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the response into a structure
	var calendarResponse struct {
		Items []GoogleEvent `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&calendarResponse); err != nil {
		return nil, err
	}

	return calendarResponse.Items, nil
}

// GoogleEvent structure to represent calendar event details
type GoogleEvent struct {
	Summary string `json:"summary"`
	Start   struct {
		DateTime string `json:"dateTime"`
	} `json:"start"`
	End struct {
		DateTime string `json:"dateTime"`
	} `json:"end"`
	Description string `json:"description"`
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Validate the state
	if r.FormValue("state") != oauthStateString {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Exchange code for token
	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	// Use token to get user information
	client := googleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	// Parse the user information
	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Display user information
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Welcome %s!</h1>", userInfo["name"])
	fmt.Fprintf(w, "<p>Email: %s</p>", userInfo["email"])

	// Store the token in session for future use
	storeTokenInSession(w, token)
}
