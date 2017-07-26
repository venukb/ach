// Copyright 2017 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package ach

import (
	"testing"
)

func mockFilePPD() *File {
	mockFile := NewFile()
	mockFile.SetHeader(mockFileHeader())
	mockBatch := mockBatchPPD()
	mockFile.AddBatch(mockBatch)
	if err := mockFile.Build(); err != nil {
		panic(err)
	}
	return mockFile
}

func TestFileError(t *testing.T) {
	err := &FileError{FieldName: "mock", Msg: "test message"}
	if err.Error() != "mock test message" {
		t.Error("FileError Error has changed formatting")
	}
}

// TestFileBatchCount if calculated count is different from control
func TestFileBatchCount(t *testing.T) {
	file := mockFilePPD()

	// More batches than the file control count.
	file.AddBatch(mockBatchPPD())
	if err := file.Validate(); err != nil {
		if e, ok := err.(*FileError); ok {
			if e.FieldName != "BatchCount" {
				t.Errorf("%T: %s", err, err)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}

func TestFileEntryAddenda(t *testing.T) {
	file := mockFilePPD()

	// more entries than the file control
	file.Control.EntryAddendaCount = 5
	if err := file.Validate(); err != nil {
		if e, ok := err.(*FileError); ok {
			if e.FieldName != "EntryAddendaCount" {
				t.Errorf("%T: %s", err, err)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}

func TestFileDebitAmount(t *testing.T) {
	file := mockFilePPD()

	// inequality in total debit amount
	file.Control.TotalDebitEntryDollarAmountInFile = 63
	if err := file.Validate(); err != nil {
		if e, ok := err.(*FileError); ok {
			if e.FieldName != "TotalDebitEntryDollarAmountInFile" {
				t.Errorf("%T: %s", err, err)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}

func TestFileCreditAmount(t *testing.T) {
	file := mockFilePPD()

	// inequality in total debit amount
	file.Control.TotalCreditEntryDollarAmountInFile = 63
	if err := file.Validate(); err != nil {
		if e, ok := err.(*FileError); ok {
			if e.FieldName != "TotalCreditEntryDollarAmountInFile" {
				t.Errorf("%T: %s", err, err)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}

func TestFileEntryHash(t *testing.T) {
	file := mockFilePPD()
	file.AddBatch(mockBatchPPD())
	file.Build()
	file.Control.EntryHash = 63
	if err := file.Validate(); err != nil {
		if e, ok := err.(*FileError); ok {
			if e.FieldName != "EntryHash" {
				t.Errorf("%T: %s", err, err)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}

func TestFileBlockCount10(t *testing.T) {
	file := NewFile().SetHeader(mockFileHeader())
	batch := NewBatchPPD()
	batch.SetHeader(mockBatchHeader())
	batch.AddEntry(mockEntryDetail())
	batch.AddEntry(mockEntryDetail())
	batch.AddEntry(mockEntryDetail())
	batch.AddEntry(mockEntryDetail())
	batch.AddEntry(mockEntryDetail())
	batch.AddEntry(mockEntryDetail())
	batch.Build()
	file.AddBatch(batch)
	if err := file.Build(); err != nil {
		t.Errorf("%T: %s", err, err)
	}

	// ensure with 10 records in file we don't get 2 for a block count
	if file.Control.BlockCount != 1 {
		t.Error("BlockCount on 10 records is not equal to 1")
	}
	// make 11th record which should produce BlockCount of 2
	file.Batches[0].AddEntry(mockEntryDetail())
	file.Batches[0].Build() // File.Build does not re-build Batches
	if err := file.Build(); err != nil {
		t.Errorf("%T: %s", err, err)
	}
	if file.Control.BlockCount != 2 {
		t.Error("BlockCount on 11 records is not equal to 2")
	}
}

func TestFileBuildBadFileHeader(t *testing.T) {
	file := NewFile().SetHeader(FileHeader{})
	if err := file.Build(); err != nil {
		if e, ok := err.(*FieldError); ok {
			if e.Msg != msgFieldInclusion {
				t.Errorf("%T: %s", err, err)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}

func TestFileBuildNoBatch(t *testing.T) {
	file := NewFile().SetHeader(mockFileHeader())
	if err := file.Build(); err != nil {
		if e, ok := err.(*FileError); ok {
			if e.FieldName != "Batchs" {
				t.Errorf("%T: %s", err, err)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}

func TestFileValidateAllBatch(t *testing.T) {
	file := mockFilePPD()
	// break the file header
	file.Batches[0].GetHeader().ODFIIdentification = 0
	if err := file.ValidateAll(); err != nil {
		if e, ok := err.(*FieldError); ok {
			if e.Msg != msgFieldInclusion {
				t.Errorf("%T: %s", err, err)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}

func TestFileValidateAllFileHeader(t *testing.T) {
	file := mockFilePPD()
	// break the file header
	file.Header.ImmediateOrigin = 0
	if err := file.ValidateAll(); err != nil {
		if e, ok := err.(*FieldError); ok {
			if e.Msg != msgFieldInclusion {
				t.Errorf("%T: %s", err, err)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}

func TestFileValidateAllFileControl(t *testing.T) {
	file := mockFilePPD()
	// break the file header
	file.Control.BatchCount = 0
	if err := file.ValidateAll(); err != nil {
		if e, ok := err.(*FieldError); ok {
			if e.Msg != msgFieldInclusion {
				t.Errorf("%T: %s", err, err)
			}
		} else {
			t.Errorf("%T: %s", err, err)
		}
	}
}
