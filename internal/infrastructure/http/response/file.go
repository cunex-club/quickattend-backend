package response

import "github.com/gofiber/fiber/v2"

func File(c *fiber.Ctx, filename string, contentType string, content []byte) error {
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	return c.Status(fiber.StatusOK).Send(content)
}

func Excel(c *fiber.Ctx, filename string, content []byte) error {
	return File(
		c,
		filename,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		content,
	)
}
