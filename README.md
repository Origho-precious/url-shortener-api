# URL Shortener Server

The server is responsible for handling user authentication, URL shortening, analytics, and other backend functionalities.

## Technologies Used

- **Golang**: Programming language used for backend development.
- **Gin**: Web framework for Golang, used for building web applications.
- **go-jwt**: Library for JSON Web Tokens (JWT) implementation in Golang, used for user authentication and session management.
- **MongoDB**: NoSQL database used for storing user account information, URL mappings, and analytics data.
- **imagekit**: Go client library for the ImageKit.io API, used for saving generated QR code images.
- **go-qr-code**: Library for generating QR codes in Golang.
- **mail-trap**: Library for send emails, this application specifically uses mailtrap templates accessed via API.
- **godotenv**: Go library for loading environment variables from a .env file.

## Getting Started:

To get started with the project, clone the repository and run:

```bash
go mod download
```

or

```bash
go mod tidy
```

Create a new file in `./server` directory called `.env` and copy & paste what is in `.env.example` into it:

```bash
# SERVER configs
PORT=
GIN_MODE=debug # Options: debug, release, test
BASE_URL=
CLIENT_URL=
URL_REDIRECT_PREFIX=
MONGO_URI=
JWT_SECRET=

# MAILTRAP configs
MAILTRAP_SENDER_EMAIL=
MAILTRAP_AUTH=

# EMAIL TEMPLATE IDs
RESET_PASSWORD_TEMPLATE_UUID=
EMAIL_VERIFICATION_TEMPLATE_UUID=

# IMAGEKIT configs
IMAGEKIT_PUBLIC_KEY=
IMAGEKIT_PRIVATE_KEY=
IMAGEKIT_URL_ENDPOINT=
```

Make sure to replace the placeholders with your actual credentials.

Start the server by running:

```bash
go run main.go
```

## API Endpoints

Sure! Here are all the endpoints defined for both the user and URL routers:

### User Endpoints

1. **POST /v1/api/users/**

   - **Description**: Register a new user.
   - **Handler**: `HandleSignup` function in the `controllers` package.
   - **Body**:
     ```json
     {
     	"email": "", // required
     	"password": "", // required
     	"fullName": "" // required
     }
     ```

2. **POST /v1/api/users/login**

   - **Description**: Log in an existing user.
   - **Handler**: `HandleLogin` function in the `controllers` package.
   - **Body**:

   ```json
   {
   	"email": "", // required
   	"password": "" // required
   }
   ```

3. **POST /v1/api/users/verify**

   - **Description**: Verify user email address.
   - **Middleware**: Requires authentication token.
   - **Handler**: `HandleEmailVerification` function in the `controllers` package.
   - **Body**:

   ```json
   {
   	"token": ""
   }
   ```

4. **GET /v1/api/users/resend-verification-token**

   - **Description**: Resend email verification token.
   - **Middleware**: Requires authentication token.
   - **Handler**: `HandleEmailVerificationTokenResend` function in the `controllers` package.

5. **POST /v1/api/users/forgot-password**

   - **Description**: Initiate forgot password flow.
   - **Handler**: `HandleForgotPassword` function in the `controllers` package.
   - **Body**:

   ```json
   {
   	"email": "" // required
   }
   ```

6. **PATCH /v1/api/users/reset-password**

   - **Description**: Reset user password.
   - **Middleware**: Requires authentication token.
   - **Handler**: `HandlePasswordReset` function in the `controllers` package.
   - **Query Params**:

   ```json
   {
   	"token": "" // required
   }
   ```

   - **Body**:

   ```json
   {
   	"password": "" // required
   }
   ```

7. **GET /v1/api/users/me**

   - **Description**: Get user profile.
   - **Middleware**: Requires authentication token.
   - **Handler**: `GetUserProfile` function in the `controllers` package.

8. **PATCH /v1/api/users/edit**

   - **Description**: Edit user full name.
   - **Middleware**: Requires authentication token.
   - **Handler**: `HandleUserFullNameEdit` function in the `controllers` package.
   - **Body**:

   ```json
   {
   	"fullName": "" // required
   }
   ```

### URL Endpoints

1. **POST /v1/api/urls/**

   - **Description**: Create a new shortened URL.
   - **Middleware**: Requires authentication token.
   - **Handler**: `HandleCreateShortUrl` function in the `controllers` package.
   - **Body**:

   ```json
   []{
   	"url": "", // required
   	"alias": "",
   	"expireDate": ""
   }
   ```

2. **GET /v1/api/urls/**

   - **Description**: Retrieve all URLs shortened by the user.
   - **Middleware**: Requires authentication token.
   - **Handler**: `GetUrlsByUserID` function in the `controllers` package.
   - **Query Params**:

   ```json
   {
   	"page": "",
   	"limit": ""
   }
   ```

3. **DELETE /v1/api/urls/:id/delete**

   - **Description**: Delete a shortened URL by its ID.
   - **Middleware**: Requires authentication token.
   - **Handler**: `HandleUrlDelete` function in the `controllers` package.

4. **GET /redirect/:slug**
   - **Description**: Redirect to the original URL associated with the given slug.
   - **Handler**: `RedirectToLongUrl` function in the `controllers` package.

## References

- **Golang**: [Official Website](https://golang.org/)
- **Gin**: [GitHub Repository](https://github.com/gin-gonic/gin)
- **go-jwt**: [GitHub Repository](https://github.com/dgrijalva/jwt-go)
- **Imagekit**: [Official Website](https://imagekit.io/)
- **go-qr-code**: [GitHub Repository](https://github.com/skip2/go-qrcode)
- **MongoDB Go Driver**: [Official Website](https://www.mongodb.com/docs/drivers/go/current/)
- **godotenv**: [GitHub Repository](https://github.com/joho/godotenv)
- **Mailtrap**: [GitHub Repository](https://mailtrap.io/home)
