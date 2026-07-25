package response

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func Excel(c *fiber.Ctx, filename string, content []byte) error {
	ascii := strings.Map(func(r rune) rune {
		if r < 0x20 || r > 0x7e || r == '"' || r == '\\' {
			return '_'
		}
		return r
	}, filename)
	if strings.TrimSpace(ascii) == "" {
		ascii = "export.xlsx"
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(
		`attachment; filename="%s"; filename*=UTF-8''%s`,
		ascii,
		url.PathEscape(filename),
	))
	return c.Status(fiber.StatusOK).Send(content)
}
