// =============================================================================
// LESSON 9: MICROSERVICES SECURITY
// =============================================================================
//
// In a monolith: one authentication check at the front door.
// In microservices: EVERY service is a potential attack surface.
//                   Traffic flows between 10-50 services.
//                   How do you secure inter-service communication?
//
// SECURITY LAYERS:
//   1. Edge Security        (API Gateway: auth, WAF, rate limiting)
//   2. Transport Security   (mTLS: encrypted service-to-service)
//   3. Authentication       (WHO are you? JWT, OAuth2, API keys)
//   4. Authorization        (WHAT can you do? RBAC, ABAC, policies)
//   5. Data Security        (encryption at rest, field-level encryption)
//   6. Zero Trust           (never trust, always verify)
// =============================================================================

package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// =============================================================================
// CONCEPT 1: JWT (JSON Web Tokens)
// =============================================================================
//
// JWT is the standard for passing identity between services.
//
// STRUCTURE: header.payload.signature (Base64URL encoded)
//
// HEADER:    {"alg":"HS256","typ":"JWT"}
// PAYLOAD:   {"sub":"user-123","name":"Vikram","role":"admin","exp":1700000000}
// SIGNATURE: HMAC-SHA256(base64(header) + "." + base64(payload), secret)
//
// FLOW:
//   1. User logs in → Auth service creates JWT → returns to client
//   2. Client sends JWT in Authorization header
//   3. API Gateway validates JWT → extracts claims → forwards to services
//   4. Services TRUST the claims (because JWT is signed)
//
// IMPORTANT:
//   ✅ JWT is signed (tamper-proof) — NOT encrypted (anyone can read payload)
//   ✅ Short-lived (15 min) + refresh token (7 days) = best practice
//   ❌ JWT cannot be revoked (until expiry) — use a token blacklist for logout

type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type JWTClaims struct {
	Sub   string   `json:"sub"` // subject (user ID)
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
	Iss   string   `json:"iss"` // issuer
	Aud   string   `json:"aud"` // audience
	Exp   int64    `json:"exp"` // expiration (unix timestamp)
	Iat   int64    `json:"iat"` // issued at
	Jti   string   `json:"jti"` // unique token ID (for blacklisting)
}

type JWTService struct {
	secret    []byte
	issuer    string
	blacklist map[string]bool // token IDs that have been revoked
	mu        sync.RWMutex
}

func NewJWTService(secret, issuer string) *JWTService {
	return &JWTService{
		secret:    []byte(secret),
		issuer:    issuer,
		blacklist: make(map[string]bool),
	}
}

func (j *JWTService) CreateToken(userID, name, email string, roles []string, ttl time.Duration) string {
	// Generate unique token ID
	jtiBytes := make([]byte, 16)
	rand.Read(jtiBytes)
	jti := hex.EncodeToString(jtiBytes)

	header := JWTHeader{Alg: "HS256", Typ: "JWT"}
	claims := JWTClaims{
		Sub:   userID,
		Name:  name,
		Email: email,
		Roles: roles,
		Iss:   j.issuer,
		Aud:   "microservices-api",
		Exp:   time.Now().Add(ttl).Unix(),
		Iat:   time.Now().Unix(),
		Jti:   jti,
	}

	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Sign with HMAC-SHA256
	signingInput := headerB64 + "." + claimsB64
	mac := hmac.New(sha256.New, j.secret)
	mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return headerB64 + "." + claimsB64 + "." + signature
}

func (j *JWTService) ValidateToken(token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Verify signature
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, j.secret)
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid claims encoding")
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("invalid claims: %w", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	// Check blacklist
	j.mu.RLock()
	if j.blacklist[claims.Jti] {
		j.mu.RUnlock()
		return nil, fmt.Errorf("token revoked")
	}
	j.mu.RUnlock()

	return &claims, nil
}

func (j *JWTService) RevokeToken(jti string) {
	j.mu.Lock()
	j.blacklist[jti] = true
	j.mu.Unlock()
}

// =============================================================================
// CONCEPT 2: API Key Authentication
// =============================================================================
//
// Simpler than JWT for service-to-service communication.
//
// HOW: Each service has an API key. Passed in header: X-API-Key
//
// RULES:
//   ✅ Store hashed (never store plaintext API keys in DB)
//   ✅ Rotate regularly
//   ✅ Scope to specific services/permissions
//   ✅ Rate limit per key

type APIKeyStore struct {
	mu   sync.RWMutex
	keys map[string]APIKeyInfo // hash(key) → info
}

type APIKeyInfo struct {
	ServiceName string
	Permissions []string
	RateLimit   int
	ExpiresAt   time.Time
}

func NewAPIKeyStore() *APIKeyStore {
	return &APIKeyStore{keys: make(map[string]APIKeyInfo)}
}

func (s *APIKeyStore) RegisterKey(key, service string, perms []string) {
	hash := hashAPIKey(key)
	s.mu.Lock()
	s.keys[hash] = APIKeyInfo{
		ServiceName: service,
		Permissions: perms,
		RateLimit:   1000,
		ExpiresAt:   time.Now().Add(90 * 24 * time.Hour), // 90 days
	}
	s.mu.Unlock()
}

func (s *APIKeyStore) Validate(key string) (*APIKeyInfo, error) {
	hash := hashAPIKey(key)
	s.mu.RLock()
	info, ok := s.keys[hash]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("invalid API key")
	}
	if time.Now().After(info.ExpiresAt) {
		return nil, fmt.Errorf("API key expired")
	}
	return &info, nil
}

func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// =============================================================================
// CONCEPT 3: RBAC (Role-Based Access Control)
// =============================================================================
//
// Users have ROLES. Roles have PERMISSIONS. Check permissions, not roles.
//
// EXAMPLE:
//   admin  → [read, write, delete, manage_users]
//   editor → [read, write]
//   viewer → [read]
//
// RBAC is simple but rigid. For complex rules, use ABAC or OPA.
//
// ABAC (Attribute-Based Access Control):
//   Rules based on attributes: user.department == resource.department
//   More flexible but more complex.
//
// OPA (Open Policy Agent):
//   Externalized policy engine. Policies written in Rego language.
//   Service asks OPA: "Can user X do action Y on resource Z?"
//   OPA evaluates policy and returns allow/deny.
//   BEST for microservices — centralized policy, decoupled from code.

type RBACService struct {
	roles     map[string][]string // role → permissions
	userRoles map[string][]string // user → roles
}

func NewRBACService() *RBACService {
	return &RBACService{
		roles: map[string][]string{
			"admin":  {"read", "write", "delete", "manage_users", "view_analytics"},
			"editor": {"read", "write"},
			"viewer": {"read"},
		},
		userRoles: make(map[string][]string),
	}
}

func (r *RBACService) AssignRole(userID, role string) {
	r.userRoles[userID] = append(r.userRoles[userID], role)
}

func (r *RBACService) HasPermission(userID, permission string) bool {
	for _, role := range r.userRoles[userID] {
		for _, perm := range r.roles[role] {
			if perm == permission {
				return true
			}
		}
	}
	return false
}

func (r *RBACService) GetPermissions(userID string) []string {
	permSet := make(map[string]bool)
	for _, role := range r.userRoles[userID] {
		for _, perm := range r.roles[role] {
			permSet[perm] = true
		}
	}
	perms := make([]string, 0, len(permSet))
	for p := range permSet {
		perms = append(perms, p)
	}
	return perms
}

// =============================================================================
// CONCEPT 4: mTLS (Mutual TLS)
// =============================================================================
//
// Regular TLS: client verifies server's certificate (HTTPS).
// Mutual TLS:  BOTH sides verify each other's certificates.
//
// WHY: In microservices, you need to verify that the CALLING service is
//      who it claims to be. Any service without a valid cert is rejected.
//
// HOW:
//   1. Certificate Authority (CA) issues certs to each service
//   2. Service A connects to Service B
//   3. Service B presents its cert → A validates
//   4. Service A presents its cert → B validates
//   5. Encrypted channel established
//
// MANAGEMENT: Service mesh (Istio, Linkerd) automates mTLS cert rotation.
//
// NOTE: We can't run a real TLS demo in a standalone Go file, but here's
// the conceptual model:

type mTLSConfig struct {
	ServiceName string
	CertFile    string // /etc/certs/service.crt
	KeyFile     string // /etc/certs/service.key
	CAFile      string // /etc/certs/ca.crt (trust store)
}

func ExplainmTLS() {
	fmt.Println("  mTLS Setup (conceptual):")
	fmt.Println("  1. Deploy CA (e.g., Vault, cert-manager)")
	fmt.Println("  2. Each service gets its own TLS cert signed by CA")
	fmt.Println("  3. Services verify peer certs on every connection")
	fmt.Println("  4. Certificates auto-rotated (Istio: every 24h)")
	fmt.Println("  5. No service can communicate without valid cert")
	fmt.Println()
	fmt.Println("  Go code for mTLS server:")
	fmt.Println("    tlsConfig := &tls.Config{")
	fmt.Println("      ClientAuth: tls.RequireAndVerifyClientCert,")
	fmt.Println("      ClientCAs:  caCertPool,")
	fmt.Println("      MinVersion: tls.VersionTLS13,")
	fmt.Println("    }")
}

// =============================================================================
// CONCEPT 5: Zero Trust Architecture
// =============================================================================
//
// TRADITIONAL: Trust internal network. Firewall protects the perimeter.
//              Once inside, everything trusts everything.
//
// ZERO TRUST: "Never trust, always verify."
//   Every request is verified, regardless of source.
//   Inside the network ≠ trusted.
//
// PRINCIPLES:
//   1. Verify explicitly (authenticate + authorize every request)
//   2. Least privilege access (minimal permissions)
//   3. Assume breach (encrypt everything, segment networks)
//
// IMPLEMENTATION:
//   ✅ mTLS between all services
//   ✅ JWT/token validation at every service (not just gateway)
//   ✅ Network policies (Kubernetes NetworkPolicy, Calico)
//   ✅ Service mesh for policy enforcement
//   ✅ Secrets management (Vault, AWS Secrets Manager)
//   ✅ Audit logging everywhere

// =============================================================================
// CONCEPT 6: OAuth2 + OIDC Flows
// =============================================================================
//
// OAuth2 = Authorization framework. Delegates access without sharing passwords.
//
// ROLES:
//   Resource Owner:  User
//   Client:          Your app (web, mobile)
//   Authorization Server: Issues tokens (Auth0, Keycloak, Okta)
//   Resource Server: Your API (microservices)
//
// FLOWS:
//   Authorization Code:    Web apps (most secure, uses redirect)
//   Authorization Code + PKCE: Mobile/SPA (no client secret)
//   Client Credentials:    Service-to-service (no user involved)
//   Device Code:           Smart TV, CLI tools
//
// OIDC (OpenID Connect):
//   Layer on top of OAuth2 that adds identity (ID token with user info).
//   OAuth2 = authorization ("what can you do")
//   OIDC   = authentication ("who are you") + authorization

type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
}

// Client Credentials Flow (service-to-service)
func ExplainClientCredentialsFlow() {
	fmt.Println("  Client Credentials Flow (Service-to-Service):")
	fmt.Println("  1. Service A sends: POST /oauth/token")
	fmt.Println("     Body: grant_type=client_credentials&client_id=X&client_secret=Y")
	fmt.Println("  2. Auth server validates client_id + client_secret")
	fmt.Println("  3. Returns access_token (JWT) with scopes")
	fmt.Println("  4. Service A calls Service B with: Authorization: Bearer <token>")
	fmt.Println("  5. Service B validates token (signature + expiry + scopes)")
}

func main() {
	// =========================================================================
	// DEMO 1: JWT Creation and Validation
	// =========================================================================
	fmt.Println("=== JWT AUTHENTICATION ===")

	jwtSvc := NewJWTService("super-secret-key-change-in-production", "auth-service")

	// Create a token
	token := jwtSvc.CreateToken("user-123", "Vikram", "vikram@dev.com",
		[]string{"admin", "editor"}, 15*time.Minute)

	fmt.Printf("  Token: %s...%s\n", token[:30], token[len(token)-20:])
	fmt.Printf("  Parts: %d\n", len(strings.Split(token, ".")))

	// Validate the token
	claims, err := jwtSvc.ValidateToken(token)
	if err != nil {
		fmt.Printf("  ✗ Validation failed: %v\n", err)
	} else {
		fmt.Printf("  ✓ Valid! User: %s, Roles: %v, Expires: %s\n",
			claims.Sub, claims.Roles,
			time.Unix(claims.Exp, 0).Format("15:04:05"))
	}

	// Tamper with token
	fmt.Println("\n  Tampering test:")
	tamperedToken := token[:len(token)-5] + "XXXXX"
	_, err = jwtSvc.ValidateToken(tamperedToken)
	fmt.Printf("  ✗ Tampered token: %v\n", err)

	// Revoke token
	fmt.Println("\n  Revocation test:")
	jwtSvc.RevokeToken(claims.Jti)
	_, err = jwtSvc.ValidateToken(token)
	fmt.Printf("  ✗ Revoked token: %v\n", err)

	// =========================================================================
	// DEMO 2: API Key Authentication
	// =========================================================================
	fmt.Println("\n=== API KEY AUTHENTICATION ===")

	keyStore := NewAPIKeyStore()
	keyStore.RegisterKey("sk_live_payment_abc123", "payment-service",
		[]string{"charge", "refund"})
	keyStore.RegisterKey("sk_live_order_xyz789", "order-service",
		[]string{"read_orders", "create_orders"})

	// Valid key
	info, err := keyStore.Validate("sk_live_payment_abc123")
	if err != nil {
		fmt.Printf("  ✗ %v\n", err)
	} else {
		fmt.Printf("  ✓ Service: %s, Permissions: %v\n", info.ServiceName, info.Permissions)
	}

	// Invalid key
	_, err = keyStore.Validate("sk_live_fake_key")
	fmt.Printf("  ✗ Fake key: %v\n", err)

	// =========================================================================
	// DEMO 3: RBAC Authorization
	// =========================================================================
	fmt.Println("\n=== RBAC AUTHORIZATION ===")

	rbac := NewRBACService()
	rbac.AssignRole("user-1", "admin")
	rbac.AssignRole("user-2", "viewer")
	rbac.AssignRole("user-3", "editor")

	tests := []struct {
		user       string
		permission string
	}{
		{"user-1", "delete"},
		{"user-1", "manage_users"},
		{"user-2", "read"},
		{"user-2", "write"},
		{"user-3", "write"},
		{"user-3", "delete"},
	}

	for _, t := range tests {
		allowed := rbac.HasPermission(t.user, t.permission)
		symbol := "✗"
		if allowed {
			symbol = "✓"
		}
		fmt.Printf("  %s %s → %s: %v\n", symbol, t.user, t.permission, allowed)
	}

	fmt.Printf("\n  user-1 permissions: %v\n", rbac.GetPermissions("user-1"))
	fmt.Printf("  user-2 permissions: %v\n", rbac.GetPermissions("user-2"))

	// =========================================================================
	// DEMO 4: mTLS Explanation
	// =========================================================================
	fmt.Println("\n=== mTLS (Mutual TLS) ===")
	ExplainmTLS()

	// =========================================================================
	// DEMO 5: OAuth2 Client Credentials
	// =========================================================================
	fmt.Println("\n=== OAUTH2 CLIENT CREDENTIALS ===")
	ExplainClientCredentialsFlow()

	// =========================================================================
	// Summary
	// =========================================================================
	fmt.Println("\n=== SECURITY LAYERS ===")
	fmt.Println("┌──────────────────────────┬────────────────────────────────────────┐")
	fmt.Println("│ Layer                    │ Implementation                         │")
	fmt.Println("├──────────────────────────┼────────────────────────────────────────┤")
	fmt.Println("│ Edge (Gateway)           │ WAF, DDoS protection, rate limiting    │")
	fmt.Println("│ Authentication           │ JWT, OAuth2/OIDC, API keys             │")
	fmt.Println("│ Authorization            │ RBAC, ABAC, OPA policies               │")
	fmt.Println("│ Transport                │ mTLS (service mesh automates this)     │")
	fmt.Println("│ Data at rest             │ AES-256 encryption, KMS                │")
	fmt.Println("│ Secrets                  │ Vault, AWS Secrets Manager             │")
	fmt.Println("│ Network                  │ K8s NetworkPolicy, security groups     │")
	fmt.Println("│ Audit                    │ Structured logging, immutable logs     │")
	fmt.Println("└──────────────────────────┴────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("ZERO TRUST CHECKLIST:")
	fmt.Println("  □ mTLS between all services")
	fmt.Println("  □ Validate auth at EVERY service (not just gateway)")
	fmt.Println("  □ Least privilege: minimal permissions per service")
	fmt.Println("  □ Network segmentation (don't let any pod talk to any pod)")
	fmt.Println("  □ Rotate secrets and certs automatically")
	fmt.Println("  □ Encrypt all data in transit and at rest")
	fmt.Println("  □ Audit log every access decision")
}
