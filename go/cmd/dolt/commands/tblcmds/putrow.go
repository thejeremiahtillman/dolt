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

package tblcmds

import (
	"context"
	"strings"

	"github.com/fatih/color"

	"github.com/liquidata-inc/dolt/go/cmd/dolt/cli"
	"github.com/liquidata-inc/dolt/go/cmd/dolt/commands"
	"github.com/liquidata-inc/dolt/go/cmd/dolt/errhand"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/env"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/row"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/rowconv"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/schema"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/table/untyped"
	"github.com/liquidata-inc/dolt/go/libraries/utils/argparser"
	"github.com/liquidata-inc/dolt/go/store/types"
)

var putRowShortDesc = "Adds or updates a row in a table"
var putRowLongDesc = "dolt table put-row will put a row in a table.  If a row already exists with a matching primary key" +
	"it will be overwritten with the new value. All required fields for rows in this table must be supplied or the command" +
	"will fail.  example usage:\n" +
	"\n" +
	"  dolt table put-row \"field0:value0\" \"field1:value1\" ... \"fieldN:valueN\"\n"
var putRowSynopsis = []string{
	"<table> <field_name:field_value>...",
}

type putRowArgs struct {
	FieldNames []string
	KVPs       map[string]string
	TableName  string
}

func parsePutRowArgs(commandStr string, args []string) *putRowArgs {
	ap := argparser.NewArgParser()
	ap.ArgListHelp["table"] = "The table being inserted into"
	ap.ArgListHelp["field_name:field_value"] = "There should be a <field_name>:<field_value> pair for each field " +
		"that you want set on this row.  If all required fields are not set, then this command will fail."
	help, usage := cli.HelpAndUsagePrinters(commandStr, putRowShortDesc, putRowLongDesc, putRowSynopsis, ap)
	apr := cli.ParseArgs(ap, args, help)

	if apr.NArg() == 0 {
		usage()
		return nil
	}

	parsedArgs := apr.Args()
	tableName := parsedArgs[0]
	fieldNames, kvps, verr := parseKVPs(parsedArgs[1:])

	if verr != nil {
		cli.PrintErrln(verr.Error())
		return nil
	}

	return &putRowArgs{fieldNames, kvps, tableName}
}

func parseKVPs(args []string) ([]string, map[string]string, errhand.VerboseError) {
	fieldNames := make([]string, len(args))
	kvps := make(map[string]string, len(args))
	for i, arg := range args {
		colonIndex := strings.IndexByte(arg, ':')

		if colonIndex != -1 {
			key := strings.TrimSpace(arg[:colonIndex])
			value := arg[colonIndex+1:]

			if key != "" {
				kvps[key] = value
				fieldNames[i] = key
			} else {
				bdr := errhand.BuildDError(`"%s" is not a valid key value pair.`, strings.TrimSpace(arg))
				bdr.AddDetails("Key value pairs must be in the format key:value, where the length of key must be at least 1 character.  \"%s\" has a length of 0 characters", strings.TrimSpace(arg))
				return nil, nil, bdr.Build()
			}
		} else {
			bdr := errhand.BuildDError(`"%s" is not a valid key value pair.`, strings.TrimSpace(arg))
			bdr.AddDetails("Key value pairs must be in the format key:value.  \"%s\" has no key value separator ':'.  ", strings.TrimSpace(arg))
			bdr.AddDetails("To set a value to empty you may use \"key:\" but not just \"key\", however leaving this key off of the command line has the same effect.")
			return nil, nil, bdr.Build()
		}
	}

	return fieldNames, kvps, nil
}

func PutRow(commandStr string, args []string, dEnv *env.DoltEnv) int {
	prArgs := parsePutRowArgs(commandStr, args)

	if prArgs == nil {
		return 1
	}

	root, err := dEnv.WorkingRoot(context.Background())
	fmt := root.VRW().Format()

	if err != nil {
		cli.PrintErrln(color.RedString("Unable to get working value."))
		return 1
	}

	tbl, ok, err := root.GetTable(context.TODO(), prArgs.TableName)

	if err != nil {
		cli.PrintErrln(color.RedString("error: failed to read tables: " + err.Error()))
		return 1
	}

	if !ok {
		cli.PrintErrln(color.RedString("Unknown table %s", prArgs.TableName))
		return 1
	}

	sch, err := tbl.GetSchema(context.TODO())

	if err != nil {
		cli.PrintErrln(color.RedString("error: failed to read schema: " + err.Error()))
		return 1
	}

	row, verr := createRow(fmt, sch, prArgs)

	if verr == nil {
		m, err := tbl.GetRowData(context.TODO())

		if err != nil {
			verr = errhand.BuildDError("error: failed to get row data.").AddCause(err).Build()
		} else {
			me := m.Edit()
			updated, err := me.Set(row.NomsMapKey(sch), row.NomsMapValue(sch)).Map(context.TODO())

			if err != nil {
				verr = errhand.BuildDError("error: failed to modify table").AddCause(err).Build()
			} else {
				tbl, err = tbl.UpdateRows(context.Background(), updated)

				if err != nil {
					verr = errhand.BuildDError("error: failed to update rows").AddCause(err).Build()
				} else {
					root, err = root.PutTable(context.Background(), dEnv.DoltDB, prArgs.TableName, tbl)

					if err != nil {
						verr = errhand.BuildDError("error: failed to write table back to database").AddCause(err).Build()
					} else {
						verr = commands.UpdateWorkingWithVErr(dEnv, root)
					}
				}
			}
		}
	}

	if verr != nil {
		cli.PrintErrln(verr.Verbose())
		return 1
	}

	cli.Println(color.CyanString("Successfully put row."))
	return 0
}

func createRow(nbf *types.NomsBinFormat, sch schema.Schema, prArgs *putRowArgs) (row.Row, errhand.VerboseError) {
	var unknownFields []string
	untypedTaggedVals := make(row.TaggedValues)
	for k, v := range prArgs.KVPs {
		if col, ok := schema.ColFromName(sch, k); ok {
			untypedTaggedVals[col.Tag] = types.String(v)
		} else {
			unknownFields = append(unknownFields, k)
		}
	}

	if len(unknownFields) > 0 {
		bdr := errhand.BuildDError("Not all supplied keys are known in this table's schema.")
		bdr.AddDetails("The fields %v were supplied but are not known in this table.", unknownFields)
		return nil, bdr.Build()
	}

	untypedSch, err := untyped.UntypeSchema(sch)

	if err != nil {
		return nil, errhand.BuildDError("error: failed to get schemas").AddCause(err).Build()
	}

	mapping, err := rowconv.TagMapping(untypedSch, sch)

	if err != nil {
		return nil, errhand.BuildDError("Failed to infer mapping").AddCause(err).Build()
	}

	rconv, err := rowconv.NewRowConverter(mapping)

	if err != nil {
		return nil, errhand.BuildDError("failed to create row converter").AddCause(err).Build()
	}

	untypedRow, err := row.New(nbf, untypedSch, untypedTaggedVals)

	if err != nil {
		return nil, errhand.BuildDError("").AddCause(err).Build()
	}

	typedRow, err := rconv.Convert(untypedRow)

	if err != nil {
		return nil, errhand.BuildDError("inserted row does not match schema").AddCause(err).Build()
	}

	if col, _ := row.GetInvalidCol(typedRow, sch); col != nil {
		bdr := errhand.BuildDError("Missing required fields.")
		bdr.AddDetails("The value for the column %s is not valid", col.Name)
		return nil, bdr.Build()
	}

	return typedRow, nil
}
