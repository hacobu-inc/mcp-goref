package refactor

import (
    "fmt"
    "go/types"
    "os"
    "path/filepath"
    "sort"

    "golang.org/x/tools/go/packages"
)
// isIdentByte reports whether b is a valid ASCII identifier character.
// Used to ensure replacements occur only at identifier boundaries.
func isIdentByte(b byte) bool {
    return b == '_' ||
        ('0' <= b && b <= '9') ||
        ('A' <= b && b <= 'Z') ||
        ('a' <= b && b <= 'z')
}

// Rename replaces oldName with newName across the module of fileArg.
func Rename(fileArg, oldName, newName string) error {
    moduleRoot, err := findModuleRoot(fileArg)
    if err != nil {
        return err
    }
    absFile, err := filepath.Abs(fileArg)
    if err != nil {
        return err
    }
    pkgs, err := loadPackages(moduleRoot)
    if err != nil {
        return err
    }
    var filePkg *packages.Package
    for _, p := range pkgs {
        for _, f := range p.GoFiles {
            if sameFile(absFile, f) {
                filePkg = p
                break
            }
        }
        if filePkg != nil {
            break
        }
    }
    if filePkg == nil {
        return fmt.Errorf("file %s not part of module packages", fileArg)
    }
    typeName, methodName := parseSymbolName(oldName)
    var matches []types.Object
    for id, obj := range filePkg.TypesInfo.Defs {
        if obj == nil {
            continue
        }
        pos := filePkg.Fset.Position(id.Pos())
        if !sameFile(absFile, pos.Filename) {
            continue
        }
        if typeName != "" {
            fn, ok := obj.(*types.Func)
            if !ok {
                continue
            }
            sig, ok := fn.Type().(*types.Signature)
            if !ok || sig.Recv() == nil {
                continue
            }
            recv := sig.Recv().Type()
            named, ok := recv.(*types.Named)
            if !ok {
                if ptr, ok2 := recv.(*types.Pointer); ok2 {
                    named, ok = ptr.Elem().(*types.Named)
                }
            }
            if ok && named.Obj().Name() == typeName && obj.Name() == methodName {
                matches = append(matches, obj)
            }
        } else if obj.Name() == methodName {
            matches = append(matches, obj)
        }
    }
    if len(matches) == 0 {
        return fmt.Errorf("symbol %s not found in file %s", oldName, fileArg)
    }
    if len(matches) > 1 {
        return fmt.Errorf("symbol %s is ambiguous (%d matches)", oldName, len(matches))
    }
    defObj := matches[0]
    if existing := filePkg.Types.Scope().Lookup(newName); existing != nil && existing != defObj {
        return fmt.Errorf("new symbol name %s conflicts with existing symbol", newName)
    }
    type occ struct{ file string; offset int }
    var occs []occ
    for id, obj := range filePkg.TypesInfo.Defs {
        if obj == defObj {
            pos := filePkg.Fset.Position(id.Pos())
            occs = append(occs, occ{file: pos.Filename, offset: pos.Offset})
        }
    }
    for _, p := range pkgs {
        for id, obj := range p.TypesInfo.Uses {
            if obj == defObj {
                pos := p.Fset.Position(id.Pos())
                occs = append(occs, occ{file: pos.Filename, offset: pos.Offset})
            }
        }
    }
    if len(occs) == 0 {
        return fmt.Errorf("no occurrences of symbol %s found", oldName)
    }
    // Group occurrences by file.
    files := map[string][]int{}
    for _, o := range occs {
        files[o.file] = append(files[o.file], o.offset)
    }
    var totalReplaced int
    // Track replacement counts per file.
    type fileReplacement struct{ file string; count int }
    var replacements []fileReplacement
    // Perform replacements in each file.
    for file, offs := range files {
        content, err := os.ReadFile(file)
        if err != nil {
            return fmt.Errorf("reading file %s: %v", file, err)
        }
        sort.Sort(sort.Reverse(sort.IntSlice(offs)))
        var actualCount int
        for _, off := range offs {
            // Ensure exact match of oldName at this offset.
            end := off + len(oldName)
            if end > len(content) || string(content[off:end]) != oldName {
                continue
            }
            // Check identifier boundaries: preceding character must not be identifier char.
            if off > 0 && isIdentByte(content[off-1]) {
                continue
            }
            // Following character must not be identifier char.
            if end < len(content) && isIdentByte(content[end]) {
                continue
            }
            // Replace oldName with newName.
            content = append(content[:off], append([]byte(newName), content[end:]...)...)
            actualCount++
        }
        if actualCount > 0 {
            if err := os.WriteFile(file, content, 0644); err != nil {
                return fmt.Errorf("writing file %s: %v", file, err)
            }
            replacements = append(replacements, fileReplacement{file, actualCount})
            totalReplaced += actualCount
        }
    }
    if totalReplaced == 0 {
        return fmt.Errorf("symbol %s not found or no valid occurrences", oldName)
    }
    // Print summary of replacements.
    fmt.Printf("Renamed '%s' -> '%s' in:\n", oldName, newName)
    for _, r := range replacements {
        rel, _ := filepath.Rel(moduleRoot, r.file)
        fmt.Printf("- %s (%d occurrences)\n", rel, r.count)
    }
    return nil
}