package models

import "testing"

func Test_ParseCheckBox_1(t *testing.T) {
    if ParseCheckBox("on") != true {
      t.Error("ParseCheckbox should return true for 'on'")
    }
}

func Tes_ParseCheckBox_2(t *testing.T) {
    if ParseCheckBox("") != false {
      t.Error("ParseCheckBox should return false for ''")
    }
}
