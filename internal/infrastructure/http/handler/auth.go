package handler

import (
	"net/http"
	"strings"

	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler interface {
	AuthCunex(c *fiber.Ctx) error
}

func validateToken(c *fiber.Ctx, token string) error {
	url := "https://jsonplaceholder.typicode.com/todos/1"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response.SendError(c, 500, response.ErrInternalError, "failed to create token validation request")
	}

	req.Header.Set("Content-type", "application/json")
	req.Header.Set("ClientId", "")
	req.Header.Set("ClientSecret", "")

	q := req.URL.Query()
	q.Add("token", token)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return response.SendError(c, 500, response.ErrInternalError, "failed to call external token validation API")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response.SendError(c, resp.StatusCode, response.ErrUnauthorized, "invalid token")
	}

	return nil
}

func (h *Handler) AuthCunex(c *fiber.Ctx) error {

	var data = struct {
		Token string `json:"token"`
	}{}

	if err := c.BodyParser(&data); err != nil {
		return response.SendError(c, 400, response.ErrBadRequest, "invalid JSON body")
	}

	if strings.TrimSpace(data.Token) == "" {
		return response.SendError(c, 400, "TOKEN_REQUIRED", "token is required")
	}

	err := validateToken(c, data.Token)
	if err != nil {
		return err
	}

	var (
		key []byte
		t   *jwt.Token
		s   string
	)

	key = []byte("HELLO")
	t = jwt.New(jwt.SigningMethodHS256)
	s, err = t.SignedString(key)
	if err != nil {
		return response.SendError(c, 500, "JWT_SIGN_FAIL", "failed to sign token")
	}

	return response.OK(c, map[string]string{
		"token": s,
	})
}
