package middleware

import (
	"strings"

	"github.com/IamMaheshGurung/privateOnsenBooking/config"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// TokenClaims defines the claims in JWT tokens
type TokenClaims struct {
	UserID  uint   `json:"userId"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	GuestID uint   `json:"guestId,omitempty"`
	jwt.StandardClaims
}

// AdminAuth middleware checks if request is from admin
func AdminAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from header or cookie
		token := extractToken(c)
		if token == "" {
			// If no token, redirect to login page
			if c.Accepts("html") == "html" {
				return c.Redirect("/admin/login")
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authentication required",
			})
		}

		// Verify token
		claims, err := verifyToken(token)
		if err != nil || claims.Role != "admin" {
			// If invalid token, redirect to login
			if c.Accepts("html") == "html" {
				return c.Redirect("/admin/login")
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid or expired token",
			})
		}

		// Set user info in locals for controllers to use
		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}

// GuestAuth middleware checks if request is from authenticated guest
func GuestAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from header or cookie
		token := extractToken(c)
		if token == "" {
			// If no token, redirect to login page
			if c.Accepts("html") == "html" {
				return c.Redirect("/guest/login")
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authentication required",
			})
		}

		// Verify token
		claims, err := verifyToken(token)
		if err != nil || claims.GuestID == 0 {
			// If invalid token, redirect to login
			if c.Accepts("html") == "html" {
				return c.Redirect("/guest/login")
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid or expired token",
			})
		}

		// Set guest info in locals for controllers to use
		c.Locals("guestID", claims.GuestID)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}

// OptionalAuth middleware checks for authentication but doesn't require it
func OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from header or cookie
		token := extractToken(c)
		if token != "" {
			// Try to verify token but don't reject if invalid
			claims, err := verifyToken(token)
			if err == nil {
				// Set user info in locals if token is valid
				c.Locals("userAuthenticated", true)
				c.Locals("userID", claims.UserID)
				c.Locals("role", claims.Role)
				c.Locals("email", claims.Email)

				if claims.GuestID > 0 {
					c.Locals("guestID", claims.GuestID)
				}
			}
		}

		return c.Next()
	}
}

// Helper function to extract token from request
func extractToken(c *fiber.Ctx) string {
	// Try to get from Authorization header first
	bearerToken := c.Get("Authorization")
	if len(bearerToken) > 7 && strings.ToUpper(bearerToken[0:7]) == "BEARER " {
		return bearerToken[7:]
	}

	// Try to get from cookie
	return c.Cookies("auth_token")
}

// Helper function to verify token
func verifyToken(tokenString string) (*TokenClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JWTSecret), nil
	})

	// Check for errors
	if err != nil {
		return nil, err
	}

	// Validate claims
	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
