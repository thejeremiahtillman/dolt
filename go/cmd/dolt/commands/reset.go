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

package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/liquidata-inc/dolt/go/cmd/dolt/cli"
	"github.com/liquidata-inc/dolt/go/cmd/dolt/errhand"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/doltdb"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/env"
	"github.com/liquidata-inc/dolt/go/libraries/doltcore/env/actions"
	"github.com/liquidata-inc/dolt/go/libraries/utils/argparser"
)

const (
	SoftResetParam = "soft"
	HardResetParam = "hard"
)

var resetShortDesc = "Resets staged tables to their HEAD state"
var resetLongDesc = `Sets the state of a table in the staging area to be that table's value at HEAD

dolt reset <tables>...
	This form resets the values for all staged <tables> to their values at HEAD. (It does not affect the working tree or
	the current branch.)

	This means that </b>dolt reset <tables></b> is the opposite of <b>dolt add <tables></b>.

	After running <b>dolt reset <tables></b> to update the staged tables, you can use <b>dolt checkout</b> to check the
	contents out of the staged tables to the working tables.

dolt reset .
	This form resets <b>all</b> staged tables to their values at HEAD. It is the opposite of <b>dolt add .</b>`

var resetSynopsis = []string{
	"<tables>...",
	"[--hard | --soft]",
}

func Reset(commandStr string, args []string, dEnv *env.DoltEnv) int {
	ctx := context.TODO()

	ap := argparser.NewArgParser()
	ap.SupportsFlag(HardResetParam, "", "Resets the working tables and staged tables. Any changes to tracked tables in the working tree since <commit> are discarded.")
	ap.SupportsFlag(SoftResetParam, "", "Does not touch the working tables, but removes all tables staged to be committed.")
	help, usage := cli.HelpAndUsagePrinters(commandStr, resetShortDesc, resetLongDesc, resetSynopsis, ap)
	apr := cli.ParseArgs(ap, args, help)

	workingRoot, stagedRoot, headRoot, verr := getAllRoots(dEnv)

	if verr == nil {
		if apr.ContainsAll(HardResetParam, SoftResetParam) {
			verr = errhand.BuildDError("error: --%s and --%s are mutually exclusive options.", HardResetParam, SoftResetParam).Build()
		} else if apr.Contains(HardResetParam) {
			verr = resetHard(ctx, dEnv, apr, workingRoot, headRoot)
		} else {
			verr = resetSoft(dEnv, apr, stagedRoot, headRoot)
		}
	}

	return HandleVErrAndExitCode(verr, usage)
}

func resetHard(ctx context.Context, dEnv *env.DoltEnv, apr *argparser.ArgParseResults, workingRoot, headRoot *doltdb.RootValue) errhand.VerboseError {
	if apr.NArg() != 0 {
		return errhand.BuildDError("--%s does not support additional params", HardResetParam).SetPrintUsage().Build()
	}

	// need to save the state of files that aren't tracked
	untrackedTables := make(map[string]*doltdb.Table)
	wTblNames, err := workingRoot.GetTableNames(ctx)

	if err != nil {
		return errhand.BuildDError("error: failed to read tables from the working set").AddCause(err).Build()
	}

	for _, tblName := range wTblNames {
		untrackedTables[tblName], _, err = workingRoot.GetTable(ctx, tblName)

		if err != nil {
			return errhand.BuildDError("error: failed to read '%s' from the working set", tblName).AddCause(err).Build()
		}
	}

	headTblNames, err := headRoot.GetTableNames(ctx)

	if err != nil {
		return errhand.BuildDError("error: failed to read tables from head").AddCause(err).Build()
	}

	for _, tblName := range headTblNames {
		delete(untrackedTables, tblName)
	}

	newWkRoot := headRoot
	for tblName, tbl := range untrackedTables {
		newWkRoot, err = newWkRoot.PutTable(ctx, dEnv.DoltDB, tblName, tbl)

		if err != nil {
			return errhand.BuildDError("error: failed to write table back to database").Build()
		}
	}

	// TODO: update working and staged in one repo_state write.
	err = dEnv.UpdateWorkingRoot(ctx, newWkRoot)

	if err != nil {
		return errhand.BuildDError("error: failed to update the working tables.").AddCause(err).Build()
	}

	_, err = dEnv.UpdateStagedRoot(ctx, headRoot)

	if err != nil {
		return errhand.BuildDError("error: failed to update the staged tables.").AddCause(err).Build()
	}

	return nil
}

func resetSoft(dEnv *env.DoltEnv, apr *argparser.ArgParseResults, stagedRoot, headRoot *doltdb.RootValue) errhand.VerboseError {
	tbls := apr.Args()

	if len(tbls) == 0 || (len(tbls) == 1 && tbls[0] == ".") {
		var err error
		tbls, err = actions.AllTables(context.TODO(), stagedRoot, headRoot)

		if err != nil {
			return errhand.BuildDError("error: failed to get all tables").AddCause(err).Build()
		}
	}

	verr := ValidateTablesWithVErr(tbls, stagedRoot, headRoot)

	if verr != nil {
		return verr
	}

	stagedRoot, verr = resetStaged(dEnv, tbls, stagedRoot, headRoot)

	if verr != nil {
		return verr
	}

	printNotStaged(dEnv, stagedRoot)
	return nil
}

func printNotStaged(dEnv *env.DoltEnv, staged *doltdb.RootValue) {
	// Printing here is best effort.  Fail silently
	working, err := dEnv.WorkingRoot(context.Background())

	if err != nil {
		return
	}

	notStaged, err := actions.NewTableDiffs(context.TODO(), working, staged)

	if err != nil {
		return
	}

	if notStaged.NumRemoved+notStaged.NumModified > 0 {
		cli.Println("Unstaged changes after reset:")

		lines := make([]string, 0, notStaged.Len())
		for _, tblName := range notStaged.Tables {
			tdt := notStaged.TableToType[tblName]

			if tdt != actions.AddedTable {
				lines = append(lines, fmt.Sprintf("%s\t%s", tblDiffTypeToShortLabel[tdt], tblName))
			}
		}

		cli.Println(strings.Join(lines, "\n"))
	}
}

func resetStaged(dEnv *env.DoltEnv, tbls []string, staged, head *doltdb.RootValue) (*doltdb.RootValue, errhand.VerboseError) {
	updatedRoot, err := staged.UpdateTablesFromOther(context.TODO(), tbls, head)

	if err != nil {
		return nil, errhand.BuildDError("error: failed to update tables").AddCause(err).Build()
	}

	return updatedRoot, UpdateStagedWithVErr(dEnv, updatedRoot)
}

func getAllRoots(dEnv *env.DoltEnv) (*doltdb.RootValue, *doltdb.RootValue, *doltdb.RootValue, errhand.VerboseError) {
	workingRoot, err := dEnv.WorkingRoot(context.Background())

	if err != nil {
		return nil, nil, nil, errhand.BuildDError("Unable to get staged.").AddCause(err).Build()
	}

	stagedRoot, err := dEnv.StagedRoot(context.Background())

	if err != nil {
		return nil, nil, nil, errhand.BuildDError("Unable to get staged.").AddCause(err).Build()
	}

	headRoot, err := dEnv.HeadRoot(context.TODO())

	if err != nil {
		return nil, nil, nil, errhand.BuildDError("Unable to get at HEAD.").AddCause(err).Build()
	}

	return workingRoot, stagedRoot, headRoot, nil
}
