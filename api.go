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
)

type GatewayHandler func(GWInput) GWOutput // the handler function signature used for AWS API Gateway
type ErrJSON map[string]string

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
	return
}

// Prepare for JWT operations by reading JWT certificate and key. Must be called by all API functions that use JWT.
func (apiimpl *APIImpl) PrepareJWT(in GWInput) {
	if apiimpl.JWTPublicKey != nil && apiimpl.JWTPrivateKey != nil {
		return
	}
	var err error
	apiimpl.JWTPublicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(in.StageVar["JWTPublicKey"]))
	if err != nil {
		panic(err)
	}
	apiimpl.JWTPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(in.StageVar["JWTPrivateKey"]))
	if err != nil {
		panic(err)
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
				r.ParseMultipartForm(32 << 20)
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
				log.Printf("Failed to serialise type %s into JSON", reflect.TypeOf(result.BodyJSON).Name())
				http.Error(w, fmt.Sprint("Serialisation error"), http.StatusInternalServerError)
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
	apiimpl.PrepareJWT(in)

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
			BodyJSON: ErrJSON{"err": "login is not accepted"},
		}
	}
	if err != nil {
		log.Printf("Database query error - %v", err)
		return GWOutput{
			Status:   http.StatusInternalServerError,
			BodyJSON: ErrJSON{"err": "server error"},
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
		BodyJSON: map[string]string{"Authorization": tokenStr},
		Header:   map[string]string{"Authorization": tokenStr},
	}
}

// Take JWT from Authorization header and respond with token's user name.
func (apiimpl *APIImpl) TestToken(in GWInput) GWOutput {
	apiimpl.PrepareJWT(in)

	token, err := jwt.ParseWithClaims(in.Param.Header["Authorization"], jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return apiimpl.JWTPublicKey, nil
	})
	if err != nil || !token.Valid {
		return GWOutput{
			Status:   http.StatusUnauthorized,
			BodyJSON: ErrJSON{"err": "not authorised"},
		}
	}
	tokenClaims := token.Claims.(jwt.MapClaims)
	return GWOutput{
		Status:   http.StatusOK,
		BodyJSON: map[string]string{JWT_ATTR_USERNAME: fmt.Sprint(tokenClaims[JWT_ATTR_USERNAME])},
	}
}
