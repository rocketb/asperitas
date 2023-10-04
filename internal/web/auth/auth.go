package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/rocketb/asperitas/internal/usecase/user"
	"github.com/rocketb/asperitas/pkg/logger"
	"github.com/rocketb/asperitas/pkg/web"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/open-policy-agent/opa/rego"
	"go.opentelemetry.io/otel/attribute"
)

// ErrForbidden is returned when auth issue is identified.
var ErrForbidden = errors.New("action is not allowed")

type User struct {
	Username string    `json:"username"`
	ID       uuid.UUID `json:"id"`
}

type Claims struct {
	jwt.RegisteredClaims
	User  User        `json:"user"`
	Roles []user.Role `json:"roles"`
}

// KeyLookup declares a method set of behavior for looking up
// private and public keys for JWT use.
type KeyLookup interface {
	PrivateKeyPEM(kid string) (pem string, err error)
	PublicKeyPEM(kid string) (pem string, err error)
}

// Config represents information required to initialize auth.
type Config struct {
	Log       *logger.Logger
	KeyLookup KeyLookup
	ActiveKID string
	DB        *sqlx.DB
}

type Auth interface {
	GenerateToken(ctx context.Context, claims Claims) (string, error)
	Authenticate(ctx context.Context, barerToeken string) (Claims, error)
	Authorize(ctx context.Context, claims Claims, userID uuid.UUID, rule string) error
}

// Auth is used to authenticate clients. It can generate a token for a
// set of user claims and recreate the claims by parsing the token.
type Usecase struct {
	log       *logger.Logger
	keyLookup KeyLookup
	kid       string
	method    jwt.SigningMethod
	parser    *jwt.Parser
	mu        sync.RWMutex
	cache     map[string]string
}

func New(cfg Config) *Usecase {
	return &Usecase{
		kid:       cfg.ActiveKID,
		log:       cfg.Log,
		keyLookup: cfg.KeyLookup,
		method:    jwt.GetSigningMethod("RS256"),
		parser:    jwt.NewParser(jwt.WithValidMethods([]string{"RS256"})),
		cache:     make(map[string]string),
	}
}

// GenerateToken generates a signed JWT token string representing the user Claims.
func (a *Usecase) GenerateToken(ctx context.Context, claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = a.kid

	_, span := web.AddSpan(ctx, "internal.web.auth.GenerateToken")
	defer span.End()

	privateKeyPEM, err := a.keyLookup.PrivateKeyPEM(a.kid)
	if err != nil {
		return "", fmt.Errorf("private key lookup: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", fmt.Errorf("parsing private key: %w", err)
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("singing token: %w", err)
	}

	return str, nil
}

// Authenticate process the token to validate the sender's token is valid.
func (a *Usecase) Authenticate(ctx context.Context, barerToken string) (Claims, error) {
	parts := strings.Split(barerToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return Claims{}, errors.New("expected authorization header format: Bearer <token>")
	}

	ctx, span := web.AddSpan(ctx, "internal.web.auth.Authenticate")
	defer span.End()

	var claims Claims
	token, _, err := a.parser.ParseUnverified(parts[1], &claims)
	if err != nil {
		return Claims{}, fmt.Errorf("parsing token: %w", err)
	}

	// Perform an extra level of authentication verification with OPA

	kidRaw, ok := token.Header["kid"]
	if !ok {
		return Claims{}, fmt.Errorf("kid missing from header: %w", err)
	}

	kid, ok := kidRaw.(string)
	if !ok {
		return Claims{}, fmt.Errorf("kid mailformed: %w", err)
	}

	pem, err := a.publicKeyLookup(kid)
	if err != nil {
		return Claims{}, fmt.Errorf("fetching public key: %w", err)
	}

	input := map[string]any{
		"Key":   pem,
		"Token": parts[1],
	}

	if err := a.opaPolicyEvaluation(ctx, opaAuthentication, RuleAuthenticate, input); err != nil {
		return Claims{}, fmt.Errorf("authentication failed: %w", err)
	}

	return claims, nil
}

// Authorize attempts to authorize the user with the provided roles, if
// none of the input roles are within the user's claims, we return an error
// otherwise the user is authorized.
func (a *Usecase) Authorize(ctx context.Context, claims Claims, userID uuid.UUID, rule string) error {
	input := map[string]any{
		"Roles":   claims.Roles,
		"Subject": claims.Subject,
		"UserID":  userID,
	}

	ctx, span := web.AddSpan(ctx, "internal.web.auth.Authorize")
	defer span.End()

	if err := a.opaPolicyEvaluation(ctx, opaAuthorization, rule, input); err != nil {
		return fmt.Errorf("rego evaluation failed: %w", err)
	}

	return nil
}

// publicKeyLookup performs a lookup for the public PEM for the specific kid.
func (a *Usecase) publicKeyLookup(kid string) (string, error) {
	pem, err := func() (string, error) {
		a.mu.RLock()
		defer a.mu.RUnlock()

		pem, ok := a.cache[kid]
		if !ok {
			return "", errors.New("not found")
		}
		return pem, nil
	}()
	if err == nil {
		return pem, nil
	}

	pem, err = a.keyLookup.PublicKeyPEM(kid)
	if err != nil {
		return "", fmt.Errorf("fetching public key: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.cache[kid] = pem

	return pem, nil
}

// opaPolicyEvaluation asks opa to evaluate the token against the specified token
// policy and public key.
func (a *Usecase) opaPolicyEvaluation(ctx context.Context, opaPolicy string, rule string, input any) error {
	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)

	ctx, span := web.AddSpan(ctx, "internal.web.auth.opaPolicyEvaluation", attribute.String("query", query))
	defer span.End()

	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", opaPolicy),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bingings results[%v] ok[%v]", results, ok)
	}

	return nil
}
