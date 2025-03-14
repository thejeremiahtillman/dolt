// Copyright 2019 Liquidata, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sqlexport

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/liquidata-inc/dolt/go/libraries/doltcore"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/row"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/schema"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/sql"
	"github.com/liquidata-inc/dolt/go/libraries/utils/filesys"
	"github.com/liquidata-inc/dolt/go/libraries/utils/iohelp"
	"github.com/liquidata-inc/dolt/go/store/types"
)

const doubleQuot = "\""

// SqlExportWriter is a TableWriter that writes SQL drop, create and insert statements to re-create a dolt table in a
// SQL database.
type SqlExportWriter struct {
	tableName       string
	sch             schema.Schema
	wr              io.WriteCloser
	writtenFirstRow bool
}

// OpenSQLExportWriter returns a new SqlWriter for the table given writing to a file with the path given.
func OpenSQLExportWriter(path string, tableName string, fs filesys.WritableFS, sch schema.Schema) (*SqlExportWriter, error) {
	err := fs.MkDirs(filepath.Dir(path))
	if err != nil {
		return nil, err
	}

	wr, err := fs.OpenForWrite(path)
	if err != nil {
		return nil, err
	}

	return &SqlExportWriter{tableName: tableName, sch: sch, wr: wr}, nil
}

// Returns the schema of this TableWriter.
func (w *SqlExportWriter) GetSchema() schema.Schema {
	return w.sch
}

// WriteRow will write a row to a table
func (w *SqlExportWriter) WriteRow(ctx context.Context, r row.Row) error {
	if err := w.maybeWriteDropCreate(); err != nil {
		return err
	}

	stmt, err := w.insertStatementForRow(r)

	if err != nil {
		return err
	}

	return iohelp.WriteLine(w.wr, stmt)
}

func (w *SqlExportWriter) maybeWriteDropCreate() error {
	if !w.writtenFirstRow {
		if err := iohelp.WriteLine(w.wr, w.dropCreateStatement()); err != nil {
			return err
		}
		w.writtenFirstRow = true
	}
	return nil
}

// Close should flush all writes, release resources being held
func (w *SqlExportWriter) Close(ctx context.Context) error {
	// exporting an empty table will not get any WriteRow calls, so write the drop / create here
	if err := w.maybeWriteDropCreate(); err != nil {
		return err
	}

	if w.wr != nil {
		return w.wr.Close()
	}
	return nil
}

func (w *SqlExportWriter) insertStatementForRow(r row.Row) (string, error) {
	var b strings.Builder
	b.WriteString("INSERT INTO ")
	b.WriteString(sql.QuoteIdentifier(w.tableName))
	b.WriteString(" ")

	b.WriteString("(")
	var seenOne bool
	err := w.sch.GetAllCols().Iter(func(tag uint64, col schema.Column) (stop bool, err error) {
		if seenOne {
			b.WriteRune(',')
		}
		b.WriteString(sql.QuoteIdentifier(col.Name))
		seenOne = true
		return false, nil
	})

	if err != nil {
		return "", err
	}

	b.WriteString(")")

	b.WriteString(" VALUES (")
	seenOne = false
	_, err = r.IterSchema(w.sch, func(tag uint64, val types.Value) (stop bool, err error) {
		if seenOne {
			b.WriteRune(',')
		}
		b.WriteString(w.sqlString(val))
		seenOne = true
		return false, nil
	})

	if err != nil {
		return "", err
	}

	b.WriteString(");")

	return b.String(), nil
}

func (w *SqlExportWriter) dropCreateStatement() string {
	var b strings.Builder
	b.WriteString("DROP TABLE IF EXISTS ")
	b.WriteString(sql.QuoteIdentifier(w.tableName))
	b.WriteString(";\n")
	b.WriteString(sql.SchemaAsCreateStmt(w.tableName, w.sch))

	return b.String()
}

func (w *SqlExportWriter) sqlString(value types.Value) string {
	if types.IsNull(value) {
		return "NULL"
	}

	switch value.Kind() {
	case types.BoolKind:
		if value.(types.Bool) {
			return "TRUE"
		} else {
			return "FALSE"
		}
	case types.UUIDKind:
		convFn := doltcore.GetConvFunc(value.Kind(), types.StringKind)
		str, _ := convFn(value)
		return doubleQuot + string(str.(types.String)) + doubleQuot
	case types.StringKind:
		s := string(value.(types.String))
		s = strings.ReplaceAll(s, doubleQuot, "\\\"")
		return doubleQuot + s + doubleQuot
	default:
		convFn := doltcore.GetConvFunc(value.Kind(), types.StringKind)
		str, _ := convFn(value)
		return string(str.(types.String))
	}
}
