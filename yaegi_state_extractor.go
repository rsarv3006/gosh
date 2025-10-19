//go:build darwin || linux

package main

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/traefik/yaegi/interp"
)

// CompletionItem represents a completion suggestion
type CompletionItem struct {
	Label         string
	Kind          string // "function", "variable", "type", "package", "constant"
	Detail        string
	Documentation string
}

// SymbolExtractor extracts symbols from Yaegi interpreter state
type SymbolExtractor struct {
	interp       *interp.Interpreter
	symbolCache  map[string][]CompletionItem
	cacheMutex   sync.RWMutex
	lastEvalHash string
}

// NewSymbolExtractor creates a new symbol extractor
func NewSymbolExtractor(interp *interp.Interpreter) *SymbolExtractor {
	return &SymbolExtractor{
		interp:      interp,
		symbolCache: make(map[string][]CompletionItem),
	}
}

// refreshIfNeeded refreshes the symbol cache if the interpreter state has changed
func (s *SymbolExtractor) refreshIfNeeded() {
	defer func() {
		if r := recover(); r != nil {
			// If yaegi panics during symbol extraction, just clear cache
			s.cacheMutex.Lock()
			s.symbolCache = make(map[string][]CompletionItem)
			s.cacheMutex.Unlock()
		}
	}()
	
	// For now, refresh on every call. In a more sophisticated implementation,
	// we could track changes more efficiently.
	s.extractSymbols()
}

// extractSymbols extracts all available symbols from the interpreter
func (s *SymbolExtractor) extractSymbols() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// Clear cache
	s.symbolCache = make(map[string][]CompletionItem)

	// Get all symbols from the interpreter with error handling
	defer func() {
		if r := recover(); r != nil {
			// If yaegi panics, just return empty cache
			s.symbolCache = make(map[string][]CompletionItem)
		}
	}()

	symbols := s.interp.Symbols("")

	for pkgName, pkgSymbols := range symbols {
		completions := []CompletionItem{}
		
		for symName, symValue := range pkgSymbols {
			if !symValue.IsValid() {
				continue
			}

			// Add recovery for each symbol too
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Skip problematic symbols
					}
				}()
				
				item := s.createCompletionItem(symName, symValue)
				completions = append(completions, item)
			}()
		}

		s.symbolCache[pkgName] = completions
	}
}

// createCompletionItem creates a CompletionItem from a reflect.Value
func (s *SymbolExtractor) createCompletionItem(name string, value reflect.Value) CompletionItem {
	item := CompletionItem{
		Label: name,
	}

	switch value.Kind() {
	case reflect.Func:
		item.Kind = "function"
		item.Detail = s.getFunctionSignature(value)
	case reflect.Struct, reflect.Ptr:
		if value.Type().String() == "*interface {}" {
			item.Kind = "variable"
		} else {
			item.Kind = "type"
			item.Detail = value.Type().String()
		}
	case reflect.Interface:
		if !value.IsNil() {
			item.Kind = "variable"
		}
	default:
		item.Kind = "variable"
		item.Detail = value.Type().String()
	}

	return item
}

// getFunctionSignature extracts a readable function signature
func (s *SymbolExtractor) getFunctionSignature(fn reflect.Value) string {
	if fn.Kind() != reflect.Func {
		return ""
	}

	fnType := fn.Type()
	
	// Build parameter list
	var params []string
	for i := 0; i < fnType.NumIn(); i++ {
		paramType := fnType.In(i)
		params = append(params, paramType.String())
	}

	// Build return type list
	var returns []string
	for i := 0; i < fnType.NumOut(); i++ {
		returnType := fnType.Out(i)
		returns = append(returns, returnType.String())
	}

	// Format signature
	paramStr := strings.Join(params, ", ")
	var returnStr string
	if len(returns) > 0 {
		returnStr = " " + strings.Join(returns, ", ")
	}

	return fmt.Sprintf("func(%s)%s", paramStr, returnStr)
}

// GetCompletionSuggestions returns completion suggestions for a given prefix
func (s *SymbolExtractor) GetCompletionSuggestions(partial string) []CompletionItem {
	var suggestions []CompletionItem

	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	for _, pkgCompletions := range s.symbolCache {
		for _, item := range pkgCompletions {
			if strings.HasPrefix(item.Label, partial) {
				suggestions = append(suggestions, item)
			}
		}
	}

	return suggestions
}

// GetFunctions returns matching function symbols
func (s *SymbolExtractor) GetFunctions(partial string) []CompletionItem {
	var functions []CompletionItem

	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	for _, pkgCompletions := range s.symbolCache {
		for _, item := range pkgCompletions {
			if item.Kind == "function" && strings.HasPrefix(item.Label, partial) {
				functions = append(functions, item)
			}
		}
	}

	return functions
}

// GetVariables returns matching variable symbols
func (s *SymbolExtractor) GetVariables(partial string) []CompletionItem {
	var variables []CompletionItem

	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	for _, pkgCompletions := range s.symbolCache {
		for _, item := range pkgCompletions {
			if item.Kind == "variable" && strings.HasPrefix(item.Label, partial) {
				variables = append(variables, item)
			}
		}
	}

	return variables
}

// GetTypes returns matching type symbols
func (s *SymbolExtractor) GetTypes(partial string) []CompletionItem {
	var types []CompletionItem

	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	for _, pkgCompletions := range s.symbolCache {
		for _, item := range pkgCompletions {
			if item.Kind == "type" && strings.HasPrefix(item.Label, partial) {
				types = append(types, item)
			}
		}
	}

	return types
}

// GetSelectorCompletions returns completions for selector expressions (e.g., "fmt.")
func (s *SymbolExtractor) GetSelectorCompletions(scope, partial string) []CompletionItem {
	var completions []CompletionItem

	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	// Look for package symbols that match the scope
	if pkgCompletions, exists := s.symbolCache[scope]; exists {
		for _, item := range pkgCompletions {
			if strings.HasPrefix(item.Label, partial) {
				completions = append(completions, item)
			}
		}
	}

	// Also check main scope for user-defined symbols
	if mainCompletions, exists := s.symbolCache["main"]; exists {
		for _, item := range mainCompletions {
			if strings.HasPrefix(item.Label, partial) {
				completions = append(completions, item)
			}
		}
	}

	return completions
}

// GetAllPackages returns all available package names
func (s *SymbolExtractor) GetAllPackages() []string {
	var packages []string

	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	for pkgName := range s.symbolCache {
		packages = append(packages, pkgName)
	}

	return packages
}

// GetPackageSymbols returns all symbols in a specific package
func (s *SymbolExtractor) GetPackageSymbols(pkgName string) []CompletionItem {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	return s.symbolCache[pkgName]
}
