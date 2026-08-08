// checksql extracts every sqlc-generated query constant from
// internal/db/*.sql.go and PREPAREs it against a real, migrated Postgres
// database. sqlc generate validates table/column names but not operators
// (a broken `||` written as `| |` still generates cleanly), and PREPARE
// catches that class of error without needing to actually execute the
// query or have a caller/test for it yet — see api/TODO.md Tier 4.
package main

import (
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type query struct {
	name string
	sql  string
	file string
	line int
}

func main() {
	dir := flag.String("dir", "internal/db", "directory of sqlc-generated *.sql.go files")
	dsn := flag.String("dsn", os.Getenv("DATABASE_URL"), "Postgres connection string, schema already migrated")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("checksql: -dsn (or DATABASE_URL) is required")
	}

	queries, err := extractQueries(*dir)
	if err != nil {
		log.Fatalf("checksql: %v", err)
	}
	if len(queries) == 0 {
		log.Fatalf("checksql: no query constants found under %s", *dir)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("checksql: connect: %v", err)
	}
	defer pool.Close()

	failures := 0
	for i, q := range queries {
		stmtName := fmt.Sprintf("checksql_%d", i)
		if _, err := pool.Exec(ctx, fmt.Sprintf("PREPARE %s AS %s", stmtName, q.sql)); err != nil {
			failures++
			fmt.Printf("FAIL %s (%s:%d)\n  %v\n\n", q.name, q.file, q.line, err)
			continue
		}
		if _, err := pool.Exec(ctx, "DEALLOCATE "+stmtName); err != nil {
			log.Fatalf("checksql: deallocate %s: %v", stmtName, err)
		}
	}

	fmt.Printf("checksql: %d query constant(s) checked, %d failed\n", len(queries), failures)
	if failures > 0 {
		os.Exit(1)
	}
}

// extractQueries walks every *.sql.go file in dir and pulls out each
// top-level "const <name> = `<sql>`" declaration — sqlc's own
// generation convention, so every such const in these files is a query,
// never anything else (enum consts live in models.go, out of scope here).
func extractQueries(dir string) ([]query, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.sql.go"))
	if err != nil {
		return nil, err
	}

	var out []query
	fset := token.NewFileSet()
	for _, path := range paths {
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.CONST {
				continue
			}
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok || valueSpec.Type != nil || len(valueSpec.Names) != 1 || len(valueSpec.Values) != 1 {
					continue
				}
				lit, ok := valueSpec.Values[0].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				sql, err := strconv.Unquote(lit.Value)
				if err != nil {
					return nil, fmt.Errorf("%s: unquote %s: %w", path, valueSpec.Names[0].Name, err)
				}
				pos := fset.Position(lit.Pos())
				out = append(out, query{
					name: valueSpec.Names[0].Name,
					sql:  sql,
					file: path,
					line: pos.Line,
				})
			}
		}
	}
	return out, nil
}
