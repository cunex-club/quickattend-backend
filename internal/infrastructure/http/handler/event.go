package handler

import (
	"strings"

	"github.com/cunex-club/quickattend-backend/internal/infrastructure/http/response"
	"github.com/gofiber/fiber/v2"
)

type EventHandler interface {
	GetEvents(*fiber.Ctx) error
}

func (h *Handler) GetEvents(c *fiber.Ctx) error {
	args := make(map[string]string)
	queryKeys := []string{"managed", "page", "pageSize", "search"}
	pairs := strings.SplitSeq(string(c.Request().URI().QueryString()), "&")

	for pair := range pairs {
		pair := strings.Split(pair, "=")
		key := pair[0]
		value := pair[1]
		for foundedKey, _ := range args {
			if key == foundedKey {
				return response.SendError(c, 400, response.ErrBadRequest, "Duplicate query parameter")
			}
		}
		unknownKey := true
		for _, queryKey := range queryKeys {
			if key == queryKey {
				unknownKey = false
			}
		}
		if unknownKey {
			return response.SendError(c, 400, response.ErrBadRequest, "Unknown URI query key")
		}
		args[key] = value
	}
	// for _, v := range queryKeys {
	// 	param := c.Request().URI().QueryArgs().PeekMulti(v)
	// 	if len(param) > 1 {
	// 		return response.SendError(c, 400, response.ErrBadRequest, "Duplicate query parameter")
	// 	}
	// 	if len(param) == 0 {
	// 		args[v] = ""
	// 	} else {
	// 		args[v] = string(param[0])
	// 	}
	// }

	// call service, pass in args and check types, limits, etc of each arg

	return response.OK(c, args)
}
