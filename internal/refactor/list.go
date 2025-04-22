package refactor

import (
    "fmt"
    "go/types"
    "path/filepath"
    "sort"

    "golang.org/x/tools/go/packages"
)

// ListRefs lists references to the given symbol in the module of fileArg.
func ListRefs(fileArg, symbolName string) error {
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
    typeName, methodName := parseSymbolName(symbolName)
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
        return fmt.Errorf("symbol %s not found in file %s", symbolName, fileArg)
    }
    if len(matches) > 1 {
        return fmt.Errorf("symbol %s is ambiguous (%d matches)", symbolName, len(matches))
    }
    defObj := matches[0]
    fset := filePkg.Fset
    type posInfo struct{ file string; line, col int }
    var results []posInfo
    for _, p := range pkgs {
        for id, obj := range p.TypesInfo.Uses {
            if obj == defObj {
                pos := fset.Position(id.Pos())
                rel, _ := filepath.Rel(moduleRoot, pos.Filename)
                results = append(results, posInfo{file: rel, line: pos.Line, col: pos.Column})
            }
        }
    }
    sort.Slice(results, func(i, j int) bool {
        if results[i].file != results[j].file {
            return results[i].file < results[j].file
        }
        if results[i].line != results[j].line {
            return results[i].line < results[j].line
        }
        return results[i].col < results[j].col
    })
    for _, r := range results {
        fmt.Printf("%s:%d:%d\n", r.file, r.line, r.col)
    }
    return nil
}