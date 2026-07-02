package tenant
import ( "net"; "net/http"; "strings" )
func ResolveSlug(req *http.Request) string {
if slug := req.Header.Get("X-Tenant-Slug"); slug != "" { return slug }
if id := req.Header.Get("X-Tenant-ID"); id != "" { return id }
host := req.Host
if h, _, err := net.SplitHostPort(host); err == nil { host = h }
if net.ParseIP(host) != nil { return "" }
parts := strings.Split(host, ".")
if len(parts) >= 3 { return parts[0] }
return ""
}
