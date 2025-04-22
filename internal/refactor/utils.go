package refactor

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "golang.org/x/tools/go/packages"
)

// parseSymbolName splits a symbol string into type and method (if present).
func parseSymbolName(symbolName string) (string, string) {
    parts := strings.SplitN(symbolName, ".", 2)
    if len(parts) == 2 {
        return parts[0], parts[1]
    }
    return "", symbolName
}

// findModuleRoot locates the module root directory by searching for go.mod.
func findModuleRoot(filePath string) (string, error) {
    abs, err := filepath.Abs(filePath)
    if err != nil {
        return "", err
    }
    dir := filepath.Dir(abs)
    for {
        if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
            return dir, nil
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            break
        }
        dir = parent
    }
    return "", fmt.Errorf("go.mod not found, cannot determine module root")
}

// loadPackages loads all packages under the module root with type info.
func loadPackages(moduleRoot string) ([]*packages.Package, error) {
   cfg := &packages.Config{
       Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
           packages.NeedTypes | packages.NeedTypesInfo,
       Dir:   moduleRoot,
       Tests: true,
   }
    pkgs, err := packages.Load(cfg, "./...")
    if err != nil {
        return nil, err
    }
    for _, p := range pkgs {
        if len(p.Errors) > 0 {
            return nil, fmt.Errorf("errors loading package %s: %v", p.ID, p.Errors)
        }
    }
    return pkgs, nil
}

// sameFile reports whether paths refer to the same file.
func sameFile(a, b string) bool {
    aa, err1 := filepath.Abs(a)
    bb, err2 := filepath.Abs(b)
    if err1 == nil && err2 == nil {
        return filepath.Clean(aa) == filepath.Clean(bb)
    }
    return a == b
}