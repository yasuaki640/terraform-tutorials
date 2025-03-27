package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/sessions"
	"github.com/yasuaki640/terraform-tutorials/aws-cognito-sample/sample-application-client/config"
	"golang.org/x/oauth2"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func HandleHome(w http.ResponseWriter, r *http.Request) {
	html := `
        <html>
        <body>
            <h1>Welcome to Cognito OIDC Go App</h1>
            <a href="/login">Login with Cognito</a>
        </body>
        </html>`
	fmt.Fprint(w, html)
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func HandleLogin(writer http.ResponseWriter, request *http.Request) {
	// Store the state in the session
	state := generateState()
	session, _ := store.Get(request, "auth-session")
	session.Values["state"] = state
	session.Save(request, writer)

	url := config.Oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(writer, request, url, http.StatusFound)
}

func HandleCallback(writer http.ResponseWriter, request *http.Request) {
	// Check if the state matches
	session, _ := store.Get(request, "auth-session")
	stored := session.Values["state"]
	received := request.URL.Query().Get("state")
	if stored != received {
		http.Error(writer, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	code := request.URL.Query().Get("code")

	// Exchange the authorization code for a token
	rawToken, err := config.Oauth2Config.Exchange(ctx, code)
	if err != nil {
		http.Error(writer, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tokenString := rawToken.AccessToken

	// Parse the token (do signature verification for your use case in production)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		fmt.Printf("Error parsing token: %v\n", err)
		return
	}

	// Check if the token is valid and extract claims
	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		http.Error(writer, "Invalid claims", http.StatusBadRequest)
		return
	}

	type ClaimsPage struct {
		AccessToken string
		Claims      jwt.MapClaims
	}

	// Prepare data for rendering the template
	pageData := ClaimsPage{
		AccessToken: tokenString,
		Claims:      claims,
	}

	// Define the HTML template
	tmpl := `
    <html>
        <body>
            <h1>User Information</h1>
            <h1>JWT Claims</h1>
            <p><strong>Access Token:</strong> {{.AccessToken}}</p>
            <ul>
                {{range $key, $value := .Claims}}
                    <li><strong>{{$key}}:</strong> {{$value}}</li>
                {{end}}
            </ul>
            <a href="/logout">Logout</a>
        </body>
    </html>`

	// Parse and execute the template
	t := template.Must(template.New("claims").Parse(tmpl))
	t.Execute(writer, pageData)
}

func HandleLogout(writer http.ResponseWriter, request *http.Request) {
	// Here, you would clear the session or cookie if stored.
	http.Redirect(writer, request, "/", http.StatusFound)
}
