package handlers

import (
    "context"
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "crudapi/config"
    "crudapi/models"
    "crudapi/services"
    "encoding/json"
    "fmt"
    "io"
)

type AuthHandler struct {
    authService  *services.AuthService
    oauthConfig  *oauth2.Config
}

func NewAuthHandler(authService *services.AuthService, cfg *config.Config) *AuthHandler {
    oauthCfg := &oauth2.Config{
        ClientID:     cfg.GoogleClientID,
        ClientSecret: cfg.GoogleClientSecret,
        RedirectURL:  cfg.GoogleRedirectURL,
        Scopes:       []string{"openid", "email", "profile"},
        Endpoint:     google.Endpoint,
    }
    return &AuthHandler{authService: authService, oauthConfig: oauthCfg}
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req models.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    resp, err := h.authService.Register(c.Request.Context(), &req)
    if err != nil {
        if errors.Is(err, services.ErrEmailTaken) {
            c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
        return
    }

    c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req models.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    resp, err := h.authService.Login(c.Request.Context(), &req)
    if err != nil {
        if errors.Is(err, services.ErrInvalidCredentials) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
            return
        }
        if errors.Is(err, services.ErrGoogleProvider) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "this account uses Google login"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
    // state is a CSRF token — in production, store this in a short-lived server-side session or signed cookie
    state := "random-state-value" // TODO: generate & store per-request
    url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
    c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
    // TODO: validate state matches what you stored
    code := c.Query("code")
    if code == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
        return
    }

    token, err := h.oauthConfig.Exchange(context.Background(), code)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange code"})
        return
    }

    // Fetch user info from Google
    client := h.oauthConfig.Client(context.Background(), token)
    resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user info"})
        return
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read user info"})
        return
    }

    var googleUser struct {
        ID        string `json:"id"`
        Email     string `json:"email"`
        Name      string `json:"name"`
        Picture   string `json:"picture"`
        Verified  bool   `json:"verified_email"`
    }
    if err := json.Unmarshal(body, &googleUser); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user info"})
        return
    }

    user := &models.User{
        Email:         googleUser.Email,
        Name:          googleUser.Name,
        AvatarURL:     googleUser.Picture,
        Provider:      models.ProviderGoogle,
        GoogleID:      googleUser.ID,
        EmailVerified: googleUser.Verified,
    }

    authResp, err := h.authService.HandleGoogleUser(c.Request.Context(), user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "auth failed"})
        return
    }

    // For mobile apps: redirect with tokens in query params (use a custom scheme)
    // For web: redirect to frontend with tokens
    c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf(
        "%s/auth/callback?access_token=%s&refresh_token=%s",
        "yourapp://", authResp.AccessToken, authResp.RefreshToken,
    ))
    // OR just return JSON if this is a pure API:
    c.JSON(http.StatusOK, authResp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
    var req models.RefreshRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    resp, err := h.authService.RefreshTokens(c.Request.Context(), req.RefreshToken)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
        return
    }

    c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
    var req models.RefreshRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    _ = h.authService.Logout(c.Request.Context(), req.RefreshToken)
    c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
    // UserID is set by the auth middleware
    userID := c.GetString("user_id")
    c.JSON(http.StatusOK, gin.H{"user_id": userID})
}