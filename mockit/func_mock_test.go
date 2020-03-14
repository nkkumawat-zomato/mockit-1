package mockit

import (
	"path/filepath"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/pasdam/mockit/matchers/argument"
	"github.com/stretchr/testify/assert"
)

func Test_funcMock_ShouldUseArgumentMatcher(t *testing.T) {
	m := NewFuncMock(t, filepath.Base).(*funcMock)
	m.With(argument.Any).Return("result")

	assert.Equal(t, "result", filepath.Base("argument-1"))
	assert.Equal(t, "result", filepath.Base("argument-2"))
	assert.Equal(t, "result", filepath.Base("argument-3"))
	assert.Equal(t, 3, len(m.calls))
	m.Verify(t, "argument-1")
	m.Verify(t, "argument-2")
	m.Verify(t, "argument-3")
}

func Test_funcMock_ShouldReturnDefaultOutputIfNoMatchingCallIsFound(t *testing.T) {
	m := NewFuncMock(t, filepath.Base).(*funcMock)
	m.With("argument-1").Return("out-1")
	m.With("argument-2").Return("out-2")
	m.With("argument-3").Return("out-3")

	assert.Equal(t, "", filepath.Base("non-matching-argument"))
	assert.Equal(t, 1, len(m.calls))
	m.Verify(t, "non-matching-argument")
}

func Test_funcMock_ShouldReturnAZeroValueIfTheMockArgumentIsNil(t *testing.T) {
	m := NewFuncMock(t, filepath.Walk).(*funcMock)
	m.With("arg", nil).Return(nil)

	assert.Nil(t, filepath.Walk("arg", nil))
	assert.Equal(t, 1, len(m.calls))
	m.Verify(t, "arg", nil)
}

func Test_funcMock_ShouldReturnExpectedOutputIfAMatchingCallIsFound(t *testing.T) {
	m := NewFuncMock(t, filepath.Base).(*funcMock)
	m.With("matching-argument").Return("some-out")

	assert.Equal(t, "some-out", filepath.Base("matching-argument"))
	assert.Equal(t, 1, len(m.calls))
	m.Verify(t, "matching-argument")
}

func Test_funcMock_ShouldDisableAndRestoreAMock(t *testing.T) {
	m := NewFuncMock(t, filepath.Base).(*funcMock)
	m.With("matching-argument").Return("some-out")

	assert.Equal(t, "some-out", filepath.Base("matching-argument"))

	m.Disable()

	assert.Equal(t, "matching-argument", filepath.Base("matching-argument"))

	m.Enable()

	assert.Equal(t, "some-out", filepath.Base("matching-argument"))

	assert.Equal(t, 2, len(m.calls))
	m.Verify(t, "matching-argument")
}

func Test_NewFuncMock(t *testing.T) {
	type args struct {
		t      *testing.T
		target interface{}
	}
	tests := []struct {
		name       string
		args       args
		defaultOut []interface{}
		shouldFail bool
	}{
		{
			name: "Non function",
			args: args{
				target: "non-function-type",
			},
			shouldFail: true,
		},
		{
			name: "Function",
			args: args{
				target: filepath.Base,
			},
			defaultOut: []interface{}{""},
			shouldFail: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := new(testing.T)
			got := NewFuncMock(mockT, tt.args.target)
			if tt.shouldFail {
				if !mockT.Failed() {
					t.Fatalf("NewFuncMock was expected to fail, but it didn't")
				}
				if got != nil {
					t.Fatalf("NewFuncMock was expected to return nil, but it was %v", got)
				}
			} else {
				if mockT.Failed() {
					t.Fatalf("NewFuncMock wasn't expected to fail, but it did")
				}
				if got == nil {
					t.Fatalf("NewFuncMock was expected to return a valid object, but it was nil")
				}
				m := got.(*funcMock)
				assert.True(t, reflect.DeepEqual(reflect.ValueOf(tt.args.target), m.target))
				assert.Equal(t, len(tt.defaultOut), len(m.defaultOut))
				for i := 0; i < len(tt.defaultOut); i++ {
					assert.Equal(t, tt.defaultOut[i], m.defaultOut[i].Interface())
				}
			}
		})
	}
}

func Test_funcMock_CallRealMethod(t *testing.T) {
	type fields struct {
		mocks []*call
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "First mock",
			fields: fields{
				mocks: nil,
			},
		},
		{
			name: "Second mock",
			fields: fields{
				mocks: []*call{&call{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &funcMock{
				mocks: tt.fields.mocks,
				currentMock: &call{
					in: []reflect.Value{reflect.ValueOf("some-arg")},
				},
			}
			f.CallRealMethod()

			expectedMockIndex := len(tt.fields.mocks)
			assert.Equal(t, expectedMockIndex+1, len(f.mocks))
			assert.Nil(t, f.mocks[expectedMockIndex].out)
			assert.Equal(t, 1, len(f.mocks[expectedMockIndex].in))
			assert.Equal(t, "some-arg", f.mocks[expectedMockIndex].in[0].String())
			assert.Nil(t, f.calls)
			assert.Nil(t, f.defaultOut)
			assert.Nil(t, f.guard)
			assert.Nil(t, f.t)
		})
	}
}

func Test_funcMock_Return(t *testing.T) {
	type args struct {
		values []interface{}
	}
	type fields struct {
		mocks []*call
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		shouldFail bool
	}{
		{
			name: "First mock",
			fields: fields{
				mocks: nil,
			},
			args: args{
				values: []interface{}{"out-1"},
			},
			shouldFail: false,
		},
		{
			name: "Second mock",
			fields: fields{
				mocks: []*call{&call{}},
			},
			args: args{
				values: []interface{}{"out-2"},
			},
			shouldFail: false,
		},
		{
			name: "Wrong return type",
			fields: fields{
				mocks: []*call{&call{}},
			},
			args: args{
				values: []interface{}{100},
			},
			shouldFail: true,
		},
		{
			name: "Not enough return values",
			fields: fields{
				mocks: []*call{&call{}},
			},
			args: args{
				values: []interface{}{},
			},
			shouldFail: true,
		},
		{
			name: "Too many return values",
			fields: fields{
				mocks: []*call{&call{}},
			},
			args: args{
				values: []interface{}{"out-0", "out-1"},
			},
			shouldFail: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := []reflect.Value{reflect.ValueOf("some-arg")}
			mockT := new(testing.T)
			f := &funcMock{
				mocks: tt.fields.mocks,
				currentMock: &call{
					in: in,
				},
				target: reflect.ValueOf(filepath.Base),
				t:      mockT,
			}
			f.Return(tt.args.values...)

			assert.Equal(t, tt.shouldFail, mockT.Failed())

			if !tt.shouldFail {
				expectedMockIndex := len(tt.fields.mocks)
				assert.Equal(t, expectedMockIndex+1, len(f.mocks))
				assert.Equal(t, in, f.mocks[expectedMockIndex].in)

				assert.Equal(t, 1, len(f.mocks[expectedMockIndex].out))
				assert.Equal(t, tt.args.values[0], f.mocks[expectedMockIndex].out[0].String())

				assert.Nil(t, f.calls)
				assert.Nil(t, f.defaultOut)
				assert.Nil(t, f.guard)
			}
		})
	}
}

func Test_funcMock_ReturnDefaults(t *testing.T) {
	type fields struct {
		mocks      []*call
		defaultOut []reflect.Value
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "First mock",
			fields: fields{
				mocks: nil,
			},
		},
		{
			name: "Second mock",
			fields: fields{
				mocks: []*call{&call{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &funcMock{
				mocks: tt.fields.mocks,
				currentMock: &call{
					in: []reflect.Value{reflect.ValueOf("some-arg")},
				},
			}
			f.ReturnDefaults()

			expectedMockIndex := len(tt.fields.mocks)
			assert.Equal(t, expectedMockIndex+1, len(f.mocks))
			assert.Nil(t, f.mocks[expectedMockIndex].out)
			assert.Equal(t, 1, len(f.mocks[expectedMockIndex].in))
			assert.Equal(t, "some-arg", f.mocks[expectedMockIndex].in[0].String())
			assert.Nil(t, f.calls)
			assert.Nil(t, f.defaultOut)
			assert.Nil(t, f.guard)
			assert.Nil(t, f.t)
		})
	}
}

func Test_funcMock_Verify(t *testing.T) {
	type fields struct {
		calls      []*call
		defaultOut []reflect.Value
		target     reflect.Value
	}
	type args struct {
		in []interface{}
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		shouldFail bool
	}{
		{
			name: "Not called",
			fields: fields{
				calls:      nil,
				defaultOut: []reflect.Value{reflect.ValueOf("default-out-value")},
				target:     reflect.ValueOf(filepath.Base),
			},
			args: args{
				in: []interface{}{"some-arg"},
			},
			shouldFail: true,
		},
		{
			name: "Called with a different argument",
			fields: fields{
				calls: []*call{&call{
					in:  []reflect.Value{reflect.ValueOf("some-arg")},
					out: []reflect.Value{reflect.ValueOf("mocked-out-value")},
				}},
				defaultOut: []reflect.Value{reflect.ValueOf("default-out-value")},
				target:     reflect.ValueOf(filepath.Base),
			},
			args: args{
				in: []interface{}{"some-other-arg"},
			},
			shouldFail: true,
		},
		{
			name: "Called",
			fields: fields{
				calls: []*call{&call{
					in:  []reflect.Value{reflect.ValueOf("some-arg")},
					out: []reflect.Value{reflect.ValueOf("mocked-out-value")},
				}},
				defaultOut: []reflect.Value{reflect.ValueOf("default-out-value")},
				target:     reflect.ValueOf(filepath.Base),
			},
			args: args{
				in: []interface{}{"some-arg"},
			},
			shouldFail: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &funcMock{
				calls:      tt.fields.calls,
				defaultOut: tt.fields.defaultOut,
				target:     tt.fields.target,
			}
			mockT := new(testing.T)
			m.Verify(mockT, tt.args.in...)
			if mockT.Failed() != tt.shouldFail {
				if tt.shouldFail {
					t.Errorf("Verify was expected to fail, but it didn't")
				} else {
					t.Errorf("Verify wasn't expected to fail, but it did")
				}
			}
		})
	}
}

func Test_funcMock_With(t *testing.T) {
	type args struct {
		in []interface{}
	}
	tests := []struct {
		name       string
		args       args
		shouldFail bool
	}{
		{
			name: "Invalid arguments",
			args: args{
				in: []interface{}{"some-in", "some-additional-in"},
			},
			shouldFail: true,
		},
		{
			name: "Valid arguments",
			args: args{
				in: []interface{}{"some-in"},
			},
			shouldFail: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := new(testing.T)
			m := &funcMock{
				target: reflect.ValueOf(filepath.Base),
				t:      mockT,
			}
			got := m.With(tt.args.in...)

			assert.Equal(t, m, got)
			if tt.shouldFail != mockT.Failed() {
				if tt.shouldFail {
					t.Errorf("Verify was expected to fail, but it didn't")
				} else {
					t.Errorf("Verify wasn't expected to fail, but it did")
				}
			}
			assert.Equal(t, 0, len(m.mocks))
			assert.NotNil(t, m.currentMock)
			if tt.shouldFail == false {
				assert.Equal(t, len(tt.args.in), len(m.currentMock.in))
				for i := 0; i < len(tt.args.in); i++ {
					assert.Equal(t, tt.args.in[i], m.currentMock.in[i].Interface())
				}
			} else {
				assert.Nil(t, m.currentMock.in)
			}
			assert.Nil(t, m.currentMock.out)
		})
	}
}

func Test_funcMock_makeCall(t *testing.T) {
	type fields struct {
		mocks      []*call
		defaultOut []reflect.Value
	}
	type args struct {
		in []reflect.Value
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      []reflect.Value
		wantCount int
	}{
		{
			name: "Default output",
			fields: fields{
				mocks:      nil,
				defaultOut: []reflect.Value{reflect.ValueOf("default-out-value")},
			},
			args: args{
				in: []reflect.Value{reflect.ValueOf("some-arg")},
			},
			want: []reflect.Value{reflect.ValueOf("default-out-value")},
		},
		{
			name: "Mocked output",
			fields: fields{
				mocks: []*call{&call{
					in:  []reflect.Value{reflect.ValueOf("some-arg")},
					out: []reflect.Value{reflect.ValueOf("mocked-out-value")},
				}},
			},
			args: args{
				in: []reflect.Value{reflect.ValueOf("some-arg")},
			},
			want: []reflect.Value{reflect.ValueOf("mocked-out-value")},
		},
		{
			name: "Real method",
			fields: fields{
				mocks: []*call{&call{
					in:  []reflect.Value{reflect.ValueOf("../mockit/func_mock_test.go")},
					out: nil,
				}},
			},
			args: args{
				in: []reflect.Value{reflect.ValueOf("../mockit/func_mock_test.go")},
			},
			want: []reflect.Value{reflect.ValueOf("func_mock_test.go")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guard := monkey.Patch(filepath.Base, func(string) string { return "wrong-result" })
			m := &funcMock{
				mocks:      tt.fields.mocks,
				defaultOut: tt.fields.defaultOut,
				target:     reflect.ValueOf(filepath.Base),
				guard:      guard,
			}
			if got := m.makeCall(tt.args.in); !callsMatch(got, tt.want, true) {
				t.Errorf("mockFunc.makeCall() = %v, want %v", got, tt.want)
			}
			assert.Equal(t, 1, len(m.calls))
			assert.Equal(t, tt.args.in, m.calls[0].in)
			assert.Nil(t, m.calls[0].out)
		})
	}
}

func Test_funcMock_makeCall_shouldRecordMultipleCalls(t *testing.T) {
	m := NewFuncMock(t, filepath.Base).(*funcMock)

	filepath.Base("arg-0")
	filepath.Base("arg-1")
	filepath.Base("arg-2")

	assert.Equal(t, 3, len(m.calls))

	assert.Equal(t, 1, len(m.calls[0].in))
	assert.Equal(t, "arg-0", m.calls[0].in[0].String())
	assert.Nil(t, m.calls[0].out)

	assert.Equal(t, 1, len(m.calls[1].in))
	assert.Equal(t, "arg-1", m.calls[1].in[0].String())
	assert.Nil(t, m.calls[1].out)

	assert.Equal(t, 1, len(m.calls[2].in))
	assert.Equal(t, "arg-2", m.calls[2].in[0].String())
	assert.Nil(t, m.calls[2].out)
}
