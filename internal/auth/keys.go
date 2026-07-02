package auth
import ( "crypto/ecdsa"; "fmt"; "strings"; "github.com/golang-jwt/jwt/v5" )
// normalizePEM restores real newlines when the key arrives with literal "\n" escapes, which is
// how docker-compose env_file / .env values are delivered. Genuine PEM strings are unaffected.
func normalizePEM(pemStr string) string { return strings.ReplaceAll(pemStr, `\n`, "\n") }
func ParsePrivateKey(pemStr string) (*ecdsa.PrivateKey, error) {
key, err := jwt.ParseECPrivateKeyFromPEM([]byte(normalizePEM(pemStr)))
if err != nil { return nil, err }
if key.Curve.Params().Name != "P-256" { return nil, fmt.Errorf("invalid curve") }
return key, nil
}
func ParsePublicKey(pemStr string) (*ecdsa.PublicKey, error) {
key, err := jwt.ParseECPublicKeyFromPEM([]byte(normalizePEM(pemStr)))
if err != nil { return nil, err }
if key.Curve.Params().Name != "P-256" { return nil, fmt.Errorf("invalid curve") }
return key, nil
}
