package response

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func File(c *fiber.Ctx, filename string, contentType string, content []byte) error {
	c.Set("Content-Type", contentType)

	// RFC 6266: provide an ASCII-only `filename` for legacy clients plus a
	// percent-encoded `filename*` so non-ASCII names (e.g. Thai event names)
	// survive transport intact instead of being garbled or dropped.
	ascii := asciiFilename(filename)
	c.Set("Content-Disposition", fmt.Sprintf(
		`attachment; filename="%s"; filename*=UTF-8''%s`,
		ascii, url.PathEscape(filename),
	))
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

// asciiFilename produces a safe ASCII fallback for the legacy `filename`
// parameter: non-printable/non-ASCII runes and characters that could break the
// header (quotes, backslash) are replaced with underscores.
func asciiFilename(name string) string {
	var b strings.Builder
	for _, r := range name {
		if r < 0x20 || r > 0x7E || r == '"' || r == '\\' {
			b.WriteByte('_')
		} else {
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "export.xlsx"
	}
	return out
}
