package middleware

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type ValidationMiddleware struct {
	logger *zap.Logger
}

func NewValidationMiddleware(logger *zap.Logger) *ValidationMiddleware {
	return &ValidationMiddleware{
		logger: logger,
	}
}

// ValidateUUID valida que un parámetro sea un UUID válido
func (vm *ValidationMiddleware) ValidateUUID(paramName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			param := r.URL.Query().Get(paramName)
			if param != "" {
				if !isValidUUID(param) {
					vm.logger.Warn("UUID inválido", zap.String("param", paramName), zap.String("value", param))
					http.Error(w, "UUID inválido", http.StatusBadRequest)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateRequiredFields valida que los campos requeridos estén presentes en el body JSON
func (vm *ValidationMiddleware) ValidateRequiredFields(requiredFields ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
				var body map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					vm.logger.Warn("Error decodificando JSON", zap.Error(err))
					http.Error(w, "JSON inválido", http.StatusBadRequest)
					return
				}

				for _, field := range requiredFields {
					if _, exists := body[field]; !exists {
						vm.logger.Warn("Campo requerido faltante", zap.String("field", field))
						http.Error(w, "Campo requerido faltante: "+field, http.StatusBadRequest)
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateStringLength valida la longitud de un string
func (vm *ValidationMiddleware) ValidateStringLength(fieldName string, minLength, maxLength int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
				var body map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					vm.logger.Warn("Error decodificando JSON", zap.Error(err))
					http.Error(w, "JSON inválido", http.StatusBadRequest)
					return
				}

				if value, exists := body[fieldName]; exists {
					if str, ok := value.(string); ok {
						if len(str) < minLength || len(str) > maxLength {
							vm.logger.Warn("Longitud de string inválida",
								zap.String("field", fieldName),
								zap.Int("length", len(str)),
								zap.Int("min", minLength),
								zap.Int("max", maxLength),
							)
							http.Error(w, "Longitud inválida para el campo: "+fieldName, http.StatusBadRequest)
							return
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateNumericRange valida que un número esté en un rango específico
func (vm *ValidationMiddleware) ValidateNumericRange(fieldName string, minValue, maxValue float64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
				var body map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					vm.logger.Warn("Error decodificando JSON", zap.Error(err))
					http.Error(w, "JSON inválido", http.StatusBadRequest)
					return
				}

				if value, exists := body[fieldName]; exists {
					switch v := value.(type) {
					case float64:
						if v < minValue || v > maxValue {
							vm.logger.Warn("Valor numérico fuera de rango",
								zap.String("field", fieldName),
								zap.Float64("value", v),
								zap.Float64("min", minValue),
								zap.Float64("max", maxValue),
							)
							http.Error(w, "Valor fuera de rango para el campo: "+fieldName, http.StatusBadRequest)
							return
						}
					case int:
						if float64(v) < minValue || float64(v) > maxValue {
							vm.logger.Warn("Valor numérico fuera de rango",
								zap.String("field", fieldName),
								zap.Int("value", v),
								zap.Float64("min", minValue),
								zap.Float64("max", maxValue),
							)
							http.Error(w, "Valor fuera de rango para el campo: "+fieldName, http.StatusBadRequest)
							return
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateEmail valida el formato de email
func (vm *ValidationMiddleware) ValidateEmail(fieldName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
				var body map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					vm.logger.Warn("Error decodificando JSON", zap.Error(err))
					http.Error(w, "JSON inválido", http.StatusBadRequest)
					return
				}

				if value, exists := body[fieldName]; exists {
					if email, ok := value.(string); ok {
						if !isValidEmail(email) {
							vm.logger.Warn("Email inválido", zap.String("field", fieldName), zap.String("email", email))
							http.Error(w, "Email inválido", http.StatusBadRequest)
							return
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ValidatePositiveInteger valida que un valor sea un entero positivo
func (vm *ValidationMiddleware) ValidatePositiveInteger(fieldName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
				var body map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					vm.logger.Warn("Error decodificando JSON", zap.Error(err))
					http.Error(w, "JSON inválido", http.StatusBadRequest)
					return
				}

				if value, exists := body[fieldName]; exists {
					switch v := value.(type) {
					case float64:
						if v <= 0 || v != float64(int(v)) {
							vm.logger.Warn("Valor no es un entero positivo",
								zap.String("field", fieldName),
								zap.Float64("value", v),
							)
							http.Error(w, "El campo debe ser un entero positivo: "+fieldName, http.StatusBadRequest)
							return
						}
					case int:
						if v <= 0 {
							vm.logger.Warn("Valor no es un entero positivo",
								zap.String("field", fieldName),
								zap.Int("value", v),
							)
							http.Error(w, "El campo debe ser un entero positivo: "+fieldName, http.StatusBadRequest)
							return
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateQueryParam valida un parámetro de query
func (vm *ValidationMiddleware) ValidateQueryParam(paramName string, validator func(string) bool, errorMessage string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			param := r.URL.Query().Get(paramName)
			if param != "" && !validator(param) {
				vm.logger.Warn("Parámetro de query inválido",
					zap.String("param", paramName),
					zap.String("value", param),
				)
				http.Error(w, errorMessage, http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ValidatePagination valida parámetros de paginación
func (vm *ValidationMiddleware) ValidatePagination() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pageStr := r.URL.Query().Get("page")
			limitStr := r.URL.Query().Get("limit")

			if pageStr != "" {
				if page, err := strconv.Atoi(pageStr); err != nil || page < 1 {
					vm.logger.Warn("Página inválida", zap.String("page", pageStr))
					http.Error(w, "Página debe ser un número mayor a 0", http.StatusBadRequest)
					return
				}
			}

			if limitStr != "" {
				if limit, err := strconv.Atoi(limitStr); err != nil || limit < 1 || limit > 100 {
					vm.logger.Warn("Límite inválido", zap.String("limit", limitStr))
					http.Error(w, "Límite debe ser un número entre 1 y 100", http.StatusBadRequest)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Funciones auxiliares

func isValidUUID(uuid string) bool {
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(pattern, strings.ToLower(uuid))
	return matched
}

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}
