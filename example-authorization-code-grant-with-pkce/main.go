package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// You need to replace these with your actual values
var (
	clientID     = "YOUR_CLIENT_ID"
	clientSecret = "YOUR_CLIENT_SECRET"
	tenantID     = "YOUR_TENANT_ID"
	redirectURL  = "YOUR_REDIRECT_URL"
)

func main() {
	// Create a new OIDC provider
	provider, err := oidc.NewProvider(context.Background(), fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", tenantID))
	if err != nil {
		log.Fatalf("Failed to create OIDC provider: %v", err)
	}

	// Configure the OAuth2 client
	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// Generate a random PKCE code verifier
	codeVerifier, err := pkce.GenerateCodeVerifier()
	if err != nil {
		log.Fatalf("Failed to generate PKCE code verifier: %v", err)
	}

	// Generate the PKCE code challenge
	codeChallenge := pkce.GenerateCodeChallenge(codeVerifier)

	// Generate the authorization URL
	authURL := config.AuthCodeURL("state",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	// Open the authorization URL in the user's default browser
	err = openbrowser.Open(authURL)
	if err != nil {
		log.Fatalf("Failed to open browser: %v", err)
	}

	// Start a local server to listen for the callback
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Get the authorization code from the request
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Authorization code missing", http.StatusBadRequest)
			return
		}

		// Exchange the authorization code for a token
		token, err := config.Exchange(context.Background(), code,
			oauth2.SetAuthURLParam("code_verifier", codeVerifier),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
			return
		}

		// Print the access token
		fmt.Fprintf(w, "Access token: %s\n", token.AccessToken)

		// You can now use the access token to access protected resources
		// ...

		// Stop the server
		os.Exit(0)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
