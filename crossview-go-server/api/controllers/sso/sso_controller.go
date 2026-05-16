package sso

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/url"

	"crossview-go-server/lib"
	"crossview-go-server/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type SSOController struct {
	logger     lib.Logger
	env        lib.Env
	ssoService services.SSOServiceInterface
}

func NewSSOController(logger lib.Logger, env lib.Env, ssoService services.SSOServiceInterface) SSOController {
	return SSOController{
		logger:     logger,
		env:        env,
		ssoService: ssoService,
	}
}

func (c *SSOController) GetStatus(ctx *gin.Context) {
	ssoConfig := c.ssoService.GetSSOStatus()
	ctx.JSON(http.StatusOK, gin.H{
		"enabled": ssoConfig.Enabled,
		"oidc": gin.H{
			"enabled": ssoConfig.OIDC.Enabled,
		},
		"saml": gin.H{
			"enabled": ssoConfig.SAML.Enabled,
		},
	})
}

func (c *SSOController) InitiateOIDC(ctx *gin.Context) {
	callbackURL := c.buildCallbackURL(ctx, "/api/auth/oidc/callback")
	c.logger.Infof("OIDC initiation: callbackURL=%s fwd-proto=%s fwd-host=%s host=%s",
		callbackURL,
		ctx.GetHeader("X-Forwarded-Proto"),
		ctx.GetHeader("X-Forwarded-Host"),
		ctx.Request.Host,
	)

	codeVerifier, codeChallenge, err := generatePKCE()
	if err != nil {
		c.logger.Errorf("OIDC PKCE generation failed: %s", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	nonceBytes := make([]byte, 16)
	rand.Read(nonceBytes)
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)

	session := sessions.Default(ctx)
	session.Set("oidcCodeVerifier", codeVerifier)
	session.Set("oidcNonce", nonce)
	if err := session.Save(); err != nil {
		c.logger.Errorf("OIDC session save failed: %s", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "session error"})
		return
	}

	authURL, err := c.ssoService.InitiateOIDC(ctx.Request.Context(), callbackURL, codeChallenge, nonce)
	if err != nil {
		c.logger.Errorf("OIDC initiation failed: %s", err.Error())
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.logger.Infof("OIDC redirect: authURL=%s", authURL)
	ctx.Redirect(http.StatusFound, authURL)
}

func generatePKCE() (verifier, challenge string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return
}

func (c *SSOController) HandleOIDCCallback(ctx *gin.Context) {
	code := ctx.Query("code")
	state := ctx.Query("state")
	errorParam := ctx.Query("error")

	if errorParam != "" {
		c.logger.Warnf("OIDC callback error: %s", errorParam)
		frontendURL := c.env.CORSOrigin
		ctx.Redirect(http.StatusFound, frontendURL+"/login?error=sso_failed")
		return
	}

	if code == "" {
		c.logger.Warn("OIDC callback missing code parameter")
		frontendURL := c.env.CORSOrigin
		ctx.Redirect(http.StatusFound, frontendURL+"/login?error=sso_failed")
		return
	}

	callbackURL := c.buildCallbackURL(ctx, "/api/auth/oidc/callback")

	session := sessions.Default(ctx)
	codeVerifier, _ := session.Get("oidcCodeVerifier").(string)
	session.Delete("oidcCodeVerifier")
	session.Delete("oidcNonce")

	user, err := c.ssoService.HandleOIDCCallback(ctx.Request.Context(), code, state, callbackURL, codeVerifier)
	if err != nil {
		c.logger.Errorf("OIDC callback failed: %s", err.Error())
		frontendURL := c.env.CORSOrigin
		ctx.Redirect(http.StatusFound, frontendURL+"/login?error=sso_failed")
		return
	}

	session.Set("userId", user.ID)
	session.Set("userRole", user.Role)
	if err := session.Save(); err != nil {
		c.logger.Errorf("Failed to save session: %s", err.Error())
		frontendURL := c.env.CORSOrigin
		ctx.Redirect(http.StatusFound, frontendURL+"/login?error=sso_failed")
		return
	}

	c.logger.Infof("OIDC login successful: userId=%d, username=%s", user.ID, user.Username)
	ctx.Redirect(http.StatusFound, c.env.PostLoginURL)
}

func (c *SSOController) InitiateSAML(ctx *gin.Context) {
	// Build callback URL dynamically from request origin
	callbackURL := c.buildCallbackURL(ctx, "/api/auth/saml/callback")

	authURL, err := c.ssoService.InitiateSAML(ctx.Request.Context(), callbackURL)
	if err != nil {
		c.logger.Errorf("SAML initiation failed: %s", err.Error())
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusFound, authURL)
}

func (c *SSOController) HandleSAMLCallback(ctx *gin.Context) {
	samlResponse := ctx.PostForm("SAMLResponse")
	if samlResponse == "" {
		c.logger.Warn("SAML callback missing SAMLResponse")
		frontendURL := c.env.CORSOrigin
		ctx.Redirect(http.StatusFound, frontendURL+"/login?error=sso_failed")
		return
	}

	// Build callback URL dynamically from request origin
	callbackURL := c.buildCallbackURL(ctx, "/api/auth/saml/callback")

	user, err := c.ssoService.HandleSAMLCallback(ctx.Request.Context(), samlResponse, callbackURL)
	if err != nil {
		c.logger.Errorf("SAML callback failed: %s", err.Error())
		frontendURL := c.env.CORSOrigin
		ctx.Redirect(http.StatusFound, frontendURL+"/login?error=sso_failed")
		return
	}

	session := sessions.Default(ctx)
	session.Set("userId", user.ID)
	session.Set("userRole", user.Role)
	if err := session.Save(); err != nil {
		c.logger.Errorf("Failed to save session: %s", err.Error())
		frontendURL := c.env.CORSOrigin
		ctx.Redirect(http.StatusFound, frontendURL+"/login?error=sso_failed")
		return
	}

	c.logger.Infof("SAML login successful: userId=%d, username=%s", user.ID, user.Username)
	ctx.Redirect(http.StatusFound, c.env.PostLoginURL)
}

// buildCallbackURL constructs the callback URL dynamically from the request
// Falls back to config value if request origin cannot be determined
func (c *SSOController) buildCallbackURL(ctx *gin.Context, callbackPath string) string {
	// Try to get the origin from the request
	scheme := "http"
	if ctx.GetHeader("X-Forwarded-Proto") == "https" || ctx.Request.TLS != nil {
		scheme = "https"
	}

	host := ctx.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = ctx.Request.Host
	}

	// If we have a valid host, build the callback URL dynamically
	if host != "" {
		callbackURL := url.URL{
			Scheme: scheme,
			Host:   host,
			Path:   callbackPath,
		}
		return callbackURL.String()
	}

	// Fallback to config value
	// Extract callback URL from config by parsing the CORS origin
	if c.env.CORSOrigin != "" {
		originURL, err := url.Parse(c.env.CORSOrigin)
		if err == nil {
			callbackURL := url.URL{
				Scheme: originURL.Scheme,
				Host:   originURL.Host,
				Path:   callbackPath,
			}
			return callbackURL.String()
		}
	}

	// Last resort: use default from config
	return "http://localhost:3001" + callbackPath
}
