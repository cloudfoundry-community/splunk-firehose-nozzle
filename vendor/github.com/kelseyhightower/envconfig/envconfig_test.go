// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package envconfig

import (
	"flag"
	"fmt"
	"os"
	"testing"
	"time"
)

type HonorDecodeInStruct struct {
	Value string
}

func (h *HonorDecodeInStruct) Decode(env string) error {
	h.Value = "decoded"
	return nil
}

type Specification struct {
	Embedded
	EmbeddedButIgnored           `ignored:"true"`
	Debug                        bool
	Port                         int
	Rate                         float32
	User                         string
	TTL                          uint32
	Timeout                      time.Duration
	AdminUsers                   []string
	MagicNumbers                 []int
	MultiWordVar                 string
	SomePointer                  *string
	SomePointerWithDefault       *string `default:"foo2baz"`
	MultiWordVarWithAlt          string  `envconfig:"MULTI_WORD_VAR_WITH_ALT"`
	MultiWordVarWithLowerCaseAlt string  `envconfig:"multi_word_var_with_lower_case_alt"`
	NoPrefixWithAlt              string  `envconfig:"SERVICE_HOST"`
	DefaultVar                   string  `default:"foobar"`
	RequiredVar                  string  `required:"true"`
	NoPrefixDefault              string  `envconfig:"BROKER" default:"127.0.0.1"`
	RequiredDefault              string  `required:"true" default:"foo2bar"`
	Ignored                      string  `ignored:"true"`
	NestedSpecification          struct {
		Property            string `envconfig:"inner"`
		PropertyWithDefault string `default:"fuzzybydefault"`
	} `envconfig:"outer"`
	AfterNested  string
	DecodeStruct HonorDecodeInStruct `envconfig:"honor"`
	Datetime     time.Time
}

type Embedded struct {
	Enabled             bool
	EmbeddedPort        int
	MultiWordVar        string
	MultiWordVarWithAlt string `envconfig:"MULTI_WITH_DIFFERENT_ALT"`
	EmbeddedAlt         string `envconfig:"EMBEDDED_WITH_ALT"`
	EmbeddedIgnored     string `ignored:"true"`
}

type EmbeddedButIgnored struct {
	FirstEmbeddedButIgnored  string
	SecondEmbeddedButIgnored string
}

func TestProcess(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEBUG", "true")
	os.Setenv("ENV_CONFIG_PORT", "8080")
	os.Setenv("ENV_CONFIG_RATE", "0.5")
	os.Setenv("ENV_CONFIG_USER", "Kelsey")
	os.Setenv("ENV_CONFIG_TIMEOUT", "2m")
	os.Setenv("ENV_CONFIG_ADMINUSERS", "John,Adam,Will")
	os.Setenv("ENV_CONFIG_MAGICNUMBERS", "5,10,20")
	os.Setenv("SERVICE_HOST", "127.0.0.1")
	os.Setenv("ENV_CONFIG_TTL", "30")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	os.Setenv("ENV_CONFIG_IGNORED", "was-not-ignored")
	os.Setenv("ENV_CONFIG_OUTER_INNER", "iamnested")
	os.Setenv("ENV_CONFIG_AFTERNESTED", "after")
	os.Setenv("ENV_CONFIG_HONOR", "honor")
	os.Setenv("ENV_CONFIG_DATETIME", "2016-08-16T18:57:05Z")
	err := Process("env_config", &s)
	if err != nil {
		t.Error(err.Error())
	}
	if s.NoPrefixWithAlt != "127.0.0.1" {
		t.Errorf("expected %v, got %v", "127.0.0.1", s.NoPrefixWithAlt)
	}
	if !s.Debug {
		t.Errorf("expected %v, got %v", true, s.Debug)
	}
	if s.Port != 8080 {
		t.Errorf("expected %d, got %v", 8080, s.Port)
	}
	if s.Rate != 0.5 {
		t.Errorf("expected %f, got %v", 0.5, s.Rate)
	}
	if s.TTL != 30 {
		t.Errorf("expected %d, got %v", 30, s.TTL)
	}
	if s.User != "Kelsey" {
		t.Errorf("expected %s, got %s", "Kelsey", s.User)
	}
	if s.Timeout != 2*time.Minute {
		t.Errorf("expected %s, got %s", 2*time.Minute, s.Timeout)
	}
	if s.RequiredVar != "foo" {
		t.Errorf("expected %s, got %s", "foo", s.RequiredVar)
	}
	if len(s.AdminUsers) != 3 ||
		s.AdminUsers[0] != "John" ||
		s.AdminUsers[1] != "Adam" ||
		s.AdminUsers[2] != "Will" {
		t.Errorf("expected %#v, got %#v", []string{"John", "Adam", "Will"}, s.AdminUsers)
	}
	if len(s.MagicNumbers) != 3 ||
		s.MagicNumbers[0] != 5 ||
		s.MagicNumbers[1] != 10 ||
		s.MagicNumbers[2] != 20 {
		t.Errorf("expected %#v, got %#v", []int{5, 10, 20}, s.MagicNumbers)
	}
	if s.Ignored != "" {
		t.Errorf("expected empty string, got %#v", s.Ignored)
	}

	if s.NestedSpecification.Property != "iamnested" {
		t.Errorf("expected '%s' string, got %#v", "iamnested", s.NestedSpecification.Property)
	}

	if s.NestedSpecification.PropertyWithDefault != "fuzzybydefault" {
		t.Errorf("expected default '%s' string, got %#v", "fuzzybydefault", s.NestedSpecification.PropertyWithDefault)
	}

	if s.AfterNested != "after" {
		t.Errorf("expected default '%s' string, got %#v", "after", s.AfterNested)
	}

	if s.DecodeStruct.Value != "decoded" {
		t.Errorf("expected default '%s' string, got %#v", "decoded", s.DecodeStruct.Value)
	}

	if expected := time.Date(2016, 8, 16, 18, 57, 05, 0, time.UTC); !s.Datetime.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.Format(time.RFC3339), s.Datetime.Format(time.RFC3339))
	}
}

func TestParseErrorBool(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEBUG", "string")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	err := Process("env_config", &s)
	v, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %v", v)
	}
	if v.FieldName != "Debug" {
		t.Errorf("expected %s, got %v", "Debug", v.FieldName)
	}
	if s.Debug != false {
		t.Errorf("expected %v, got %v", false, s.Debug)
	}
}

func TestParseErrorFloat32(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_RATE", "string")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	err := Process("env_config", &s)
	v, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %v", v)
	}
	if v.FieldName != "Rate" {
		t.Errorf("expected %s, got %v", "Rate", v.FieldName)
	}
	if s.Rate != 0 {
		t.Errorf("expected %v, got %v", 0, s.Rate)
	}
}

func TestParseErrorInt(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_PORT", "string")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	err := Process("env_config", &s)
	v, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %v", v)
	}
	if v.FieldName != "Port" {
		t.Errorf("expected %s, got %v", "Port", v.FieldName)
	}
	if s.Port != 0 {
		t.Errorf("expected %v, got %v", 0, s.Port)
	}
}

func TestParseErrorUint(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_TTL", "-30")
	err := Process("env_config", &s)
	v, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %v", v)
	}
	if v.FieldName != "TTL" {
		t.Errorf("expected %s, got %v", "TTL", v.FieldName)
	}
	if s.TTL != 0 {
		t.Errorf("expected %v, got %v", 0, s.TTL)
	}
}

func TestErrInvalidSpecification(t *testing.T) {
	m := make(map[string]string)
	err := Process("env_config", &m)
	if err != ErrInvalidSpecification {
		t.Errorf("expected %v, got %v", ErrInvalidSpecification, err)
	}
}

func TestUnsetVars(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("USER", "foo")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	// If the var is not defined the non-prefixed version should not be used
	// unless the struct tag says so
	if s.User != "" {
		t.Errorf("expected %q, got %q", "", s.User)
	}
}

func TestAlternateVarNames(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_MULTI_WORD_VAR", "foo")
	os.Setenv("ENV_CONFIG_MULTI_WORD_VAR_WITH_ALT", "bar")
	os.Setenv("ENV_CONFIG_MULTI_WORD_VAR_WITH_LOWER_CASE_ALT", "baz")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	// Setting the alt version of the var in the environment has no effect if
	// the struct tag is not supplied
	if s.MultiWordVar != "" {
		t.Errorf("expected %q, got %q", "", s.MultiWordVar)
	}

	// Setting the alt version of the var in the environment correctly sets
	// the value if the struct tag IS supplied
	if s.MultiWordVarWithAlt != "bar" {
		t.Errorf("expected %q, got %q", "bar", s.MultiWordVarWithAlt)
	}

	// Alt value is not case sensitive and is treated as all uppercase
	if s.MultiWordVarWithLowerCaseAlt != "baz" {
		t.Errorf("expected %q, got %q", "baz", s.MultiWordVarWithLowerCaseAlt)
	}
}

func TestRequiredVar(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foobar")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.RequiredVar != "foobar" {
		t.Errorf("expected %s, got %s", "foobar", s.RequiredVar)
	}
}

func TestBlankDefaultVar(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "requiredvalue")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.DefaultVar != "foobar" {
		t.Errorf("expected %s, got %s", "foobar", s.DefaultVar)
	}

	if *s.SomePointerWithDefault != "foo2baz" {
		t.Errorf("expected %s, got %s", "foo2baz", *s.SomePointerWithDefault)
	}
}

func TestNonBlankDefaultVar(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEFAULTVAR", "nondefaultval")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "requiredvalue")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.DefaultVar != "nondefaultval" {
		t.Errorf("expected %s, got %s", "nondefaultval", s.DefaultVar)
	}
}

func TestExplicitBlankDefaultVar(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEFAULTVAR", "")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "")

	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.DefaultVar != "" {
		t.Errorf("expected %s, got %s", "\"\"", s.DefaultVar)
	}
}

func TestAlternateNameDefaultVar(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("BROKER", "betterbroker")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.NoPrefixDefault != "betterbroker" {
		t.Errorf("expected %q, got %q", "betterbroker", s.NoPrefixDefault)
	}

	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.NoPrefixDefault != "127.0.0.1" {
		t.Errorf("expected %q, got %q", "127.0.0.1", s.NoPrefixDefault)
	}
}

func TestRequiredDefault(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.RequiredDefault != "foo2bar" {
		t.Errorf("expected %q, got %q", "foo2bar", s.RequiredDefault)
	}
}

func TestPointerFieldBlank(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if s.SomePointer != nil {
		t.Errorf("expected <nil>, got %q", *s.SomePointer)
	}
}

func TestMustProcess(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_DEBUG", "true")
	os.Setenv("ENV_CONFIG_PORT", "8080")
	os.Setenv("ENV_CONFIG_RATE", "0.5")
	os.Setenv("ENV_CONFIG_USER", "Kelsey")
	os.Setenv("SERVICE_HOST", "127.0.0.1")
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	MustProcess("env_config", &s)

	defer func() {
		if err := recover(); err != nil {
			return
		}

		t.Error("expected panic")
	}()
	m := make(map[string]string)
	MustProcess("env_config", &m)
}

func TestEmbeddedStruct(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "required")
	os.Setenv("ENV_CONFIG_ENABLED", "true")
	os.Setenv("ENV_CONFIG_EMBEDDEDPORT", "1234")
	os.Setenv("ENV_CONFIG_MULTIWORDVAR", "foo")
	os.Setenv("ENV_CONFIG_MULTI_WORD_VAR_WITH_ALT", "bar")
	os.Setenv("ENV_CONFIG_MULTI_WITH_DIFFERENT_ALT", "baz")
	os.Setenv("ENV_CONFIG_EMBEDDED_WITH_ALT", "foobar")
	os.Setenv("ENV_CONFIG_SOMEPOINTER", "foobaz")
	os.Setenv("ENV_CONFIG_EMBEDDED_IGNORED", "was-not-ignored")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}
	if !s.Enabled {
		t.Errorf("expected %v, got %v", true, s.Enabled)
	}
	if s.EmbeddedPort != 1234 {
		t.Errorf("expected %d, got %v", 1234, s.EmbeddedPort)
	}
	if s.MultiWordVar != "foo" {
		t.Errorf("expected %s, got %s", "foo", s.MultiWordVar)
	}
	if s.Embedded.MultiWordVar != "foo" {
		t.Errorf("expected %s, got %s", "foo", s.Embedded.MultiWordVar)
	}
	if s.MultiWordVarWithAlt != "bar" {
		t.Errorf("expected %s, got %s", "bar", s.MultiWordVarWithAlt)
	}
	if s.Embedded.MultiWordVarWithAlt != "baz" {
		t.Errorf("expected %s, got %s", "baz", s.Embedded.MultiWordVarWithAlt)
	}
	if s.EmbeddedAlt != "foobar" {
		t.Errorf("expected %s, got %s", "foobar", s.EmbeddedAlt)
	}
	if *s.SomePointer != "foobaz" {
		t.Errorf("expected %s, got %s", "foobaz", *s.SomePointer)
	}
	if s.EmbeddedIgnored != "" {
		t.Errorf("expected empty string, got %#v", s.Ignored)
	}
}

func TestEmbeddedButIgnoredStruct(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "required")
	os.Setenv("ENV_CONFIG_FIRSTEMBEDDEDBUTIGNORED", "was-not-ignored")
	os.Setenv("ENV_CONFIG_SECONDEMBEDDEDBUTIGNORED", "was-not-ignored")
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}
	if s.FirstEmbeddedButIgnored != "" {
		t.Errorf("expected empty string, got %#v", s.Ignored)
	}
	if s.SecondEmbeddedButIgnored != "" {
		t.Errorf("expected empty string, got %#v", s.Ignored)
	}
}

func TestNonPointerFailsProperly(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "snap")

	err := Process("env_config", s)
	if err != ErrInvalidSpecification {
		t.Errorf("non-pointer should fail with ErrInvalidSpecification, was instead %s", err)
	}
}

func TestCustomValueFields(t *testing.T) {
	var s struct {
		Foo    string
		Bar    bracketed
		Baz    quoted
		Struct setterStruct
	}

	// Set would panic when the receiver is nil,
	// so make sure it has an initial value to replace.
	s.Baz = quoted{new(bracketed)}

	os.Clearenv()
	os.Setenv("ENV_CONFIG_FOO", "foo")
	os.Setenv("ENV_CONFIG_BAR", "bar")
	os.Setenv("ENV_CONFIG_BAZ", "baz")
	os.Setenv("ENV_CONFIG_STRUCT", "inner")

	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if want := "foo"; s.Foo != want {
		t.Errorf("foo: got %#q, want %#q", s.Foo, want)
	}

	if want := "[bar]"; s.Bar.String() != want {
		t.Errorf("bar: got %#q, want %#q", s.Bar, want)
	}

	if want := `["baz"]`; s.Baz.String() != want {
		t.Errorf(`baz: got %#q, want %#q`, s.Baz, want)
	}

	if want := `setterstruct{"inner"}`; s.Struct.Inner != want {
		t.Errorf(`Struct.Inner: got %#q, want %#q`, s.Struct.Inner, want)
	}
}

func TestCustomPointerFields(t *testing.T) {
	var s struct {
		Foo    string
		Bar    *bracketed
		Baz    *quoted
		Struct *setterStruct
	}

	// Set would panic when the receiver is nil,
	// so make sure they have initial values to replace.
	s.Bar = new(bracketed)
	s.Baz = &quoted{new(bracketed)}

	os.Clearenv()
	os.Setenv("ENV_CONFIG_FOO", "foo")
	os.Setenv("ENV_CONFIG_BAR", "bar")
	os.Setenv("ENV_CONFIG_BAZ", "baz")
	os.Setenv("ENV_CONFIG_STRUCT", "inner")

	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}

	if want := "foo"; s.Foo != want {
		t.Errorf("foo: got %#q, want %#q", s.Foo, want)
	}

	if want := "[bar]"; s.Bar.String() != want {
		t.Errorf("bar: got %#q, want %#q", s.Bar, want)
	}

	if want := `["baz"]`; s.Baz.String() != want {
		t.Errorf(`baz: got %#q, want %#q`, s.Baz, want)
	}

	if want := `setterstruct{"inner"}`; s.Struct.Inner != want {
		t.Errorf(`Struct.Inner: got %#q, want %#q`, s.Struct.Inner, want)
	}
}

func TestEmptyPrefixUsesFieldNames(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("REQUIREDVAR", "foo")

	err := Process("", &s)
	if err != nil {
		t.Errorf("Process failed: %s", err)
	}

	if s.RequiredVar != "foo" {
		t.Errorf(
			`RequiredVar not populated correctly: expected "foo", got %q`,
			s.RequiredVar,
		)
	}
}

func TestNestedStructVarName(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "required")
	val := "found with only short name"
	os.Setenv("INNER", val)
	if err := Process("env_config", &s); err != nil {
		t.Error(err.Error())
	}
	if s.NestedSpecification.Property != val {
		t.Errorf("expected %s, got %s", val, s.NestedSpecification.Property)
	}
}

func TestTextUnmarshalerError(t *testing.T) {
	var s Specification
	os.Clearenv()
	os.Setenv("ENV_CONFIG_REQUIREDVAR", "foo")
	os.Setenv("ENV_CONFIG_DATETIME", "I'M NOT A DATE")

	err := Process("env_config", &s)

	v, ok := err.(*ParseError)
	if !ok {
		t.Errorf("expected ParseError, got %v", v)
	}
	if v.FieldName != "Datetime" {
		t.Errorf("expected %s, got %v", "Debug", v.FieldName)
	}
	if s.Debug != false {
		t.Errorf("expected %v, got %v", false, s.Debug)
	}
}

type bracketed string

func (b *bracketed) Set(value string) error {
	*b = bracketed("[" + value + "]")
	return nil
}

func (b bracketed) String() string {
	return string(b)
}

// quoted is used to test the precedence of Decode over Set.
// The sole field is a flag.Value rather than a setter to validate that
// all flag.Value implementations are also Setter implementations.
type quoted struct{ flag.Value }

func (d quoted) Decode(value string) error {
	return d.Set(`"` + value + `"`)
}

type setterStruct struct {
	Inner string
}

func (ss *setterStruct) Set(value string) error {
	ss.Inner = fmt.Sprintf("setterstruct{%q}", value)
	return nil
}