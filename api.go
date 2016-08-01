package main

import (
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
	"github.com/zenazn/goji/web"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	METHOD_GET        = "GET"
	METHOD_POST       = "POST"
	JWT_EXPIRE_AFTER  = time.Hour * 24 // Tokens expire after 24 hours
	JWT_ATTR_EXPIRE   = "expire"       // expire is a Unix timestamp
	JWT_ATTR_USERNAME = "username"     // username is a string

	JWT_PUBLIC_KEY = `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAPGRDaLmC6cLMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTYwNzI1MTA0OTM2WhcNMTcwNzI1MTA0OTM2WjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAq0z1mnpP4fLfQ3nqIjrtH4RhclHhgJoCLkT5YBl9hyXTliq9pGnl8Bzx
rlcoKizhv0HGeLNQOiox/tV0zLakQ5gXfMDE7KSy+vBDJWG6bBmdC6cMjECSKssQ
GTwHdFtctV/U/0Ih79dYoBqW1yRB8qQCzm0fdC6UDnYZncqvhv2u4g7CyOEIO+Q5
dJm9LQVnLXvOXT3IoAiW2RUnbJ2aeA7hDYMZ/pcoasnMn+U5dGSngOsjWyk2T1Y1
IaH2f84QwruRu26BPNEGht0juKTCT18e8Awn3ZLpkhr2GwNTzykrxzf/SyIai+Uk
k+gGGvJmZNQoELEhw7E9vXm2gxXAdwIDAQABo1AwTjAdBgNVHQ4EFgQUpBZtve6F
6NZCUMndKL1hD3A34zYwHwYDVR0jBBgwFoAUpBZtve6F6NZCUMndKL1hD3A34zYw
DAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQUFAAOCAQEAjAkqxdWCW7ygo1acXo1o
gUa1TpwG39Ff6v3NrlQriO6CQsIRCckmImddYZcpZwaYCLrBCJbKh/YULW15tTbA
uIlHLWQEalzrWfw0kIDVIs6QXj3W+dOLAalgFi41rNE8hxSyFHCBGVKjE2Y526xb
iC5nEFjRoF7ez5dy6O/shiYxzepIB1Nx6CUZpYG+WDPHn9OzwgaIn/nyiXDWmFH2
zYSamiw46obPUmu/HNKql4S2iJkOq5H1uqHQjnzUc0rv7CN42Wxhn0UD9RO+W7+L
aIP1gdIMJNgAFPRfpm5JVDkWzowBS7m3PY+ruSTos4mimqvZSO60jUFfshfpwQ+3
Pw==
-----END CERTIFICATE-----`
	JWT_PRIVATE_KEY = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCrTPWaek/h8t9D
eeoiOu0fhGFyUeGAmgIuRPlgGX2HJdOWKr2kaeXwHPGuVygqLOG/QcZ4s1A6KjH+
1XTMtqRDmBd8wMTspLL68EMlYbpsGZ0LpwyMQJIqyxAZPAd0W1y1X9T/QiHv11ig
GpbXJEHypALObR90LpQOdhmdyq+G/a7iDsLI4Qg75Dl0mb0tBWcte85dPcigCJbZ
FSdsnZp4DuENgxn+lyhqycyf5Tl0ZKeA6yNbKTZPVjUhofZ/zhDCu5G7boE80QaG
3SO4pMJPXx7wDCfdkumSGvYbA1PPKSvHN/9LIhqL5SST6AYa8mZk1CgQsSHDsT29
ebaDFcB3AgMBAAECggEAU+fVYX5JxI33SBDeWzfr0AVCygFLaHeHW+yTDbxOnTUt
B6AV1gO9CjjTNKciWE41oT3xnkuOn37tkDo0BNXtbeKAlq3Bh3xA4uNusE/HRY3i
O8PuRICYV/exAftCV38s0PaI2SMmhlk/4uRDQExVNSma6kvPHVR3VwIIGB8gjQjG
RVZHJPE+lr44RAZKLpcUrHYjeOxVWxfuatxwLVxInxIx3C5yMZlT4+b6qzdTNUCU
hDvoEiZMgTZZPdLmmqIMzcXvowYz5U+VanCEDCTx0kc/FdU+5xVrIGfRi9rZnQ7f
Tg3hLplVRHBQY01SJWY0TWhfkhuFgBmzohORTuNRuQKBgQDUrMYWTYD/EgqPaRDU
uw2ARIHShD1XUZFW4w4s4ThHQB/uJOb92CPc3K5i0Xqf2xH9a6AC2sr/qvLmnIuG
57hIKggFqu82zRLTPCc33Q4VXD71PkJU05drTmmmvDI+X5Z4eP00aBCWLLq5YRCv
ozst3ll+Vj9izCM2ROnME8efzQKBgQDOMnjs9OiPhgRST9/pZzNDt4CNNFKzY/i9
tbYfIGz9MPDhGxaJW25UtVgjF4mkgwjyaIwchYRN0iYOcNSXoaZQ3y1PwEmlDOYs
IViCHT7TPC641YVc0UdfPmXUI1A5AyS1a2RjKZwSpP8kBrwVjgS9D+Vh4RTl7VRX
ZX4lD+W1UwKBgBlYJZsOzWqYOc3xVWIkkG1SvK3buHupasqR8GSEynIjQCrfFu/1
TADMA7QfBp/6OWCb7MuqSzrAooW87hu7jYh8CcyzHCLJuY6Wwo2zuDPvdElBjCIT
vR26kHigQNSSC5p7wKD4LdHXrsDcwmJL74d90ehuWstpTGDxQXNigA2ZAoGAGOJE
b6w6qJ9uxBQ5nGxE5oYtsFzBIj8NVK+qM+Vw4blXSINBXAA5t2VPJqT/imf52288
gXCnf9C9oP6C2W27qYTVbgtxl8aPvIGlscYfv9RCezHhb0seRuM73LcKRmcXtgEo
00LBQArDc7CQYDWMYtiZQQ+tuvXCOO3ZpFVfzlsCgYEAnxzou/29ymiz+SSvZ3xB
/vN3aYI9tLmzT9FGnBg/pr3VooQayJ43YI+IrMeigS4Ouxju8nKOCN4HvN7xYpge
d1QpYh7WlfS9UQaiFBl2sbYSRGP/Mdug8Nzp7odKPWWHbK/dK3KagPZ3nrSteHMC
aqP46UBsU+jCR+XT1biGlsU=
-----END PRIVATE KEY-----`
)

type GatewayHandler func(GWInput) GWOutput // the handler function signature used for AWS API Gateway

// API implementations.
type APIImpl struct {
	Routes        map[string]GatewayHandler
	HTTPStageVars map[string]string // stage variables used by conventional HTTP handlers
	JWTPublicKey  *rsa.PublicKey
	JWTPrivateKey *rsa.PrivateKey
}

func NewAPIImpl(httpStageVars map[string]string) (ret *APIImpl) {
	ret = &APIImpl{HTTPStageVars: httpStageVars}
	ret.Routes = map[string]GatewayHandler{
		METHOD_GET + "/":                  ret.Greeting,
		METHOD_POST + "/login/{username}": ret.Authenticate,
		METHOD_GET + "/login":             ret.TestToken,
	}
	var err error
	ret.JWTPublicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(JWT_PUBLIC_KEY))
	if err != nil {
		panic(err)
	}
	ret.JWTPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(JWT_PRIVATE_KEY))
	if err != nil {
		panic(err)
	}
	return
}

// Prepare for JWT operations by reading JWT certificate and key. Must be called by all API functions that use JWT.
func (apiimpl *APIImpl) PrepareJWT(in GWInput) {
	if apiimpl.JWTPublicKey != nil && apiimpl.JWTPrivateKey != nil {
		return
	}
}

// Transform API implementations to conventional HTTP handlers and install the routes on web mux.
func (apiimpl *APIImpl) InstallMuxRoutes(mux *web.Mux) {
	for endpoint := range apiimpl.Routes {
		gwHandler := apiimpl.Routes[endpoint]
		slash := strings.Index(endpoint, "/")
		method := endpoint[0:slash]
		route := endpoint[slash:]
		// Transform the way URL parameters are specified from /{param} to /:param
		route = strings.Replace(route, "{", ":", -1)
		route = strings.Replace(route, "}", "", -1)
		// Transform gateway handler into a conventional HTTP handler
		conventionalHandler := func(c web.C, w http.ResponseWriter, r *http.Request) {
			// Construct AWS API gateway input from the request
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Print("IO failure in HTTP handler")
				return
			}
			// Header is a tricky one - AWS API gateway only accepts single header values
			gwHeaders := map[string]string{}
			for key, vals := range r.Header {
				if len(vals) > 0 {
					gwHeaders[key] = strings.Join(vals, ", ")
				}
			}
			// Form value is also a tricky one - AWS API gateway only accepts single form values
			if r.Form == nil {
				r.ParseMultipartForm(32 << 20) // 32 MB max
			}
			gwFormValues := map[string]string{}
			for key, vals := range r.Form {
				if len(vals) > 0 {
					gwFormValues[key] = vals[0]
				}
			}
			gwInput := GWInput{
				Body: body,
				Param: GWParams{
					Path:   c.URLParams,
					Query:  gwFormValues,
					Header: gwHeaders,
				},
				StageVar: apiimpl.HTTPStageVars,
				Context: GWRequest{
					IP:     r.RemoteAddr,
					Method: strings.ToLower(method),
					Stage:  "",
					Path:   route,
				},
			}
			// Call original gateway handler and write the result to HTTP response
			result := gwHandler(gwInput)
			for key, val := range result.Header {
				w.Header().Set(key, val)
			}
			w.WriteHeader(result.Status)
			// enable CORS for Swagger UI
			NoCacheAllowCORS(w)
			w.Header().Set("Content-Type", "application/json")
			if jsonBytes, err := json.Marshal(result.BodyJSON); err != nil {
				log.Panicf("Failed to serialise type %s into JSON", reflect.TypeOf(result.BodyJSON).Name())
			} else {
				w.WriteHeader(result.Status)
				w.Write(jsonBytes)
			}
		}
		switch method {
		case METHOD_GET:
			mux.Get(route, conventionalHandler)
		case METHOD_POST:
			mux.Post(route, conventionalHandler)
		default:
			log.Panicf("Does not work for method %s", method)
		}
		log.Printf("Install handler '%s' on '%s'", method, route)
	}
}

// Take no input parameter and simply respond with a greeting message.
func (apiimpl *APIImpl) Greeting(_ GWInput) GWOutput {
	return GWOutput{Status: http.StatusOK, BodyJSON: "Have a nice day"}
}

// Validate user credentials and respond with a JWT if credentials are acceptable.
func (apiimpl *APIImpl) Authenticate(in GWInput) GWOutput {
	username := in.Param.Path["username"]
	password := in.Param.Query["password"]
	// Validate username and password (not hashed for this demo) with database
	err := DoSQL(GetSQLConfigFromAPIGateway(in), func(db *sqlx.DB) error {
		var foo int
		return db.Get(&foo, "select 1 from users where username = ? and password = ?", username, password)
	})
	if err == sql.ErrNoRows {
		return GWOutput{
			Status:   http.StatusUnauthorized,
			BodyJSON: "credentials are not accepted",
		}
	}
	if err != nil {
		log.Printf("Database query error - %v", err)
		return GWOutput{
			Status:   http.StatusInternalServerError,
			BodyJSON: fmt.Sprintf("database query error - %v", err),
		}
	}
	// Sign JWT and return it in header
	var tokenStr string
	tokenObj := jwt.New(jwt.SigningMethodRS256)
	tokenClaims := jwt.MapClaims{
		JWT_ATTR_EXPIRE:   strconv.Itoa(int(time.Now().Add(JWT_EXPIRE_AFTER).Unix())),
		JWT_ATTR_USERNAME: username,
	}
	tokenObj.Claims = tokenClaims
	if tokenStr, err = tokenObj.SignedString(apiimpl.JWTPrivateKey); err != nil {
		panic(err)
	}
	return GWOutput{
		Status:   http.StatusOK,
		BodyJSON: map[string]string{"Authorization": "Bearer " + tokenStr},
		Header:   map[string]string{"Authorization": "Bearer " + tokenStr},
	}
}

// Take JWT from Authorization header and respond with token's user name.
func (apiimpl *APIImpl) TestToken(in GWInput) GWOutput {
	headerAuthorization := strings.TrimPrefix(in.Param.Header["Authorization"], "Bearer ")
	token, err := jwt.ParseWithClaims(headerAuthorization, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return apiimpl.JWTPublicKey, nil
	})
	if err != nil || !token.Valid {
		return GWOutput{
			Status:   http.StatusUnauthorized,
			BodyJSON: "token is not accepted",
		}
	}
	tokenClaims := token.Claims.(jwt.MapClaims)
	return GWOutput{
		Status:   http.StatusOK,
		BodyJSON: map[string]string{JWT_ATTR_USERNAME: fmt.Sprint(tokenClaims[JWT_ATTR_USERNAME])},
	}
}
