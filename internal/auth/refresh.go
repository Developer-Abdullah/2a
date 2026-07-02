package auth
import ( "crypto/rand"; "crypto/sha256"; "encoding/base64"; "encoding/hex" )
func GenerateRefreshToken() (string, string, error) {
b := make([]byte, 32)
if _, err := rand.Read(b); err != nil { return "", "", err }
plain := base64.URLEncoding.EncodeToString(b)
hash := sha256.Sum256([]byte(plain))
return plain, hex.EncodeToString(hash[:]), nil
}
