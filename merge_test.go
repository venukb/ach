// Licensed to The Moov Authors under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. The Moov Authors licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package ach

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/igrmk/treemap/v2"
	"github.com/stretchr/testify/require"
)

func lineCount(f *File) int {
	lines := 2 // FileHeader, FileControl
	for i := range f.Batches {
		lines += 2 // BatchHeader, BatchControl
		entries := f.Batches[i].GetEntries()
		for j := range entries {
			lines++
			if entries[j].Addenda02 != nil {
				lines++
			}
			lines += len(entries[j].Addenda05)
			if entries[j].Addenda98 != nil {
				lines++
			}
			if entries[j].Addenda98Refused != nil {
				lines++
			}
			if entries[j].Addenda99 != nil {
				lines++
			}
			if entries[j].Addenda99Dishonored != nil {
				lines++
			}
			if entries[j].Addenda99Contested != nil {
				lines++
			}
		}
	}
	for i := range f.IATBatches {
		lines += 2 // IATBatchHeader, BatchControl
		for j := range f.IATBatches[i].Entries {
			lines++
			if f.IATBatches[i].Entries[j].Addenda10 != nil {
				lines++
			}
			if f.IATBatches[i].Entries[j].Addenda11 != nil {
				lines++
			}
			if f.IATBatches[i].Entries[j].Addenda12 != nil {
				lines++
			}
			if f.IATBatches[i].Entries[j].Addenda13 != nil {
				lines++
			}
			if f.IATBatches[i].Entries[j].Addenda14 != nil {
				lines++
			}
			if f.IATBatches[i].Entries[j].Addenda15 != nil {
				lines++
			}
			if f.IATBatches[i].Entries[j].Addenda16 != nil {
				lines++
			}

			lines += len(f.IATBatches[i].Entries[j].Addenda17)
			lines += len(f.IATBatches[i].Entries[j].Addenda18)

			if f.IATBatches[i].Entries[j].Addenda98 != nil {
				lines++
			}
			if f.IATBatches[i].Entries[j].Addenda99 != nil {
				lines++
			}
		}
	}
	return lines
}

func filesAreEqual(f1, f2 *File) error {
	// File Header
	if f1.Header.ImmediateOrigin != f2.Header.ImmediateOrigin {
		return fmt.Errorf("f1.Header.ImmediateOrigin=%s vs f2.Header.ImmediateOrigin=%s", f1.Header.ImmediateOrigin, f2.Header.ImmediateOrigin)
	}
	if f1.Header.ImmediateDestination != f2.Header.ImmediateDestination {
		return fmt.Errorf("f1.Header.ImmediateDestination=%s vs f2.Header.ImmediateDestination=%s", f1.Header.ImmediateDestination, f2.Header.ImmediateDestination)
	}

	// Batches
	if len(f1.Batches) != len(f2.Batches) {
		return fmt.Errorf("len(f1.Batches)=%d vs len(f2.Batches)=%d", len(f1.Batches), len(f2.Batches))
	}
	for i := range f1.Batches {
		for j := range f2.Batches {
			if f1.Batches[i].Equal(f2.Batches[j]) {
				goto next
			}
		}
		return fmt.Errorf("unable to find batch in f2: %v", f1.Batches[i])
	next:
		// check the next batch
	}

	// IATBatches
	if len(f1.IATBatches) != len(f2.IATBatches) {
		return fmt.Errorf("len(f1.IATBatches)=%d vs len(f2.IATBatches)=%d", len(f1.IATBatches), len(f2.IATBatches))
	}

	// File Control
	if f1.Control.EntryAddendaCount != f2.Control.EntryAddendaCount {
		return fmt.Errorf("f1.Control.EntryAddendaCount=%d vs f2.Control.EntryAddendaCount=%d", f1.Control.EntryAddendaCount, f2.Control.EntryAddendaCount)
	}
	if f1.Control.TotalDebitEntryDollarAmountInFile != f2.Control.TotalDebitEntryDollarAmountInFile {
		return fmt.Errorf("f1.Control.TotalDebitEntryDollarAmountInFile=%d vs f2.Control.TotalDebitEntryDollarAmountInFile=%d", f1.Control.TotalDebitEntryDollarAmountInFile, f2.Control.TotalDebitEntryDollarAmountInFile)
	}
	if f1.Control.TotalCreditEntryDollarAmountInFile != f2.Control.TotalCreditEntryDollarAmountInFile {
		return fmt.Errorf("f1.Control.TotalCreditEntryDollarAmountInFile=%d vs f2.Control.TotalCreditEntryDollarAmountInFile=%d", f1.Control.TotalCreditEntryDollarAmountInFile, f2.Control.TotalCreditEntryDollarAmountInFile)
	}

	return nil
}

func TestMergeFiles__filesAreEqual(t *testing.T) {
	file, err := readACHFilepath(filepath.Join("test", "testdata", "ppd-debit.ach"))
	require.NoError(t, err)

	// compare a file against itself
	err = filesAreEqual(file, file)
	require.NoError(t, err)

	// break the equality
	f2 := *file
	f2.Header.ImmediateOrigin = "12"
	err = filesAreEqual(file, &f2)
	require.Error(t, err)
}

func TestMergeFiles__identity(t *testing.T) {
	file, err := readACHFilepath(filepath.Join("test", "testdata", "ppd-debit.ach"))
	require.NoError(t, err)

	out, err := MergeFiles([]*File{file})
	require.NoError(t, err)
	require.Len(t, out, 1)

	err = filesAreEqual(file, out[0])
	require.NoError(t, err)

	for _, f := range out {
		require.NoError(t, f.Validate())
	}

	t.Run("multiple", func(t *testing.T) {
		out, err := MergeFiles([]*File{file, file, file, file})
		require.NoError(t, err)
		require.Len(t, out, 1)

		for i := range out[0].Batches {
			entries := out[0].Batches[i].GetEntries()
			require.Equal(t, 1, len(entries))
		}
		for i := range out {
			require.NoError(t, out[i].Create())
			require.NoError(t, out[i].Validate())
		}
	})
}

func TestMergeFiles__together(t *testing.T) {
	f1, err := readACHFilepath(filepath.Join("test", "testdata", "ppd-debit.ach"))
	require.NoError(t, err)

	f2, err := readACHFilepath(filepath.Join("test", "testdata", "web-debit.ach"))
	require.NoError(t, err)

	f2.Header = f1.Header // replace Header so they're merged into one file

	if len(f1.Batches) != 1 || len(f2.Batches) != 3 {
		t.Errorf("did batch counts change? f1:%d f2:%d", len(f1.Batches), len(f2.Batches))
	}

	out, err := MergeFiles([]*File{f1, f2})
	require.NoError(t, err)

	if len(out) != 1 {
		t.Errorf("got %d merged ACH files", len(out))
	}
	if len(out[0].Batches) != 4 {
		t.Errorf("got %d batches", len(out[0].Batches))
	}

	for _, f := range out {
		require.NoError(t, f.Validate())
	}
}

func TestMergeFiles__apart(t *testing.T) {
	f1, err := readACHFilepath(filepath.Join("test", "testdata", "ppd-debit.ach"))
	require.NoError(t, err)

	f2, err := readACHFilepath(filepath.Join("test", "testdata", "web-debit.ach"))
	require.NoError(t, err)

	out, err := MergeFiles([]*File{f1, f2})
	require.NoError(t, err)

	if len(out) != 2 {
		t.Errorf("got %d merged ACH files", len(out))
	}
	if len(out[0].Batches) != 1 {
		t.Errorf("got %d batches", len(out[0].Batches))
	}
	if len(out[1].Batches) != 3 {
		t.Errorf("got %d batches", len(out[1].Batches))
	}

	for _, f := range out {
		require.NoError(t, f.Validate())
	}
}

func BenchmarkLineCount(b *testing.B) {
	newACHFile := func() *File {
		// Nacha files have a max of 10,000 lines and a batch is
		// a header, entries, and control.
		batches, err := rand.Int(rand.Reader, big.NewInt(3000))
		if err != nil {
			b.Fatal(err)
		}

		file := NewFile()
		file.SetHeader(mockFileHeader())
		file.Control = mockFileControl()

		for i := 0; i < int(batches.Int64()+1); i++ {
			file.AddBatch(mockBatchPPD(b))
		}
		if err := file.Create(); err != nil {
			b.Fatal(err)
		}
		return file
	}

	for i := 0; i < b.N; i++ {
		b.StopTimer() // pause timer so we can init our ACH file
		file := newACHFile()
		b.StartTimer() // resume benchmark

		// Count lines in our file
		lineCount(file)
	}
}

func TestMergeFiles__lineCount(t *testing.T) {
	file, err := readACHFilepath(filepath.Join("test", "testdata", "ppd-debit.ach"))
	require.NoError(t, err)
	require.NoError(t, file.Create())

	if n := lineCount(file); n != 5 {
		// We've optimized small file line counts to bypass writing out the file
		// into plain text as it's costly.
		t.Errorf("did we change optimizations? n=%d", n)
	}

	// Add 100 batches to file and get a real line count
	populateFileWithMockBatches(t, 100, file)

	require.NoError(t, file.Create())

	if n := lineCount(file); n != 305 {
		t.Errorf("unexpected line count of %d", n)
	}

	// Remove BatchCount and still properly count lines
	file.Control.BatchCount = 0
	if n := lineCount(file); n != 305 {
		t.Errorf("unexpected error n=%d", n)
	}

	// Ensure merge won't create files over the line limit
	f2, err := readACHFilepath(filepath.Join("test", "testdata", "web-debit.ach"))
	require.NoError(t, err)
	f2.Header = file.Header // replace Header so they're merged into one file
	require.NoError(t, f2.Create())

	output, err := MergeFilesWith([]*File{file, f2}, Conditions{
		MaxLines: 100,
	})
	require.NoError(t, err)
	require.Len(t, output, 2)
	require.Equal(t, 100, lineCount(output[0]))
	require.Equal(t, 23, lineCount(output[1]))
}

// TestMergeFiles__splitFiles generates a file over the 10k line limit and attempts to merge
// another file into it only to come away with two files after merging.
func TestMergeFiles__splitFiles(t *testing.T) {
	file, err := readACHFilepath(filepath.Join("test", "testdata", "ppd-debit.ach"))
	require.NoError(t, err)

	file.Control = NewFileControl()
	require.NoError(t, file.Create())

	if len(file.Batches) != 1 {
		t.Fatalf("unexpected batch count of %d", len(file.Batches))
	}

	// Add a bunch of batches so it's over the line limit
	// somewhere between 3-4k Batches exceed the 10k line limit
	populateFileWithMockBatches(t, 4000, file)

	require.NoError(t, file.Create())

	// Read another file to merge
	f2, err := readACHFilepath(filepath.Join("test", "testdata", "web-debit.ach"))
	require.NoError(t, err)

	f2.Header = file.Header // replace Header so they're merged into one file
	require.NoError(t, f2.Create())

	// read a third file
	f3, err := readACHFilepath(filepath.Join("test", "testdata", "20110805A.ach"))
	require.NoError(t, err)

	f3.Header = file.Header // replace Header so they're merged into one file
	require.NoError(t, f3.Create())

	traceNumbersBefore := countTraceNumbers(file, f2, f3)

	out, err := MergeFiles([]*File{file, f2, f3})
	if err != nil || len(out) != 1 {
		t.Fatalf("got %d files, error=%v", len(out), err)
	}
	require.Equal(t, 7, len(out[0].Batches))

	traceNumbersAfter := countTraceNumbers(out...)
	if traceNumbersBefore != traceNumbersAfter {
		t.Fatalf("found %d of %d trace numbers", traceNumbersBefore, traceNumbersAfter)
	}

	for _, f := range out {
		if err := f.Validate(); err != nil {
			t.Fatalf("invalid file: %v", err)
		}

		min, err := f.FlattenBatches()
		require.NoError(t, err)
		require.NoError(t, min.Validate())
	}
}

func TestMergeFiles__dollarAmount(t *testing.T) {
	file, err := readACHFilepath(filepath.Join("test", "testdata", "ppd-debit.ach"))
	require.NoError(t, err)
	require.NoError(t, file.Create())

	if n := lineCount(file); n != 5 {
		// We've optimized small file line counts to bypass writing out the file
		// into plain text as it's costly.
		t.Errorf("did we change optimizations? n=%d", n)
	}

	// Add 100 batches to file and get a real line count
	populateFileWithMockBatches(t, 100, file)

	// Verify our file's contents
	require.NoError(t, file.Create())
	require.Equal(t, 305, lineCount(file))
	require.Equal(t, 101, countTraceNumbers(file))

	mergedFiles, err := MergeFilesWith([]*File{file}, Conditions{
		MaxDollarAmount: 1000000, // $10,000.00
	})
	require.NoError(t, err)
	require.Len(t, mergedFiles, 101)
	require.Equal(t, 101, countTraceNumbers(mergedFiles...))

	for i := range mergedFiles {
		// With our static cases each file has one Batch
		require.Equal(t, 1, len(mergedFiles[i].Batches))

		entryCount := len(mergedFiles[i].Batches[0].GetEntries())
		require.Equal(t, 1, entryCount)
	}
}

func TestMergeFiles__dollarAmount2(t *testing.T) {
	file, err := readACHFilepath(filepath.Join("test", "testdata", "ppd-debit.ach"))
	require.NoError(t, err)
	require.NoError(t, file.Create())

	if n := lineCount(file); n != 5 {
		// We've optimized small file line counts to bypass writing out the file
		// into plain text as it's costly.
		t.Errorf("did we change optimizations? n=%d", n)
	}

	// Add 100 batches to file and get a real line count
	populateFileWithMockBatches(t, 100, file)

	// Verify our file's contents
	require.NoError(t, file.Create())
	require.Equal(t, 305, lineCount(file))
	require.Equal(t, 101, countTraceNumbers(file))

	mergedFiles, err := MergeFilesWith([]*File{file}, Conditions{
		MaxDollarAmount: 33_000_000_00,
	})
	require.NoError(t, err)
	require.Len(t, mergedFiles, 4)
	require.Equal(t, 101, countTraceNumbers(mergedFiles...))

	require.Equal(t, 2, len(mergedFiles[0].Batches))
	require.Equal(t, 1, len(mergedFiles[1].Batches))
	require.Equal(t, 1, len(mergedFiles[2].Batches))
	require.Equal(t, 1, len(mergedFiles[3].Batches))

	entryCount := len(mergedFiles[0].Batches[0].GetEntries())
	require.Equal(t, 1, entryCount)
	entryCount = len(mergedFiles[0].Batches[1].GetEntries())
	require.Equal(t, 32, entryCount)

	entryCount = len(mergedFiles[1].Batches[0].GetEntries())
	require.Equal(t, 33, entryCount)

	entryCount = len(mergedFiles[2].Batches[0].GetEntries())
	require.Equal(t, 33, entryCount)

	entryCount = len(mergedFiles[3].Batches[0].GetEntries())
	require.Equal(t, 2, entryCount)
}

func countTraceNumbers(files ...*File) int {
	var total int
	for f := range files {
		for b := range files[f].Batches {
			total += len(files[f].Batches[b].GetEntries())
		}
	}
	return total
}

func TestMergeFiles__invalid(t *testing.T) {
	f1, err := readACHFilepath(filepath.Join("test", "testdata", "ppd-debit.ach"))
	if err != nil {
		t.Fatal(err)
	}
	f1.Header.ImmediateOrigin = "0000000000" // make file invalid

	f2, err := readACHFilepath(filepath.Join("test", "testdata", "web-debit.ach"))
	if err != nil {
		t.Fatal(err)
	}
	f2.Header = f1.Header

	out, err := MergeFiles([]*File{f1, f2})
	if len(out) != 0 || err == nil {
		t.Errorf("expected error: len(out)=%d error=%v", len(out), err)
	}
}

func populateFileWithMockBatches(t testing.TB, numBatches int, file *File) {
	lastBatchIdx := len(file.Batches) - 1
	var startSeq = file.Batches[lastBatchIdx].GetHeader().BatchNumber + 1
	var entryDetail = file.Batches[0].GetEntries()[0]

	for i := startSeq; i < (numBatches + startSeq); i++ {
		header := mockBatchHeader()
		header.StandardEntryClassCode = "PPD"
		header.ServiceClassCode = 225
		header.CompanyName = "Example Company"
		header.CompanyIdentification = "132465"
		header.CompanyEntryDescription = "Example Description"
		header.ODFIIdentification = "12104288"
		batch, err := NewBatch(header)
		if err != nil {
			t.Fatal(err)
		}

		ed := *entryDetail
		n, _ := strconv.Atoi(ed.TraceNumber)
		ed.TraceNumber = strconv.Itoa(n + i + 1e5)
		batch.AddEntry(&ed)

		batch.GetHeader().BatchNumber = i
		batch.GetControl().BatchNumber = i

		if err := batch.Create(); err != nil {
			t.Fatal(err)
		}

		file.AddBatch(batch)
	}
}

func TestMergeFiles__ValidateOpts(t *testing.T) {
	f1, err := readACHFilepath(filepath.Join("test", "testdata", "ppd-debit.ach"))
	require.NoError(t, err)

	f1.SetValidation(&ValidateOpts{
		CustomReturnCodes: true,
	})

	f2, err := readACHFilepath(filepath.Join("test", "testdata", "web-debit.ach"))
	require.NoError(t, err)

	f2.Header = f1.Header
	f2.SetValidation(&ValidateOpts{
		AllowInvalidAmounts: true,
	})

	merged, err := MergeFiles([]*File{f1, f2})
	require.NoError(t, err)
	require.Len(t, merged, 1)

	opts := merged[0].GetValidation()
	require.False(t, opts.SkipAll)
	require.True(t, opts.CustomReturnCodes)
	require.True(t, opts.AllowInvalidAmounts)
}

func TestMergeFilesHelpers(t *testing.T) {
	t.Run("pickOutFile", func(t *testing.T) {
		fh := mockFileHeader()
		var input *outFile

		output := pickOutFile(fh, input)
		require.Equal(t, fh, output.header)
		require.Empty(t, output.batches)
		require.Nil(t, output.next)

		input = &outFile{
			header: mockFileHeader(),
		}
		require.Equal(t, input, pickOutFile(fh, input))

		fh2 := mockFileHeader()
		fh2.ImmediateOrigin = "123456780"
		output = pickOutFile(fh2, input)
		require.Equal(t, output, input.next) // verify the chain continues
		require.Equal(t, fh2, output.header)
		require.Empty(t, output.batches)
		require.Nil(t, output.next)

		fh3 := mockFileHeader()
		fh3.ImmediateDestination = "123456780"
		output = pickOutFile(fh3, input)
		require.Equal(t, fh3, output.header)
	})

	t.Run("findOutBatch", func(t *testing.T) {
		bh := mockBatchPPDHeader()
		var batches []*batch
		var entry *EntryDetail

		output := findOutBatch(bh, batches, entry)
		require.Nil(t, output)

		// find the batch
		batches = append(batches, &batch{
			header:  *bh,
			entries: treemap.New[string, *EntryDetail](),
		})
		output = findOutBatch(bh, batches, entry)
		require.Equal(t, batches[0], output)

		// add an entry to the batch
		traceNumber := "123456780000000001"
		batches[0].entries.Set(traceNumber, &EntryDetail{
			ID:          "1",
			TraceNumber: traceNumber,
		})
		require.True(t, batches[0].entries.Contains(traceNumber))

		// exclude the batch when the trace number is found
		output = findOutBatch(bh, batches, &EntryDetail{
			TraceNumber: traceNumber,
		})
		require.Nil(t, output)
	})
}
