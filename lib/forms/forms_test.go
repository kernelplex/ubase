package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type MyInt int

type Embedded struct {
	ES string `json:"es"`
	EI int    `json:"ei"`
}

type Inner struct {
	X int `json:"x"`
}

type TestStruct struct {
	S   string    `json:"s"`
	I   int       `json:"i"`
	I8  int8      `json:"i8"`
	I64 int64     `json:"i64"`
	MI  MyInt     `json:"myInt"`
	U   uint      `json:"u"`
	U16 uint16    `json:"u16"`
	F32 float32   `json:"f32"`
	F64 float64   `json:"f64"`
	B   bool      `json:"b"`
	T   time.Time `json:"t"`

	PS *string    `json:"ps"`
	PI *int       `json:"pi"`
	PB *bool      `json:"pb"`
	PT *time.Time `json:"pt"`

	SS []string    `json:"ss"`
	SI []int       `json:"si"`
	SU []uint      `json:"su"`
	SF []float64   `json:"sf"`
	SB []bool      `json:"sb"`
	ST []time.Time `json:"st"`

	Embedded

	Inner Inner `json:"inner"`
}

func TestParseFormToStruct_AllTypes(t *testing.T) {
	vals := url.Values{}
	// scalars
	vals.Set("s", "  hello  ")
	vals.Set("i", "42")
	vals.Set("i8", "8")
	vals.Set("i64", "64")
	vals.Set("myInt", "123")
	vals.Set("u", "7")
	vals.Set("u16", "16")
	vals.Set("f32", "3.14")
	vals.Set("f64", "2.71828")
	vals.Set("b", "on")
	vals.Set("t", "2024-09-01T12:34:56Z")

	// pointers
	vals.Set("ps", " hi ")
	vals.Set("pi", "99")
	vals.Set("pb", "0")
	vals.Set("pt", "2024-01-02")

	// slices (with some empties and spaces)
	vals["ss"] = []string{"a", "", " b "}
	vals["si"] = []string{"1", "2", " 3 "}
	vals["su"] = []string{"5", "6"}
	vals["sf"] = []string{"1.5", "2.5"}
	vals["sb"] = []string{"true", "false", "1", "0", "on", "off", "yes", "no", "y", "n"}
	vals["st"] = []string{"2024-09-01T12:34:56Z", "2024-01-02", "2024-01-02 03:04:05"}

	// embedded
	vals.Set("es", "emb")
	vals.Set("ei", "77")

	// non-time struct value provided: should be ignored without error
	vals.Set("inner", "ignored")

	req := httptest.NewRequest(http.MethodPost, "http://example.test/submit", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var dst TestStruct
	if err := ParseFormToStruct(req, &dst); err != nil {
		t.Fatalf("ParseFormToStruct returned error: %v", err)
	}

	if dst.S != "hello" { // trimmed
		t.Errorf("S got %q, want %q", dst.S, "hello")
	}
	if dst.I != 42 || dst.I8 != 8 || dst.I64 != 64 || dst.MI != MyInt(123) {
		t.Errorf("int fields not parsed correctly: %+v", dst)
	}
	if dst.U != 7 || dst.U16 != 16 {
		t.Errorf("uint fields not parsed correctly: U=%d U16=%d", dst.U, dst.U16)
	}
	if !floatAlmostEqual32(dst.F32, 3.14, 1e-5) || !floatAlmostEqual64(dst.F64, 2.71828, 1e-9) {
		t.Errorf("float fields not parsed correctly: F32=%f F64=%f", dst.F32, dst.F64)
	}
	if dst.B != true {
		t.Errorf("bool field not parsed correctly: B=%v", dst.B)
	}

	tExpected, _ := time.Parse(time.RFC3339, "2024-09-01T12:34:56Z")
	if !dst.T.Equal(tExpected) {
		t.Errorf("time field not parsed correctly: got %v want %v", dst.T, tExpected)
	}

	if dst.PS == nil || *dst.PS != "hi" { // trimmed
		t.Errorf("pointer string not parsed: %v", dst.PS)
	}
	if dst.PI == nil || *dst.PI != 99 {
		t.Errorf("pointer int not parsed: %v", dst.PI)
	}
	if dst.PB == nil || *dst.PB != false {
		t.Errorf("pointer bool not parsed: %v", dst.PB)
	}
	ptExpected, _ := time.Parse("2006-01-02", "2024-01-02")
	if dst.PT == nil || !dst.PT.Equal(ptExpected) {
		t.Errorf("pointer time not parsed: got %v want %v", dst.PT, ptExpected)
	}

	// slices
	if len(dst.SS) != 2 || dst.SS[0] != "a" || dst.SS[1] != "b" {
		t.Errorf("[]string parsed incorrectly: %#v", dst.SS)
	}
	if len(dst.SI) != 3 || dst.SI[0] != 1 || dst.SI[1] != 2 || dst.SI[2] != 3 {
		t.Errorf("[]int parsed incorrectly: %#v", dst.SI)
	}
	if len(dst.SU) != 2 || dst.SU[0] != 5 || dst.SU[1] != 6 {
		t.Errorf("[]uint parsed incorrectly: %#v", dst.SU)
	}
	if len(dst.SF) != 2 || !floatAlmostEqual64(dst.SF[0], 1.5, 1e-9) || !floatAlmostEqual64(dst.SF[1], 2.5, 1e-9) {
		t.Errorf("[]float parsed incorrectly: %#v", dst.SF)
	}
	expBools := []bool{true, false, true, false, true, false, true, false, true, false}
	if len(dst.SB) != len(expBools) {
		t.Fatalf("[]bool length mismatch: got %d want %d", len(dst.SB), len(expBools))
	}
	for i := range expBools {
		if dst.SB[i] != expBools[i] {
			t.Errorf("[]bool[%d] = %v want %v", i, dst.SB[i], expBools[i])
		}
	}
	if len(dst.ST) != 3 {
		t.Fatalf("[]time length mismatch: got %d", len(dst.ST))
	}
	st0, _ := time.Parse(time.RFC3339, "2024-09-01T12:34:56Z")
	st1, _ := time.Parse("2006-01-02", "2024-01-02")
	st2, _ := time.Parse("2006-01-02 15:04:05", "2024-01-02 03:04:05")
	if !dst.ST[0].Equal(st0) || !dst.ST[1].Equal(st1) || !dst.ST[2].Equal(st2) {
		t.Errorf("[]time parsed incorrectly: %#v", dst.ST)
	}

	// embedded fields
	if dst.ES != "emb" || dst.EI != 77 {
		t.Errorf("embedded fields parsed incorrectly: ES=%q EI=%d", dst.ES, dst.EI)
	}

	// inner struct remains zero (ignored)
	if dst.Inner.X != 0 {
		t.Errorf("non-time struct should be skipped; got %+v", dst.Inner)
	}
}

func TestParseFormToStruct_Errors(t *testing.T) {
	vals := url.Values{}
	vals.Set("i", "not-an-int")
	req := httptest.NewRequest(http.MethodPost, "http://example.test/submit", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var dst struct {
		I int `json:"i"`
	}
	if err := ParseFormToStruct(req, &dst); err == nil {
		t.Fatalf("expected error for invalid int, got nil")
	}
}

func floatAlmostEqual32(a, b float32, eps float32) bool {
	if a > b {
		return float32(a-b) < eps
	}
	return float32(b-a) < eps
}

func floatAlmostEqual64(a, b float64, eps float64) bool {
	if a > b {
		return a-b < eps
	}
	return b-a < eps
}

func TestParseFormToStruct_EmptyPointersAndSlices(t *testing.T) {
	vals := url.Values{}
	// Explicitly include empty values; they should be trimmed out and ignored.
	vals["ps"] = []string{"", "  "}
	vals["pi"] = []string{"  "}
	vals["pb"] = []string{""}
	vals["pt"] = []string{"  "}
	vals["ss"] = []string{" ", ""}
	vals["si"] = []string{"", "  "}
	req := httptest.NewRequest(http.MethodPost, "http://example.test/empty", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var dst TestStruct
	if err := ParseFormToStruct(req, &dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.PS != nil || dst.PI != nil || dst.PB != nil || dst.PT != nil {
		t.Errorf("pointer fields should remain nil when inputs are empty: PS=%v PI=%v PB=%v PT=%v", dst.PS, dst.PI, dst.PB, dst.PT)
	}
	if len(dst.SS) != 0 {
		t.Errorf("[]string should be empty (or nil), got %#v", dst.SS)
	}
	if len(dst.SI) != 0 {
		t.Errorf("[]int should be empty (or nil), got %#v", dst.SI)
	}
}

func TestParseFormToStruct_InvalidTypesErrors(t *testing.T) {
	cases := []struct {
		name    string
		key     string
		val     string
		makeDst func() any
	}{
		{
			name: "invalid uint",
			key:  "u",
			val:  "-1",
			makeDst: func() any {
				return &struct {
					U uint `json:"u"`
				}{}
			},
		},
		{
			name: "invalid float",
			key:  "f64",
			val:  "abc",
			makeDst: func() any {
				return &struct {
					F float64 `json:"f64"`
				}{}
			},
		},
		{
			name: "invalid bool",
			key:  "b",
			val:  "maybe",
			makeDst: func() any {
				return &struct {
					B bool `json:"b"`
				}{}
			},
		},
		{
			name: "invalid time",
			key:  "t",
			val:  "not-a-time",
			makeDst: func() any {
				return &struct {
					T time.Time `json:"t"`
				}{}
			},
		},
		{
			name: "invalid slice element",
			key:  "si",
			val:  "x",
			makeDst: func() any {
				return &struct {
					SI []int `json:"si"`
				}{}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vals := url.Values{}
			vals[tc.key] = []string{tc.val}
			req := httptest.NewRequest(http.MethodPost, "http://example.test/invalid", strings.NewReader(vals.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			dst := tc.makeDst()
			if err := ParseFormToStruct(req, dst); err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestParseFormToStruct_ScalarUsesFirstValue(t *testing.T) {
	vals := url.Values{}
	vals["i"] = []string{"1", "2", "3"}
	req := httptest.NewRequest(http.MethodPost, "http://example.test/multi", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var dst struct {
		I int `json:"i"`
	}
	if err := ParseFormToStruct(req, &dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.I != 1 {
		t.Errorf("expected first value 1, got %d", dst.I)
	}
}
