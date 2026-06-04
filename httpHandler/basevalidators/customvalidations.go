package basevalidators

import (
	"reflect"
	"regexp"
	"sync"
	"time"
	"unicode"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var getInstance = sync.OnceValue(func() *CustomValidator {
	instance := &CustomValidator{}
	instance.specialChars = map[rune]bool{'!': true, '@': true, '#': true, '$': true, '%': true, '^': true, '&': true, '*': true, '(': true, ')': true, '-': true, '_': true, '+': true, '=': true, '<': true, '>': true, '?': true, '/': true, '{': true, '}': true, '[': true, ']': true, '|': true}
	en := en.New()
	uni := ut.New(en, en)
	instance.trans, _ = uni.GetTranslator("en")
	instance.v = validator.New(validator.WithRequiredStructEnabled())
	instance.RegisterCustomValidators()
	return instance
})

type CustomValidator struct {
	trans        ut.Translator
	v            *validator.Validate
	specialChars map[rune]bool
}

func GetInstance() *CustomValidator {
	return getInstance()
}

func (cv *CustomValidator) GetTrans() ut.Translator {
	return cv.trans
}

func (cv *CustomValidator) GetValidator() *validator.Validate {
	return cv.v
}

func (cv *CustomValidator) RegisterCustomValidators() {
	cv.v.RegisterValidation("password", cv.PasswordValidator)
	cv.v.RegisterValidation("addressLines", cv.AddressLinesValidator)

	cv.v.RegisterValidation("diffValues", cv.DifferentValuesForKeys)
	cv.v.RegisterValidation("dob", cv.ValidateDOB)

	cv.v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Tag.Get("field")
	})

	cv.v.RegisterTranslation("required", cv.trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} must have a value!", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())

		return t
	})

	cv.v.RegisterTranslation("password", cv.trans, func(ut ut.Translator) error {
		return ut.Add("password", "{0} must be 8 characters long and should contain combination of special characters, digits, small and capital letters", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("password", fe.Field())

		return t
	})

	cv.v.RegisterTranslation("required_without", cv.trans, func(ut ut.Translator) error {
		return ut.Add("required_without", "{0} atleast must have a value!", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required_without", fe.Field())

		return t
	})

	cv.v.RegisterTranslation("min", cv.trans, func(ut ut.Translator) error {
		return ut.Add("min", "{0} must have minimum values! {1}", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("min", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("max", cv.trans, func(ut ut.Translator) error {
		return ut.Add("max", "{0} must have maximum values! {1}", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("max", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("password", cv.trans, func(ut ut.Translator) error {
		return ut.Add("password", "{0} must be minimum 8 characters long and maximum 64 characters and should contain combination of special characters, digits, small and capital letters", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("password", fe.Field())

		return t
	})

	cv.v.RegisterTranslation("email", cv.trans, func(ut ut.Translator) error {
		return ut.Add("email", "{0} syntax is not correct", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("email", fe.Field())

		return t
	})

	cv.v.RegisterTranslation("gt", cv.trans, func(ut ut.Translator) error {
		return ut.Add("gt", "{0} must be greater than! {1}", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("gt", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("gte", cv.trans, func(ut ut.Translator) error {
		return ut.Add("gte", "{0} must be greater than equal to {1}", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("gte", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("arraylength", cv.trans, func(ut ut.Translator) error {
		return ut.Add("arraylength", "{0} must be json array with at least one value", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("arraylength", fe.Field())

		return t
	})

	cv.v.RegisterTranslation("oneof", cv.trans, func(ut ut.Translator) error {
		return ut.Add("oneof", "{0} can only be {1}", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("oneof", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("lte", cv.trans, func(ut ut.Translator) error {
		return ut.Add("lte", "{0} must be lesser than equal to {1}", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("lte", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("alpha", cv.trans, func(ut ut.Translator) error {
		return ut.Add("alpha", "{0} must have all characters", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("alpha", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("e164", cv.trans, func(ut ut.Translator) error {
		return ut.Add("e164", "{0} must have e164 format", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("e164", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("alphanum", cv.trans, func(ut ut.Translator) error {
		return ut.Add("alphanum", "{0} must have characters and numeric only", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("alphanum", fe.Field(), fe.Param())

		return t
	})

	cv.v.RegisterTranslation("addressLines", cv.trans, func(ut ut.Translator) error {
		return ut.Add("addressLines", "{0} must have a maximum of 3 lines and each line must not exceed 100 characters", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("addressLines", fe.Field())
		return t
	})

	cv.v.RegisterTranslation("dob", cv.trans, func(ut ut.Translator) error {
		return ut.Add("dob", "{0} must have yyyy-mm-dd and should not be in future or more than 110 years", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("addressLines", fe.Field())
		return t
	})
	// Register a custom translation for URL validation
	cv.v.RegisterTranslation("url", cv.trans,
		// This function adds a custom message to the translator for the "url" validation tag.
		func(ut ut.Translator) error {
			// The custom message indicates that the field must contain a valid URL.
			return ut.Add("url", "{0} must be a valid URL", true)
		},
		// This function is called when a validation error occurs.
		// It retrieves the translated message and formats it with the field name.
		func(ut ut.Translator, fe validator.FieldError) string {
			// Get the translated message for the "url" tag, using the field name that failed validation.
			t, _ := ut.T("url", fe.Field())
			// Return the translated error message.
			return t
		})

	cv.v.RegisterValidation("dateformat", cv.ValidateDateFormat)

	cv.v.RegisterTranslation("dateformat", cv.trans, func(ut ut.Translator) error {
		return ut.Add("dateformat", "{0} field must be in the format yyyy-mm-dd and should be a valid date", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("dateformat", fe.Field())
		return t
	})

	cv.v.RegisterValidation("timeRange", cv.ValidateTimeRange)

	cv.v.RegisterTranslation("timeRange", cv.trans, func(ut ut.Translator) error {
		return ut.Add("timeRange", "{0}: endTime must be greater than StartTime, not greater than the current time and time values should not be zero", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("timeRange", fe.Field())
		return t
	})

	cv.v.RegisterValidation("dateMatch", cv.ValidateDatesMatch)

	cv.v.RegisterTranslation("dateMatch", cv.trans, func(ut ut.Translator) error {
		return ut.Add("dateMatch", "{0} field must have the same date as StartTime and EndTime", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("dateMatch", fe.Field())
		return t
	})

	cv.v.RegisterValidation("future", cv.FutureValidation)
	cv.v.RegisterTranslation("future", cv.trans, func(ut ut.Translator) error {
		return ut.Add("future", "{0} field must have future time.", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("future", fe.Field())
		return t
	})
}

func (cv *CustomValidator) PasswordValidator(fl validator.FieldLevel) bool {
	// Use a regular expression to enforce password criteria
	password := fl.Field().String()
	var (
		hasLower   bool
		hasUpper   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
			hasDigit = true
		case cv.isSpecialCharacter(char):
			hasSpecial = true
		}
	}
	return hasLower && hasUpper && hasDigit && hasSpecial && len(password) >= 8 && len(password) <= 64
}

func (cv *CustomValidator) isSpecialCharacter(char rune) bool {
	_, ok := cv.specialChars[char]
	return ok
}
func (cv *CustomValidator) CreateErrors(err error) map[string][]string {
	returnMap := make(map[string][]string)
	errs := err.(validator.ValidationErrors)

	for _, e := range errs {
		returnMap[e.Field()] = []string{e.Translate(cv.GetTrans())}
	}
	return returnMap
}

func (cv *CustomValidator) AddressLinesValidator(fl validator.FieldLevel) bool {
	lines, ok := fl.Field().Interface().([]string)
	if !ok {
		return false
	}
	if len(lines) > 3 {
		return false // No more than 3 lines allowed
	}
	for _, line := range lines {
		if len(line) > 100 {
			return false // Each line must not exceed 100 characters
		}
	}
	return true
}

func (cv *CustomValidator) DifferentValuesForKeys(fl validator.FieldLevel) bool {
	// Get the struct value
	structValue := fl.Top()

	// Create a map to keep track of seen values
	seenValues := make(map[string]bool)

	// Iterate over the struct fields
	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		if field.Kind() == reflect.String {
			value := field.String()
			if seenValues[value] {
				return false // Duplicate value found
			}
			seenValues[value] = true
		}
	}

	return true
}

func (cv *CustomValidator) ValidateDateFormat(fl validator.FieldLevel) bool {
	date := fl.Field().String()

	// Check if the format is "yyyy-mm-dd" using regular expression
	re := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !re.MatchString(date) {
		return false
	}

	// Parse the date to ensure it's valid
	_, err := time.Parse("2006-01-02", date)
	return err == nil
}

// ValidateDOB checks if the date of birth is in a valid format and within a reasonable age range
func (cv *CustomValidator) ValidateDOB(fl validator.FieldLevel) bool {
	dobStr := fl.Field().String()
	dob, err := time.Parse("2006-01-02", dobStr) // Adjust the format as needed
	if err != nil {
		return false
	}

	// Check if the date is in the past
	if dob.After(time.Now()) {
		return false
	}

	// Check if the age is within a reasonable range (e.g., 0 to 110 years)
	age := time.Now().Year() - dob.Year()
	if age < 0 || age > 110 {
		return false
	}

	return true
}

func (cv *CustomValidator) ValidateDatesMatch(fl validator.FieldLevel) bool {
	// Retrieve the current struct using reflection
	structValue := fl.Parent()

	// Access the StartTime, EndTime, and Date fields
	startTimeField := structValue.FieldByName("StartTime")
	endTimeField := structValue.FieldByName("EndTime")
	dateField := structValue.FieldByName("Date")

	if !startTimeField.IsValid() || !endTimeField.IsValid() || !dateField.IsValid() {
		return false // Fields do not exist or have incorrect types
	}

	if startTimeField.Kind() != reflect.Int64 || endTimeField.Kind() != reflect.Int64 || dateField.Kind() != reflect.String {
		return false // Fields are not of the expected types
	}

	startTime := startTimeField.Int()
	endTime := endTimeField.Int()
	date := dateField.String()

	// Convert startTime and endTime from epoch to "yyyy-mm-dd" format
	startTimeDate := time.Unix(startTime, 0).UTC().Format("2006-01-02")
	endTimeDate := time.Unix(endTime, 0).UTC().Format("2006-01-02")

	// Check if startTimeDate, endTimeDate, and dateField have the same date
	if startTimeDate != date || endTimeDate != date {
		return false // Dates do not match
	}

	return true
}

func (cv *CustomValidator) ValidateTimeRange(fl validator.FieldLevel) bool {
	// Retrieve the current struct using reflection
	structValue := fl.Parent()

	// Access the StartTime and EndTime fields
	startTimeField := structValue.FieldByName("StartTime")
	endTimeField := structValue.FieldByName("EndTime")

	if !startTimeField.IsValid() || !endTimeField.IsValid() {
		return false // Fields do not exist or have incorrect types
	}

	if startTimeField.Kind() != reflect.Int64 || endTimeField.Kind() != reflect.Int64 {
		return false // Fields are not of type int64
	}

	startTime := startTimeField.Int()
	endTime := endTimeField.Int()

	// Check if the time values are zero
	if startTime == 0 || endTime == 0 {
		return false // Time values should not be zero
	}

	// Validate that EndTime is greater than StartTime and not greater than current time
	if endTime < startTime || time.Now().Unix() < endTime {
		return false
	}

	return true
}

// Custom validation function to check if StartTime is in the future
func (cv *CustomValidator) FutureValidation(fl validator.FieldLevel) bool {
	startTime := fl.Field().Int()
	currentTime := time.Now().Unix()
	return startTime > currentTime
}
