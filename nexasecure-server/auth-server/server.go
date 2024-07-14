package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var db *sql.DB

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type VerifyAuthRequest struct {
	AuthToken   string `json:"authToken"`
	UserAddress string `json:"userAddress"`
	Signature   string `json:"signature"`
}

func main() {
	var dbPort string = "1234"

	fmt.Println("PORT is set to", dbPort)
	connStr := "postgres://nexaUser:nexaPass@localhost:5432/nexasecure-db?sslmode=disable"

	var errDB error
	db, errDB = sql.Open("postgres", connStr)
	if errDB != nil {
		log.Fatal("Database connection error:", errDB)
		return
	}

	dbConnErr := db.Ping()
	if dbConnErr != nil {
		log.Fatal("Database connection failed \n", dbConnErr)
		return
	}

	fmt.Println("Router is initiating to", dbPort)
	MainRouter := chi.NewRouter()
	MainRouter.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	fmt.Println("Initiating the Auth Router at ", dbPort)
	AuthRoutes := chi.NewRouter()
	AuthRoutes.Get("/healthz", HandlerReadiness)
	AuthRoutes.Get("/error", HandlerError)
	AuthRoutes.Post("/login", HandlerLogin)
	AuthRoutes.Post("/logout", HandlerLogout)
	MainRouter.Mount("/auth", AuthRoutes)

	Server := &http.Server{
		Addr:    ":" + dbPort,
		Handler: MainRouter,
	}

	fmt.Println("Server is running on port", dbPort)
	serverErr := Server.ListenAndServe()
	if serverErr != nil {
		log.Fatal(serverErr)
	}
}

func responseWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func responseWithError(w http.ResponseWriter, code int, message string) {
	if code > 499 {
		log.Fatal(message)
	}

	type errResponse struct {
		Error string `json:"error"`
	}

	responseWithJSON(w, code, errResponse{Error: message})
}

func HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func HandlerError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	errorMessage := map[string]string{"error": "Something went wrong"}
	json.NewEncoder(w).Encode(errorMessage)
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type NFTauth struct {
	UserAddress string `json:"userAddress"`
	AuthToken   string `json:"authToken"`
	Signature   string `json:"signature"`
}

type User struct {
	uniqueid uuid.UUID
	Username string
	Password string
	// Add more fields as needed
}

func HandlerLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("HandlerLogin called")
	//check request method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//checking weather the hardware wallet is online or not
	checkHardwarewallet, walleterr := checkServer("http://127.0.0.1:3030")
	if walleterr != nil {
		http.Error(w, walleterr.Error(), http.StatusUnauthorized)
		return
	}
	if checkHardwarewallet {
		http.Error(w, "Connect the hardware wallet", http.StatusUnauthorized)
		return
	}

	checkineme, memerr := checkServer("http://127.0.0.1:3031")
	if memerr != nil {
		http.Error(w, memerr.Error(), http.StatusUnauthorized)
		return

	}

	if checkineme {
		http.Error(w, "Wallet is not working properly", http.StatusUnauthorized)
		return

	}

	//check for the login credintial
	var credentials Credentials
	crentialRequestErr := json.NewDecoder(r.Body).Decode(&credentials)
	if crentialRequestErr != nil {
		http.Error(w, crentialRequestErr.Error(), http.StatusBadRequest)
		return
	}

	//check the database for the user's detail
	user, err := getUserByUsername(credentials.Username)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println("Error querying database for user:", err)
		return
	}

	if user.Password != credentials.Password {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	//2FA - generate the OTP
	err = sendOTP(credentials.Username)
	if err != nil {
		http.Error(w, "OTP generation failed", http.StatusInternalServerError)
		return
	}

	//	waiting for the user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter OTP: ")
	otp, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading OTP: %v\n", err)
		return
	}
	otp = strings.TrimSpace(otp)

	//validate the OTP
	isvalidOTP, err := validateOTP(credentials.Username, otp)
	if err != nil {
		fmt.Printf("Error validating OTP: %v\n", err) // Ensure you have an http.ResponseWriter 'w'
		return
	}

	if !isvalidOTP {
		http.Error(w, "Invalid OTP", http.StatusUnauthorized)
		return
	}
	//get the account details from the wallet
	authAccountURL := "http://localhost:3030/get_accounts"
	authAccountResponse, authAccErr := fetchData(authAccountURL)
	if authAccErr != nil {
		http.Error(w, "Error while fetching the account from the wallet", http.StatusInternalServerError)
		return
	}

	if authAccountResponse == nil {
		http.Error(w, "No account connected with the wallet", http.StatusInternalServerError)
		return
	}

	var nftAuthCre VerifyAuthRequest
	if !verifyAuth(nftAuthCre.AuthToken, nftAuthCre.UserAddress, nftAuthCre.Signature) {
		http.Error(w, "Authorization failed", http.StatusUnauthorized)
		return
	}

	authNFTURL := "http://localhost:3030/get_token"
	authTokenRes, authTokenErr := fetchData(authNFTURL)
	if authTokenErr != nil {
		http.Error(w, "Error while fetching the token from the wallet", http.StatusInternalServerError)
		return
	}

	if authTokenRes == nil {
		http.Error(w, "No token found: User not authorized", http.StatusUnauthorized)
		return
	}

	verifyAuthTokenURL := "http://localhost:3020/login"
	verifyAuthTokenRes, verificationErr := fetchData(verifyAuthTokenURL)
	if verificationErr != nil {
		http.Error(w, "Error during authentication", http.StatusBadRequest)
		return
	}

	var verifyAuthToken bool
	err = json.Unmarshal(verifyAuthTokenRes, &verifyAuthToken)
	if err != nil {
		http.Error(w, "Error parsing authentication response", http.StatusInternalServerError)
		return
	}

	if !verifyAuthToken {
		http.Error(w, "Authentication Error: Unauthorized access !!", http.StatusUnauthorized)
		return
	}

	requestTokenResponse, err := requestToken(nftAuthCre.UserAddress)
	if err != nil {
		http.Error(w, "Error requesting token", http.StatusInternalServerError)
		return
	}

	addTokenURL := "http://127.0.0.1:3031/set-token"
	req, err := http.NewRequest("POST", addTokenURL, bytes.NewBuffer(requestTokenResponse))
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error sending request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("Error from server: %s", string(bodyBytes)), resp.StatusCode)
		return
	}

	err = LogEvent("login", map[string]interface{}{"username": credentials.Username})
	if err != nil {
		responseWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to log login event"})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login success"))
}

func extractTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
			return tokenParts[1]
		}
	}

	cookie, err := r.Cookie("auth_token")
	if err == nil {
		return cookie.Value
	}

	return "Token not found"
}

func LogEvent(eventType string, details interface{}) error {
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer logFile.Close()

	logEntry := fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), eventType)
	if details != nil {
		detailsJSON, err := json.Marshal(details)
		if err != nil {
			return err
		}
		logEntry += string(detailsJSON) + "\n"
	}

	if _, err := logFile.WriteString(logEntry); err != nil {
		return err
	}

	return nil
}

func getUserIdFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["user_id"].(string)
		return userID, nil
	}

	return "", errors.New("Invalid token or claims")
}

var (
	blacklistMutex sync.Mutex
	blacklist      = make(map[string]struct{})
)

func InvalidateToken(token string) error {
	if IsTokenInvalid(token) {
		return errors.New("Token already invalidated")
	}
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()
	blacklist[token] = struct{}{}

	return nil
}

func IsTokenInvalid(token string) bool {
	blacklistMutex.Lock()
	defer blacklistMutex.Unlock()
	_, invalid := blacklist[token]
	return invalid
}

func HandlerLogout(w http.ResponseWriter, r *http.Request) {
	// Get account from the wallet
	accountAddress, err := fetchData("http://localhost:3030/get_accounts")
	if err != nil {
		http.Error(w, "Error while fetching the account from the wallet", http.StatusInternalServerError)
		return
	}
	if accountAddress == nil {
		http.Error(w, "No account connected with the wallet", http.StatusInternalServerError)
		return
	}

	// Get and remove the request token
	requestTokenURL := "http://localhost:3031/get-token"
	resp, err := http.Get(requestTokenURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to retrieve request token", resp.StatusCode)
		return
	}

	requestToken, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading token response body", http.StatusInternalServerError)
		return
	}

	if len(requestToken) == 0 {
		http.Error(w, "No token found - server error", http.StatusUnauthorized)
		return
	}

	// Convert accountAddress to string before using it in the map literal
	payload := map[string]string{
		"accountaddress": string(accountAddress),
		"requesttoken":   string(requestToken),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Error creating JSON payload", http.StatusInternalServerError)
		return
	}

	// Send API call to mintnewAuthURL with JSON data
	mintnewAuthURL := "http://localhost:3020/logout"
	req, err := http.NewRequest("POST", mintnewAuthURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		http.Error(w, "Error creating request to mintnewAuthURL", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		http.Error(w, "Error sending request to mintnewAuthURL", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to logout", resp.StatusCode)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logout successful"))

}

func getUserByUsername(usern string) (User, error) {
	var user User
	err := db.QueryRow("SELECT username, password FROM users WHERE username = $1", usern).Scan(&user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("no user found with username: %s", usern)
		}
		return User{}, err
	}
	return user, nil
}

func checkServer(url string) (bool, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false, err // Return error if there was an issue making the request
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300, nil
}

type Payload struct {
	Username string `json:"username"`
}

func sendOTP(username string) error {
	url := "http://localhost:5000/gen-otp"

	payload := Payload{
		Username: username,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		fmt.Printf("Response: %s\n", body)
		return errors.New("OTP generation failed")
	}

	fmt.Printf("Response: %s\n", body)
	return nil
}

type OTPValidateRequest struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
}

func validateOTP(username, otp string) (bool, error) {
	url := "http://localhost:5000/verify-otp"
	payload := OTPValidateRequest{
		Username: username,
		OTP:      otp,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("error marshaling JSON: %v", err)
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("error creating HTTP request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return false, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("error: received non-200 response code: %d", resp.StatusCode)
	}

	fmt.Println("OTP validated successfully")
	return true, nil
}

func verifyAuth(authToken, userAddress, signature string) bool {
	verifyAuthURL := "http://localhost:3020/login"

	verifyAuthRequest := map[string]string{
		"authToken":   authToken,
		"userAddress": userAddress,
		"signature":   signature,
	}

	jsonPayload, err := json.Marshal(verifyAuthRequest)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return false
	}

	req, err := http.NewRequest("POST", verifyAuthURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Verification failed:", resp.Status)
		return false
	}

	return true
}

func requestToken(userAddress string) ([]byte, error) {
	requestTokenURL := "http://localhost:3020/request"

	requestTokenPayload := map[string]string{
		"userAddress": userAddress,
	}

	jsonRequestPayload, err := json.Marshal(requestTokenPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", requestTokenURL, bytes.NewBuffer(jsonRequestPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("error from server: %s", string(bodyBytes))
	}

	return ioutil.ReadAll(resp.Body)
}

func fetchData(apiEndpoint string) ([]byte, error) {
	resp, err := http.Get(apiEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received non-200 response code: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}
