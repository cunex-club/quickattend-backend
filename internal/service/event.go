package service

type EventService interface {
	ValidateQRCode(code string) (errMsg string)
}

func (s *service) ValidateQRCode(code string) string {
	if code == "" {
		return "Missing URL path parameter 'qrcode'"
	}
	length := 0
	numbers := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	for _, r := range code {
		isDigit := false
		for _, num := range numbers {
			if string(r) == num {
				isDigit = true
			}
		}
		if !isDigit {
			return "URL path parameter 'qrcode' contains non-number character(s)"
		}

		length += 1
		if length > 10 {
			break
		}
	}
	if length != 10 {
		return "URL path parameter 'qrcode' must have length of 10"
	}

	return ""
}
