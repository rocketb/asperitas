package validate

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// validate holds the settings and caches for validating request struct values.
var validate *validator.Validate

var translator ut.Translator

func init() {

	// Instantiate a validator
	validate = validator.New()

	// Create a translator for english so the error messages are
	// humman-readable.
	translator, _ = ut.New(en.New(), en.New()).GetTranslator("en")

	// Register the english error messages for use.
	en_translations.RegisterDefaultTranslations(validate, translator)

	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// Check validate the provided model against it's declared tags.
func Check(val any) error {
	if err := validate.Struct(val); err != nil {

		// Use a type assertion to get the ral error value.
		verrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}

		var fields FieldErrors
		for _, e := range verrors {
			field := FieldError{
				Field: e.Field(),
				Error: e.Translate(translator),
			}
			fields = append(fields, field)
		}
		return fields
	}

	return nil
}
