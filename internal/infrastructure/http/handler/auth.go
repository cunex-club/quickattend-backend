package handler

import (
	"os"
	"strconv"
	"strings"

	"github.com/cunex-club/quickattend-backend/internal/entity"
	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler interface {
	AuthCunex(c *fiber.Ctx) error
	AuthUser(c *fiber.Ctx) error
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

	userResponse, err := h.Service.Auth.ValidateCUNEXToken(data.Token)
	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	convRefId, convErr := strconv.ParseUint(userResponse.RefId, 10, 64)
	if convErr != nil {
		return response.SendError(c, 500, response.ErrInternalError, "failed to convert RefId")
	}

	user := entity.User{
		RefID:       convRefId,
		FirstnameTH: userResponse.FirstNameTH,
		SurnameTH:   userResponse.LastNameTH,
		TitleTH:     "",
		FirstnameEN: userResponse.FirstnameEN,
		SurnameEN:   userResponse.LastNameEN,
		TitleEN:     "",
	}
	createdUser, createUserErr := h.Service.Auth.CreateUserIfNotExists(&user)
	if createUserErr != nil {
		return response.SendError(c, createUserErr.Status, createUserErr.Code, createUserErr.Message)
	}

	var (
		key []byte
		t   *jwt.Token
	)

	t = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"uuid": createdUser.ID,
		})

	JwtKey, JwtKeyExists := os.LookupEnv("JwtKey")
	if !JwtKeyExists {
		return response.SendError(c, 500, "JWT_SIGN_KEY_NOT_FOUND", "JWT signing key not configured")
	}

	key = []byte(JwtKey)
	access_token, signErr := t.SignedString(key)
	if signErr != nil {
		return response.SendError(c, 500, "JWT_SIGN_FAIL", "failed to sign token")
	}

	return response.OK(c, map[string]string{
		"access_token": access_token,
	})
}

func (h *Handler) AuthUser(c *fiber.Ctx) error {

	header := c.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return response.SendError(c, 401, response.ErrUnauthorized, "missing Authorization header")
	}

	tokenStr := strings.TrimPrefix(header, "Bearer ")
	results, err := h.Service.Auth.GetUserService(tokenStr)

	if err != nil {
		return response.SendError(c, err.Status, err.Code, err.Message)
	}

	return response.OK(c, results)
}
