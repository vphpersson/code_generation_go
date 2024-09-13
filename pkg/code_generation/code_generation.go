package code_generation

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var defaultPackagePath = "main"

type ImportSet map[string]bool

func (importSet ImportSet) Generate() string {
	if len(importSet) == 0 {
		return ""
	}

	var importEntries []string
	for importPath := range importSet {
		importEntries = append(importEntries, fmt.Sprintf("\t\"%s\"", importPath))
	}
	return fmt.Sprintf("import (\n%s\n)", strings.Join(importEntries, "\n"))
}

func GenerateLiteral(value reflect.Value, importSet ImportSet) (string, ImportSet, error) {
	if !value.IsValid() {
		return "", nil, errors.New("invalid value provided")
	}

	switch value.Kind() {
	case reflect.String:
		return strconv.Quote(value.String()), importSet, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprint(value.Int()), importSet, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprint(value.Uint()), importSet, nil
	case reflect.Struct:
		pkgPath := value.Type().PkgPath()
		// TODO: Is this check reasonable?
		if pkgPath != "" && pkgPath != defaultPackagePath {
			if importSet == nil {
				importSet = make(map[string]bool)
			}
			importSet[pkgPath] = true
		}
		return processStruct(value, importSet)
	case reflect.Slice, reflect.Array:
		return processSlice(value, importSet)
	case reflect.Map:
		return processMap(value, importSet)
	case reflect.Ptr:
		return processPointer(value, importSet)
	case reflect.Func:
		if !value.IsNil() {
			return "", nil, errors.New("function fields are not supported")
		}
		return "nil", importSet, nil
	default:
		return fmt.Sprintf("%v", value.Interface()), importSet, nil
	}
}

func processStruct(value reflect.Value, importSet ImportSet) (string, ImportSet, error) {
	typ := value.Type()

	name := typ.Name()
	pkgPath := typ.PkgPath()
	if pkgPath != defaultPackagePath {
		pkgPathSlice := strings.Split(pkgPath, "/")
		name = pkgPathSlice[len(pkgPathSlice)-1] + "." + name
	}

	result := fmt.Sprintf("%s{\n", name)
	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := value.Field(i)

		var fieldLiteral string
		var err error

		fieldLiteral, importSet, err = GenerateLiteral(fieldValue, importSet)
		if err != nil {
			return "", nil, err
		}

		result += fmt.Sprintf("    %s: %s,\n", field.Name, fieldLiteral)
	}
	result += "}"
	return result, importSet, nil
}

func processSlice(value reflect.Value, importSet ImportSet) (string, ImportSet, error) {
	elements := make([]string, value.Len())
	for i := 0; i < value.Len(); i++ {
		var elemLiteral string
		var err error

		elemLiteral, importSet, err = GenerateLiteral(value.Index(i), importSet)
		if err != nil {
			return "", nil, err
		}

		elements[i] = elemLiteral
	}
	return fmt.Sprintf("[]%s{%s}", value.Type().Elem(), strings.Join(elements, ", ")), importSet, nil
}

func processMap(value reflect.Value, importSet ImportSet) (string, ImportSet, error) {
	mapKeys := value.MapKeys()
	elements := make([]string, len(mapKeys))
	for i, key := range mapKeys {
		var keyLiteral string
		var err error
		keyLiteral, importSet, err = GenerateLiteral(key, importSet)
		if err != nil {
			return "", nil, err
		}
		mapValue := value.MapIndex(key)
		var valueLiteral string
		valueLiteral, importSet, err = GenerateLiteral(mapValue, importSet)
		if err != nil {
			return "", nil, err
		}
		elements[i] = fmt.Sprintf("%s: %s", keyLiteral, valueLiteral)
	}
	return fmt.Sprintf("map[%s]%s{%s}", value.Type().Key(), value.Type().Elem(), strings.Join(elements, ", ")), importSet, nil
}

func processPointer(value reflect.Value, importSet ImportSet) (string, ImportSet, error) {
	if value.IsNil() {
		return "nil", importSet, nil
	}

	var literal string
	var err error
	literal, importSet, err = GenerateLiteral(value.Elem(), importSet)
	if err != nil {
		return "", nil, err
	}

	return "&" + literal, importSet, nil
}

func init() {
	if goPackage := os.Getenv("GOPACKAGE"); goPackage != "" {
		defaultPackagePath = goPackage
	}
}
