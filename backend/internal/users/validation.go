package users

import (
	"regexp"
	"strings"
)

var (
	emailPattern    = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	e164Pattern     = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)
	telegramPattern = regexp.MustCompile(`^@?[A-Za-z0-9_]{3,32}$`)
)

type normalizedCreateInput struct {
	Username       string
	Password       string
	Email          *string
	Language       string
	FirstName      *string
	LastName       *string
	DisplayName    *string
	PhoneNumber    *string
	WhatsappNumber *string
	TelegramHandle *string
	IsActive       bool
	Roles          []string
}

type normalizedUpdateInput struct {
	Username       *string
	Password       *string
	Email          *string
	Language       *string
	FirstName      *string
	LastName       *string
	DisplayName    *string
	PhoneNumber    *string
	WhatsappNumber *string
	TelegramHandle *string
	IsActive       *bool
	Roles          *[]string
}

func normalizeCreateInput(in CreateInput) (normalizedCreateInput, map[string]any) {
	issues := map[string]any{}

	username := strings.TrimSpace(in.Username)
	if username == "" {
		appendIssue(issues, "username", "is required")
	} else if len(username) > 64 {
		appendIssue(issues, "username", "must be 64 characters or fewer")
	}

	password := strings.TrimSpace(in.Password)
	if password == "" {
		appendIssue(issues, "password", "is required")
	} else if len(in.Password) > 256 {
		appendIssue(issues, "password", "must be 256 characters or fewer")
	}

	email := normalizeOptionalString(in.Email, 254)
	if in.Email != nil && exceedsLen(*in.Email, 254) {
		appendIssue(issues, "email", "must be 254 characters or fewer")
	}
	if in.Email != nil && email != nil && !emailPattern.MatchString(*email) {
		appendIssue(issues, "email", "must be a valid email address")
	}

	language := "English"
	if in.Language != nil {
		value := strings.TrimSpace(*in.Language)
		if value != "" {
			language = value
		}
	}
	if len(language) > 32 {
		appendIssue(issues, "language", "must be 32 characters or fewer")
	}

	firstName := normalizeOptionalString(in.FirstName, 120)
	lastName := normalizeOptionalString(in.LastName, 120)
	displayName := normalizeOptionalString(in.DisplayName, 160)
	if in.FirstName != nil && exceedsLen(*in.FirstName, 120) {
		appendIssue(issues, "firstName", "must be 120 characters or fewer")
	}
	if in.LastName != nil && exceedsLen(*in.LastName, 120) {
		appendIssue(issues, "lastName", "must be 120 characters or fewer")
	}
	if in.DisplayName != nil && exceedsLen(*in.DisplayName, 160) {
		appendIssue(issues, "displayName", "must be 160 characters or fewer")
	}
	if displayName == nil {
		displayName = deriveDisplayName(firstName, lastName, username)
	}

	phone := normalizeOptionalString(in.PhoneNumber, 20)
	if in.PhoneNumber != nil && exceedsLen(*in.PhoneNumber, 20) {
		appendIssue(issues, "phoneNumber", "must be 20 characters or fewer")
	}
	if in.PhoneNumber != nil && phone != nil && !e164Pattern.MatchString(*phone) {
		appendIssue(issues, "phoneNumber", "must be E.164 format, e.g. +15551234567")
	}

	whatsapp := normalizeOptionalString(in.WhatsappNumber, 20)
	if in.WhatsappNumber != nil && exceedsLen(*in.WhatsappNumber, 20) {
		appendIssue(issues, "whatsappNumber", "must be 20 characters or fewer")
	}
	if in.WhatsappNumber != nil && whatsapp != nil && !e164Pattern.MatchString(*whatsapp) {
		appendIssue(issues, "whatsappNumber", "must be E.164 format, e.g. +15551234567")
	}

	telegram := normalizeTelegram(in.TelegramHandle)
	if in.TelegramHandle != nil && exceedsLen(*in.TelegramHandle, 33) {
		appendIssue(issues, "telegramHandle", "must be 32 characters or fewer")
	}
	if in.TelegramHandle != nil && telegram != nil && !telegramPattern.MatchString(*in.TelegramHandle) {
		appendIssue(issues, "telegramHandle", "must contain letters, numbers, or underscore")
	}

	roles := normalizeRoles(in.Roles)

	return normalizedCreateInput{
		Username:       username,
		Password:       in.Password,
		Email:          email,
		Language:       language,
		FirstName:      firstName,
		LastName:       lastName,
		DisplayName:    displayName,
		PhoneNumber:    phone,
		WhatsappNumber: whatsapp,
		TelegramHandle: telegram,
		IsActive:       in.IsActive,
		Roles:          roles,
	}, issues
}

func normalizeUpdateInput(in UpdateInput, existing UserRecord) (normalizedUpdateInput, map[string]any) {
	issues := map[string]any{}
	out := normalizedUpdateInput{
		IsActive: in.IsActive,
		Roles:    nil,
	}

	if in.Username != nil {
		value := strings.TrimSpace(*in.Username)
		if value == "" {
			appendIssue(issues, "username", "cannot be empty")
		} else if len(value) > 64 {
			appendIssue(issues, "username", "must be 64 characters or fewer")
		} else {
			out.Username = &value
		}
	}

	if in.Password != nil {
		value := strings.TrimSpace(*in.Password)
		if value == "" {
			appendIssue(issues, "password", "cannot be empty")
		} else if len(*in.Password) > 256 {
			appendIssue(issues, "password", "must be 256 characters or fewer")
		} else {
			password := *in.Password
			out.Password = &password
		}
	}

	if in.Email != nil {
		email := normalizeOptionalString(in.Email, 254)
		if exceedsLen(*in.Email, 254) {
			appendIssue(issues, "email", "must be 254 characters or fewer")
		}
		if email != nil && !emailPattern.MatchString(*email) {
			appendIssue(issues, "email", "must be a valid email address")
		}
		out.Email = email
	}

	if in.Language != nil {
		value := strings.TrimSpace(*in.Language)
		if value == "" {
			value = "English"
		}
		if len(value) > 32 {
			appendIssue(issues, "language", "must be 32 characters or fewer")
		} else {
			out.Language = &value
		}
	}

	firstName := normalizeOptionalString(in.FirstName, 120)
	lastName := normalizeOptionalString(in.LastName, 120)
	displayName := normalizeOptionalString(in.DisplayName, 160)
	if in.FirstName != nil && exceedsLen(*in.FirstName, 120) {
		appendIssue(issues, "firstName", "must be 120 characters or fewer")
	}
	if in.LastName != nil && exceedsLen(*in.LastName, 120) {
		appendIssue(issues, "lastName", "must be 120 characters or fewer")
	}
	if in.DisplayName != nil && exceedsLen(*in.DisplayName, 160) {
		appendIssue(issues, "displayName", "must be 160 characters or fewer")
	}
	if in.FirstName != nil {
		out.FirstName = firstName
	}
	if in.LastName != nil {
		out.LastName = lastName
	}
	if in.DisplayName != nil {
		if displayName == nil {
			derived := deriveDisplayName(
				coalesceOptionalString(firstName, existing.FirstName),
				coalesceOptionalString(lastName, existing.LastName),
				coalesceString(out.Username, existing.Username),
			)
			out.DisplayName = derived
		} else {
			out.DisplayName = displayName
		}
	}

	if in.PhoneNumber != nil {
		phone := normalizeOptionalString(in.PhoneNumber, 20)
		if exceedsLen(*in.PhoneNumber, 20) {
			appendIssue(issues, "phoneNumber", "must be 20 characters or fewer")
		}
		if phone != nil && !e164Pattern.MatchString(*phone) {
			appendIssue(issues, "phoneNumber", "must be E.164 format, e.g. +15551234567")
		}
		out.PhoneNumber = phone
	}

	if in.WhatsappNumber != nil {
		whatsapp := normalizeOptionalString(in.WhatsappNumber, 20)
		if exceedsLen(*in.WhatsappNumber, 20) {
			appendIssue(issues, "whatsappNumber", "must be 20 characters or fewer")
		}
		if whatsapp != nil && !e164Pattern.MatchString(*whatsapp) {
			appendIssue(issues, "whatsappNumber", "must be E.164 format, e.g. +15551234567")
		}
		out.WhatsappNumber = whatsapp
	}

	if in.TelegramHandle != nil {
		telegram := normalizeTelegram(in.TelegramHandle)
		if exceedsLen(*in.TelegramHandle, 33) {
			appendIssue(issues, "telegramHandle", "must be 32 characters or fewer")
		}
		if telegram != nil && !telegramPattern.MatchString(*in.TelegramHandle) {
			appendIssue(issues, "telegramHandle", "must contain letters, numbers, or underscore")
		}
		out.TelegramHandle = telegram
	}

	if in.Roles != nil {
		clean := normalizeRoles(*in.Roles)
		out.Roles = &clean
	}

	return out, issues
}

func appendIssue(issues map[string]any, field, message string) {
	raw, ok := issues[field]
	if !ok {
		issues[field] = []string{message}
		return
	}
	values, ok := raw.([]string)
	if !ok {
		values = []string{}
	}
	issues[field] = append(values, message)
}

func normalizeOptionalString(value *string, maxLen int) *string {
	_ = maxLen
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeTelegram(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	trimmed = strings.TrimPrefix(trimmed, "@")
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func deriveDisplayName(firstName, lastName *string, username string) *string {
	parts := []string{}
	if firstName != nil && strings.TrimSpace(*firstName) != "" {
		parts = append(parts, strings.TrimSpace(*firstName))
	}
	if lastName != nil && strings.TrimSpace(*lastName) != "" {
		parts = append(parts, strings.TrimSpace(*lastName))
	}
	if len(parts) > 0 {
		value := strings.TrimSpace(strings.Join(parts, " "))
		if value != "" {
			return &value
		}
	}
	user := strings.TrimSpace(username)
	if user == "" {
		return nil
	}
	return &user
}

func coalesceString(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
}

func coalesceOptionalString(value, fallback *string) *string {
	if value != nil {
		return value
	}
	return fallback
}

func exceedsLen(value string, maxLen int) bool {
	return len(strings.TrimSpace(value)) > maxLen
}
