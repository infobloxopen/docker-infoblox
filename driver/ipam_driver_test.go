package main

import(
  "testing"
  "os"
  "io/ioutil"
)


func TestDirExists(t *testing.T){
    tempDir, _ := ioutil.TempDir(os.TempDir(), "")
    defer os.Remove(tempDir)

    exists, err := dirExists(tempDir)

    if err != nil{
        t.Errorf("Unexpected Error Excepted: nil, Actual: %v", err)
    }

    if !exists{
        t.Errorf("Expected: Directory %s to exist, Actual: Directory %s does not exists", tempDir, tempDir)
    }
}


func TestDirExistsNegative(t *testing.T){
    tempDir := "/dirDoesNotExists"

    exists, err := dirExists(tempDir)

    if err != nil{
        t.Errorf("Unexpected Error Excepted: nil, Actual: %v", err)
    }

    if exists{
        t.Errorf("Expected: Directory %s not to exist, Actual: Directory %s exists", tempDir, tempDir)
    }
}
