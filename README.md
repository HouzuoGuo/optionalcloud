# optionalcloud
An experiment running Go program as AWS Lambda to route and serve HTTP APIs on AWS API Gateway, while retaining the freedom of running the same HTTP APIs using custom
HTTP servers without the involvement of AWS API Gateway.

# Features
- GET / - Display a greeting message.
- POST /login/{username}?password= - Authenticate a user using username/password combination, return JWT if authorised.
- GET /login - Return user name if the Authorization header in the incoming request contains a valid JWT.
- Run stand-alone HTTP server to serve the same APIs.
