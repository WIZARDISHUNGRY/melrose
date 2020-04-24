package dsl

import (
	"testing"

	"github.com/emicklei/melrose"
)

func TestIsCompatible(t *testing.T) {
	if got, want := true, IsCompatibleSyntax("1.0"); got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := false, IsCompatibleSyntax("2.0"); got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := false, IsCompatibleSyntax("1.1"); got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func TestNestedFunctions(t *testing.T) {
	e := NewEvaluator(NewVariableStore(), melrose.NoLooper)
	input := `pitch(1,repeat(1,reverse(join(note('E'),sequence('F G')))))`
	_, err := e.EvaluateExpression(input)
	if err != nil {
		t.Error(err)
	}
}

func TestMulitLineEvaluate(t *testing.T) {
	e := NewEvaluator(NewVariableStore(), melrose.NoLooper)
	input := `sequence("
		C D E C 
		C D E C 
		E F 2G
		E F 2G 
		8G 8A 8G 8F E C 
		8G 8A 8G 8F E C
		2C 2G3 2C
		2C 2G3 2C
		")`
	_, err := e.EvaluateStatement(input)
	if err != nil {
		t.Error(err)
	}
}

func Test_isAssignment(t *testing.T) {
	type args struct {
		entry string
	}
	tests := []struct {
		name           string
		args           args
		wantVarname    string
		wantExpression string
		wantOk         bool
	}{
		{"a=1",
			args{"a=1"},
			"a",
			"1",
			true,
		},
		{" a = note('=')",
			args{" a = note('=')"},
			"a",
			"note('=')",
			true,
		},
		{"multi line",
			args{`j2 = join(  repeat(2,i2) )`},
			"j2",
			"join(  repeat(2,i2) )",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVarname, gotExpression, gotOk := IsAssignment(tt.args.entry)
			if gotVarname != tt.wantVarname {
				t.Errorf("isAssignment() gotVarname = %v, want %v", gotVarname, tt.wantVarname)
			}
			if gotExpression != tt.wantExpression {
				t.Errorf("isAssignment() gotExpression = %v, want %v", gotExpression, tt.wantExpression)
			}
			if gotOk != tt.wantOk {
				t.Errorf("isAssignment() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
